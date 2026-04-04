package api

import (
	"net/http"
	"sort"

	"flowpilot/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type attributionEntry struct {
	Label  string  `json:"label"`
	Value  float64 `json:"value"`
	Pct    float64 `json:"pct"`
	Change float64 `json:"change"`
}

type performanceResponse struct {
	ByAsset  []attributionEntry  `json:"by_asset"`
	ByType   []attributionEntry  `json:"by_type"`
	BySource []attributionEntry  `json:"by_source"`
	History  []historyPoint      `json:"history"`
}

type historyPoint struct {
	Date     string  `json:"date"`
	NetWorth float64 `json:"net_worth"`
}

func (d *Deps) handlePerformance(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current positions
	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	if len(positions) == 0 {
		c.JSON(http.StatusOK, performanceResponse{
			ByAsset:  []attributionEntry{},
			ByType:   []attributionEntry{},
			BySource: []attributionEntry{},
			History:  []historyPoint{},
		})
		return
	}

	// Compute total net worth
	var netWorth float64
	for _, p := range positions {
		netWorth += p.MarketValue
	}

	// Get last 2 snapshots to compute changes
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(2)
	cursor, err := d.DB.Snapshots().Find(ctx, bson.D{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch snapshots"})
		return
	}
	var recentSnaps []models.Snapshot
	if err := cursor.All(ctx, &recentSnaps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode snapshots"})
		return
	}

	// Build previous position price map for change calculation
	prevPrices := make(map[string]float64)
	if len(recentSnaps) >= 2 {
		prevSnap := recentSnaps[1]
		prevPositions, _ := d.Snapshot.GetPositions(ctx, prevSnap.ID)
		for _, pp := range prevPositions {
			prevPrices[pp.Symbol] = pp.MarkPrice
		}
	}

	// Attribution by asset (individual positions)
	byAsset := make([]attributionEntry, 0, len(positions))
	for _, p := range positions {
		change := 0.0
		if prev, ok := prevPrices[p.Symbol]; ok && prev > 0 {
			change = ((p.MarkPrice - prev) / prev) * 100
		}
		byAsset = append(byAsset, attributionEntry{
			Label:  p.Symbol,
			Value:  p.MarketValue,
			Pct:    (p.MarketValue / netWorth) * 100,
			Change: change,
		})
	}
	sort.Slice(byAsset, func(i, j int) bool { return byAsset[i].Value > byAsset[j].Value })

	// Attribution by type
	typeValue := make(map[string]float64)
	typeChange := make(map[string]float64)
	typeCount := make(map[string]int)
	for _, p := range positions {
		typeValue[p.AssetType] += p.MarketValue
		if prev, ok := prevPrices[p.Symbol]; ok && prev > 0 {
			typeChange[p.AssetType] += ((p.MarkPrice - prev) / prev) * 100
			typeCount[p.AssetType]++
		}
	}
	byType := make([]attributionEntry, 0)
	for t, v := range typeValue {
		avgChange := 0.0
		if typeCount[t] > 0 {
			avgChange = typeChange[t] / float64(typeCount[t])
		}
		byType = append(byType, attributionEntry{
			Label:  t,
			Value:  v,
			Pct:    (v / netWorth) * 100,
			Change: avgChange,
		})
	}
	sort.Slice(byType, func(i, j int) bool { return byType[i].Value > byType[j].Value })

	// Attribution by source
	sourceValue := make(map[string]float64)
	for _, p := range positions {
		sourceValue[p.Source] += p.MarketValue
	}
	bySource := make([]attributionEntry, 0)
	for s, v := range sourceValue {
		bySource = append(bySource, attributionEntry{
			Label:  s,
			Value:  v,
			Pct:    (v / netWorth) * 100,
			Change: 0,
		})
	}
	sort.Slice(bySource, func(i, j int) bool { return bySource[i].Value > bySource[j].Value })

	// Snapshot history for chart — deduplicate by date (keep latest per day)
	histOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetLimit(500)
	histCursor, err := d.DB.Snapshots().Find(ctx, bson.D{}, histOpts)
	history := make([]historyPoint, 0)
	if err == nil {
		var snaps []models.Snapshot
		if histCursor.All(ctx, &snaps) == nil {
			byDate := make(map[string]models.Snapshot)
			dateOrder := make([]string, 0)
			for _, s := range snaps {
				// Skip incomplete syncs with negligible net worth
				if s.NetWorth < 100 {
					continue
				}
				key := s.CreatedAt.Format("2006-01-02")
				if _, exists := byDate[key]; !exists {
					dateOrder = append(dateOrder, key)
				}
				byDate[key] = s // latest per day wins
			}
			for _, key := range dateOrder {
				s := byDate[key]
				history = append(history, historyPoint{
					Date:     s.CreatedAt.Format("Jan 2"),
					NetWorth: s.NetWorth,
				})
			}
		}
	}

	c.JSON(http.StatusOK, performanceResponse{
		ByAsset:  byAsset,
		ByType:   byType,
		BySource: bySource,
		History:  history,
	})
}
