package portfolio

import (
	"context"
	"time"

	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CostBasisStore struct {
	db *db.MongoDB
}

func NewCostBasisStore(db *db.MongoDB) *CostBasisStore {
	return &CostBasisStore{db: db}
}

func (s *CostBasisStore) Get(ctx context.Context, symbol, source, accountID string) (*models.CostBasis, error) {
	filter := bson.D{
		{Key: "symbol", Value: symbol},
		{Key: "source", Value: source},
		{Key: "account_id", Value: accountID},
	}
	var cb models.CostBasis
	err := s.db.CostBasis().FindOne(ctx, filter).Decode(&cb)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &cb, nil
}

func (s *CostBasisStore) Upsert(ctx context.Context, symbol, source, accountID string, avgCost float64) error {
	filter := bson.D{
		{Key: "symbol", Value: symbol},
		{Key: "source", Value: source},
		{Key: "account_id", Value: accountID},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "symbol", Value: symbol},
			{Key: "source", Value: source},
			{Key: "account_id", Value: accountID},
			{Key: "avg_cost", Value: avgCost},
			{Key: "updated_at", Value: time.Now()},
		}},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := s.db.CostBasis().UpdateOne(ctx, filter, update, opts)
	return err
}
