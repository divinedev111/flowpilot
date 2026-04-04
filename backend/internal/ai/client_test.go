package ai

import (
	"testing"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"github.com/stretchr/testify/assert"
)

func TestBuildDigestPrompt(t *testing.T) {
	metrics := portfolio.Metrics{
		NetWorth:      100000,
		PositionCount: 5,
		ExposureByType: map[string]float64{
			"equity": 70,
			"crypto": 30,
		},
		ConcentrationBySymbol: map[string]float64{
			"AAPL": 40,
			"BTC":  30,
		},
	}

	diff := &models.Diff{
		NetWorthChange: 5000,
		TopMovers: []models.Mover{
			{Symbol: "AAPL", ChangePercent: 10, OldValue: 36000, NewValue: 40000},
		},
		AddedSymbols:   []string{"ETH"},
		RemovedSymbols: []string{"GOOG"},
	}

	alerts := []models.AlertEvent{
		{Severity: "high", Message: "Crypto exposure exceeded 25%"},
	}

	prompt := buildDigestPrompt(metrics, diff, alerts)

	assert.Contains(t, prompt, "$100000.00")
	assert.Contains(t, prompt, "equity: 70.0%")
	assert.Contains(t, prompt, "AAPL: 10.0%")
	assert.Contains(t, prompt, "$5000.00")
	assert.Contains(t, prompt, "ETH")
	assert.Contains(t, prompt, "GOOG")
	assert.Contains(t, prompt, "Crypto exposure exceeded 25%")
}

func TestBuildPortfolioContext(t *testing.T) {
	metrics := portfolio.Metrics{
		NetWorth:       50000,
		PositionCount:  2,
		ExposureByType: map[string]float64{"equity": 100},
	}
	positions := []models.Position{
		{Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab"},
	}

	ctx := BuildPortfolioContext(metrics, positions)
	assert.Contains(t, ctx, "$50000.00")
	assert.Contains(t, ctx, "AAPL")
	assert.Contains(t, ctx, "schwab")
}

func TestBuildDigestPrompt_NilDiff(t *testing.T) {
	metrics := portfolio.Metrics{
		NetWorth:              10000,
		PositionCount:         1,
		ExposureByType:        map[string]float64{"crypto": 100},
		ConcentrationBySymbol: map[string]float64{"BTC": 100},
	}
	prompt := buildDigestPrompt(metrics, nil, nil)
	assert.Contains(t, prompt, "$10000.00")
	assert.NotContains(t, prompt, "Changes Since")
}
