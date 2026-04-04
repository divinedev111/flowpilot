package api

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"flowpilot/internal/connectors"
	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"
	snapshotpkg "flowpilot/internal/snapshot"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (d *Deps) handleSync(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Fetch from all connectors concurrently
	var allPositions []models.Position
	var mu sync.Mutex
	var wg sync.WaitGroup
	var fetchErrors []string

	for _, conn := range d.Connectors {
		wg.Add(1)
		go func(conn connectors.Connector) {
			defer wg.Done()
			positions, err := conn.FetchPositions(ctx)
			if err != nil {
				log.Printf("connector %s error: %v", conn.Name(), err)
				mu.Lock()
				fetchErrors = append(fetchErrors, conn.Name()+": "+err.Error())
				mu.Unlock()
				return
			}
			mu.Lock()
			allPositions = append(allPositions, positions...)
			mu.Unlock()
		}(conn)
	}
	wg.Wait()

	if len(allPositions) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "no positions fetched",
			"details": fetchErrors,
		})
		return
	}

	// 2. Get previous snapshot for day-change computation and diffing
	prevSnap, prevErr := d.Snapshot.GetLatest(ctx)

	// Build previous price map for day change
	prevPrices := make(map[string]float64)
	if prevErr == nil && prevSnap != nil {
		prevPositions, err := d.Snapshot.GetPositions(ctx, prevSnap.ID)
		if err == nil {
			// Build a lookup using just symbol since SnapshotPosition
			// doesn't have source/account_id.
			// For positions that share a symbol across sources, we use
			// the current positions to build the key with source+account_id.
			symbolPriceMap := portfolio.BuildPrevPriceMap(prevPositions)
			for _, p := range allPositions {
				key := p.Symbol + "|" + p.Source + "|" + p.AccountID
				if price, ok := symbolPriceMap[p.Symbol]; ok {
					prevPrices[key] = price
				}
			}
		}
	}

	// 3. Enrich positions with P&L data
	allPositions = portfolio.EnrichPnL(ctx, allPositions, d.CostBasisStore, prevPrices)

	// 4. Clean stale positions for sources we're syncing, then upsert
	syncedSources := make(map[string]bool)
	for _, p := range allPositions {
		syncedSources[p.Source] = true
	}
	for source := range syncedSources {
		if err := d.Store.DeleteBySource(ctx, source); err != nil {
			log.Printf("warning: failed to clean stale %s positions: %v", source, err)
		}
	}
	if err := d.Store.UpsertPositions(ctx, allPositions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store positions"})
		return
	}

	// 5. Compute metrics (includes P&L aggregates)
	metrics := portfolio.ComputeMetrics(allPositions)

	// 6. Create new snapshot
	snap, err := d.Snapshot.Create(ctx, allPositions, metrics.NetWorth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create snapshot"})
		return
	}

	// 7. Compute diff if previous snapshot exists
	var diff *models.Diff
	if prevErr == nil && prevSnap != nil {
		prevPositions, err := d.Snapshot.GetPositions(ctx, prevSnap.ID)
		if err == nil {
			newPositions, _ := d.Snapshot.GetPositions(ctx, snap.ID)
			computedDiff := snapshotpkg.ComputeDiff(prevSnap.ID, snap.ID, prevSnap.NetWorth, snap.NetWorth, prevPositions, newPositions)
			diff = &computedDiff

			// Store diff
			_, _ = d.DB.Diffs().InsertOne(ctx, diff)
		}
	} else if prevErr != nil && prevErr != mongo.ErrNoDocuments {
		log.Printf("warning: failed to get previous snapshot: %v", prevErr)
	}

	// 8. Evaluate alert rules
	var alertEvents []models.AlertEvent
	if d.Alerts != nil {
		events, err := d.Alerts.Evaluate(ctx, metrics, diff)
		if err != nil {
			log.Printf("alert evaluation error: %v", err)
		} else {
			alertEvents = events
		}
	}

	// 9. Dispatch notifications for triggered alerts
	if d.Notifier != nil && len(alertEvents) > 0 {
		go d.Notifier.Dispatch(context.Background(), alertEvents)
	}

	// 10. Generate AI digest
	if d.AI != nil {
		go func() {
			summary, riskNotes, err := d.AI.GenerateDigest(context.Background(), metrics, diff, alertEvents)
			if err != nil {
				log.Printf("AI digest error: %v", err)
				return
			}
			digest := models.AIDigest{
				SnapshotID: snap.ID,
				Summary:    summary,
				RiskNotes:  riskNotes,
				Timestamp:  time.Now(),
			}
			if _, err := d.DB.AIDigests().InsertOne(context.Background(), digest); err != nil {
				log.Printf("AI digest store error: %v", err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshot":          snap,
		"metrics":           metrics,
		"diff":              diff,
		"positions":         len(allPositions),
		"fetch_errors":      fetchErrors,
		"alerts_triggered":  len(alertEvents),
	})
}
