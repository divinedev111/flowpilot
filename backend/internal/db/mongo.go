package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// In the v2 driver, mongo.Connect does not accept a context;
// the context is used only for the ping.
func Connect(ctx context.Context, uri string) (*MongoDB, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &MongoDB{Client: client}, nil
}

func (m *MongoDB) Init(dbName string) {
	m.DB = m.Client.Database(dbName)
}

func (m *MongoDB) Disconnect(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) PositionsCurrent() *mongo.Collection {
	return m.DB.Collection("positions_current")
}

func (m *MongoDB) Snapshots() *mongo.Collection {
	return m.DB.Collection("snapshots")
}

func (m *MongoDB) SnapshotPositions() *mongo.Collection {
	return m.DB.Collection("snapshot_positions")
}

func (m *MongoDB) Diffs() *mongo.Collection {
	return m.DB.Collection("diffs")
}

func (m *MongoDB) AlertRules() *mongo.Collection {
	return m.DB.Collection("alert_rules")
}

func (m *MongoDB) AlertEvents() *mongo.Collection {
	return m.DB.Collection("alert_events")
}

func (m *MongoDB) AIDigests() *mongo.Collection {
	return m.DB.Collection("ai_digests")
}

func (m *MongoDB) CostBasis() *mongo.Collection {
	return m.DB.Collection("cost_basis")
}

func (m *MongoDB) NewsCache() *mongo.Collection {
	return m.DB.Collection("news_cache")
}

func (m *MongoDB) Watchlist() *mongo.Collection {
	return m.DB.Collection("watchlist")
}

func (m *MongoDB) OAuthTokens() *mongo.Collection {
	return m.DB.Collection("oauth_tokens")
}

func (m *MongoDB) Notes() *mongo.Collection {
	return m.DB.Collection("notes")
}

func (m *MongoDB) Strategies() *mongo.Collection {
	return m.DB.Collection("strategies")
}

func (m *MongoDB) StrategyAssignments() *mongo.Collection {
	return m.DB.Collection("strategy_assignments")
}

func (m *MongoDB) PolicyRules() *mongo.Collection {
	return m.DB.Collection("policy_rules")
}

func (m *MongoDB) PolicyViolations() *mongo.Collection {
	return m.DB.Collection("policy_violations")
}

func (m *MongoDB) BacktestResults() *mongo.Collection {
	return m.DB.Collection("backtest_results")
}

func (m *MongoDB) PriceCache() *mongo.Collection {
	return m.DB.Collection("price_cache")
}

func (m *MongoDB) CorrelationCache() *mongo.Collection {
	return m.DB.Collection("correlation_cache")
}

func (m *MongoDB) AuditLogs() *mongo.Collection {
	return m.DB.Collection("audit_logs")
}

func (m *MongoDB) PredictionMarkets() *mongo.Collection {
	return m.DB.Collection("prediction_markets")
}
