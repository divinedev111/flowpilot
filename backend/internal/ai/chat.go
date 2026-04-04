package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/anthropics/anthropic-sdk-go"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"
)

func BuildPortfolioContext(metrics portfolio.Metrics, positions []models.Position) string {
	ctx := fmt.Sprintf("Current Portfolio State:\n- Net Worth: $%.2f\n- Positions: %d\n\n", metrics.NetWorth, metrics.PositionCount)

	ctx += "Exposure by Asset Type:\n"
	for t, pct := range metrics.ExposureByType {
		ctx += fmt.Sprintf("- %s: %.1f%%\n", t, pct)
	}

	ctx += "\nHoldings:\n"
	for _, p := range positions {
		ctx += fmt.Sprintf("- %s: %.4f units @ $%.2f = $%.2f (%s, %s)\n",
			p.Symbol, p.Quantity, p.MarkPrice, p.MarketValue, p.AssetType, p.Source)
	}

	return ctx
}

// ChatStream sends a user message and streams the Claude response via a writer.
// The writer receives SSE-formatted events: "data: {json}\n\n" for each token chunk
// and "data: [DONE]\n\n" at the end.
func (c *Client) ChatStream(ctx context.Context, conversationID string, userMessage string, portfolioContext string, w io.Writer) error {
	c.mu.Lock()
	history := c.conversations[conversationID]
	// Add user message to history
	history = append(history, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))
	// Trim to max messages
	if len(history) > maxConversationMessages {
		history = history[len(history)-maxConversationMessages:]
	}
	c.conversations[conversationID] = history
	c.mu.Unlock()

	fullSystem := systemPrompt + "\n\n" + portfolioContext

	stream := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: fullSystem},
		},
		Messages: history,
	})
	defer stream.Close()

	assistantText := ""
	for stream.Next() {
		event := stream.Current()
		// For content_block_delta events, the Delta.Text field contains the text chunk.
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			chunk := event.Delta.Text
			assistantText += chunk
			// Write SSE event
			sseData := map[string]string{"text": chunk}
			jsonBytes, _ := json.Marshal(sseData)
			fmt.Fprintf(w, "data: %s\n\n", jsonBytes)
			if flusher, ok := w.(interface{ Flush() }); ok {
				flusher.Flush()
			}
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("claude stream: %w", err)
	}

	// Write done event
	fmt.Fprintf(w, "data: [DONE]\n\n")
	if flusher, ok := w.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	// Store assistant response in conversation history
	c.mu.Lock()
	c.conversations[conversationID] = append(c.conversations[conversationID],
		anthropic.NewAssistantMessage(anthropic.NewTextBlock(assistantText)))
	c.mu.Unlock()

	return nil
}
