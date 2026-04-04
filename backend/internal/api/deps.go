package api

import (
	"flowpilot/internal/ai"
	"flowpilot/internal/alerts"
	"flowpilot/internal/audit"
	"flowpilot/internal/config"
	"flowpilot/internal/connectors"
	"flowpilot/internal/correlation"
	"flowpilot/internal/db"
	"flowpilot/internal/news"
	"flowpilot/internal/notify"
	"flowpilot/internal/policy"
	"flowpilot/internal/polymarket"
	"flowpilot/internal/portfolio"
	"flowpilot/internal/snapshot"
)

type Deps struct {
	Config         *config.Config
	DB             *db.MongoDB
	Connectors     []connectors.Connector
	Store          *portfolio.Store
	CostBasisStore *portfolio.CostBasisStore
	Snapshot       *snapshot.Engine
	Alerts         *alerts.Engine
	Policies       *policy.Engine
	Notifier       *notify.Dispatcher
	AI             *ai.Client
	News           *news.Service
	Audit          *audit.Logger
	Polymarket     *polymarket.Service
	Prices         *correlation.PriceService
}
