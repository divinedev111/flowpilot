package snapshot

import (
	"context"
	"os"
	"testing"
	"time"

	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
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

	t.Cleanup(func() {
		mongo.DB.Drop(context.Background())
		mongo.Disconnect(context.Background())
	})

	return mongo
}

func TestCreateSnapshot(t *testing.T) {
	mongoDB := setupTestDB(t)
	engine := NewEngine(mongoDB)
	ctx := context.Background()

	positions := []models.Position{
		{Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab", AccountID: "123", Timestamp: time.Now()},
		{Symbol: "BTC", Quantity: 1, MarkPrice: 60000, MarketValue: 60000, AssetType: "crypto", Source: "coinbase", AccountID: "456", Timestamp: time.Now()},
	}

	snap, err := engine.Create(ctx, positions, 61500.0)
	require.NoError(t, err)
	assert.Equal(t, 61500.0, snap.NetWorth)
	assert.False(t, snap.ID.IsZero())

	// Verify snapshot positions were stored
	cursor, err := mongoDB.SnapshotPositions().Find(ctx, bson.D{})
	require.NoError(t, err)
	var snapPositions []models.SnapshotPosition
	require.NoError(t, cursor.All(ctx, &snapPositions))
	assert.Len(t, snapPositions, 2)
}

func TestGetLatest(t *testing.T) {
	mongoDB := setupTestDB(t)
	engine := NewEngine(mongoDB)
	ctx := context.Background()

	// Create two snapshots with different net worths
	positions := []models.Position{
		{Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab", AccountID: "123", Timestamp: time.Now()},
	}

	_, err := engine.Create(ctx, positions, 1500.0)
	require.NoError(t, err)

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	snap2, err := engine.Create(ctx, positions, 2000.0)
	require.NoError(t, err)

	latest, err := engine.GetLatest(ctx)
	require.NoError(t, err)
	assert.Equal(t, snap2.NetWorth, latest.NetWorth)
	assert.Equal(t, snap2.ID, latest.ID)
}

func TestGetPositions(t *testing.T) {
	mongoDB := setupTestDB(t)
	engine := NewEngine(mongoDB)
	ctx := context.Background()

	positions := []models.Position{
		{Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab", AccountID: "123", Timestamp: time.Now()},
		{Symbol: "BTC", Quantity: 1, MarkPrice: 60000, MarketValue: 60000, AssetType: "crypto", Source: "coinbase", AccountID: "456", Timestamp: time.Now()},
	}

	snap, err := engine.Create(ctx, positions, 61500.0)
	require.NoError(t, err)

	snapPositions, err := engine.GetPositions(ctx, snap.ID)
	require.NoError(t, err)
	assert.Len(t, snapPositions, 2)

	// Verify position data was correctly stored
	symbols := make(map[string]float64)
	for _, sp := range snapPositions {
		symbols[sp.Symbol] = sp.MarkPrice
		assert.Equal(t, snap.ID, sp.SnapshotID)
	}
	assert.Equal(t, 150.0, symbols["AAPL"])
	assert.Equal(t, 60000.0, symbols["BTC"])
}
