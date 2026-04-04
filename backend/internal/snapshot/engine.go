package snapshot

import (
	"context"
	"time"

	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Engine struct {
	db *db.MongoDB
}

func NewEngine(db *db.MongoDB) *Engine {
	return &Engine{db: db}
}

// Create persists a new Snapshot with the given net worth and stores a
// SnapshotPosition for each position at their current mark prices.
func (e *Engine) Create(ctx context.Context, positions []models.Position, netWorth float64) (*models.Snapshot, error) {
	snap := models.Snapshot{
		CreatedAt: time.Now(),
		NetWorth:  netWorth,
	}

	res, err := e.db.Snapshots().InsertOne(ctx, snap)
	if err != nil {
		return nil, err
	}
	snap.ID = res.InsertedID.(bson.ObjectID)

	if len(positions) > 0 {
		snapPositions := make([]models.SnapshotPosition, len(positions))
		for i, p := range positions {
			snapPositions[i] = models.SnapshotPosition{
				SnapshotID: snap.ID,
				Symbol:     p.Symbol,
				Quantity:   p.Quantity,
				MarkPrice:  p.MarkPrice,
			}
		}

		if _, err := e.db.SnapshotPositions().InsertMany(ctx, snapPositions); err != nil {
			return nil, err
		}
	}

	return &snap, nil
}

func (e *Engine) GetLatest(ctx context.Context) (*models.Snapshot, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var snap models.Snapshot
	err := e.db.Snapshots().FindOne(ctx, bson.D{}, opts).Decode(&snap)
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (e *Engine) GetPositions(ctx context.Context, snapshotID bson.ObjectID) ([]models.SnapshotPosition, error) {
	cursor, err := e.db.SnapshotPositions().Find(ctx, bson.D{{Key: "snapshot_id", Value: snapshotID}})
	if err != nil {
		return nil, err
	}
	var positions []models.SnapshotPosition
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}
