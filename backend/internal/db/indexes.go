package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (m *MongoDB) EnsureIndexes(ctx context.Context) error {
	indexes := map[*mongo.Collection][]mongo.IndexModel{
		m.PositionsCurrent(): {
			{
				Keys: bson.D{
					{Key: "symbol", Value: 1},
					{Key: "source", Value: 1},
					{Key: "account_id", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{Keys: bson.D{{Key: "asset_type", Value: 1}}},
		},
		m.SnapshotPositions(): {
			{Keys: bson.D{{Key: "snapshot_id", Value: 1}}},
		},
		m.Diffs(): {
			{Keys: bson.D{{Key: "to_snapshot", Value: 1}}},
		},
		m.AlertEvents(): {
			{Keys: bson.D{{Key: "timestamp", Value: -1}}},
			{Keys: bson.D{{Key: "rule_id", Value: 1}}},
		},
		m.AIDigests(): {
			{Keys: bson.D{{Key: "snapshot_id", Value: 1}}},
		},
		m.CostBasis(): {
			{
				Keys: bson.D{
					{Key: "symbol", Value: 1},
					{Key: "source", Value: 1},
					{Key: "account_id", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		},
		m.AuditLogs(): {
			{Keys: bson.D{{Key: "timestamp", Value: -1}}},
			{Keys: bson.D{{Key: "event_type", Value: 1}}},
		},
		m.PredictionMarkets(): {
			{Keys: bson.D{{Key: "last_synced", Value: -1}}},
			{Keys: bson.D{{Key: "category", Value: 1}}},
		},
	}

	for coll, idxModels := range indexes {
		for _, idx := range idxModels {
			if _, err := coll.Indexes().CreateOne(ctx, idx); err != nil {
				return err
			}
		}
	}

	return nil
}
