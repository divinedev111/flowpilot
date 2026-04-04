package snapshot

import (
	"math"
	"sort"
	"time"

	"flowpilot/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ComputeDiff compares two sets of snapshot positions and produces a Diff
// containing net worth change, top movers sorted by absolute change percent,
// and lists of added/removed symbols.
func ComputeDiff(fromID, toID bson.ObjectID, oldNetWorth, newNetWorth float64, oldPositions, newPositions []models.SnapshotPosition) models.Diff {
	oldMap := make(map[string]models.SnapshotPosition)
	for _, p := range oldPositions {
		oldMap[p.Symbol] = p
	}

	newMap := make(map[string]models.SnapshotPosition)
	for _, p := range newPositions {
		newMap[p.Symbol] = p
	}

	var movers []models.Mover
	var added, removed []string

	// Find changes and new positions
	for sym, np := range newMap {
		op, existed := oldMap[sym]
		if !existed {
			added = append(added, sym)
			movers = append(movers, models.Mover{
				Symbol:   sym,
				OldValue: 0,
				NewValue: np.Quantity * np.MarkPrice,
			})
			continue
		}

		oldVal := op.Quantity * op.MarkPrice
		newVal := np.Quantity * np.MarkPrice
		if oldVal > 0 {
			changePct := ((newVal - oldVal) / oldVal) * 100
			movers = append(movers, models.Mover{
				Symbol:        sym,
				ChangePercent: changePct,
				OldValue:      oldVal,
				NewValue:      newVal,
			})
		}
	}

	// Find removed positions
	for sym, op := range oldMap {
		if _, exists := newMap[sym]; !exists {
			removed = append(removed, sym)
			movers = append(movers, models.Mover{
				Symbol:        sym,
				ChangePercent: -100,
				OldValue:      op.Quantity * op.MarkPrice,
				NewValue:      0,
			})
		}
	}

	// Sort movers by absolute change percent descending
	sort.Slice(movers, func(i, j int) bool {
		return math.Abs(movers[i].ChangePercent) > math.Abs(movers[j].ChangePercent)
	})

	// Keep top 10 movers
	if len(movers) > 10 {
		movers = movers[:10]
	}

	return models.Diff{
		FromSnapshot:   fromID,
		ToSnapshot:     toID,
		NetWorthChange: newNetWorth - oldNetWorth,
		TopMovers:      movers,
		AddedSymbols:   added,
		RemovedSymbols: removed,
		CreatedAt:      time.Now(),
	}
}
