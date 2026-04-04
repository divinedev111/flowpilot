package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	polyconn "flowpilot/internal/connectors/polymarket"
	"flowpilot/internal/connectors/schwab"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *Deps) handleSchwabAuth(c *gin.Context) {
	sc := d.getSchwabClient()
	if sc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schwab connector not configured"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, sc.AuthURL())
}

func (d *Deps) handleSchwabCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	sc := d.getSchwabClient()
	if sc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schwab connector not configured"})
		return
	}

	if err := sc.ExchangeCode(c.Request.Context(), code); err != nil {
		log.Printf("Schwab token exchange error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "token exchange failed: " + err.Error()})
		return
	}

	log.Println("Schwab OAuth tokens acquired successfully")

	frontendPort := "3002"
	if p := d.Config.FrontendPort; p != "" {
		frontendPort = p
	}
	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:"+frontendPort+"/")
}

func (d *Deps) handleAuthStatus(c *gin.Context) {
	schwabConnected := false
	if sc := d.getSchwabClient(); sc != nil {
		schwabConnected = sc.HasTokens()
	}

	polymarketWallet := ""
	if pc := d.getPolymarketClient(); pc != nil {
		polymarketWallet = pc.WalletAddress()
	}

	c.JSON(http.StatusOK, gin.H{
		"schwab":             schwabConnected,
		"coinbase":           d.hasCoinbase(),
		"polymarket":         polymarketWallet != "",
		"polymarket_wallet":  polymarketWallet,
	})
}

func (d *Deps) handleSavePolymarketWallet(c *gin.Context) {
	var body struct {
		Wallet string `json:"wallet"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	wallet := strings.TrimSpace(body.Wallet)
	if wallet == "" {
		// Disconnect: remove wallet and connector
		ctx := c.Request.Context()
		_, _ = d.DB.OAuthTokens().DeleteOne(ctx, bson.M{"provider": "polymarket"})

		// Remove from connectors
		if pc := d.getPolymarketClient(); pc != nil {
			pc.SetWalletAddress("")
		}

		c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
		return
	}

	// Save to MongoDB
	ctx := c.Request.Context()
	filter := bson.M{"provider": "polymarket"}
	update := bson.M{"$set": bson.M{
		"provider":     "polymarket",
		"access_token": wallet,
		"updated_at":   time.Now(),
	}}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := d.DB.OAuthTokens().UpdateOne(ctx, filter, update, opts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save wallet"})
		return
	}

	// Update or add polymarket connector
	if pc := d.getPolymarketClient(); pc != nil {
		pc.SetWalletAddress(wallet)
	} else {
		d.Connectors = append(d.Connectors, polyconn.NewClient(wallet))
	}

	log.Printf("Polymarket wallet saved: %s", wallet)
	c.JSON(http.StatusOK, gin.H{"status": "connected", "wallet": wallet})
}

func (d *Deps) handleDisconnectSchwab(c *gin.Context) {
	ctx := c.Request.Context()
	_, _ = d.DB.OAuthTokens().DeleteOne(ctx, bson.M{"provider": "schwab"})

	if sc := d.getSchwabClient(); sc != nil {
		sc.SetTokens("", "")
	}

	log.Println("Schwab disconnected")
	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

func (d *Deps) handleDisconnectCoinbase(c *gin.Context) {
	for i, conn := range d.Connectors {
		if conn.Name() == "coinbase" {
			d.Connectors = append(d.Connectors[:i], d.Connectors[i+1:]...)
			break
		}
	}

	log.Println("Coinbase disconnected (runtime only, will reconnect on restart if API key is set)")
	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

func (d *Deps) getSchwabClient() *schwab.Client {
	for _, conn := range d.Connectors {
		if sc, ok := conn.(*schwab.Client); ok {
			return sc
		}
	}
	return nil
}

func (d *Deps) getPolymarketClient() *polyconn.Client {
	for _, conn := range d.Connectors {
		if pc, ok := conn.(*polyconn.Client); ok {
			return pc
		}
	}
	return nil
}

func (d *Deps) hasCoinbase() bool {
	for _, conn := range d.Connectors {
		if conn.Name() == "coinbase" {
			return true
		}
	}
	return false
}

// handleGetAuditLogs returns recent audit events, optionally filtered by event_type.
// Query params:
//
//	?event_type=auth.login  — filter by event type
//	?limit=50               — limit results (default 100)
func (d *Deps) handleGetAuditLogs(c *gin.Context) {
	eventType := c.Query("event_type")

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	events, err := d.Audit.GetEvents(c.Request.Context(), eventType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
