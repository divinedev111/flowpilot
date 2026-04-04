package api

import (
	"net/http"
	"time"

	"flowpilot/internal/models"
	"flowpilot/internal/portfolio"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (d *Deps) handleGetAlertRules(c *gin.Context) {
	rules, err := d.Alerts.GetRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

func (d *Deps) handleCreateAlertRule(c *gin.Context) {
	var rule models.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.Enabled = true
	rule.CreatedAt = time.Now()

	created, err := d.Alerts.CreateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rule"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rule": created})
}

func (d *Deps) handleGetAlertEvents(c *gin.Context) {
	events, err := d.Alerts.GetEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (d *Deps) handleGetPolicies(c *gin.Context) {
	ctx := c.Request.Context()

	rules, err := d.Policies.GetRules(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch policy rules"})
		return
	}

	violations, err := d.Policies.GetViolations(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch violations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules":      rules,
		"violations": violations,
	})
}

func (d *Deps) handleCreatePolicy(c *gin.Context) {
	var rule models.PolicyRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.Enabled = true
	rule.CreatedAt = time.Now()

	created, err := d.Policies.CreateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create policy rule"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rule": created})
}

func (d *Deps) handleUpdatePolicy(c *gin.Context) {
	id := c.Param("id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	var rule models.PolicyRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := d.Policies.UpdateRule(c.Request.Context(), oid, rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update policy rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (d *Deps) handleDeletePolicy(c *gin.Context) {
	id := c.Param("id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	if err := d.Policies.DeleteRule(c.Request.Context(), oid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete policy rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (d *Deps) handleCheckPolicies(c *gin.Context) {
	ctx := c.Request.Context()

	positions, err := d.Store.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}

	metrics := portfolio.ComputeMetrics(positions)

	violations, err := d.Policies.CheckCompliance(ctx, metrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check compliance"})
		return
	}

	status := "passing"
	if len(violations) > 0 {
		status = "violations_detected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"violations": violations,
		"metrics":    metrics,
	})
}
