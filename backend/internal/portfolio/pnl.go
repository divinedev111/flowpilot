package portfolio

import (
	"context"
	"log"

	"flowpilot/internal/models"
)

// EnrichPnL computes P&L fields for each position by looking up cost basis
// and comparing against previous snapshot prices for day change.
// If no cost basis exists for a position, the current mark price is stored
// as the initial cost basis.
func EnrichPnL(ctx context.Context, positions []models.Position, cbStore *CostBasisStore, prevPrices map[string]float64) []models.Position {
	enriched := make([]models.Position, len(positions))
	for i, p := range positions {
		enriched[i] = p

		// Look up or initialize cost basis
		cb, err := cbStore.Get(ctx, p.Symbol, p.Source, p.AccountID)
		if err != nil {
			log.Printf("cost basis lookup error for %s: %v", p.Symbol, err)
			continue
		}

		var avgCost float64
		if cb == nil {
			// First sync: set current price as cost basis
			avgCost = p.MarkPrice
			if err := cbStore.Upsert(ctx, p.Symbol, p.Source, p.AccountID, avgCost); err != nil {
				log.Printf("cost basis upsert error for %s: %v", p.Symbol, err)
			}
		} else {
			avgCost = cb.AvgCost
		}

		enriched[i].CostBasis = avgCost

		// Total P&L
		enriched[i].TotalPnL = (p.MarkPrice - avgCost) * p.Quantity
		if avgCost > 0 {
			enriched[i].TotalPnLPct = ((p.MarkPrice - avgCost) / avgCost) * 100
		}

		// Day change: compare to previous snapshot price
		key := p.Symbol + "|" + p.Source + "|" + p.AccountID
		if prevPrice, ok := prevPrices[key]; ok && prevPrice > 0 {
			enriched[i].DayChange = (p.MarkPrice - prevPrice) * p.Quantity
			enriched[i].DayChangePct = ((p.MarkPrice - prevPrice) / prevPrice) * 100
		}
	}
	return enriched
}

// BuildPrevPriceMap builds a lookup map from snapshot positions keyed by
// "symbol|source|account_id". Since SnapshotPosition doesn't carry source
// and account_id, we use just the symbol as the key.
func BuildPrevPriceMap(prevPositions []models.SnapshotPosition) map[string]float64 {
	m := make(map[string]float64, len(prevPositions))
	for _, sp := range prevPositions {
		// SnapshotPosition only has Symbol, so we key by symbol with empty source/account
		// The sync handler will build a better map using current positions to match.
		m[sp.Symbol] = sp.MarkPrice
	}
	return m
}
