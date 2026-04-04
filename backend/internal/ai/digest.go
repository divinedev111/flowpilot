package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"
)

func (c *Client) GenerateDigest(ctx context.Context, metrics portfolio.Metrics, diff *models.Diff, alertEvents []models.AlertEvent) (summary string, riskNotes string, err error) {
	prompt := buildDigestPrompt(metrics, diff, alertEvents)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("claude digest: %w", err)
	}

	responseText := ""
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	summary, riskNotes = parseDigestResponse(responseText)

	return summary, riskNotes, nil
}

func parseDigestResponse(response string) (summary string, riskNotes string) {
	// Try common delimiters
	delimiters := []string{"RISK NOTES:", "Risk Notes:", "RISK:", "Risk Factors:", "Key Risks:", "KEY RISKS:"}
	for _, delim := range delimiters {
		idx := strings.Index(response, delim)
		if idx >= 0 {
			summary = strings.TrimSpace(response[:idx])
			riskNotes = strings.TrimSpace(response[idx+len(delim):])
			return summary, riskNotes
		}
	}

	// Try splitting on numbered sections: "2." or "2)"
	if idx := strings.Index(response, "\n2."); idx >= 0 {
		summary = strings.TrimSpace(response[:idx])
		riskNotes = strings.TrimSpace(response[idx+3:])
		return summary, riskNotes
	}
	if idx := strings.Index(response, "\n2)"); idx >= 0 {
		summary = strings.TrimSpace(response[:idx])
		riskNotes = strings.TrimSpace(response[idx+3:])
		return summary, riskNotes
	}

	// Fallback: everything is summary
	return strings.TrimSpace(response), ""
}

func buildDigestPrompt(metrics portfolio.Metrics, diff *models.Diff, alertEvents []models.AlertEvent) string {
	prompt := fmt.Sprintf("Portfolio Overview:\n- Net Worth: $%.2f\n- Positions: %d\n\n", metrics.NetWorth, metrics.PositionCount)

	prompt += "Exposure by Asset Type:\n"
	for assetType, pct := range metrics.ExposureByType {
		prompt += fmt.Sprintf("- %s: %.1f%%\n", assetType, pct)
	}

	prompt += "\nTop Holdings by Concentration:\n"
	for symbol, pct := range metrics.ConcentrationBySymbol {
		prompt += fmt.Sprintf("- %s: %.1f%%\n", symbol, pct)
	}

	if diff != nil {
		prompt += fmt.Sprintf("\nChanges Since Last Snapshot:\n- Net Worth Change: $%.2f\n", diff.NetWorthChange)
		if len(diff.TopMovers) > 0 {
			prompt += "Top Movers:\n"
			for _, m := range diff.TopMovers {
				prompt += fmt.Sprintf("- %s: %.1f%% ($%.2f → $%.2f)\n", m.Symbol, m.ChangePercent, m.OldValue, m.NewValue)
			}
		}
		if len(diff.AddedSymbols) > 0 {
			prompt += fmt.Sprintf("New Positions: %v\n", diff.AddedSymbols)
		}
		if len(diff.RemovedSymbols) > 0 {
			prompt += fmt.Sprintf("Removed Positions: %v\n", diff.RemovedSymbols)
		}
	}

	if len(alertEvents) > 0 {
		prompt += "\nTriggered Alerts:\n"
		for _, e := range alertEvents {
			prompt += fmt.Sprintf("- [%s] %s\n", e.Severity, e.Message)
		}
	}

	prompt += "\nPlease provide:\n1. A concise portfolio summary\n2. Key risk factors and observations"
	return prompt
}
