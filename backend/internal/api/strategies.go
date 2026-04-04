package api

import (
	"net/http"
	"time"

	"flowpilot/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type strategyWithMetrics struct {
	models.Strategy `bson:",inline"`
	TotalValue      float64  `json:"total_value"`
	AllocationPct   float64  `json:"allocation_pct"`
	PositionCount   int      `json:"position_count"`
	Symbols         []string `json:"symbols"`
}

type strategyDetail struct {
	models.Strategy `bson:",inline"`
	TotalValue      float64           `json:"total_value"`
	AllocationPct   float64           `json:"allocation_pct"`
	PositionCount   int               `json:"position_count"`
	Symbols         []string          `json:"symbols"`
	Positions       []models.Position `json:"positions"`
}

func (d *Deps) handleGetStrategies(c *gin.Context) {
	ctx := c.Request.Context()

	// Fetch all strategies sorted by creation date.
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := d.DB.Strategies().Find(ctx, bson.D{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch strategies"})
		return
	}
	var strategies []models.Strategy
	if err := cursor.All(ctx, &strategies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode strategies"})
		return
	}

	// Fetch all assignments.
	aCursor, err := d.DB.StrategyAssignments().Find(ctx, bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch assignments"})
		return
	}
	var assignments []models.StrategyAssignment
	if err := aCursor.All(ctx, &assignments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode assignments"})
		return
	}

	// Build a map of strategyID -> []symbol.
	assignMap := make(map[bson.ObjectID][]string)
	for _, a := range assignments {
		assignMap[a.StrategyID] = append(assignMap[a.StrategyID], a.Symbol)
	}

	// Fetch all current positions.
	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	// Build a map of symbol -> total market value.
	symbolValue := make(map[string]float64)
	var totalPortfolioValue float64
	for _, p := range positions {
		symbolValue[p.Symbol] += p.MarketValue
		totalPortfolioValue += p.MarketValue
	}

	// Compute metrics per strategy.
	result := make([]strategyWithMetrics, 0, len(strategies))
	for _, s := range strategies {
		symbols := assignMap[s.ID]
		if symbols == nil {
			symbols = []string{}
		}

		var totalValue float64
		for _, sym := range symbols {
			totalValue += symbolValue[sym]
		}

		var allocationPct float64
		if totalPortfolioValue > 0 {
			allocationPct = (totalValue / totalPortfolioValue) * 100
		}

		posCount := 0
		for _, sym := range symbols {
			if _, ok := symbolValue[sym]; ok {
				posCount++
			}
		}

		result = append(result, strategyWithMetrics{
			Strategy:      s,
			TotalValue:    totalValue,
			AllocationPct: allocationPct,
			PositionCount: posCount,
			Symbols:       symbols,
		})
	}

	c.JSON(http.StatusOK, gin.H{"strategies": result})
}

func (d *Deps) handleGetStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy id"})
		return
	}

	var strategy models.Strategy
	err = d.DB.Strategies().FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&strategy)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch strategy"})
		return
	}

	// Fetch assignments for this strategy.
	aCursor, err := d.DB.StrategyAssignments().Find(ctx, bson.D{{Key: "strategy_id", Value: oid}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch assignments"})
		return
	}
	var assignments []models.StrategyAssignment
	if err := aCursor.All(ctx, &assignments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode assignments"})
		return
	}

	symbols := make([]string, 0, len(assignments))
	for _, a := range assignments {
		symbols = append(symbols, a.Symbol)
	}

	// Fetch all current positions.
	allPositions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	// Build symbol set for fast lookup.
	symbolSet := make(map[string]bool, len(symbols))
	for _, sym := range symbols {
		symbolSet[sym] = true
	}

	// Filter positions belonging to this strategy and compute total portfolio value.
	var strategyPositions []models.Position
	var totalPortfolioValue float64
	var totalValue float64
	for _, p := range allPositions {
		totalPortfolioValue += p.MarketValue
		if symbolSet[p.Symbol] {
			strategyPositions = append(strategyPositions, p)
			totalValue += p.MarketValue
		}
	}

	if strategyPositions == nil {
		strategyPositions = []models.Position{}
	}

	var allocationPct float64
	if totalPortfolioValue > 0 {
		allocationPct = (totalValue / totalPortfolioValue) * 100
	}

	detail := strategyDetail{
		Strategy:      strategy,
		TotalValue:    totalValue,
		AllocationPct: allocationPct,
		PositionCount: len(strategyPositions),
		Symbols:       symbols,
		Positions:     strategyPositions,
	}

	c.JSON(http.StatusOK, detail)
}

func (d *Deps) handleCreateStrategy(c *gin.Context) {
	var strategy models.Strategy
	if err := c.ShouldBindJSON(&strategy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	strategy.CreatedAt = now
	strategy.UpdatedAt = now

	ctx := c.Request.Context()
	result, err := d.DB.Strategies().InsertOne(ctx, strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create strategy"})
		return
	}

	strategy.ID = result.InsertedID.(bson.ObjectID)
	c.JSON(http.StatusCreated, gin.H{"strategy": strategy})
}

func (d *Deps) handleUpdateStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy id"})
		return
	}

	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := bson.D{{Key: "updated_at", Value: time.Now()}}
	if body.Name != nil {
		updates = append(updates, bson.E{Key: "name", Value: *body.Name})
	}
	if body.Description != nil {
		updates = append(updates, bson.E{Key: "description", Value: *body.Description})
	}
	if body.Color != nil {
		updates = append(updates, bson.E{Key: "color", Value: *body.Color})
	}

	filter := bson.D{{Key: "_id", Value: oid}}
	update := bson.D{{Key: "$set", Value: updates}}

	res := d.DB.Strategies().FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After))
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update strategy"})
		return
	}

	var updated models.Strategy
	if err := res.Decode(&updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode strategy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"strategy": updated})
}

func (d *Deps) handleDeleteStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy id"})
		return
	}

	// Delete the strategy.
	res, err := d.DB.Strategies().DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete strategy"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}

	// Delete all assignments for this strategy.
	_, _ = d.DB.StrategyAssignments().DeleteMany(ctx, bson.D{{Key: "strategy_id", Value: oid}})

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (d *Deps) handleAssignStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy id"})
		return
	}

	var body struct {
		Symbol string `json:"symbol" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the strategy exists.
	count, err := d.DB.Strategies().CountDocuments(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}

	// Check if this symbol is already assigned to this strategy.
	existing, err := d.DB.StrategyAssignments().CountDocuments(ctx, bson.D{
		{Key: "strategy_id", Value: oid},
		{Key: "symbol", Value: body.Symbol},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check assignment"})
		return
	}
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "symbol already assigned to this strategy"})
		return
	}

	assignment := models.StrategyAssignment{
		StrategyID: oid,
		Symbol:     body.Symbol,
		AssignedAt: time.Now(),
	}

	result, err := d.DB.StrategyAssignments().InsertOne(ctx, assignment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign symbol"})
		return
	}

	assignment.ID = result.InsertedID.(bson.ObjectID)
	c.JSON(http.StatusCreated, gin.H{"assignment": assignment})
}

func (d *Deps) handleUnassignStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	symbol := c.Param("symbol")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy id"})
		return
	}

	res, err := d.DB.StrategyAssignments().DeleteOne(ctx, bson.D{
		{Key: "strategy_id", Value: oid},
		{Key: "symbol", Value: symbol},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unassign symbol"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
