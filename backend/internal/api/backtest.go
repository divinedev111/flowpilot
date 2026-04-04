package api

import (
	"net/http"
	"sort"
	"time"

	"flowpilot/internal/alerts"
	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"
	snapshotpkg "flowpilot/internal/snapshot"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type backtestRequest struct {
	Rules    []models.AlertRule `json:"rules"`
	FromDate string             `json:"from_date"`
	ToDate   string             `json:"to_date"`
}

type backtestEvent struct {
	RuleType     string `json:"rule_type"`
	Severity     string `json:"severity"`
	Message      string `json:"message"`
	Evidence     string `json:"evidence"`
	SnapshotDate string `json:"snapshot_date"`
}

type backtestResponse struct {
	Events            []backtestEvent `json:"events"`
	TotalTriggers     int             `json:"total_triggers"`
	SnapshotsAnalyzed int             `json:"snapshots_analyzed"`
}

func (d *Deps) handleBacktest(c *gin.Context) {
	var req backtestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Rules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one rule is required"})
		return
	}

	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date format, use YYYY-MM-DD"})
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_date format, use YYYY-MM-DD"})
		return
	}
	// Include the entire to_date day.
	toDate = toDate.Add(24*time.Hour - time.Nanosecond)

	ctx := c.Request.Context()

	// Load snapshots in date range, sorted by created_at ascending.
	filter := bson.D{
		{Key: "created_at", Value: bson.D{{Key: "$gte", Value: fromDate}}},
		{Key: "created_at", Value: bson.D{{Key: "$lte", Value: toDate}}},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := d.DB.Snapshots().Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch snapshots"})
		return
	}

	var snapshots []models.Snapshot
	if err := cursor.All(ctx, &snapshots); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode snapshots"})
		return
	}

	if len(snapshots) < 2 {
		c.JSON(http.StatusOK, backtestResponse{
			Events:            []backtestEvent{},
			TotalTriggers:     0,
			SnapshotsAnalyzed: len(snapshots),
		})
		return
	}

	// Mark all rules as enabled for backtesting.
	for i := range req.Rules {
		req.Rules[i].Enabled = true
	}

	var allEvents []backtestEvent

	// Iterate through consecutive snapshot pairs.
	for i := 1; i < len(snapshots); i++ {
		prev := snapshots[i-1]
		curr := snapshots[i]

		// Load positions for both snapshots.
		prevPositions, err := d.Snapshot.GetPositions(ctx, prev.ID)
		if err != nil {
			continue
		}
		currPositions, err := d.Snapshot.GetPositions(ctx, curr.ID)
		if err != nil {
			continue
		}

		// Reconstruct metrics from current snapshot positions.
		modelPositions := snapshotPositionsToPositions(currPositions)
		metrics := portfolio.ComputeMetrics(modelPositions)

		// Compute diff between the two snapshots.
		diff := snapshotpkg.ComputeDiff(prev.ID, curr.ID, prev.NetWorth, curr.NetWorth, prevPositions, currPositions)
		if prev.NetWorth > 0 {
			diff.NetWorthChange = ((curr.NetWorth - prev.NetWorth) / prev.NetWorth) * 100
		}

		// Evaluate alert rules against reconstructed state.
		alertEvents := alerts.EvaluateRulesForBacktest(req.Rules, metrics, &diff)

		for _, evt := range alertEvents {
			allEvents = append(allEvents, backtestEvent{
				RuleType:     ruleTypeForEvent(evt, req.Rules),
				Severity:     evt.Severity,
				Message:      evt.Message,
				Evidence:     evt.Evidence,
				SnapshotDate: curr.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	// Sort events by date.
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].SnapshotDate < allEvents[j].SnapshotDate
	})

	c.JSON(http.StatusOK, backtestResponse{
		Events:            allEvents,
		TotalTriggers:     len(allEvents),
		SnapshotsAnalyzed: len(snapshots),
	})
}

// snapshotPositionsToPositions converts snapshot positions to standard positions
// so we can compute metrics.
func snapshotPositionsToPositions(sps []models.SnapshotPosition) []models.Position {
	positions := make([]models.Position, len(sps))
	for i, sp := range sps {
		positions[i] = models.Position{
			Symbol:      sp.Symbol,
			Quantity:    sp.Quantity,
			MarkPrice:   sp.MarkPrice,
			MarketValue: sp.Quantity * sp.MarkPrice,
		}
	}
	return positions
}

// ruleTypeForEvent looks up the rule type for an alert event by matching its
// RuleID against the provided rules.
func ruleTypeForEvent(evt models.AlertEvent, rules []models.AlertRule) string {
	for _, r := range rules {
		if r.ID == evt.RuleID {
			return r.RuleType
		}
	}
	return "unknown"
}
