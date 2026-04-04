package api

import (
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"

	"flowpilot/internal/portfolio"

	"github.com/gin-gonic/gin"
)

type AllocationAdjustment struct {
	AssetType string  `json:"asset_type" binding:"required"`
	TargetPct float64 `json:"target_pct" binding:"required"`
}

type ScenarioRequest struct {
	Adjustments []AllocationAdjustment `json:"adjustments" binding:"required"`
}

func (d *Deps) handleScenario(c *gin.Context) {
	var req ScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Get current positions and metrics
	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	currentMetrics := portfolio.ComputeMetrics(positions)
	if currentMetrics.NetWorth == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no portfolio data available"})
		return
	}

	// Compute hypothetical metrics
	hypothetical := computeHypothetical(currentMetrics, req.Adjustments)

	// Build comparison prompt for AI
	prompt := buildScenarioPrompt(currentMetrics, hypothetical, req.Adjustments)

	// Get AI commentary (non-streaming for scenarios)
	commentary := ""
	if d.AI != nil {
		message, err := d.AI.GetClient().Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5,
			MaxTokens: 1024,
			System: []anthropic.TextBlockParam{
				{Text: "You are a portfolio analyst for FlowPilot Finance. Analyze the proposed allocation changes and their implications."},
			},
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		})
		if err == nil {
			for _, block := range message.Content {
				if block.Type == "text" {
					commentary += block.Text
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current":      currentMetrics,
		"hypothetical": hypothetical,
		"adjustments":  req.Adjustments,
		"commentary":   commentary,
	})
}

// computeHypothetical takes current metrics and target allocations,
// returns what the metrics would look like after rebalancing.
func computeHypothetical(current portfolio.Metrics, adjustments []AllocationAdjustment) portfolio.Metrics {
	hyp := portfolio.Metrics{
		NetWorth:              current.NetWorth,
		PositionCount:         current.PositionCount,
		ExposureByType:        make(map[string]float64),
		ConcentrationBySymbol: make(map[string]float64),
	}

	// Copy current exposure
	for k, v := range current.ExposureByType {
		hyp.ExposureByType[k] = v
	}

	// Apply adjustments
	for _, adj := range adjustments {
		hyp.ExposureByType[adj.AssetType] = adj.TargetPct
	}

	// Copy concentration (simplified -- in a real system we'd recompute per-symbol)
	for k, v := range current.ConcentrationBySymbol {
		hyp.ConcentrationBySymbol[k] = v
	}

	return hyp
}

func buildScenarioPrompt(current, hypothetical portfolio.Metrics, adjustments []AllocationAdjustment) string {
	prompt := fmt.Sprintf("Current Portfolio: $%.2f net worth\n\nCurrent Allocation:\n", current.NetWorth)
	for t, pct := range current.ExposureByType {
		prompt += fmt.Sprintf("- %s: %.1f%%\n", t, pct)
	}

	prompt += "\nProposed Changes:\n"
	for _, adj := range adjustments {
		currentPct := current.ExposureByType[adj.AssetType]
		prompt += fmt.Sprintf("- %s: %.1f%% -> %.1f%%\n", adj.AssetType, currentPct, adj.TargetPct)
	}

	prompt += "\nHypothetical Allocation:\n"
	for t, pct := range hypothetical.ExposureByType {
		prompt += fmt.Sprintf("- %s: %.1f%%\n", t, pct)
	}

	prompt += "\nPlease analyze:\n1. What these changes would mean for portfolio risk\n2. How this compares to the current allocation\n3. Any potential concerns or benefits"
	return prompt
}
