[![CI](https://github.com/divinedev111/flowpilot/actions/workflows/ci.yml/badge.svg)](https://github.com/divinedev111/flowpilot/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# FlowPilot

A self-hosted portfolio operations platform that aggregates holdings from Charles Schwab, Coinbase, and Polymarket into a single dashboard. Tracks changes over time via snapshots, fires configurable alerts when thresholds are crossed, and uses Claude AI to explain what changed.

## What It Does

- **Unified portfolio view** across stocks (Schwab), crypto (Coinbase), and prediction markets (Polymarket)
- **Automated snapshots** with day-over-day diffing and P&L tracking
- **Configurable alerts** -- exposure thresholds, concentration limits, position changes, net worth swings
- **AI digest** -- Claude summarizes portfolio changes and flags risks after each sync
- **Correlation analysis** -- NxN Pearson correlation matrix using free Binance + Yahoo Finance price data
- **Risk dashboard** -- concentration, drawdown, HHI diversification score, exposure breakdowns
- **Scenario simulation** -- adjust target allocations and get AI commentary on the rebalance
- **Notifications** -- alerts dispatched via email (SMTP), Telegram, and Discord
- **Research notes** -- rich text editor with tagging and symbol association
- **Investment strategies** -- named groupings with position assignment and allocation tracking
- **Compliance policies** -- rules for max concentration, min positions, max exposure
- **Prediction markets** -- browse Polymarket markets relevant to your holdings, view wallet positions via MetaMask
- **AI chat** -- ask natural language questions about your portfolio with full context streaming via SSE

## Architecture

```
Schwab API ──┐
Coinbase API ─┤── Connectors (concurrent) ── Normalize ── Portfolio Engine
Polymarket ───┘                                               │
                                                              ├── Snapshot Engine ── Diff Engine
                                                              ├── Alert Engine ── Notification Dispatcher
                                                              ├── Policy Engine
                                                              └── AI Layer (digest, chat, insights)
```

**Key design decision:** All financial math is deterministic. AI never computes numbers -- it only receives pre-computed metrics and explains them. If AI fails, rule-based insights still return.

### Backend
- Go with Gin HTTP framework
- MongoDB for persistence (snapshots, positions, alerts, cost basis, price cache)
- Anthropic SDK for Claude API integration
- Self-signed TLS for local HTTPS
- Dependency injection via a `Deps` struct

### Frontend
- React + TypeScript with Vite
- Recharts for charts, Tiptap for rich text editing
- 13 pages: Dashboard, Performance, Risk, Correlation, Predictions, Alerts, Scenarios, Notes, Strategies, Policies, Watchlist, News, Settings

## Setup

### Prerequisites

- Go 1.25+
- Node.js 20+
- MongoDB Atlas account (or local MongoDB)
- Anthropic API key

### Configuration

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

### Required Environment Variables

| Variable | Description |
|---|---|
| `MONGODB_URI` | MongoDB connection string |
| `MONGODB_NAME` | Database name (default: `flowpilot`) |
| `ANTHROPIC_API_KEY` | Claude API key for AI features |
| `PORT` | Backend port (default: `8080`) |
| `FRONTEND_PORT` | Frontend dev server port (default: `3000`) |

### Connector Setup

#### Schwab

1. Register at [developer.schwab.com](https://developer.schwab.com)
2. Create a Trader API app and get your client ID/secret
3. Set `SCHWAB_CLIENT_ID`, `SCHWAB_CLIENT_SECRET`, `SCHWAB_REDIRECT_URI`
4. Visit `/api/auth/schwab` in the browser to complete OAuth

#### Coinbase

1. Create a CDP API key at [portal.cdp.coinbase.com](https://portal.cdp.coinbase.com)
2. Select Ed25519 key type
3. Set `COINBASE_API_KEY` (key ID) and `COINBASE_API_SECRET` (base64-encoded private key)

#### Polymarket

No API key needed -- positions are on-chain. Connect your MetaMask wallet in the Predictions page. The wallet address is saved to MongoDB.

### Generate TLS Certificates

The backend runs over HTTPS with self-signed certs:

```bash
cd backend
openssl req -x509 -newkey rsa:2048 -keyout server.key -out server.crt -days 365 -nodes -subj '/CN=localhost'
```

### Running

Use the included script:

```bash
./flowpilot.sh start    # builds backend, starts both services
./flowpilot.sh stop     # stops both
./flowpilot.sh status   # check what's running
./flowpilot.sh logs backend   # tail backend logs
```

Or run manually:

```bash
# Backend
cd backend && go build -o backend ./cmd/server && cd .. && ./backend/backend

# Frontend
cd frontend && npm install && npx vite --port 3002
```

### Running Tests

```bash
cd backend && go test ./...
```

Integration tests (snapshot, portfolio store) require `MONGODB_URI` to be set. They are skipped otherwise.

## Module Rename Note

The Go module is currently `flowpilot`. If you rename the repository, update `backend/go.mod` and all import paths accordingly. A find-and-replace of `flowpilot/internal/` across all `.go` files will handle it.

## Project Structure

```
backend/
  cmd/server/          main.go entrypoint
  internal/
    ai/                Claude integration (chat, digest, insights)
    alerts/            Alert rule evaluation engine
    api/               HTTP handlers and router
    audit/             Request audit logging middleware
    config/            Environment variable loading
    connectors/        Schwab, Coinbase, Polymarket data fetchers
    correlation/       Price fetching (Binance/Yahoo) and Pearson matrix
    crypto/            AES-256-GCM encryption utilities
    db/                MongoDB client, collections, indexes
    models/            Domain types (Position, Snapshot, AlertRule, etc.)
    news/              Google News RSS + AI sentiment scoring
    notify/            Email, Telegram, Discord notification dispatchers
    policy/            Compliance rule evaluation
    polymarket/        Polymarket Gamma API client and market matching
    portfolio/         Metrics computation, P&L enrichment, cost basis
    snapshot/          Snapshot creation, diff computation
frontend/
  src/
    api/               API client
    components/        Reusable UI components
    pages/             Route pages (Dashboard, Risk, Alerts, etc.)
    types/             TypeScript interfaces
    utils/             Formatting utilities
```
