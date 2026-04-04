package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (d *Deps) handleNewsFeed(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current positions
	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	if len(positions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"items":             []any{},
			"generated_at":      nil,
			"holdings_analyzed": []string{},
		})
		return
	}

	// Extract symbols: stock/crypto tickers + prediction market topics
	symbolSet := make(map[string]bool)
	for _, p := range positions {
		if p.AssetType == "prediction_market" {
			// Extract key topic from prediction market title for news search
			topic := extractPredictionTopic(p.Symbol)
			if topic != "" && !symbolSet[topic] {
				symbolSet[topic] = true
			}
			continue
		}
		sym := strings.ToUpper(p.Symbol)
		if sym != "" {
			symbolSet[sym] = true
		}
	}
	symbols := make([]string, 0, len(symbolSet))
	for sym := range symbolSet {
		symbols = append(symbols, sym)
	}

	if d.News == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "news service not available"})
		return
	}

	feed, err := d.News.GetFeed(ctx, symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate news: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// extractPredictionTopic extracts a searchable topic from a Polymarket position symbol.
// Symbols look like "Will John Cornyn win the 2026 Texas Repu... [Yes]"
func extractPredictionTopic(symbol string) string {
	// Remove outcome bracket like "[Yes]", "[No]", "[Up]", "[Down]"
	if idx := strings.LastIndex(symbol, " ["); idx > 0 {
		symbol = symbol[:idx]
	}
	// Remove "Will " prefix and trailing "..."
	symbol = strings.TrimPrefix(symbol, "Will ")
	symbol = strings.TrimSuffix(symbol, "...")
	// Remove "be the", "beat quarterly", etc. to get the key entity
	// Just take the first meaningful chunk (company/person name)
	symbol = strings.TrimSpace(symbol)
	if len(symbol) > 40 {
		symbol = symbol[:40]
	}
	if len(symbol) < 5 {
		return ""
	}
	return symbol
}

func (d *Deps) handleSymbolNews(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	if d.News == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "news service not available"})
		return
	}

	ctx := c.Request.Context()
	feed, err := d.News.GetSymbolNews(ctx, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate news: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, feed)
}
