package alerts

import (
	"context"
	"fmt"
	"math"
	"time"

	"flowpilot/internal/db"
	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Engine struct {
	db *db.MongoDB
}

func NewEngine(db *db.MongoDB) *Engine {
	return &Engine{db: db}
}

func (e *Engine) Evaluate(ctx context.Context, metrics portfolio.Metrics, diff *models.Diff) ([]models.AlertEvent, error) {
	cursor, err := e.db.AlertRules().Find(ctx, bson.D{{Key: "enabled", Value: true}})
	if err != nil {
		return nil, fmt.Errorf("load alert rules: %w", err)
	}

	var rules []models.AlertRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, fmt.Errorf("decode alert rules: %w", err)
	}

	events := evaluateRules(rules, metrics, diff)
	if len(events) == 0 {
		return events, nil
	}

	docs := make([]interface{}, len(events))
	for i := range events {
		docs[i] = events[i]
	}
	if _, err := e.db.AlertEvents().InsertMany(ctx, docs); err != nil {
		return nil, fmt.Errorf("store alert events: %w", err)
	}

	return events, nil
}

func (e *Engine) GetRules(ctx context.Context) ([]models.AlertRule, error) {
	cursor, err := e.db.AlertRules().Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find alert rules: %w", err)
	}

	var rules []models.AlertRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, fmt.Errorf("decode alert rules: %w", err)
	}
	return rules, nil
}

func (e *Engine) CreateRule(ctx context.Context, rule models.AlertRule) (*models.AlertRule, error) {
	rule.CreatedAt = time.Now()
	result, err := e.db.AlertRules().InsertOne(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("insert alert rule: %w", err)
	}
	rule.ID = result.InsertedID.(bson.ObjectID)
	return &rule, nil
}

func (e *Engine) GetEvents(ctx context.Context) ([]models.AlertEvent, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(100)

	cursor, err := e.db.AlertEvents().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find alert events: %w", err)
	}

	var events []models.AlertEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("decode alert events: %w", err)
	}
	return events, nil
}

// EvaluateRulesForBacktest is the exported wrapper around evaluateRules for
// backtesting purposes.
func EvaluateRulesForBacktest(rules []models.AlertRule, metrics portfolio.Metrics, diff *models.Diff) []models.AlertEvent {
	return evaluateRules(rules, metrics, diff)
}

// evaluateRules is a pure function that checks each rule against the current
// metrics and diff, returning triggered alert events. It requires no database
// access, making it straightforward to unit-test.
func evaluateRules(rules []models.AlertRule, metrics portfolio.Metrics, diff *models.Diff) []models.AlertEvent {
	var events []models.AlertEvent
	now := time.Now()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		switch rule.RuleType {
		case "exposure_threshold":
			events = evalExposureThreshold(rule, metrics, now, events)
		case "position_change_percent":
			events = evalPositionChangePercent(rule, diff, now, events)
		case "concentration_threshold":
			events = evalConcentrationThreshold(rule, metrics, now, events)
		case "net_worth_change":
			events = evalNetWorthChange(rule, diff, now, events)
		}
	}

	return events
}

// evalExposureThreshold fires when an asset type's allocation exceeds the
// threshold percentage.
func evalExposureThreshold(rule models.AlertRule, metrics portfolio.Metrics, now time.Time, events []models.AlertEvent) []models.AlertEvent {
	actual, ok := metrics.ExposureByType[rule.TargetAsset]
	if !ok || actual <= rule.Threshold {
		return events
	}

	return append(events, models.AlertEvent{
		RuleID:    rule.ID,
		Severity:  severity(actual, rule.Threshold),
		Message:   fmt.Sprintf("%s exposure at %.1f%% exceeds %.1f%% threshold", rule.TargetAsset, actual, rule.Threshold),
		Evidence:  fmt.Sprintf("exposure_%s=%.2f", rule.TargetAsset, actual),
		Timestamp: now,
	})
}

// evalPositionChangePercent fires when any position's value changed more than
// the threshold percentage since the last snapshot.
func evalPositionChangePercent(rule models.AlertRule, diff *models.Diff, now time.Time, events []models.AlertEvent) []models.AlertEvent {
	if diff == nil {
		return events
	}

	for _, m := range diff.TopMovers {
		if rule.TargetAsset != "" && m.Symbol != rule.TargetAsset {
			continue
		}

		absChange := math.Abs(m.ChangePercent)
		if absChange <= rule.Threshold {
			continue
		}

		events = append(events, models.AlertEvent{
			RuleID:    rule.ID,
			Severity:  severity(absChange, rule.Threshold),
			Message:   fmt.Sprintf("%s changed %.1f%% (threshold %.1f%%)", m.Symbol, m.ChangePercent, rule.Threshold),
			Evidence:  fmt.Sprintf("symbol=%s change=%.2f old=%.2f new=%.2f", m.Symbol, m.ChangePercent, m.OldValue, m.NewValue),
			Timestamp: now,
		})
	}

	return events
}

// evalConcentrationThreshold fires when a single position exceeds the
// threshold percentage of total portfolio value.
func evalConcentrationThreshold(rule models.AlertRule, metrics portfolio.Metrics, now time.Time, events []models.AlertEvent) []models.AlertEvent {
	for symbol, pct := range metrics.ConcentrationBySymbol {
		if rule.TargetAsset != "" && symbol != rule.TargetAsset {
			continue
		}

		if pct <= rule.Threshold {
			continue
		}

		events = append(events, models.AlertEvent{
			RuleID:    rule.ID,
			Severity:  severity(pct, rule.Threshold),
			Message:   fmt.Sprintf("%s concentration at %.1f%% exceeds %.1f%% threshold", symbol, pct, rule.Threshold),
			Evidence:  fmt.Sprintf("symbol=%s concentration=%.2f", symbol, pct),
			Timestamp: now,
		})
	}

	return events
}

// evalNetWorthChange fires when the total portfolio value changed more than
// the threshold percentage since the last snapshot.
func evalNetWorthChange(rule models.AlertRule, diff *models.Diff, now time.Time, events []models.AlertEvent) []models.AlertEvent {
	if diff == nil {
		return events
	}

	absChange := math.Abs(diff.NetWorthChange)
	if absChange <= rule.Threshold {
		return events
	}

	return append(events, models.AlertEvent{
		RuleID:    rule.ID,
		Severity:  severity(absChange, rule.Threshold),
		Message:   fmt.Sprintf("net worth changed %.1f%% (threshold %.1f%%)", diff.NetWorthChange, rule.Threshold),
		Evidence:  fmt.Sprintf("net_worth_change=%.2f", diff.NetWorthChange),
		Timestamp: now,
	})
}

// severity determines alert severity based on how far the actual value
// exceeds the threshold.
func severity(actual, threshold float64) string {
	if threshold == 0 {
		return "critical"
	}
	ratio := actual / threshold
	switch {
	case ratio > 2.0:
		return "critical"
	case ratio > 1.5:
		return "high"
	default:
		return "medium"
	}
}
