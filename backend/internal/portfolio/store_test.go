package portfolio

import (
	"context"
	"os"
	"testing"
	"time"

	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.MongoDB {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongo, err := db.Connect(ctx, uri)
	require.NoError(t, err)
	mongo.Init("flowpilot_test")
	mongo.EnsureIndexes(ctx)

	t.Cleanup(func() {
		mongo.DB.Drop(context.Background())
		mongo.Disconnect(context.Background())
	})

	return mongo
}

func TestUpsertAndGetAll(t *testing.T) {
	mongoDB := setupTestDB(t)
	store := NewStore(mongoDB)
	ctx := context.Background()

	positions := []models.Position{
		{Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab", AccountID: "123", Timestamp: time.Now()},
		{Symbol: "BTC", Quantity: 1, MarkPrice: 60000, MarketValue: 60000, AssetType: "crypto", Source: "coinbase", AccountID: "456", Timestamp: time.Now()},
	}

	err := store.UpsertPositions(ctx, positions)
	require.NoError(t, err)

	all, err := store.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Upsert again with updated price — should not duplicate
	positions[0].MarkPrice = 160
	positions[0].MarketValue = 1600
	err = store.UpsertPositions(ctx, positions)
	require.NoError(t, err)

	all, err = store.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
	for _, p := range all {
		if p.Symbol == "AAPL" {
			assert.Equal(t, 160.0, p.MarkPrice)
		}
	}
}
