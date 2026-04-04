package api

import (
	"net/http"

	"flowpilot/internal/audit"

	"github.com/gin-gonic/gin"
)

func NewRouter(deps *Deps) *gin.Engine {
	r := gin.Default()

	// Audit middleware (logs every request except health checks)
	if deps.Audit != nil {
		r.Use(audit.Middleware(deps.Audit))
	}

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		api.POST("/sync", deps.handleSync)
		api.GET("/portfolio", deps.handleGetPortfolio)
		api.GET("/snapshots", deps.handleListSnapshots)
		api.GET("/snapshots/:id/diff", deps.handleGetSnapshotDiff)
		api.GET("/alerts/rules", deps.handleGetAlertRules)
		api.POST("/alerts/rules", deps.handleCreateAlertRule)
		api.GET("/alerts/events", deps.handleGetAlertEvents)
		api.POST("/chat", deps.handleChat)
		api.POST("/scenarios", deps.handleScenario)
		api.GET("/risk", deps.handleRisk)
		api.GET("/performance", deps.handlePerformance)

		// AI Digest & Insights
		api.GET("/digest/latest", deps.handleLatestDigest)
		api.GET("/digest/history", deps.handleDigestHistory)
		api.POST("/insights", deps.handleInsights)

		// News
		api.GET("/news", deps.handleNewsFeed)
		api.GET("/news/symbol/:symbol", deps.handleSymbolNews)

		// Watchlist
		api.GET("/watchlist", deps.handleGetWatchlist)
		api.POST("/watchlist", deps.handleAddToWatchlist)
		api.DELETE("/watchlist/:id", deps.handleRemoveFromWatchlist)

		// Auth
		api.GET("/auth/schwab", deps.handleSchwabAuth)
		api.GET("/auth/schwab/callback", deps.handleSchwabCallback)
		api.GET("/auth/status", deps.handleAuthStatus)
		api.POST("/auth/polymarket", deps.handleSavePolymarketWallet)
		api.POST("/auth/schwab/disconnect", deps.handleDisconnectSchwab)
		api.POST("/auth/coinbase/disconnect", deps.handleDisconnectCoinbase)

		// Notes
		api.GET("/notes", deps.handleGetNotes)
		api.GET("/notes/:id", deps.handleGetNote)
		api.POST("/notes", deps.handleCreateNote)
		api.PUT("/notes/:id", deps.handleUpdateNote)
		api.DELETE("/notes/:id", deps.handleDeleteNote)

		// Strategies
		api.GET("/strategies", deps.handleGetStrategies)
		api.GET("/strategies/:id", deps.handleGetStrategy)
		api.POST("/strategies", deps.handleCreateStrategy)
		api.PUT("/strategies/:id", deps.handleUpdateStrategy)
		api.DELETE("/strategies/:id", deps.handleDeleteStrategy)
		api.POST("/strategies/:id/assign", deps.handleAssignStrategy)
		api.DELETE("/strategies/:id/assign/:symbol", deps.handleUnassignStrategy)

		// Policies
		api.GET("/policies", deps.handleGetPolicies)
		api.POST("/policies", deps.handleCreatePolicy)
		api.PUT("/policies/:id", deps.handleUpdatePolicy)
		api.DELETE("/policies/:id", deps.handleDeletePolicy)
		api.POST("/policies/check", deps.handleCheckPolicies)

		// Backtesting
		api.POST("/backtest", deps.handleBacktest)

		// Correlation
		api.GET("/correlation", deps.handleCorrelation)

		// Audit
		api.GET("/audit", deps.handleGetAuditLogs)

		// Predictions (Polymarket)
		api.GET("/predictions", deps.handlePredictions)
		api.GET("/predictions/positions", deps.handleWalletPositions)
	}

	return r
}
