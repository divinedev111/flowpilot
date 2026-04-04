package audit

import (
	"context"
	"time"

	"flowpilot/internal/db"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Event struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp  time.Time     `bson:"timestamp" json:"timestamp"`
	EventType  string        `bson:"event_type" json:"event_type"`   // "auth.login", "portfolio.sync", "alert.create", etc.
	Action     string        `bson:"action" json:"action"`           // "create", "read", "update", "delete"
	Resource   string        `bson:"resource" json:"resource"`       // "oauth_tokens", "positions", "alert_rules"
	ResourceID string        `bson:"resource_id,omitempty" json:"resource_id,omitempty"`
	Details    bson.M        `bson:"details,omitempty" json:"details,omitempty"`
	IPAddress  string        `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	Success    bool          `bson:"success" json:"success"`
	ErrorMsg   string        `bson:"error_msg,omitempty" json:"error_msg,omitempty"`
}

type Logger struct {
	db *db.MongoDB
}

func NewLogger(db *db.MongoDB) *Logger {
	return &Logger{db: db}
}

func (l *Logger) Log(ctx context.Context, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	_, err := l.db.AuditLogs().InsertOne(ctx, event)
	return err
}

// GetEvents returns recent audit events, optionally filtered by event type.
// If eventType is empty, all events are returned. Results are sorted by
// timestamp descending and limited to the given count.
func (l *Logger) GetEvents(ctx context.Context, eventType string, limit int) ([]Event, error) {
	filter := bson.M{}
	if eventType != "" {
		filter["event_type"] = eventType
	}

	if limit <= 0 {
		limit = 100
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := l.db.AuditLogs().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}
