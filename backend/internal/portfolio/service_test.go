package portfolio

import (
	"testing"
	"time"

	"flowpilot/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestComputeMetrics(t *testing.T) {
	positions := []models.Position{
		{Symbol: "AAPL", MarketValue: 5000, AssetType: "equity", Timestamp: time.Now()},
		{Symbol: "MSFT", MarketValue: 3000, AssetType: "equity", Timestamp: time.Now()},
		{Symbol: "BTC", MarketValue: 2000, AssetType: "crypto", Timestamp: time.Now()},
	}

	metrics := ComputeMetrics(positions)

	assert.Equal(t, 10000.0, metrics.NetWorth)
	assert.Equal(t, 80.0, metrics.ExposureByType["equity"])
	assert.Equal(t, 20.0, metrics.ExposureByType["crypto"])
	assert.Equal(t, 50.0, metrics.ConcentrationBySymbol["AAPL"])
	assert.Equal(t, 3, metrics.PositionCount)
}

func TestComputeMetrics_Empty(t *testing.T) {
	metrics := ComputeMetrics(nil)
	assert.Equal(t, 0.0, metrics.NetWorth)
	assert.Equal(t, 0, metrics.PositionCount)
}
