package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AIDigest struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SnapshotID bson.ObjectID `bson:"snapshot_id" json:"snapshot_id"`
	Summary    string        `bson:"summary" json:"summary"`
	RiskNotes  string        `bson:"risk_notes" json:"risk_notes"`
	Timestamp  time.Time     `bson:"timestamp" json:"timestamp"`
}

type NewsItem struct {
	ID          string    `bson:"id" json:"id"`
	Title       string    `bson:"title" json:"title"`
	Summary     string    `bson:"summary" json:"summary"`
	Source      string    `bson:"source" json:"source"`
	URL         string    `bson:"url" json:"url"`
	Symbols     []string  `bson:"symbols" json:"symbols"`
	Sentiment   string    `bson:"sentiment" json:"sentiment"`     // positive, negative, neutral
	Impact      string    `bson:"impact" json:"impact"`           // high, medium, low
	Relevance   float64   `bson:"relevance" json:"relevance"`     // 0-1 score
	PublishedAt time.Time `bson:"published_at" json:"published_at"`
	Category    string    `bson:"category" json:"category"`       // earnings, macro, sector, crypto
}

type NewsFeed struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Items           []NewsItem    `bson:"items" json:"items"`
	GeneratedAt     time.Time     `bson:"generated_at" json:"generated_at"`
	HoldingsAnalyzed []string     `bson:"holdings_analyzed" json:"holdings_analyzed"`
}

type Note struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol    string        `bson:"symbol" json:"symbol"`   // "" for general notes
	Title     string        `bson:"title" json:"title"`
	Content   string        `bson:"content" json:"content"` // Tiptap JSON string
	Tags      []string      `bson:"tags" json:"tags"`
	Pinned    bool          `bson:"pinned" json:"pinned"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

type PredictionMarket struct {
	ID             string    `bson:"_id" json:"id"`
	Source         string    `bson:"source" json:"source"`
	Title          string    `bson:"title" json:"title"`
	Description    string    `bson:"description,omitempty" json:"description,omitempty"`
	OutcomeYes     float64   `bson:"outcome_yes" json:"outcome_yes"`
	OutcomeNo      float64   `bson:"outcome_no" json:"outcome_no"`
	Volume         float64   `bson:"volume" json:"volume"`
	EndDate        string    `bson:"end_date" json:"end_date"`
	Category       string    `bson:"category" json:"category"`
	RelatedSymbols []string  `bson:"related_symbols" json:"related_symbols"`
	URL            string    `bson:"url" json:"url"`
	LastSynced     time.Time `bson:"last_synced" json:"last_synced"`
}
