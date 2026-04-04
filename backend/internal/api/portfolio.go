package api

import (
	"net/http"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *Deps) handleGetPortfolio(c *gin.Context) {
	ctx := c.Request.Context()

	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	metrics := portfolio.ComputeMetrics(positions)

	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
		"metrics":   metrics,
	})
}

func (d *Deps) handleListSnapshots(c *gin.Context) {
	ctx := c.Request.Context()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50)
	cursor, err := d.DB.Snapshots().Find(ctx, bson.D{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch snapshots"})
		return
	}
	var snapshots []models.Snapshot
	if err := cursor.All(ctx, &snapshots); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"snapshots": snapshots})
}

func (d *Deps) handleGetSnapshotDiff(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid snapshot id"})
		return
	}

	var diff models.Diff
	err = d.DB.Diffs().FindOne(ctx, bson.D{{Key: "to_snapshot", Value: oid}}).Decode(&diff)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "diff not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"diff": diff})
}
