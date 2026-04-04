package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"flowpilot/internal/ai"
	"flowpilot/internal/alerts"
	"flowpilot/internal/api"
	"flowpilot/internal/audit"
	"flowpilot/internal/config"
	"flowpilot/internal/connectors"
	"flowpilot/internal/connectors/coinbase"
	polyconn "flowpilot/internal/connectors/polymarket"
	"flowpilot/internal/connectors/schwab"
	"flowpilot/internal/correlation"
	"flowpilot/internal/db"
	"flowpilot/internal/news"
	"flowpilot/internal/notify"
	"flowpilot/internal/policy"
	"flowpilot/internal/polymarket"
	"flowpilot/internal/portfolio"
	"flowpilot/internal/snapshot"
)

type oauthToken struct {
	Provider     string    `bson:"provider"`
	AccessToken  string    `bson:"access_token"`
	RefreshToken string    `bson:"refresh_token"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func saveOAuthTokens(ctx context.Context, mongoDB *db.MongoDB, provider, access, refresh string) error {
	filter := bson.M{"provider": provider}
	update := bson.M{"$set": oauthToken{
		Provider:     provider,
		AccessToken:  access,
		RefreshToken: refresh,
		UpdatedAt:    time.Now(),
	}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := mongoDB.OAuthTokens().UpdateOne(ctx, filter, update, opts)
	return err
}

func loadOAuthTokens(ctx context.Context, mongoDB *db.MongoDB, provider string) (string, string, error) {
	var tok oauthToken
	err := mongoDB.OAuthTokens().FindOne(ctx, bson.M{"provider": provider}).Decode(&tok)
	if err != nil {
		return "", "", err
	}
	return tok.AccessToken, tok.RefreshToken, nil
}

func main() {
	// Load .env file if present (not required — env vars can come from anywhere).
	// Try multiple paths: project root (when run from root), ../../.env (when run via go run from cmd/server).
	loaded := false
	for _, p := range []string{".env", "../../.env"} {
		if err := godotenv.Load(p); err == nil {
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoDB, err := db.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Disconnect(context.Background())

	mongoDB.Init(cfg.MongoDB)

	if err := mongoDB.EnsureIndexes(ctx); err != nil {
		log.Fatalf("failed to create indexes: %v", err)
	}

	// Init connectors
	var conns []connectors.Connector

	if cfg.SchwabClientID != "" {
		sc := schwab.NewClient(cfg.SchwabClientID, cfg.SchwabClientSecret, cfg.SchwabRedirectURI)

		// Persist tokens to MongoDB on exchange/refresh
		sc.SetTokenSaver(func(access, refresh string) {
			saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer saveCancel()
			if err := saveOAuthTokens(saveCtx, mongoDB, "schwab", access, refresh); err != nil {
				log.Printf("Failed to persist Schwab tokens: %v", err)
			} else {
				log.Println("Schwab tokens persisted to MongoDB")
			}
		})

		// Restore tokens from MongoDB if available
		if access, refresh, err := loadOAuthTokens(ctx, mongoDB, "schwab"); err == nil && access != "" {
			sc.SetTokens(access, refresh)
			log.Println("Schwab tokens restored from MongoDB")
		}

		conns = append(conns, sc)
	}
	if cfg.CoinbaseAPIKey != "" {
		conns = append(conns, coinbase.NewClient(cfg.CoinbaseAPIKey, cfg.CoinbaseAPISecret))
	}

	// Restore Polymarket wallet from MongoDB if saved
	if polyWallet, _, err := loadOAuthTokens(ctx, mongoDB, "polymarket"); err == nil && polyWallet != "" {
		conns = append(conns, polyconn.NewClient(polyWallet))
		log.Printf("Polymarket wallet restored: %s", polyWallet)
	}

	alertEngine := alerts.NewEngine(mongoDB)

	var notifiers []notify.Notifier

	if cfg.SMTPHost != "" && cfg.AlertEmailTo != "" {
		notifiers = append(notifiers, notify.NewEmailNotifier(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.AlertEmailTo))
	}
	if cfg.TelegramToken != "" && cfg.TelegramChatID != "" {
		notifiers = append(notifiers, notify.NewTelegramNotifier(cfg.TelegramToken, cfg.TelegramChatID))
	}
	if cfg.DiscordWebhook != "" {
		notifiers = append(notifiers, notify.NewDiscordNotifier(cfg.DiscordWebhook))
	}

	dispatcher := notify.NewDispatcher(notifiers...)

	aiClient := ai.NewClient(cfg.AnthropicAPIKey)
	newsService := news.NewService(aiClient, mongoDB)
	auditLogger := audit.NewLogger(mongoDB)
	policyEngine := policy.NewEngine(mongoDB)
	polymarketSvc := polymarket.NewService(mongoDB)

	pricesSvc := correlation.NewPriceService(mongoDB)

	deps := &api.Deps{
		Config:         cfg,
		DB:             mongoDB,
		Connectors:     conns,
		Store:          portfolio.NewStore(mongoDB),
		CostBasisStore: portfolio.NewCostBasisStore(mongoDB),
		Snapshot:       snapshot.NewEngine(mongoDB),
		Alerts:         alertEngine,
		Policies:       policyEngine,
		Notifier:       dispatcher,
		AI:             aiClient,
		News:           newsService,
		Audit:          auditLogger,
		Polymarket:     polymarketSvc,
		Prices:         pricesSvc,
	}

	router := api.NewRouter(deps)

	log.Printf("FlowPilot Finance starting on %s (HTTPS)", cfg.Port)
	log.Printf("Connectors enabled: %d", len(conns))
	log.Printf("Notification channels: %d", len(notifiers))
	if err := router.RunTLS(cfg.Port, "backend/server.crt", "backend/server.key"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
