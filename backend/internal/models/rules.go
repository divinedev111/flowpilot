package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AlertRule struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	RuleType    string        `bson:"rule_type" json:"rule_type"`
	Threshold   float64       `bson:"threshold" json:"threshold"`
	TargetAsset string        `bson:"target_asset" json:"target_asset"`
	Enabled     bool          `bson:"enabled" json:"enabled"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
}

type AlertEvent struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	RuleID    bson.ObjectID `bson:"rule_id" json:"rule_id"`
	Severity  string        `bson:"severity" json:"severity"`
	Message   string        `bson:"message" json:"message"`
	Evidence  string        `bson:"evidence" json:"evidence"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
}

type PolicyRule struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	RuleType     string        `bson:"rule_type" json:"rule_type"` // max_concentration, min_positions, max_exposure, max_position_count
	ThresholdPct float64       `bson:"threshold_pct" json:"threshold_pct"`
	AssetType    string        `bson:"asset_type,omitempty" json:"asset_type,omitempty"` // for exposure rules
	TargetSymbol string        `bson:"target_symbol,omitempty" json:"target_symbol,omitempty"`
	Severity     string        `bson:"severity" json:"severity"` // warning, critical
	Enabled      bool          `bson:"enabled" json:"enabled"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
}

type PolicyViolation struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	RuleID       bson.ObjectID `bson:"rule_id" json:"rule_id"`
	RuleName     string        `bson:"rule_name" json:"rule_name"`
	Severity     string        `bson:"severity" json:"severity"`
	Message      string        `bson:"message" json:"message"`
	CurrentValue float64       `bson:"current_value" json:"current_value"`
	Threshold    float64       `bson:"threshold" json:"threshold"`
	Symbols      []string      `bson:"symbols,omitempty" json:"symbols,omitempty"`
	Timestamp    time.Time     `bson:"timestamp" json:"timestamp"`
	Resolved     bool          `bson:"resolved" json:"resolved"`
}
