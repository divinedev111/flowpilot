package api

import (
	"net/http"

	"flowpilot/internal/ai"
	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ChatRequest struct {
	Message        string `json:"message" binding:"required"`
	ConversationID string `json:"conversation_id"`
}

func (d *Deps) handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ConversationID == "" {
		req.ConversationID = "default"
	}

	// Get current portfolio for context
	positions, _ := d.Store.GetAll(c.Request.Context())
	metrics := portfolio.ComputeMetrics(positions)
	portfolioCtx := ai.BuildPortfolioContext(metrics, positions)

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	if err := d.AI.ChatStream(c.Request.Context(), req.ConversationID, req.Message, portfolioCtx, c.Writer); err != nil {
		// If streaming already started, we can't send a JSON error.
		// Log it via SSE instead.
		c.SSEvent("error", err.Error())
	}
}

func (d *Deps) handleLatestDigest(c *gin.Context) {
	ctx := c.Request.Context()

	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	var digest models.AIDigest
	err := d.DB.AIDigests().FindOne(ctx, bson.D{}, opts).Decode(&digest)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no digests found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"digest": digest})
}

func (d *Deps) handleDigestHistory(c *gin.Context) {
	ctx := c.Request.Context()

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(10)
	cursor, err := d.DB.AIDigests().Find(ctx, bson.D{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query digests"})
		return
	}

	var digests []models.AIDigest
	if err := cursor.All(ctx, &digests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode digests"})
		return
	}

	if digests == nil {
		digests = []models.AIDigest{}
	}

	c.JSON(http.StatusOK, gin.H{"digests": digests})
}

func (d *Deps) handleInsights(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current positions and compute metrics
	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	if len(positions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"insights":       []any{},
			"overall_health": "unknown",
			"generated_at":   nil,
		})
		return
	}

	metrics := portfolio.ComputeMetrics(positions)

	// Get latest diff for context
	var diff *models.Diff
	diffOpts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var latestDiff models.Diff
	if err := d.DB.Diffs().FindOne(ctx, bson.D{}, diffOpts).Decode(&latestDiff); err == nil {
		diff = &latestDiff
	}

	if d.AI == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI client not configured"})
		return
	}

	result, err := d.AI.GenerateInsights(ctx, metrics, positions, diff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate insights: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
