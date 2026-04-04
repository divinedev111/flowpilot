package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Strategy struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Color       string        `bson:"color" json:"color"` // hex color for UI
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

type StrategyAssignment struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	StrategyID bson.ObjectID `bson:"strategy_id" json:"strategy_id"`
	Symbol     string        `bson:"symbol" json:"symbol"`
	AssignedAt time.Time     `bson:"assigned_at" json:"assigned_at"`
}

type WatchlistItem struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol    string        `bson:"symbol" json:"symbol"`
	Name      string        `bson:"name" json:"name"`
	AssetType string        `bson:"asset_type" json:"asset_type"`
	Notes     string        `bson:"notes" json:"notes"`
	AddedAt   time.Time     `bson:"added_at" json:"added_at"`
}
