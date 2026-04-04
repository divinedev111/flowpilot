package portfolio

import (
	"context"

	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	db *db.MongoDB
}

func NewStore(db *db.MongoDB) *Store {
	return &Store{db: db}
}

// UpsertPositions inserts or updates positions keyed by (symbol, source, account_id).
// If a position with the same key already exists, it is updated in place.
func (s *Store) UpsertPositions(ctx context.Context, positions []models.Position) error {
	for _, p := range positions {
		filter := bson.D{
			{Key: "symbol", Value: p.Symbol},
			{Key: "source", Value: p.Source},
			{Key: "account_id", Value: p.AccountID},
		}
		update := bson.D{{Key: "$set", Value: p}}
		opts := options.UpdateOne().SetUpsert(true)

		if _, err := s.db.PositionsCurrent().UpdateOne(ctx, filter, update, opts); err != nil {
			return err
		}
	}
	return nil
}

// DeleteBySource removes all positions for a given source (e.g. "polymarket", "coinbase").
// Called before upserting to clean up stale positions that connectors no longer return.
func (s *Store) DeleteBySource(ctx context.Context, source string) error {
	_, err := s.db.PositionsCurrent().DeleteMany(ctx, bson.M{"source": source})
	return err
}

func (s *Store) GetAll(ctx context.Context) ([]models.Position, error) {
	cursor, err := s.db.PositionsCurrent().Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var positions []models.Position
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}
