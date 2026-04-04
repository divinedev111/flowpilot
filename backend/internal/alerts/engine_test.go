package alerts

import (
	"testing"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ruleID() bson.ObjectID {
	return bson.NewObjectID()
}

func TestExposureThreshold_Triggers(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "exposure_threshold", Threshold: 30.0, TargetAsset: "crypto", Enabled: true},
	}
	metrics := portfolio.Metrics{
		NetWorth:              100000,
		ExposureByType:        map[string]float64{"crypto": 35.0, "equity": 65.0},
		ConcentrationBySymbol: map[string]float64{},
	}

	events := evaluateRules(rules, metrics, nil)

	assert.Len(t, events, 1)
	assert.Equal(t, rules[0].ID, events[0].RuleID)
	assert.Equal(t, "medium", events[0].Severity)
	assert.Contains(t, events[0].Message, "crypto")
	assert.Contains(t, events[0].Message, "35.0%")
}

func TestConcentrationThreshold_Triggers(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "concentration_threshold", Threshold: 25.0, TargetAsset: "AAPL", Enabled: true},
	}
	metrics := portfolio.Metrics{
		NetWorth:              100000,
		ExposureByType:        map[string]float64{},
		ConcentrationBySymbol: map[string]float64{"AAPL": 60.0, "MSFT": 40.0},
	}

	events := evaluateRules(rules, metrics, nil)

	assert.Len(t, events, 1)
	assert.Equal(t, rules[0].ID, events[0].RuleID)
	assert.Equal(t, "critical", events[0].Severity) // 60/25 = 2.4 > 2x → critical
	assert.Contains(t, events[0].Message, "AAPL")
	assert.Contains(t, events[0].Message, "60.0%")
}

func TestNetWorthChange_Triggers(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "net_worth_change", Threshold: 5.0, Enabled: true},
	}
	metrics := portfolio.Metrics{}
	diff := &models.Diff{
		NetWorthChange: 8.0,
	}

	events := evaluateRules(rules, metrics, diff)

	assert.Len(t, events, 1)
	assert.Equal(t, rules[0].ID, events[0].RuleID)
	assert.Equal(t, "high", events[0].Severity) // 8/5 = 1.6 > 1.5x → high
	assert.Contains(t, events[0].Message, "8.0%")
}

func TestPositionChangePercent_Triggers(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "position_change_percent", Threshold: 10.0, Enabled: true},
	}
	metrics := portfolio.Metrics{}
	diff := &models.Diff{
		TopMovers: []models.Mover{
			{Symbol: "TSLA", ChangePercent: 15.0, OldValue: 1000, NewValue: 1150},
			{Symbol: "AAPL", ChangePercent: 3.0, OldValue: 2000, NewValue: 2060},
		},
	}

	events := evaluateRules(rules, metrics, diff)

	assert.Len(t, events, 1)
	assert.Equal(t, rules[0].ID, events[0].RuleID)
	assert.Equal(t, "medium", events[0].Severity) // 15/10 = 1.5 exactly, not >1.5, so medium
	assert.Contains(t, events[0].Message, "TSLA")
	assert.Contains(t, events[0].Message, "15.0%")
}

func TestDisabledRule_DoesNotTrigger(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "exposure_threshold", Threshold: 30.0, TargetAsset: "crypto", Enabled: false},
	}
	metrics := portfolio.Metrics{
		ExposureByType: map[string]float64{"crypto": 50.0},
	}

	events := evaluateRules(rules, metrics, nil)
	assert.Len(t, events, 0)
}

func TestRuleWithinThreshold_DoesNotTrigger(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "exposure_threshold", Threshold: 30.0, TargetAsset: "crypto", Enabled: true},
		{ID: ruleID(), RuleType: "concentration_threshold", Threshold: 50.0, TargetAsset: "AAPL", Enabled: true},
		{ID: ruleID(), RuleType: "net_worth_change", Threshold: 10.0, Enabled: true},
		{ID: ruleID(), RuleType: "position_change_percent", Threshold: 20.0, Enabled: true},
	}
	metrics := portfolio.Metrics{
		ExposureByType:        map[string]float64{"crypto": 25.0},
		ConcentrationBySymbol: map[string]float64{"AAPL": 40.0},
	}
	diff := &models.Diff{
		NetWorthChange: 3.0,
		TopMovers: []models.Mover{
			{Symbol: "AAPL", ChangePercent: 5.0, OldValue: 1000, NewValue: 1050},
		},
	}

	events := evaluateRules(rules, metrics, diff)
	assert.Len(t, events, 0)
}

func TestSeverityLevels(t *testing.T) {
	assert.Equal(t, "medium", severity(31, 30))    // 1.03x
	assert.Equal(t, "high", severity(46, 30))       // 1.53x
	assert.Equal(t, "critical", severity(61, 30))   // 2.03x
	assert.Equal(t, "critical", severity(10, 0))    // zero threshold
}

func TestPositionChangePercent_NegativeChange(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "position_change_percent", Threshold: 10.0, Enabled: true},
	}
	diff := &models.Diff{
		TopMovers: []models.Mover{
			{Symbol: "TSLA", ChangePercent: -15.0, OldValue: 1000, NewValue: 850},
		},
	}

	events := evaluateRules(rules, portfolio.Metrics{}, diff)

	assert.Len(t, events, 1)
	assert.Contains(t, events[0].Message, "TSLA")
	assert.Contains(t, events[0].Message, "-15.0%")
}

func TestPositionChangePercent_TargetAssetFilter(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "position_change_percent", Threshold: 10.0, TargetAsset: "TSLA", Enabled: true},
	}
	diff := &models.Diff{
		TopMovers: []models.Mover{
			{Symbol: "TSLA", ChangePercent: 15.0, OldValue: 1000, NewValue: 1150},
			{Symbol: "AAPL", ChangePercent: 20.0, OldValue: 2000, NewValue: 2400},
		},
	}

	events := evaluateRules(rules, portfolio.Metrics{}, diff)

	assert.Len(t, events, 1)
	assert.Contains(t, events[0].Message, "TSLA")
}

func TestConcentrationThreshold_AllSymbols(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "concentration_threshold", Threshold: 25.0, TargetAsset: "", Enabled: true},
	}
	metrics := portfolio.Metrics{
		ConcentrationBySymbol: map[string]float64{"AAPL": 60.0, "MSFT": 10.0, "GOOG": 30.0},
	}

	events := evaluateRules(rules, metrics, nil)

	// AAPL at 60% and GOOG at 30% both exceed 25%
	assert.Len(t, events, 2)
}

func TestNilDiff_NoCrash(t *testing.T) {
	rules := []models.AlertRule{
		{ID: ruleID(), RuleType: "net_worth_change", Threshold: 5.0, Enabled: true},
		{ID: ruleID(), RuleType: "position_change_percent", Threshold: 10.0, Enabled: true},
	}

	events := evaluateRules(rules, portfolio.Metrics{}, nil)
	assert.Len(t, events, 0)
}
