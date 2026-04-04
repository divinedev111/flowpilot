package api

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type topPosition struct {
	Symbol string  `json:"symbol"`
	Pct    float64 `json:"pct"`
}

type concentration struct {
	TopPosition  topPosition   `json:"top_position"`
	Top3Pct      float64       `json:"top_3_pct"`
	Top5Pct      float64       `json:"top_5_pct"`
	HHI          float64       `json:"hhi"`
	TopPositions []topPosition `json:"top_positions"`
}

type exposure struct {
	ByType   map[string]float64 `json:"by_type"`
	BySource map[string]float64 `json:"by_source"`
}

type drawdown struct {
	MaxDrawdownPct     float64 `json:"max_drawdown_pct"`
	CurrentDrawdownPct float64 `json:"current_drawdown_pct"`
	PeakValue          float64 `json:"peak_value"`
	CurrentValue       float64 `json:"current_value"`
}

type riskResponse struct {
	Concentration        concentration `json:"concentration"`
	Exposure             exposure      `json:"exposure"`
	Drawdown             drawdown      `json:"drawdown"`
	DiversificationScore int           `json:"diversification_score"`
	RiskLevel            string        `json:"risk_level"`
	Warnings             []string      `json:"warnings"`
	History              []historyPoint `json:"history"`
}

func (d *Deps) handleRisk(c *gin.Context) {
	ctx := c.Request.Context()

	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	metrics := portfolio.ComputeMetrics(positions)

	conc := computeConcentration(metrics)
	exp := computeExposure(positions, metrics.NetWorth)

	// Fetch snapshots for drawdown + history (last 30, ordered oldest first)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(30)
	cursor, err := d.DB.Snapshots().Find(ctx, bson.D{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch snapshots"})
		return
	}
	var snapshots []models.Snapshot
	if err := cursor.All(ctx, &snapshots); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode snapshots"})
		return
	}

	// Reverse so oldest is first
	for i, j := 0, len(snapshots)-1; i < j; i, j = i+1, j-1 {
		snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
	}

	// Filter out snapshots with negligible net worth (incomplete syncs)
	var validSnapshots []models.Snapshot
	for _, s := range snapshots {
		if s.NetWorth >= 100 {
			validSnapshots = append(validSnapshots, s)
		}
	}

	dd := computeDrawdown(validSnapshots, metrics.NetWorth)

	// Deduplicate snapshots by date (keep latest per day)
	byDate := make(map[string]models.Snapshot)
	dateOrder := make([]string, 0)
	for _, s := range validSnapshots {
		key := s.CreatedAt.Format("2006-01-02")
		if _, exists := byDate[key]; !exists {
			dateOrder = append(dateOrder, key)
		}
		byDate[key] = s // latest per day wins
	}
	history := make([]historyPoint, 0, len(dateOrder))
	for _, key := range dateOrder {
		s := byDate[key]
		history = append(history, historyPoint{
			Date:     s.CreatedAt.Format("2006-01-02"),
			NetWorth: math.Round(s.NetWorth*100) / 100,
		})
	}

	// Diversification score: 100 - (HHI / 100), clamped 0-100
	score := int(math.Round(100 - conc.HHI/100))
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	// Risk level
	riskLevel := "low"
	switch {
	case score < 25:
		riskLevel = "critical"
	case score < 50:
		riskLevel = "high"
	case score < 75:
		riskLevel = "moderate"
	}

	// Warnings
	warnings := computeWarnings(conc, exp)

	resp := riskResponse{
		Concentration:        conc,
		Exposure:             exp,
		Drawdown:             dd,
		DiversificationScore: score,
		RiskLevel:            riskLevel,
		Warnings:             warnings,
		History:              history,
	}

	c.JSON(http.StatusOK, resp)
}

type symbolWeight struct {
	Symbol string
	Pct    float64
}

func computeConcentration(m portfolio.Metrics) concentration {
	if m.NetWorth == 0 {
		return concentration{
			TopPosition: topPosition{},
		}
	}

	// Sort symbols by concentration descending
	weights := make([]symbolWeight, 0, len(m.ConcentrationBySymbol))
	for sym, pct := range m.ConcentrationBySymbol {
		weights = append(weights, symbolWeight{Symbol: sym, Pct: math.Round(pct*10) / 10})
	}
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].Pct > weights[j].Pct
	})

	var top topPosition
	if len(weights) > 0 {
		top = topPosition{Symbol: weights[0].Symbol, Pct: weights[0].Pct}
	}

	var top3, top5, hhi float64
	for i, w := range weights {
		if i < 3 {
			top3 += w.Pct
		}
		if i < 5 {
			top5 += w.Pct
		}
		hhi += w.Pct * w.Pct
	}

	// Build top 5 positions list
	limit := 5
	if len(weights) < limit {
		limit = len(weights)
	}
	topPositions := make([]topPosition, limit)
	for i := 0; i < limit; i++ {
		topPositions[i] = topPosition{Symbol: weights[i].Symbol, Pct: weights[i].Pct}
	}

	return concentration{
		TopPosition:  top,
		Top3Pct:      math.Round(top3*10) / 10,
		Top5Pct:      math.Round(top5*10) / 10,
		HHI:          math.Round(hhi*10) / 10,
		TopPositions: topPositions,
	}
}

func computeExposure(positions []models.Position, netWorth float64) exposure {
	byType := make(map[string]float64)
	bySource := make(map[string]float64)

	if netWorth == 0 {
		return exposure{ByType: byType, BySource: bySource}
	}

	typeValue := make(map[string]float64)
	sourceValue := make(map[string]float64)

	for _, p := range positions {
		typeValue[p.AssetType] += p.MarketValue
		sourceValue[p.Source] += p.MarketValue
	}

	for t, v := range typeValue {
		byType[t] = math.Round((v/netWorth)*1000) / 10
	}
	for s, v := range sourceValue {
		bySource[s] = math.Round((v/netWorth)*1000) / 10
	}

	return exposure{ByType: byType, BySource: bySource}
}

func computeDrawdown(snapshots []models.Snapshot, currentNetWorth float64) drawdown {
	if len(snapshots) == 0 {
		return drawdown{
			CurrentValue: currentNetWorth,
			PeakValue:    currentNetWorth,
		}
	}

	peak := snapshots[0].NetWorth
	maxDD := 0.0

	for _, s := range snapshots {
		if s.NetWorth > peak {
			peak = s.NetWorth
		}
		if peak > 0 {
			dd := (s.NetWorth - peak) / peak * 100
			if dd < maxDD {
				maxDD = dd
			}
		}
	}

	// Current drawdown relative to peak
	currentDD := 0.0
	if peak > 0 {
		currentDD = (currentNetWorth - peak) / peak * 100
	}

	return drawdown{
		MaxDrawdownPct:     math.Round(maxDD*10) / 10,
		CurrentDrawdownPct: math.Round(currentDD*10) / 10,
		PeakValue:          math.Round(peak*100) / 100,
		CurrentValue:       math.Round(currentNetWorth*100) / 100,
	}
}

func computeWarnings(conc concentration, exp exposure) []string {
	var warnings []string

	// Position concentration > 15%
	if conc.TopPosition.Pct > 15 {
		warnings = append(warnings, "Top position "+conc.TopPosition.Symbol+
			" represents "+formatPct(conc.TopPosition.Pct)+"% of portfolio (>15% threshold)")
	}

	// Type exposure > 40%
	for t, pct := range exp.ByType {
		if pct > 40 {
			warnings = append(warnings, capitalize(t)+" exposure at "+formatPct(pct)+"% (>40% threshold)")
		}
	}

	// Crypto > 25%
	if crypto, ok := exp.ByType["crypto"]; ok && crypto > 25 {
		warnings = append(warnings, "Crypto exposure at "+formatPct(crypto)+"% (>25% threshold)")
	}

	if len(warnings) == 0 {
		warnings = []string{}
	}

	return warnings
}

func formatPct(v float64) string {
	s := math.Round(v*10) / 10
	if s == float64(int(s)) {
		return fmt.Sprintf("%.0f", s)
	}
	return fmt.Sprintf("%.1f", s)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Replace underscores with spaces and title-case each word
	words := strings.Split(s, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
