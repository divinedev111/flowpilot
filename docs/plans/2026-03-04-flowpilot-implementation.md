# FlowPilot Finance Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the backend for FlowPilot Finance — a Go monolith that syncs portfolio holdings from Schwab and Coinbase, computes metrics, creates snapshots, detects changes, evaluates alerts, and generates AI digests.

**Architecture:** Single Go binary using Gin for HTTP, MongoDB Atlas for storage, and the Anthropic Go SDK for AI. All financial math is deterministic in Go. AI is used only for explanations and chat.

**Tech Stack:** Go 1.25, Gin, MongoDB Go Driver v2, Anthropic Go SDK, `net/smtp`, `net/http` (Telegram/Discord webhooks)

---

## Phase 1: Backend Initialization

### Task 1: Initialize Go module and install dependencies

**Files:**
- Create: `backend/cmd/server/main.go` (placeholder)
- Create: `backend/go.mod`

**Step 1: Create backend directory and init Go module**

```bash
mkdir -p backend/cmd/server
cd backend
go mod init flowpilot
```

**Step 2: Install dependencies**

```bash
go get github.com/gin-gonic/gin
go get go.mongodb.org/mongo-driver/v2/mongo
go get github.com/anthropics/anthropic-sdk-go
go get github.com/stretchr/testify
```

**Step 3: Create placeholder main.go**

```go
// backend/cmd/server/main.go
package main

func main() {
    println("FlowPilot Finance starting...")
}
```

**Step 4: Verify it compiles**

Run: `cd backend && go build ./cmd/server`
Expected: No errors

**Step 5: Commit**

```bash
git init
git add -A
git commit -m "chore: init Go module with dependencies"
```

---

### Task 2: Create config package

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/config_test.go`

**Step 1: Write the test**

```go
// backend/internal/config/config_test.go
package config

import (
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoad_RequiredFields(t *testing.T) {
    os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
    os.Setenv("ANTHROPIC_API_KEY", "test-key")
    defer os.Unsetenv("MONGODB_URI")
    defer os.Unsetenv("ANTHROPIC_API_KEY")

    cfg, err := Load()
    require.NoError(t, err)
    assert.Equal(t, "mongodb://localhost:27017", cfg.MongoURI)
    assert.Equal(t, "test-key", cfg.AnthropicAPIKey)
    assert.Equal(t, ":8080", cfg.Port)
}

func TestLoad_MissingMongoURI(t *testing.T) {
    os.Unsetenv("MONGODB_URI")
    os.Setenv("ANTHROPIC_API_KEY", "test-key")
    defer os.Unsetenv("ANTHROPIC_API_KEY")

    _, err := Load()
    assert.Error(t, err)
}

func TestLoad_CustomPort(t *testing.T) {
    os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
    os.Setenv("ANTHROPIC_API_KEY", "test-key")
    os.Setenv("PORT", "3000")
    defer os.Unsetenv("MONGODB_URI")
    defer os.Unsetenv("ANTHROPIC_API_KEY")
    defer os.Unsetenv("PORT")

    cfg, err := Load()
    require.NoError(t, err)
    assert.Equal(t, ":3000", cfg.Port)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/config/ -v`
Expected: FAIL — package doesn't exist yet

**Step 3: Write implementation**

```go
// backend/internal/config/config.go
package config

import (
    "fmt"
    "os"
)

type Config struct {
    Port           string
    MongoURI       string
    MongoDB        string
    AnthropicAPIKey string

    // Schwab
    SchwabClientID     string
    SchwabClientSecret string
    SchwabRedirectURI  string

    // Coinbase
    CoinbaseAPIKey    string
    CoinbaseAPISecret string

    // Notifications (all optional)
    SMTPHost       string
    SMTPPort       string
    SMTPUser       string
    SMTPPass       string
    AlertEmailTo   string
    TelegramToken  string
    TelegramChatID string
    DiscordWebhook string
}

func Load() (*Config, error) {
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        return nil, fmt.Errorf("MONGODB_URI is required")
    }

    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    port = ":" + port

    dbName := os.Getenv("MONGODB_NAME")
    if dbName == "" {
        dbName = "flowpilot"
    }

    return &Config{
        Port:               port,
        MongoURI:           mongoURI,
        MongoDB:            dbName,
        AnthropicAPIKey:    apiKey,
        SchwabClientID:     os.Getenv("SCHWAB_CLIENT_ID"),
        SchwabClientSecret: os.Getenv("SCHWAB_CLIENT_SECRET"),
        SchwabRedirectURI:  os.Getenv("SCHWAB_REDIRECT_URI"),
        CoinbaseAPIKey:     os.Getenv("COINBASE_API_KEY"),
        CoinbaseAPISecret:  os.Getenv("COINBASE_API_SECRET"),
        SMTPHost:           os.Getenv("SMTP_HOST"),
        SMTPPort:           os.Getenv("SMTP_PORT"),
        SMTPUser:           os.Getenv("SMTP_USER"),
        SMTPPass:           os.Getenv("SMTP_PASS"),
        AlertEmailTo:       os.Getenv("ALERT_EMAIL_TO"),
        TelegramToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
        TelegramChatID:     os.Getenv("TELEGRAM_CHAT_ID"),
        DiscordWebhook:     os.Getenv("DISCORD_WEBHOOK_URL"),
    }, nil
}
```

**Step 4: Run tests**

Run: `cd backend && go test ./internal/config/ -v`
Expected: All 3 tests PASS

**Step 5: Commit**

```bash
git add backend/internal/config/
git commit -m "feat: add config package with env loading"
```

---

### Task 3: Create .env.example

**Files:**
- Create: `.env.example`

**Step 1: Create env template**

```env
# Required
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net
MONGODB_NAME=flowpilot
ANTHROPIC_API_KEY=sk-ant-...
PORT=8080

# Schwab Trader API
SCHWAB_CLIENT_ID=
SCHWAB_CLIENT_SECRET=
SCHWAB_REDIRECT_URI=http://localhost:8080/api/auth/schwab/callback

# Coinbase API
COINBASE_API_KEY=
COINBASE_API_SECRET=

# Email notifications (optional)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
ALERT_EMAIL_TO=

# Telegram notifications (optional)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Discord notifications (optional)
DISCORD_WEBHOOK_URL=
```

**Step 2: Add .env to .gitignore**

```gitignore
# .gitignore
.env
backend/backend
```

**Step 3: Commit**

```bash
git add .env.example .gitignore
git commit -m "chore: add .env.example and .gitignore"
```

---

### Task 4: Create Gin server with health endpoint

**Files:**
- Create: `backend/internal/api/router.go`
- Create: `backend/internal/api/router_test.go`
- Modify: `backend/cmd/server/main.go`

**Step 1: Write the test**

```go
// backend/internal/api/router_test.go
package api

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
    router := NewRouter(nil)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/health", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var body map[string]string
    err := json.Unmarshal(w.Body.Bytes(), &body)
    require.NoError(t, err)
    assert.Equal(t, "ok", body["status"])
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/api/ -v`
Expected: FAIL

**Step 3: Write router implementation**

```go
// backend/internal/api/router.go
package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "flowpilot/internal/config"
)

func NewRouter(cfg *config.Config) *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    {
        api.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })
    }

    return r
}
```

**Step 4: Run test**

Run: `cd backend && go test ./internal/api/ -v`
Expected: PASS

**Step 5: Wire up main.go**

```go
// backend/cmd/server/main.go
package main

import (
    "log"

    "flowpilot/internal/api"
    "flowpilot/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    router := api.NewRouter(cfg)

    log.Printf("FlowPilot Finance starting on %s", cfg.Port)
    if err := router.Run(cfg.Port); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
```

**Step 6: Verify it compiles**

Run: `cd backend && go build ./cmd/server`
Expected: No errors

**Step 7: Commit**

```bash
git add backend/internal/api/ backend/cmd/server/main.go
git commit -m "feat: add Gin server with health endpoint"
```

---

## Phase 2: MongoDB Schema & Connection

### Task 5: Create model structs

**Files:**
- Create: `backend/internal/models/position.go`
- Create: `backend/internal/models/snapshot.go`
- Create: `backend/internal/models/alert.go`
- Create: `backend/internal/models/ai.go`

**Step 1: Create Position model**

```go
// backend/internal/models/position.go
package models

import (
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
)

type Position struct {
    ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
    Symbol      string        `bson:"symbol" json:"symbol"`
    Quantity    float64       `bson:"quantity" json:"quantity"`
    MarkPrice   float64       `bson:"mark_price" json:"mark_price"`
    MarketValue float64       `bson:"market_value" json:"market_value"`
    AssetType   string        `bson:"asset_type" json:"asset_type"`
    Source      string        `bson:"source" json:"source"`
    AccountID   string        `bson:"account_id" json:"account_id"`
    Timestamp   time.Time     `bson:"timestamp" json:"timestamp"`
}
```

**Step 2: Create Snapshot models**

```go
// backend/internal/models/snapshot.go
package models

import (
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
)

type Snapshot struct {
    ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt time.Time     `bson:"created_at" json:"created_at"`
    NetWorth  float64       `bson:"net_worth" json:"net_worth"`
}

type SnapshotPosition struct {
    ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
    SnapshotID bson.ObjectID `bson:"snapshot_id" json:"snapshot_id"`
    Symbol     string        `bson:"symbol" json:"symbol"`
    Quantity   float64       `bson:"quantity" json:"quantity"`
    MarkPrice  float64       `bson:"mark_price" json:"mark_price"`
}

type Mover struct {
    Symbol        string  `bson:"symbol" json:"symbol"`
    ChangePercent float64 `bson:"change_percent" json:"change_percent"`
    OldValue      float64 `bson:"old_value" json:"old_value"`
    NewValue      float64 `bson:"new_value" json:"new_value"`
}

type Diff struct {
    ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
    FromSnapshot   bson.ObjectID `bson:"from_snapshot" json:"from_snapshot"`
    ToSnapshot     bson.ObjectID `bson:"to_snapshot" json:"to_snapshot"`
    NetWorthChange float64       `bson:"net_worth_change" json:"net_worth_change"`
    TopMovers      []Mover       `bson:"top_movers" json:"top_movers"`
    AddedSymbols   []string      `bson:"added_symbols" json:"added_symbols"`
    RemovedSymbols []string      `bson:"removed_symbols" json:"removed_symbols"`
    CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
}
```

**Step 3: Create Alert models**

```go
// backend/internal/models/alert.go
package models

import (
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
)

type AlertRule struct {
    ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
    RuleType    string        `bson:"rule_type" json:"rule_type"`
    Threshold   float64       `bson:"threshold" json:"threshold"`
    TargetAsset string        `bson:"target_asset" json:"target_asset"`
    Enabled     bool          `bson:"enabled" json:"enabled"`
    CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
}

type AlertEvent struct {
    ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
    RuleID    bson.ObjectID `bson:"rule_id" json:"rule_id"`
    Severity  string        `bson:"severity" json:"severity"`
    Message   string        `bson:"message" json:"message"`
    Evidence  string        `bson:"evidence" json:"evidence"`
    Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
}
```

**Step 4: Create AI digest model**

```go
// backend/internal/models/ai.go
package models

import (
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
)

type AIDigest struct {
    ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
    SnapshotID bson.ObjectID `bson:"snapshot_id" json:"snapshot_id"`
    Summary    string        `bson:"summary" json:"summary"`
    RiskNotes  string        `bson:"risk_notes" json:"risk_notes"`
    Timestamp  time.Time     `bson:"timestamp" json:"timestamp"`
}
```

**Step 5: Verify compilation**

Run: `cd backend && go build ./internal/models/`
Expected: No errors

**Step 6: Commit**

```bash
git add backend/internal/models/
git commit -m "feat: add MongoDB model structs"
```

---

### Task 6: Create MongoDB connection package

**Files:**
- Create: `backend/internal/db/mongo.go`
- Create: `backend/internal/db/mongo_test.go`

**Step 1: Write the test**

```go
// backend/internal/db/mongo_test.go
package db

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

func TestConnect_Integration(t *testing.T) {
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        t.Skip("MONGODB_URI not set, skipping integration test")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := Connect(ctx, uri)
    require.NoError(t, err)
    require.NotNil(t, client)

    defer client.Disconnect(ctx)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/db/ -v`
Expected: FAIL or SKIP

**Step 3: Write implementation**

```go
// backend/internal/db/mongo.go
package db

import (
    "context"
    "fmt"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoDB struct {
    Client *mongo.Client
    DB     *mongo.Database
}

func Connect(ctx context.Context, uri string) (*MongoDB, error) {
    client, err := mongo.Connect(options.Client().ApplyURI(uri))
    if err != nil {
        return nil, fmt.Errorf("mongo connect: %w", err)
    }

    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return nil, fmt.Errorf("mongo ping: %w", err)
    }

    return &MongoDB{Client: client}, nil
}

func (m *MongoDB) Init(dbName string) {
    m.DB = m.Client.Database(dbName)
}

func (m *MongoDB) Disconnect(ctx context.Context) error {
    return m.Client.Disconnect(ctx)
}

// Collection accessors
func (m *MongoDB) PositionsCurrent() *mongo.Collection {
    return m.DB.Collection("positions_current")
}

func (m *MongoDB) Snapshots() *mongo.Collection {
    return m.DB.Collection("snapshots")
}

func (m *MongoDB) SnapshotPositions() *mongo.Collection {
    return m.DB.Collection("snapshot_positions")
}

func (m *MongoDB) Diffs() *mongo.Collection {
    return m.DB.Collection("diffs")
}

func (m *MongoDB) AlertRules() *mongo.Collection {
    return m.DB.Collection("alert_rules")
}

func (m *MongoDB) AlertEvents() *mongo.Collection {
    return m.DB.Collection("alert_events")
}

func (m *MongoDB) AIDigests() *mongo.Collection {
    return m.DB.Collection("ai_digests")
}
```

**Step 4: Run test with MONGODB_URI set**

Run: `cd backend && MONGODB_URI="<your-atlas-uri>" go test ./internal/db/ -v`
Expected: PASS (or SKIP if no URI)

**Step 5: Commit**

```bash
git add backend/internal/db/
git commit -m "feat: add MongoDB connection package with collection accessors"
```

---

### Task 7: Create indexes and wire DB into server

**Files:**
- Create: `backend/internal/db/indexes.go`
- Modify: `backend/cmd/server/main.go`

**Step 1: Create index setup**

```go
// backend/internal/db/indexes.go
package db

import (
    "context"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (m *MongoDB) EnsureIndexes(ctx context.Context) error {
    indexes := map[*mongo.Collection][]mongo.IndexModel{
        m.PositionsCurrent(): {
            {
                Keys:    bson.D{{"symbol", 1}, {"source", 1}, {"account_id", 1}},
                Options: options.Index().SetUnique(true),
            },
            {Keys: bson.D{{"asset_type", 1}}},
        },
        m.SnapshotPositions(): {
            {Keys: bson.D{{"snapshot_id", 1}}},
        },
        m.Diffs(): {
            {Keys: bson.D{{"to_snapshot", 1}}},
        },
        m.AlertEvents(): {
            {Keys: bson.D{{"timestamp", -1}}},
            {Keys: bson.D{{"rule_id", 1}}},
        },
        m.AIDigests(): {
            {Keys: bson.D{{"snapshot_id", 1}}},
        },
    }

    for coll, idxModels := range indexes {
        for _, idx := range idxModels {
            if _, err := coll.Indexes().CreateOne(ctx, idx); err != nil {
                return err
            }
        }
    }

    return nil
}
```

**Step 2: Update main.go to connect to MongoDB on startup**

```go
// backend/cmd/server/main.go
package main

import (
    "context"
    "log"
    "time"

    "flowpilot/internal/api"
    "flowpilot/internal/config"
    "flowpilot/internal/db"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongo, err := db.Connect(ctx, cfg.MongoURI)
    if err != nil {
        log.Fatalf("failed to connect to MongoDB: %v", err)
    }
    defer mongo.Disconnect(context.Background())

    mongo.Init(cfg.MongoDB)

    if err := mongo.EnsureIndexes(ctx); err != nil {
        log.Fatalf("failed to create indexes: %v", err)
    }

    router := api.NewRouter(cfg)

    log.Printf("FlowPilot Finance starting on %s", cfg.Port)
    if err := router.Run(cfg.Port); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
```

**Step 3: Verify compilation**

Run: `cd backend && go build ./cmd/server`
Expected: No errors

**Step 4: Commit**

```bash
git add backend/internal/db/indexes.go backend/cmd/server/main.go
git commit -m "feat: add MongoDB indexes and wire DB into server startup"
```

---

## Phase 3: Portfolio Sync Workflow

### Task 8: Create connector interface

**Files:**
- Create: `backend/internal/connectors/connector.go`

**Step 1: Define the interface**

```go
// backend/internal/connectors/connector.go
package connectors

import (
    "context"

    "flowpilot/internal/models"
)

type Connector interface {
    Name() string
    FetchPositions(ctx context.Context) ([]models.Position, error)
}
```

**Step 2: Verify compilation**

Run: `cd backend && go build ./internal/connectors/`
Expected: No errors

**Step 3: Commit**

```bash
git add backend/internal/connectors/
git commit -m "feat: add connector interface"
```

---

### Task 9: Implement Schwab connector

**Files:**
- Create: `backend/internal/connectors/schwab/schwab.go`
- Create: `backend/internal/connectors/schwab/schwab_test.go`

**Step 1: Write the test**

```go
// backend/internal/connectors/schwab/schwab_test.go
package schwab

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFetchPositions(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/accounts" {
            json.NewEncoder(w).Encode([]AccountResponse{
                {SecuritiesAccount: SecuritiesAccount{
                    AccountNumber: "12345",
                    Positions: []SchwabPosition{
                        {
                            Instrument:   Instrument{Symbol: "AAPL", AssetType: "EQUITY"},
                            LongQuantity: 10,
                            MarketValue:  1500.0,
                            AveragePrice: 150.0,
                        },
                        {
                            Instrument:   Instrument{Symbol: "MSFT", AssetType: "EQUITY"},
                            LongQuantity: 5,
                            MarketValue:  2000.0,
                            AveragePrice: 400.0,
                        },
                    },
                }},
            })
            return
        }
        w.WriteHeader(404)
    }))
    defer server.Close()

    c := &Client{
        baseURL:     server.URL,
        accessToken: "test-token",
        httpClient:  server.Client(),
    }

    positions, err := c.FetchPositions(context.Background())
    require.NoError(t, err)
    assert.Len(t, positions, 2)
    assert.Equal(t, "AAPL", positions[0].Symbol)
    assert.Equal(t, float64(10), positions[0].Quantity)
    assert.Equal(t, 1500.0, positions[0].MarketValue)
    assert.Equal(t, "schwab", positions[0].Source)
    assert.Equal(t, "equity", positions[0].AssetType)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/connectors/schwab/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// backend/internal/connectors/schwab/schwab.go
package schwab

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "flowpilot/internal/models"
)

const defaultBaseURL = "https://api.schwabapi.com/trader/v1"

type Client struct {
    baseURL      string
    accessToken  string
    refreshToken string
    clientID     string
    clientSecret string
    redirectURI  string
    httpClient   *http.Client
}

func NewClient(clientID, clientSecret, redirectURI string) *Client {
    return &Client{
        baseURL:      defaultBaseURL,
        clientID:     clientID,
        clientSecret: clientSecret,
        redirectURI:  redirectURI,
        httpClient:   &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *Client) Name() string { return "schwab" }

func (c *Client) SetTokens(access, refresh string) {
    c.accessToken = access
    c.refreshToken = refresh
}

// Schwab API response types
type Instrument struct {
    Symbol    string `json:"symbol"`
    AssetType string `json:"assetType"`
}

type SchwabPosition struct {
    Instrument   Instrument `json:"instrument"`
    LongQuantity float64    `json:"longQuantity"`
    MarketValue  float64    `json:"marketValue"`
    AveragePrice float64    `json:"averagePrice"`
}

type SecuritiesAccount struct {
    AccountNumber string           `json:"accountNumber"`
    Positions     []SchwabPosition `json:"positions"`
}

type AccountResponse struct {
    SecuritiesAccount SecuritiesAccount `json:"securitiesAccount"`
}

func (c *Client) FetchPositions(ctx context.Context) ([]models.Position, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/accounts", nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    req.URL.RawQuery = "fields=positions"

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("schwab fetch accounts: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("schwab API returned %d", resp.StatusCode)
    }

    var accounts []AccountResponse
    if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
        return nil, fmt.Errorf("schwab decode: %w", err)
    }

    now := time.Now()
    var positions []models.Position
    for _, acct := range accounts {
        for _, p := range acct.SecuritiesAccount.Positions {
            markPrice := 0.0
            if p.LongQuantity > 0 {
                markPrice = p.MarketValue / p.LongQuantity
            }
            positions = append(positions, models.Position{
                Symbol:      p.Instrument.Symbol,
                Quantity:    p.LongQuantity,
                MarkPrice:   markPrice,
                MarketValue: p.MarketValue,
                AssetType:   strings.ToLower(p.Instrument.AssetType),
                Source:      "schwab",
                AccountID:   acct.SecuritiesAccount.AccountNumber,
                Timestamp:   now,
            })
        }
    }

    return positions, nil
}
```

**Step 4: Run tests**

Run: `cd backend && go test ./internal/connectors/schwab/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/connectors/schwab/
git commit -m "feat: add Schwab connector with position fetching"
```

---

### Task 10: Implement Coinbase connector

**Files:**
- Create: `backend/internal/connectors/coinbase/coinbase.go`
- Create: `backend/internal/connectors/coinbase/coinbase_test.go`

**Step 1: Write the test**

```go
// backend/internal/connectors/coinbase/coinbase_test.go
package coinbase

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFetchPositions(t *testing.T) {
    callCount := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/v2/accounts" {
            callCount++
            json.NewEncoder(w).Encode(AccountsResponse{
                Data: []Account{
                    {
                        ID:       "acct-1",
                        Currency: Currency{Code: "BTC", Name: "Bitcoin"},
                        Balance:  Balance{Amount: "1.5", Currency: "BTC"},
                    },
                    {
                        ID:       "acct-2",
                        Currency: Currency{Code: "ETH", Name: "Ethereum"},
                        Balance:  Balance{Amount: "10.0", Currency: "ETH"},
                    },
                },
            })
            return
        }
        if r.URL.Path == "/v2/prices/BTC-USD/spot" {
            json.NewEncoder(w).Encode(PriceResponse{Data: PriceData{Amount: "60000.00"}})
            return
        }
        if r.URL.Path == "/v2/prices/ETH-USD/spot" {
            json.NewEncoder(w).Encode(PriceResponse{Data: PriceData{Amount: "3000.00"}})
            return
        }
        w.WriteHeader(404)
    }))
    defer server.Close()

    c := &Client{
        baseURL:    server.URL,
        apiKey:     "test-key",
        apiSecret:  "test-secret",
        httpClient: server.Client(),
    }

    positions, err := c.FetchPositions(context.Background())
    require.NoError(t, err)
    assert.Len(t, positions, 2)
    assert.Equal(t, "BTC", positions[0].Symbol)
    assert.Equal(t, 1.5, positions[0].Quantity)
    assert.Equal(t, 60000.0, positions[0].MarkPrice)
    assert.Equal(t, 90000.0, positions[0].MarketValue)
    assert.Equal(t, "crypto", positions[0].AssetType)
    assert.Equal(t, "coinbase", positions[0].Source)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/connectors/coinbase/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// backend/internal/connectors/coinbase/coinbase.go
package coinbase

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "flowpilot/internal/models"
)

const defaultBaseURL = "https://api.coinbase.com"

type Client struct {
    baseURL    string
    apiKey     string
    apiSecret  string
    httpClient *http.Client
}

func NewClient(apiKey, apiSecret string) *Client {
    return &Client{
        baseURL:    defaultBaseURL,
        apiKey:     apiKey,
        apiSecret:  apiSecret,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *Client) Name() string { return "coinbase" }

// Coinbase API response types
type Currency struct {
    Code string `json:"code"`
    Name string `json:"name"`
}

type Balance struct {
    Amount   string `json:"amount"`
    Currency string `json:"currency"`
}

type Account struct {
    ID       string   `json:"id"`
    Currency Currency `json:"currency"`
    Balance  Balance  `json:"balance"`
}

type AccountsResponse struct {
    Data []Account `json:"data"`
}

type PriceData struct {
    Amount string `json:"amount"`
}

type PriceResponse struct {
    Data PriceData `json:"data"`
}

func (c *Client) doGet(ctx context.Context, path string, out interface{}) error {
    req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+c.apiKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("coinbase API returned %d for %s", resp.StatusCode, path)
    }

    return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) FetchPositions(ctx context.Context) ([]models.Position, error) {
    var acctResp AccountsResponse
    if err := c.doGet(ctx, "/v2/accounts", &acctResp); err != nil {
        return nil, fmt.Errorf("coinbase fetch accounts: %w", err)
    }

    now := time.Now()
    var positions []models.Position
    for _, acct := range acctResp.Data {
        qty, err := strconv.ParseFloat(acct.Balance.Amount, 64)
        if err != nil || qty <= 0 {
            continue
        }

        price, err := c.getSpotPrice(ctx, acct.Currency.Code)
        if err != nil {
            continue
        }

        positions = append(positions, models.Position{
            Symbol:      acct.Currency.Code,
            Quantity:    qty,
            MarkPrice:   price,
            MarketValue: qty * price,
            AssetType:   "crypto",
            Source:      "coinbase",
            AccountID:   acct.ID,
            Timestamp:   now,
        })
    }

    return positions, nil
}

func (c *Client) getSpotPrice(ctx context.Context, currency string) (float64, error) {
    var priceResp PriceResponse
    path := fmt.Sprintf("/v2/prices/%s-USD/spot", currency)
    if err := c.doGet(ctx, path, &priceResp); err != nil {
        return 0, err
    }
    return strconv.ParseFloat(priceResp.Data.Amount, 64)
}
```

**Step 4: Run tests**

Run: `cd backend && go test ./internal/connectors/coinbase/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/connectors/coinbase/
git commit -m "feat: add Coinbase connector with position fetching"
```

---

### Task 11: Create portfolio service (normalization + metrics)

**Files:**
- Create: `backend/internal/portfolio/service.go`
- Create: `backend/internal/portfolio/service_test.go`

**Step 1: Write the test**

```go
// backend/internal/portfolio/service_test.go
package portfolio

import (
    "testing"
    "time"

    "flowpilot/internal/models"
    "github.com/stretchr/testify/assert"
)

func TestComputeMetrics(t *testing.T) {
    positions := []models.Position{
        {Symbol: "AAPL", MarketValue: 5000, AssetType: "equity", Timestamp: time.Now()},
        {Symbol: "MSFT", MarketValue: 3000, AssetType: "equity", Timestamp: time.Now()},
        {Symbol: "BTC", MarketValue: 2000, AssetType: "crypto", Timestamp: time.Now()},
    }

    metrics := ComputeMetrics(positions)

    assert.Equal(t, 10000.0, metrics.NetWorth)
    assert.Equal(t, 80.0, metrics.ExposureByType["equity"])
    assert.Equal(t, 20.0, metrics.ExposureByType["crypto"])
    assert.Equal(t, 50.0, metrics.ConcentrationBySymbol["AAPL"])
    assert.Equal(t, 3, metrics.PositionCount)
}

func TestComputeMetrics_Empty(t *testing.T) {
    metrics := ComputeMetrics(nil)
    assert.Equal(t, 0.0, metrics.NetWorth)
    assert.Equal(t, 0, metrics.PositionCount)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/portfolio/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// backend/internal/portfolio/service.go
package portfolio

import (
    "flowpilot/internal/models"
)

type Metrics struct {
    NetWorth              float64            `json:"net_worth"`
    ExposureByType        map[string]float64 `json:"exposure_by_type"`
    ConcentrationBySymbol map[string]float64 `json:"concentration_by_symbol"`
    PositionCount         int                `json:"position_count"`
}

func ComputeMetrics(positions []models.Position) Metrics {
    if len(positions) == 0 {
        return Metrics{
            ExposureByType:        make(map[string]float64),
            ConcentrationBySymbol: make(map[string]float64),
        }
    }

    var netWorth float64
    typeValue := make(map[string]float64)
    symbolValue := make(map[string]float64)

    for _, p := range positions {
        netWorth += p.MarketValue
        typeValue[p.AssetType] += p.MarketValue
        symbolValue[p.Symbol] += p.MarketValue
    }

    exposure := make(map[string]float64)
    for t, v := range typeValue {
        exposure[t] = (v / netWorth) * 100
    }

    concentration := make(map[string]float64)
    for s, v := range symbolValue {
        concentration[s] = (v / netWorth) * 100
    }

    return Metrics{
        NetWorth:              netWorth,
        ExposureByType:        exposure,
        ConcentrationBySymbol: concentration,
        PositionCount:         len(positions),
    }
}
```

**Step 4: Run tests**

Run: `cd backend && go test ./internal/portfolio/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/portfolio/
git commit -m "feat: add portfolio metrics computation"
```

---

### Task 12: Create snapshot engine

**Files:**
- Create: `backend/internal/snapshot/engine.go`
- Create: `backend/internal/snapshot/engine_test.go`

**Step 1: Write the test**

```go
// backend/internal/snapshot/engine_test.go
package snapshot

import (
    "context"
    "os"
    "testing"
    "time"

    "flowpilot/internal/db"
    "flowpilot/internal/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/v2/bson"
)

func setupTestDB(t *testing.T) *db.MongoDB {
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        t.Skip("MONGODB_URI not set")
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongo, err := db.Connect(ctx, uri)
    require.NoError(t, err)
    mongo.Init("flowpilot_test")

    t.Cleanup(func() {
        mongo.DB.Drop(context.Background())
        mongo.Disconnect(context.Background())
    })

    return mongo
}

func TestCreateSnapshot(t *testing.T) {
    mongoDB := setupTestDB(t)
    engine := NewEngine(mongoDB)
    ctx := context.Background()

    positions := []models.Position{
        {Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab", AccountID: "123", Timestamp: time.Now()},
        {Symbol: "BTC", Quantity: 1, MarkPrice: 60000, MarketValue: 60000, AssetType: "crypto", Source: "coinbase", AccountID: "456", Timestamp: time.Now()},
    }

    snap, err := engine.Create(ctx, positions, 61500.0)
    require.NoError(t, err)
    assert.Equal(t, 61500.0, snap.NetWorth)
    assert.False(t, snap.ID.IsZero())

    // Verify snapshot positions were stored
    cursor, err := mongoDB.SnapshotPositions().Find(ctx, bson.D{{"snapshot_id", snap.ID}})
    require.NoError(t, err)
    var snapPositions []models.SnapshotPosition
    require.NoError(t, cursor.All(ctx, &snapPositions))
    assert.Len(t, snapPositions, 2)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/snapshot/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// backend/internal/snapshot/engine.go
package snapshot

import (
    "context"
    "time"

    "flowpilot/internal/db"
    "flowpilot/internal/models"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Engine struct {
    db *db.MongoDB
}

func NewEngine(db *db.MongoDB) *Engine {
    return &Engine{db: db}
}

func (e *Engine) Create(ctx context.Context, positions []models.Position, netWorth float64) (*models.Snapshot, error) {
    snap := models.Snapshot{
        CreatedAt: time.Now(),
        NetWorth:  netWorth,
    }

    res, err := e.db.Snapshots().InsertOne(ctx, snap)
    if err != nil {
        return nil, err
    }
    snap.ID = res.InsertedID.(bson.ObjectID)

    var snapPositions []interface{}
    for _, p := range positions {
        snapPositions = append(snapPositions, models.SnapshotPosition{
            SnapshotID: snap.ID,
            Symbol:     p.Symbol,
            Quantity:   p.Quantity,
            MarkPrice:  p.MarkPrice,
        })
    }

    if len(snapPositions) > 0 {
        if _, err := e.db.SnapshotPositions().InsertMany(ctx, snapPositions); err != nil {
            return nil, err
        }
    }

    return &snap, nil
}

func (e *Engine) GetLatest(ctx context.Context) (*models.Snapshot, error) {
    opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})
    var snap models.Snapshot
    err := e.db.Snapshots().FindOne(ctx, bson.D{}, opts).Decode(&snap)
    if err != nil {
        return nil, err
    }
    return &snap, nil
}

func (e *Engine) GetPositions(ctx context.Context, snapshotID bson.ObjectID) ([]models.SnapshotPosition, error) {
    cursor, err := e.db.SnapshotPositions().Find(ctx, bson.D{{"snapshot_id", snapshotID}})
    if err != nil {
        return nil, err
    }
    var positions []models.SnapshotPosition
    if err := cursor.All(ctx, &positions); err != nil {
        return nil, err
    }
    return positions, nil
}
```

**Step 4: Run tests**

Run: `cd backend && MONGODB_URI="<your-atlas-uri>" go test ./internal/snapshot/ -v`
Expected: PASS (or SKIP)

**Step 5: Commit**

```bash
git add backend/internal/snapshot/
git commit -m "feat: add snapshot engine with create and query"
```

---

### Task 13: Create diff engine

**Files:**
- Create: `backend/internal/snapshot/diff.go`
- Create: `backend/internal/snapshot/diff_test.go`

**Step 1: Write the test**

```go
// backend/internal/snapshot/diff_test.go
package snapshot

import (
    "testing"

    "flowpilot/internal/models"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/v2/bson"
)

func TestComputeDiff(t *testing.T) {
    fromID := bson.NewObjectID()
    toID := bson.NewObjectID()

    oldPositions := []models.SnapshotPosition{
        {SnapshotID: fromID, Symbol: "AAPL", Quantity: 10, MarkPrice: 150},
        {SnapshotID: fromID, Symbol: "GOOG", Quantity: 5, MarkPrice: 100},
    }

    newPositions := []models.SnapshotPosition{
        {SnapshotID: toID, Symbol: "AAPL", Quantity: 10, MarkPrice: 170},
        {SnapshotID: toID, Symbol: "BTC", Quantity: 1, MarkPrice: 60000},
    }

    diff := ComputeDiff(fromID, toID, 2000.0, 61700.0, oldPositions, newPositions)

    assert.Equal(t, fromID, diff.FromSnapshot)
    assert.Equal(t, toID, diff.ToSnapshot)
    assert.Equal(t, 59700.0, diff.NetWorthChange)
    assert.Contains(t, diff.AddedSymbols, "BTC")
    assert.Contains(t, diff.RemovedSymbols, "GOOG")
    assert.True(t, len(diff.TopMovers) > 0)
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/snapshot/ -v -run TestComputeDiff`
Expected: FAIL

**Step 3: Write implementation**

```go
// backend/internal/snapshot/diff.go
package snapshot

import (
    "math"
    "sort"
    "time"

    "flowpilot/internal/models"
    "go.mongodb.org/mongo-driver/v2/bson"
)

func ComputeDiff(fromID, toID bson.ObjectID, oldNetWorth, newNetWorth float64, oldPositions, newPositions []models.SnapshotPosition) models.Diff {
    oldMap := make(map[string]models.SnapshotPosition)
    for _, p := range oldPositions {
        oldMap[p.Symbol] = p
    }

    newMap := make(map[string]models.SnapshotPosition)
    for _, p := range newPositions {
        newMap[p.Symbol] = p
    }

    var movers []models.Mover
    var added, removed []string

    // Find changes and new positions
    for sym, np := range newMap {
        op, existed := oldMap[sym]
        if !existed {
            added = append(added, sym)
            movers = append(movers, models.Mover{
                Symbol:   sym,
                OldValue: 0,
                NewValue: np.Quantity * np.MarkPrice,
            })
            continue
        }

        oldVal := op.Quantity * op.MarkPrice
        newVal := np.Quantity * np.MarkPrice
        if oldVal > 0 {
            changePct := ((newVal - oldVal) / oldVal) * 100
            movers = append(movers, models.Mover{
                Symbol:        sym,
                ChangePercent: changePct,
                OldValue:      oldVal,
                NewValue:      newVal,
            })
        }
    }

    // Find removed positions
    for sym, op := range oldMap {
        if _, exists := newMap[sym]; !exists {
            removed = append(removed, sym)
            movers = append(movers, models.Mover{
                Symbol:        sym,
                ChangePercent: -100,
                OldValue:      op.Quantity * op.MarkPrice,
                NewValue:      0,
            })
        }
    }

    // Sort movers by absolute change percent descending
    sort.Slice(movers, func(i, j int) bool {
        return math.Abs(movers[i].ChangePercent) > math.Abs(movers[j].ChangePercent)
    })

    // Keep top 10 movers
    if len(movers) > 10 {
        movers = movers[:10]
    }

    return models.Diff{
        FromSnapshot:   fromID,
        ToSnapshot:     toID,
        NetWorthChange: newNetWorth - oldNetWorth,
        TopMovers:      movers,
        AddedSymbols:   added,
        RemovedSymbols: removed,
        CreatedAt:      time.Now(),
    }
}
```

**Step 4: Run tests**

Run: `cd backend && go test ./internal/snapshot/ -v -run TestComputeDiff`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/snapshot/diff.go backend/internal/snapshot/diff_test.go
git commit -m "feat: add diff engine for snapshot comparison"
```

---

### Task 14: Create portfolio store (upsert positions)

**Files:**
- Create: `backend/internal/portfolio/store.go`
- Create: `backend/internal/portfolio/store_test.go`

**Step 1: Write the test**

```go
// backend/internal/portfolio/store_test.go
package portfolio

import (
    "context"
    "os"
    "testing"
    "time"

    "flowpilot/internal/db"
    "flowpilot/internal/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/v2/bson"
)

func setupTestDB(t *testing.T) *db.MongoDB {
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        t.Skip("MONGODB_URI not set")
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongo, err := db.Connect(ctx, uri)
    require.NoError(t, err)
    mongo.Init("flowpilot_test")
    mongo.EnsureIndexes(ctx)

    t.Cleanup(func() {
        mongo.DB.Drop(context.Background())
        mongo.Disconnect(context.Background())
    })

    return mongo
}

func TestUpsertAndGetAll(t *testing.T) {
    mongoDB := setupTestDB(t)
    store := NewStore(mongoDB)
    ctx := context.Background()

    positions := []models.Position{
        {Symbol: "AAPL", Quantity: 10, MarkPrice: 150, MarketValue: 1500, AssetType: "equity", Source: "schwab", AccountID: "123", Timestamp: time.Now()},
        {Symbol: "BTC", Quantity: 1, MarkPrice: 60000, MarketValue: 60000, AssetType: "crypto", Source: "coinbase", AccountID: "456", Timestamp: time.Now()},
    }

    err := store.UpsertPositions(ctx, positions)
    require.NoError(t, err)

    all, err := store.GetAll(ctx)
    require.NoError(t, err)
    assert.Len(t, all, 2)

    // Upsert again with updated price — should not duplicate
    positions[0].MarkPrice = 160
    positions[0].MarketValue = 1600
    err = store.UpsertPositions(ctx, positions)
    require.NoError(t, err)

    all, err = store.GetAll(ctx)
    require.NoError(t, err)
    assert.Len(t, all, 2)
    for _, p := range all {
        if p.Symbol == "AAPL" {
            assert.Equal(t, 160.0, p.MarkPrice)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/portfolio/ -v -run TestUpsert`
Expected: FAIL

**Step 3: Write implementation**

```go
// backend/internal/portfolio/store.go
package portfolio

import (
    "context"

    "flowpilot/internal/db"
    "flowpilot/internal/models"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
    db *db.MongoDB
}

func NewStore(db *db.MongoDB) *Store {
    return &Store{db: db}
}

func (s *Store) UpsertPositions(ctx context.Context, positions []models.Position) error {
    for _, p := range positions {
        filter := bson.D{
            {"symbol", p.Symbol},
            {"source", p.Source},
            {"account_id", p.AccountID},
        }
        update := bson.D{{"$set", p}}
        opts := options.UpdateOne().SetUpsert(true)

        if _, err := s.db.PositionsCurrent().UpdateOne(ctx, filter, update, opts); err != nil {
            return err
        }
    }
    return nil
}

func (s *Store) GetAll(ctx context.Context) ([]models.Position, error) {
    cursor, err := s.db.PositionsCurrent().Find(ctx, bson.D{})
    if err != nil {
        return nil, err
    }
    var positions []models.Position
    if err := cursor.All(ctx, &positions); err != nil {
        return nil, err
    }
    return positions, nil
}
```

**Step 4: Run tests**

Run: `cd backend && MONGODB_URI="<your-atlas-uri>" go test ./internal/portfolio/ -v`
Expected: PASS (or SKIP)

**Step 5: Commit**

```bash
git add backend/internal/portfolio/store.go backend/internal/portfolio/store_test.go
git commit -m "feat: add portfolio store with upsert and query"
```

---

### Task 15: Wire up sync workflow endpoint

**Files:**
- Create: `backend/internal/api/sync.go`
- Create: `backend/internal/api/deps.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/cmd/server/main.go`

**Step 1: Create dependencies container**

```go
// backend/internal/api/deps.go
package api

import (
    "flowpilot/internal/config"
    "flowpilot/internal/connectors"
    "flowpilot/internal/db"
    "flowpilot/internal/portfolio"
    "flowpilot/internal/snapshot"
)

type Deps struct {
    Config     *config.Config
    DB         *db.MongoDB
    Connectors []connectors.Connector
    Store      *portfolio.Store
    Snapshot   *snapshot.Engine
}
```

**Step 2: Create sync handler**

```go
// backend/internal/api/sync.go
package api

import (
    "log"
    "net/http"
    "sync"

    "flowpilot/internal/models"
    "flowpilot/internal/portfolio"
    "flowpilot/internal/snapshot"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

func (d *Deps) handleSync(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Fetch from all connectors concurrently
    var allPositions []models.Position
    var mu sync.Mutex
    var wg sync.WaitGroup
    var fetchErrors []string

    for _, conn := range d.Connectors {
        wg.Add(1)
        go func(conn connectors.Connector) {
            defer wg.Done()
            positions, err := conn.FetchPositions(ctx)
            if err != nil {
                log.Printf("connector %s error: %v", conn.Name(), err)
                mu.Lock()
                fetchErrors = append(fetchErrors, conn.Name()+": "+err.Error())
                mu.Unlock()
                return
            }
            mu.Lock()
            allPositions = append(allPositions, positions...)
            mu.Unlock()
        }(conn)
    }
    wg.Wait()

    if len(allPositions) == 0 {
        c.JSON(http.StatusBadGateway, gin.H{
            "error":   "no positions fetched",
            "details": fetchErrors,
        })
        return
    }

    // 2. Upsert positions
    if err := d.Store.UpsertPositions(ctx, allPositions); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store positions"})
        return
    }

    // 3. Compute metrics
    metrics := portfolio.ComputeMetrics(allPositions)

    // 4. Get previous snapshot for diffing
    prevSnap, prevErr := d.Snapshot.GetLatest(ctx)

    // 5. Create new snapshot
    snap, err := d.Snapshot.Create(ctx, allPositions, metrics.NetWorth)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create snapshot"})
        return
    }

    // 6. Compute diff if previous snapshot exists
    var diff *models.Diff
    if prevErr == nil && prevSnap != nil {
        prevPositions, err := d.Snapshot.GetPositions(ctx, prevSnap.ID)
        if err == nil {
            newPositions, _ := d.Snapshot.GetPositions(ctx, snap.ID)
            d := snapshot.ComputeDiff(prevSnap.ID, snap.ID, prevSnap.NetWorth, snap.NetWorth, prevPositions, newPositions)
            diff = &d

            // Store diff
            _, _ = d.DB.Diffs().InsertOne(ctx, diff)
        }
    } else if prevErr != nil && prevErr != mongo.ErrNoDocuments {
        log.Printf("warning: failed to get previous snapshot: %v", prevErr)
    }

    c.JSON(http.StatusOK, gin.H{
        "snapshot":     snap,
        "metrics":      metrics,
        "diff":         diff,
        "positions":    len(allPositions),
        "fetch_errors": fetchErrors,
    })
}
```

**Note:** This handler has a bug — the diff variable `d` shadows the package import. This will need fixing during implementation. The `d.DB.Diffs()` line should use the Deps DB reference instead. The correct approach:

```go
// In the diff section, use the Deps receiver:
if prevErr == nil && prevSnap != nil {
    prevPositions, err := d.Snapshot.GetPositions(ctx, prevSnap.ID)
    if err == nil {
        newPositions, _ := d.Snapshot.GetPositions(ctx, snap.ID)
        computedDiff := snapshot.ComputeDiff(prevSnap.ID, snap.ID, prevSnap.NetWorth, snap.NetWorth, prevPositions, newPositions)
        diff = &computedDiff
        _, _ = d.DB.Diffs().InsertOne(ctx, diff)
    }
}
```

**Step 3: Update router to accept Deps**

```go
// backend/internal/api/router.go
package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func NewRouter(deps *Deps) *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    {
        api.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })
        api.POST("/sync", deps.handleSync)
    }

    return r
}
```

**Step 4: Update main.go to wire everything**

```go
// backend/cmd/server/main.go
package main

import (
    "context"
    "log"
    "time"

    "flowpilot/internal/api"
    "flowpilot/internal/config"
    "flowpilot/internal/connectors"
    "flowpilot/internal/connectors/coinbase"
    "flowpilot/internal/connectors/schwab"
    "flowpilot/internal/db"
    "flowpilot/internal/portfolio"
    "flowpilot/internal/snapshot"
)

func main() {
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
        conns = append(conns, schwab.NewClient(cfg.SchwabClientID, cfg.SchwabClientSecret, cfg.SchwabRedirectURI))
    }
    if cfg.CoinbaseAPIKey != "" {
        conns = append(conns, coinbase.NewClient(cfg.CoinbaseAPIKey, cfg.CoinbaseAPISecret))
    }

    deps := &api.Deps{
        Config:     cfg,
        DB:         mongoDB,
        Connectors: conns,
        Store:      portfolio.NewStore(mongoDB),
        Snapshot:   snapshot.NewEngine(mongoDB),
    }

    router := api.NewRouter(deps)

    log.Printf("FlowPilot Finance starting on %s", cfg.Port)
    log.Printf("Connectors enabled: %d", len(conns))
    if err := router.Run(cfg.Port); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
```

**Step 5: Verify compilation**

Run: `cd backend && go build ./cmd/server`
Expected: No errors

**Step 6: Update router test**

The router test from Task 4 needs updating since `NewRouter` now takes `*Deps` instead of `*config.Config`:

```go
// backend/internal/api/router_test.go
package api

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
    router := NewRouter(&Deps{})

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/health", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var body map[string]string
    err := json.Unmarshal(w.Body.Bytes(), &body)
    require.NoError(t, err)
    assert.Equal(t, "ok", body["status"])
}
```

**Step 7: Run all tests**

Run: `cd backend && go test ./... -v`
Expected: All unit tests PASS, integration tests SKIP (unless MONGODB_URI is set)

**Step 8: Commit**

```bash
git add backend/internal/api/ backend/cmd/server/main.go
git commit -m "feat: wire up sync workflow endpoint with connectors, store, snapshot, and diff"
```

---

### Task 16: Add GET /api/portfolio endpoint

**Files:**
- Create: `backend/internal/api/portfolio.go`
- Modify: `backend/internal/api/router.go`

**Step 1: Create handler**

```go
// backend/internal/api/portfolio.go
package api

import (
    "net/http"

    "flowpilot/internal/portfolio"
    "github.com/gin-gonic/gin"
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
```

**Step 2: Register route in router.go**

Add to the `api` group in `NewRouter`:

```go
api.GET("/portfolio", deps.handleGetPortfolio)
```

**Step 3: Verify compilation**

Run: `cd backend && go build ./cmd/server`
Expected: No errors

**Step 4: Commit**

```bash
git add backend/internal/api/portfolio.go backend/internal/api/router.go
git commit -m "feat: add GET /api/portfolio endpoint"
```

---

### Task 17: Add GET /api/snapshots endpoints

**Files:**
- Create: `backend/internal/api/snapshots.go`
- Modify: `backend/internal/api/router.go`

**Step 1: Create handlers**

```go
// backend/internal/api/snapshots.go
package api

import (
    "net/http"

    "flowpilot/internal/models"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *Deps) handleListSnapshots(c *gin.Context) {
    ctx := c.Request.Context()

    opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(50)
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
    err = d.DB.Diffs().FindOne(ctx, bson.D{{"to_snapshot", oid}}).Decode(&diff)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "diff not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"diff": diff})
}
```

**Step 2: Register routes in router.go**

Add to the `api` group:

```go
api.GET("/snapshots", deps.handleListSnapshots)
api.GET("/snapshots/:id/diff", deps.handleGetSnapshotDiff)
```

**Step 3: Verify compilation**

Run: `cd backend && go build ./cmd/server`
Expected: No errors

**Step 4: Commit**

```bash
git add backend/internal/api/snapshots.go backend/internal/api/router.go
git commit -m "feat: add snapshot list and diff endpoints"
```

---

## Summary

After completing all 17 tasks, you will have:

- A Go backend with Gin serving on `:8080`
- Config loading from environment variables
- MongoDB Atlas connection with 7 collections and indexes
- Schwab and Coinbase connectors fetching real positions
- Portfolio metrics computation (net worth, exposure, concentration)
- Snapshot creation and historical tracking
- Diff engine comparing snapshots
- Full sync workflow: `POST /api/sync` → fetch → normalize → store → snapshot → diff
- Read endpoints: `GET /api/portfolio`, `GET /api/snapshots`, `GET /api/snapshots/:id/diff`

**Not yet implemented (future phases):**
- Alert rule evaluation and `alert_rules`/`alert_events` endpoints
- Notification dispatch (email, Telegram, Discord)
- AI digest generation and `POST /api/chat` with SSE streaming
- Scenario simulation (`POST /api/scenarios`)
- Frontend (React + Vite)
