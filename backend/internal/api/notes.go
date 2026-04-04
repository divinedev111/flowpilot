package api

import (
	"net/http"
	"time"

	"flowpilot/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *Deps) handleGetNotes(c *gin.Context) {
	ctx := c.Request.Context()

	filter := bson.D{}
	if symbol := c.Query("symbol"); symbol != "" {
		filter = bson.D{{Key: "symbol", Value: symbol}}
	}

	// Sort by pinned first (descending so true comes first), then updated_at descending
	opts := options.Find().SetSort(bson.D{
		{Key: "pinned", Value: -1},
		{Key: "updated_at", Value: -1},
	})

	cursor, err := d.DB.Notes().Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notes"})
		return
	}
	var notes []models.Note
	if err := cursor.All(ctx, &notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func (d *Deps) handleGetNote(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var note models.Note
	err = d.DB.Notes().FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&note)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"note": note})
}

func (d *Deps) handleCreateNote(c *gin.Context) {
	var note models.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now
	if note.Tags == nil {
		note.Tags = []string{}
	}

	result, err := d.DB.Notes().InsertOne(c.Request.Context(), note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
		return
	}
	note.ID = result.InsertedID.(bson.ObjectID)

	c.JSON(http.StatusCreated, gin.H{"note": note})
}

func (d *Deps) handleUpdateNote(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var update models.Note
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update.UpdatedAt = time.Now()
	if update.Tags == nil {
		update.Tags = []string{}
	}

	setFields := bson.D{
		{Key: "title", Value: update.Title},
		{Key: "content", Value: update.Content},
		{Key: "symbol", Value: update.Symbol},
		{Key: "tags", Value: update.Tags},
		{Key: "pinned", Value: update.Pinned},
		{Key: "updated_at", Value: update.UpdatedAt},
	}

	result := d.DB.Notes().FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: oid}},
		bson.D{{Key: "$set", Value: setFields}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var note models.Note
	if err := result.Decode(&note); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"note": note})
}

func (d *Deps) handleDeleteNote(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	result, err := d.DB.Notes().DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
