package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Position struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol       string        `bson:"symbol" json:"symbol"`
	Quantity     float64       `bson:"quantity" json:"quantity"`
	MarkPrice    float64       `bson:"mark_price" json:"mark_price"`
	MarketValue  float64       `bson:"market_value" json:"market_value"`
	AssetType    string        `bson:"asset_type" json:"asset_type"`
	Source       string        `bson:"source" json:"source"`
	AccountID    string        `bson:"account_id" json:"account_id"`
	Timestamp    time.Time     `bson:"timestamp" json:"timestamp"`
	CostBasis    float64       `bson:"cost_basis" json:"cost_basis"`
	DayChange    float64       `bson:"day_change" json:"day_change"`
	DayChangePct float64       `bson:"day_change_pct" json:"day_change_pct"`
	TotalPnL     float64       `bson:"total_pnl" json:"total_pnl"`
	TotalPnLPct  float64       `bson:"total_pnl_pct" json:"total_pnl_pct"`
}

type Snapshot struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	NetWorth  float64       `bson:"net_worth" json:"net_worth"`
}

type SnapshotPosition struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SnapshotID bson.ObjectID `bson:"snapshot_id" json:"snapshot_id"`
	Symbol     string        `bson:"symbol" json:"symbol"`
	Quantity   float64       `bson:"quantity" json:"quantity"`
	MarkPrice  float64       `bson:"mark_price" json:"mark_price"`
}

type Mover struct {
	Symbol        string  `bson:"symbol" json:"symbol"`
	ChangePercent float64 `bson:"change_percent" json:"change_percent"`
	OldValue      float64 `bson:"old_value" json:"old_value"`
	NewValue      float64 `bson:"new_value" json:"new_value"`
}

type Diff struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	FromSnapshot   bson.ObjectID `bson:"from_snapshot" json:"from_snapshot"`
	ToSnapshot     bson.ObjectID `bson:"to_snapshot" json:"to_snapshot"`
	NetWorthChange float64       `bson:"net_worth_change" json:"net_worth_change"`
	TopMovers      []Mover       `bson:"top_movers" json:"top_movers"`
	AddedSymbols   []string      `bson:"added_symbols" json:"added_symbols"`
	RemovedSymbols []string      `bson:"removed_symbols" json:"removed_symbols"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
}

type CostBasis struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol    string        `bson:"symbol" json:"symbol"`
	Source    string        `bson:"source" json:"source"`
	AccountID string       `bson:"account_id" json:"account_id"`
	AvgCost   float64      `bson:"avg_cost" json:"avg_cost"`
	UpdatedAt time.Time    `bson:"updated_at" json:"updated_at"`
}
