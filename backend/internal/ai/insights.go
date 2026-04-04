package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"
)

type Insight struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

type InsightsResult struct {
	Insights      []Insight `json:"insights"`
	OverallHealth string    `json:"overall_health"`
	GeneratedAt   time.Time `json:"generated_at"`
}

func (c *Client) GenerateInsights(ctx context.Context, metrics portfolio.Metrics, positions []models.Position, diff *models.Diff) (*InsightsResult, error) {
	// Start with rule-based insights (no AI needed)
	insights := generateRuleBasedInsights(metrics, positions)

	// Generate AI-powered insights
	aiInsights, overallHealth, err := c.generateAIInsights(ctx, metrics, positions, diff)
	if err != nil {
		// If AI fails, still return rule-based insights
		health := computeHealthFromRules(insights)
		return &InsightsResult{
			Insights:      insights,
			OverallHealth: health,
			GeneratedAt:   time.Now(),
		}, nil
	}

	insights = append(insights, aiInsights...)

	return &InsightsResult{
		Insights:      insights,
		OverallHealth: overallHealth,
		GeneratedAt:   time.Now(),
	}, nil
}

func generateRuleBasedInsights(metrics portfolio.Metrics, positions []models.Position) []Insight {
	var insights []Insight

	// Check position concentration > 20%
	for symbol, pct := range metrics.ConcentrationBySymbol {
		if pct > 20 {
			insights = append(insights, Insight{
				Type:           "concentration_warning",
				Severity:       "high",
				Title:          "High Single-Position Risk",
				Description:    fmt.Sprintf("%s represents %.1f%% of your portfolio, exceeding the 20%% concentration threshold.", symbol, pct),
				Recommendation: fmt.Sprintf("Consider reducing %s exposure below 20%% to limit single-position risk.", symbol),
			})
		}
	}

	// Check asset type > 50%
	for assetType, pct := range metrics.ExposureByType {
		if pct > 50 {
			insights = append(insights, Insight{
				Type:           "allocation_drift",
				Severity:       "medium",
				Title:          fmt.Sprintf("Heavy %s Allocation", assetType),
				Description:    fmt.Sprintf("%.1f%% of your portfolio is allocated to %s, exceeding 50%%.", pct, assetType),
				Recommendation: "Review if this concentration aligns with your target allocation and risk tolerance.",
			})
		}
	}

	// Low diversification: fewer than 5 positions
	if metrics.PositionCount > 0 && metrics.PositionCount < 5 {
		insights = append(insights, Insight{
			Type:           "low_diversification",
			Severity:       "medium",
			Title:          "Low Diversification",
			Description:    fmt.Sprintf("Your portfolio has only %d position(s). Limited diversification increases overall risk.", metrics.PositionCount),
			Recommendation: "Consider adding positions across different asset classes and sectors to improve diversification.",
		})
	}

	// Zero-allocation positions
	for symbol, pct := range metrics.ConcentrationBySymbol {
		if pct == 0 {
			insights = append(insights, Insight{
				Type:           "zero_allocation",
				Severity:       "low",
				Title:          "Zero-Value Position",
				Description:    fmt.Sprintf("%s has 0%% portfolio allocation (zero market value).", symbol),
				Recommendation: "Consider removing this position or investigating why the value is zero.",
			})
		}
	}

	return insights
}

type aiInsightsResponse struct {
	Insights      []Insight `json:"insights"`
	OverallHealth string    `json:"overall_health"`
}

func (c *Client) generateAIInsights(ctx context.Context, metrics portfolio.Metrics, positions []models.Position, diff *models.Diff) ([]Insight, string, error) {
	prompt := buildInsightsPrompt(metrics, positions, diff)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: insightsSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, "", fmt.Errorf("claude insights: %w", err)
	}

	responseText := ""
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	var result aiInsightsResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, "", fmt.Errorf("parse insights response: %w", err)
	}

	return result.Insights, result.OverallHealth, nil
}

const insightsSystemPrompt = `You are a portfolio risk analyst for FlowPilot Finance. Analyze the portfolio data and return structured insights about anomalies, risks, and opportunities.

You MUST respond with ONLY valid JSON in this exact format (no markdown, no code fences):
{
  "insights": [
    {
      "type": "concentration_warning|allocation_drift|correlation_risk|drawdown_alert|position_anomaly|sector_risk",
      "severity": "high|medium|low",
      "title": "Short descriptive title",
      "description": "Detailed explanation of the finding",
      "recommendation": "Actionable suggestion"
    }
  ],
  "overall_health": "good|fair|poor"
}

Focus on detecting:
- Unusual concentration patterns beyond simple thresholds
- Abnormal portfolio shifts compared to previous state
- Large correlation exposure (e.g., multiple correlated assets)
- Sudden drawdowns or value changes
- Position-level anomalies (unusual quantities, prices)
- Sector or asset class imbalances

Only include meaningful insights. If the portfolio is well-balanced, return fewer insights with "good" health.`

func buildInsightsPrompt(metrics portfolio.Metrics, positions []models.Position, diff *models.Diff) string {
	prompt := fmt.Sprintf("Analyze this portfolio for anomalies and risks:\n\nOverview:\n- Net Worth: $%.2f\n- Total Positions: %d\n\n", metrics.NetWorth, metrics.PositionCount)

	prompt += "Exposure by Asset Type:\n"
	for assetType, pct := range metrics.ExposureByType {
		prompt += fmt.Sprintf("- %s: %.1f%%\n", assetType, pct)
	}

	prompt += "\nPosition Details:\n"
	for _, p := range positions {
		prompt += fmt.Sprintf("- %s: %.4f units @ $%.2f = $%.2f (%s, %s)", p.Symbol, p.Quantity, p.MarkPrice, p.MarketValue, p.AssetType, p.Source)
		if p.DayChangePct != 0 {
			prompt += fmt.Sprintf(" [day: %+.1f%%]", p.DayChangePct)
		}
		if p.TotalPnLPct != 0 {
			prompt += fmt.Sprintf(" [total P&L: %+.1f%%]", p.TotalPnLPct)
		}
		prompt += "\n"
	}

	prompt += "\nConcentration by Symbol:\n"
	for symbol, pct := range metrics.ConcentrationBySymbol {
		prompt += fmt.Sprintf("- %s: %.1f%%\n", symbol, pct)
	}

	if diff != nil {
		prompt += fmt.Sprintf("\nRecent Changes:\n- Net Worth Change: $%.2f\n", diff.NetWorthChange)
		if len(diff.TopMovers) > 0 {
			prompt += "Top Movers:\n"
			for _, m := range diff.TopMovers {
				prompt += fmt.Sprintf("- %s: %.1f%% ($%.2f -> $%.2f)\n", m.Symbol, m.ChangePercent, m.OldValue, m.NewValue)
			}
		}
		if len(diff.AddedSymbols) > 0 {
			prompt += fmt.Sprintf("New Positions: %v\n", diff.AddedSymbols)
		}
		if len(diff.RemovedSymbols) > 0 {
			prompt += fmt.Sprintf("Removed Positions: %v\n", diff.RemovedSymbols)
		}
	}

	prompt += "\nReturn JSON insights focusing on risks not covered by simple threshold checks (position > 20%, asset type > 50%, < 5 positions). Look for correlations, sector risks, and anomalies."
	return prompt
}

func computeHealthFromRules(insights []Insight) string {
	highCount := 0
	medCount := 0
	for _, i := range insights {
		switch i.Severity {
		case "high":
			highCount++
		case "medium":
			medCount++
		}
	}
	if highCount >= 2 {
		return "poor"
	}
	if highCount >= 1 || medCount >= 2 {
		return "fair"
	}
	return "good"
}
