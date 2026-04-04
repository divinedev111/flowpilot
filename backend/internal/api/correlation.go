package api

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"flowpilot/internal/correlation"
	"flowpilot/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type correlationCacheEntry struct {
	ID         bson.ObjectID              `bson:"_id,omitempty"`
	Symbols    []string                   `bson:"symbols"`
	Matrix     [][]float64                `bson:"matrix"`
	PeriodDays int                        `bson:"period_days"`
	TopPairs   []correlation.CorrelatedPair `bson:"top_pairs"`
	ComputedAt time.Time                  `bson:"computed_at"`
}

func (d *Deps) handleCorrelation(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for a cached result less than 24h old
	cached, err := d.getCachedCorrelation(ctx)
	if err == nil && cached != nil {
		resp := gin.H{
			"symbols":     cached.Symbols,
			"matrix":      cached.Matrix,
			"period_days": cached.PeriodDays,
			"top_pairs":   cached.TopPairs,
		}
		if cached.PeriodDays > 0 && cached.PeriodDays < 20 {
			resp["warning"] = fmt.Sprintf("Only %d overlapping trading days found. At least 20 are needed for statistically meaningful correlations. Results may be unreliable.", cached.PeriodDays)
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// Get ALL positions (crypto + equity)
	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	symbols, assetTypes := extractSymbolsWithTypes(positions)
	if len(symbols) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"symbols":     []string{},
			"matrix":      [][]float64{},
			"period_days": 0,
			"top_pairs":   []correlation.CorrelatedPair{},
		})
		return
	}

	// Fetch prices — route crypto to Binance, stocks to Yahoo
	priceSvc := correlation.NewPriceService(d.DB)
	priceMap := make(map[string][]correlation.DayPrice)

	for _, sym := range symbols {
		isCrypto := assetTypes[sym]
		prices, err := priceSvc.GetDailyPrices(ctx, sym, isCrypto)
		if err != nil {
			continue
		}
		priceMap[sym] = prices
	}

	if len(priceMap) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"symbols":     []string{},
			"matrix":      [][]float64{},
			"period_days": 0,
			"top_pairs":   []correlation.CorrelatedPair{},
		})
		return
	}

	alignedSymbols, returns, periodDays := correlation.AlignReturns(priceMap)
	if returns == nil || len(returns) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"symbols":     alignedSymbols,
			"matrix":      [][]float64{},
			"period_days": 0,
			"top_pairs":   []correlation.CorrelatedPair{},
			"warning":     "Not enough overlapping price data to compute correlations.",
		})
		return
	}

	var warning string
	if periodDays < 20 {
		warning = fmt.Sprintf("Only %d overlapping trading days found. At least 20 are needed for statistically meaningful correlations. Results may be unreliable.", periodDays)
	}

	result := correlation.BuildCorrelationMatrix(alignedSymbols, returns)
	result.PeriodDays = periodDays

	for i := range result.Matrix {
		for j := range result.Matrix[i] {
			result.Matrix[i][j] = math.Round(result.Matrix[i][j]*1000) / 1000
		}
	}

	topPairs := correlation.TopPairs(result, 10)
	_ = d.cacheCorrelation(ctx, result, topPairs)

	resp := gin.H{
		"symbols":     result.Symbols,
		"matrix":      result.Matrix,
		"period_days": result.PeriodDays,
		"top_pairs":   topPairs,
	}
	if warning != "" {
		resp["warning"] = warning
	}
	c.JSON(http.StatusOK, resp)
}

// extractSymbolsWithTypes returns unique symbols and a map of symbol→isCrypto.
// Excludes prediction_market positions (no tradeable price data available).
func extractSymbolsWithTypes(positions []models.Position) ([]string, map[string]bool) {
	seen := make(map[string]bool)
	assetTypes := make(map[string]bool) // true = crypto
	var symbols []string
	for _, p := range positions {
		if p.MarketValue == 0 {
			continue
		}
		// Skip prediction market positions — they don't have price data on exchanges
		if p.AssetType == "prediction_market" {
			continue
		}
		if !seen[p.Symbol] {
			seen[p.Symbol] = true
			isCrypto := p.AssetType == "crypto" || p.AssetType == "cryptocurrency"
			assetTypes[p.Symbol] = isCrypto
			symbols = append(symbols, p.Symbol)
		}
	}
	return symbols, assetTypes
}

func (d *Deps) getCachedCorrelation(ctx context.Context) (*correlationCacheEntry, error) {
	var entry correlationCacheEntry
	filter := bson.D{
		{Key: "computed_at", Value: bson.D{{Key: "$gt", Value: time.Now().Add(-24 * time.Hour)}}},
	}
	err := d.DB.CorrelationCache().FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (d *Deps) cacheCorrelation(ctx context.Context, result *correlation.CorrelationResult, topPairs []correlation.CorrelatedPair) error {
	_, _ = d.DB.CorrelationCache().DeleteMany(ctx, bson.D{})

	entry := correlationCacheEntry{
		Symbols:    result.Symbols,
		Matrix:     result.Matrix,
		PeriodDays: result.PeriodDays,
		TopPairs:   topPairs,
		ComputedAt: result.ComputedAt,
	}

	_, err := d.DB.CorrelationCache().InsertOne(ctx, entry)
	return err
}
