package api

import (
	"net/http"
	"strings"
	"time"

	"flowpilot/internal/models"
	"flowpilot/internal/polymarket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *Deps) handleGetWatchlist(c *gin.Context) {
	ctx := c.Request.Context()

	opts := options.Find().SetSort(bson.D{{Key: "added_at", Value: -1}})
	cursor, err := d.DB.Watchlist().Find(ctx, bson.D{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch watchlist"})
		return
	}
	defer cursor.Close(ctx)

	var items []models.WatchlistItem
	if err := cursor.All(ctx, &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode watchlist"})
		return
	}
	if items == nil {
		items = []models.WatchlistItem{}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (d *Deps) handleAddToWatchlist(c *gin.Context) {
	var item models.WatchlistItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item.AddedAt = time.Now()

	result, err := d.DB.Watchlist().InsertOne(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item"})
		return
	}
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		item.ID = oid
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (d *Deps) handleRemoveFromWatchlist(c *gin.Context) {
	idStr := c.Param("id")
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := d.DB.Watchlist().DeleteOne(c.Request.Context(), bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handlePredictions(c *gin.Context) {
	ctx := c.Request.Context()

	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	symbolSet := make(map[string]bool)
	for _, p := range positions {
		symbolSet[p.Symbol] = true
	}
	var symbols []string
	for s := range symbolSet {
		symbols = append(symbols, s)
	}

	markets, err := d.Polymarket.FetchRelevantMarkets(ctx, symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch prediction markets"})
		return
	}

	if markets == nil {
		markets = []models.PredictionMarket{}
	}

	c.JSON(http.StatusOK, gin.H{
		"markets": markets,
	})
}

func (d *Deps) handleWalletPositions(c *gin.Context) {
	wallet := strings.TrimSpace(c.Query("wallet"))
	if wallet == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet address required"})
		return
	}

	ctx := c.Request.Context()
	positions, summary, err := polymarket.FetchWalletPositions(ctx, wallet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if positions == nil {
		positions = []polymarket.WalletPosition{}
	}

	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
		"summary":   summary,
	})
}
