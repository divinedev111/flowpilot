package snapshot

import (
	"testing"

	"flowpilot/internal/models"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestComputeDiff(t *testing.T) {
	fromID := bson.NewObjectID()
	toID := bson.NewObjectID()

	oldPositions := []models.SnapshotPosition{
		{SnapshotID: fromID, Symbol: "AAPL", Quantity: 10, MarkPrice: 150},
		{SnapshotID: fromID, Symbol: "GOOG", Quantity: 5, MarkPrice: 100},
	}

	newPositions := []models.SnapshotPosition{
		{SnapshotID: toID, Symbol: "AAPL", Quantity: 10, MarkPrice: 170},
		{SnapshotID: toID, Symbol: "BTC", Quantity: 1, MarkPrice: 60000},
	}

	diff := ComputeDiff(fromID, toID, 2000.0, 61700.0, oldPositions, newPositions)

	assert.Equal(t, fromID, diff.FromSnapshot)
	assert.Equal(t, toID, diff.ToSnapshot)
	assert.Equal(t, 59700.0, diff.NetWorthChange)
	assert.Contains(t, diff.AddedSymbols, "BTC")
	assert.Contains(t, diff.RemovedSymbols, "GOOG")
	assert.True(t, len(diff.TopMovers) > 0)
}
