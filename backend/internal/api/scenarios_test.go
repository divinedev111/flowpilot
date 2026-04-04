package api

import (
	"testing"

	"flowpilot/internal/portfolio"

	"github.com/stretchr/testify/assert"
)

func TestComputeHypothetical(t *testing.T) {
	current := portfolio.Metrics{
		NetWorth:      100000,
		PositionCount: 5,
		ExposureByType: map[string]float64{
			"equity": 70,
			"crypto": 30,
		},
		ConcentrationBySymbol: map[string]float64{
			"AAPL": 40,
		},
	}

	adjustments := []AllocationAdjustment{
		{AssetType: "crypto", TargetPct: 20},
		{AssetType: "equity", TargetPct: 80},
	}

	hyp := computeHypothetical(current, adjustments)

	assert.Equal(t, 100000.0, hyp.NetWorth)
	assert.Equal(t, 20.0, hyp.ExposureByType["crypto"])
	assert.Equal(t, 80.0, hyp.ExposureByType["equity"])
}

func TestBuildScenarioPrompt(t *testing.T) {
	current := portfolio.Metrics{
		NetWorth: 50000,
		ExposureByType: map[string]float64{
			"equity": 60,
			"crypto": 40,
		},
	}
	hyp := portfolio.Metrics{
		ExposureByType: map[string]float64{
			"equity": 80,
			"crypto": 20,
		},
	}
	adjs := []AllocationAdjustment{
		{AssetType: "crypto", TargetPct: 20},
	}

	prompt := buildScenarioPrompt(current, hyp, adjs)
	assert.Contains(t, prompt, "$50000.00")
	assert.Contains(t, prompt, "crypto")
}
