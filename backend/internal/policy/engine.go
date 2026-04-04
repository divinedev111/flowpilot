package policy

import (
	"context"
	"fmt"
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

func (e *Engine) GetRules(ctx context.Context) ([]models.PolicyRule, error) {
	cursor, err := e.db.PolicyRules().Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find policy rules: %w", err)
	}

	var rules []models.PolicyRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, fmt.Errorf("decode policy rules: %w", err)
	}
	return rules, nil
}

func (e *Engine) CreateRule(ctx context.Context, rule models.PolicyRule) (*models.PolicyRule, error) {
	rule.CreatedAt = time.Now()
	result, err := e.db.PolicyRules().InsertOne(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("insert policy rule: %w", err)
	}
	rule.ID = result.InsertedID.(bson.ObjectID)
	return &rule, nil
}

func (e *Engine) UpdateRule(ctx context.Context, id bson.ObjectID, rule models.PolicyRule) error {
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: rule.Name},
		{Key: "rule_type", Value: rule.RuleType},
		{Key: "threshold_pct", Value: rule.ThresholdPct},
		{Key: "asset_type", Value: rule.AssetType},
		{Key: "target_symbol", Value: rule.TargetSymbol},
		{Key: "severity", Value: rule.Severity},
		{Key: "enabled", Value: rule.Enabled},
	}}}
	_, err := e.db.PolicyRules().UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
	if err != nil {
		return fmt.Errorf("update policy rule: %w", err)
	}
	return nil
}

func (e *Engine) DeleteRule(ctx context.Context, id bson.ObjectID) error {
	_, err := e.db.PolicyRules().DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return fmt.Errorf("delete policy rule: %w", err)
	}
	return nil
}

// CheckCompliance loads enabled rules and evaluates them against the provided
// metrics, returning any violations found.
func (e *Engine) CheckCompliance(ctx context.Context, metrics portfolio.Metrics) ([]models.PolicyViolation, error) {
	cursor, err := e.db.PolicyRules().Find(ctx, bson.D{{Key: "enabled", Value: true}})
	if err != nil {
		return nil, fmt.Errorf("load policy rules: %w", err)
	}

	var rules []models.PolicyRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, fmt.Errorf("decode policy rules: %w", err)
	}

	violations := Evaluate(rules, metrics)

	if len(violations) > 0 {
		docs := make([]interface{}, len(violations))
		for i := range violations {
			docs[i] = violations[i]
		}
		if _, err := e.db.PolicyViolations().InsertMany(ctx, docs); err != nil {
			return nil, fmt.Errorf("store policy violations: %w", err)
		}
	}

	return violations, nil
}

func (e *Engine) GetViolations(ctx context.Context) ([]models.PolicyViolation, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(100)

	cursor, err := e.db.PolicyViolations().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find policy violations: %w", err)
	}

	var violations []models.PolicyViolation
	if err := cursor.All(ctx, &violations); err != nil {
		return nil, fmt.Errorf("decode policy violations: %w", err)
	}
	return violations, nil
}

// Evaluate is a pure function that checks each enabled rule against the current
// metrics, returning any policy violations. It requires no database access,
// making it straightforward to unit-test.
func Evaluate(rules []models.PolicyRule, metrics portfolio.Metrics) []models.PolicyViolation {
	var violations []models.PolicyViolation
	now := time.Now()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		switch rule.RuleType {
		case "max_concentration":
			violations = evalMaxConcentration(rule, metrics, now, violations)
		case "min_positions":
			violations = evalMinPositions(rule, metrics, now, violations)
		case "max_exposure":
			violations = evalMaxExposure(rule, metrics, now, violations)
		case "max_position_count":
			violations = evalMaxPositionCount(rule, metrics, now, violations)
		}
	}

	return violations
}

// evalMaxConcentration fires when any single position exceeds the threshold
// percentage of total portfolio value.
func evalMaxConcentration(rule models.PolicyRule, metrics portfolio.Metrics, now time.Time, violations []models.PolicyViolation) []models.PolicyViolation {
	var violatingSymbols []string
	var maxPct float64

	for symbol, pct := range metrics.ConcentrationBySymbol {
		if rule.TargetSymbol != "" && symbol != rule.TargetSymbol {
			continue
		}
		if pct > rule.ThresholdPct {
			violatingSymbols = append(violatingSymbols, symbol)
			if pct > maxPct {
				maxPct = pct
			}
		}
	}

	if len(violatingSymbols) == 0 {
		return violations
	}

	return append(violations, models.PolicyViolation{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Severity:     rule.Severity,
		Message:      fmt.Sprintf("Position concentration exceeds %.1f%% limit", rule.ThresholdPct),
		CurrentValue: maxPct,
		Threshold:    rule.ThresholdPct,
		Symbols:      violatingSymbols,
		Timestamp:    now,
	})
}

// evalMinPositions fires when the portfolio has fewer positions than the
// threshold requires.
func evalMinPositions(rule models.PolicyRule, metrics portfolio.Metrics, now time.Time, violations []models.PolicyViolation) []models.PolicyViolation {
	count := float64(metrics.PositionCount)
	if count >= rule.ThresholdPct {
		return violations
	}

	return append(violations, models.PolicyViolation{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Severity:     rule.Severity,
		Message:      fmt.Sprintf("Portfolio has %d positions, minimum required is %.0f", metrics.PositionCount, rule.ThresholdPct),
		CurrentValue: count,
		Threshold:    rule.ThresholdPct,
		Timestamp:    now,
	})
}

// evalMaxExposure fires when an asset type's allocation exceeds the threshold
// percentage.
func evalMaxExposure(rule models.PolicyRule, metrics portfolio.Metrics, now time.Time, violations []models.PolicyViolation) []models.PolicyViolation {
	if rule.AssetType == "" {
		return violations
	}

	actual, ok := metrics.ExposureByType[rule.AssetType]
	if !ok || actual <= rule.ThresholdPct {
		return violations
	}

	return append(violations, models.PolicyViolation{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Severity:     rule.Severity,
		Message:      fmt.Sprintf("%s exposure at %.1f%% exceeds %.1f%% limit", rule.AssetType, actual, rule.ThresholdPct),
		CurrentValue: actual,
		Threshold:    rule.ThresholdPct,
		Symbols:      []string{rule.AssetType},
		Timestamp:    now,
	})
}

// evalMaxPositionCount fires when the portfolio has more positions than the
// threshold allows.
func evalMaxPositionCount(rule models.PolicyRule, metrics portfolio.Metrics, now time.Time, violations []models.PolicyViolation) []models.PolicyViolation {
	count := float64(metrics.PositionCount)
	if count <= rule.ThresholdPct {
		return violations
	}

	return append(violations, models.PolicyViolation{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Severity:     rule.Severity,
		Message:      fmt.Sprintf("Portfolio has %d positions, maximum allowed is %.0f", metrics.PositionCount, rule.ThresholdPct),
		CurrentValue: count,
		Threshold:    rule.ThresholdPct,
		Timestamp:    now,
	})
}
