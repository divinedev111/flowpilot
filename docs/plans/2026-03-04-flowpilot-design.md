# FlowPilot Finance — Design Document

Date: 2026-03-04

## Overview

FlowPilot Finance is a personal portfolio automation and analytics platform. It aggregates holdings from Schwab and Coinbase, tracks changes over time, triggers alerts with multi-channel notifications, and provides a conversational AI interface powered by Claude for portfolio analysis.

## Architecture

Go monolith using Gin web framework. Single binary handles all logic: API endpoints, connectors, metrics, alerts, notifications, and AI calls.

## Project Structure

```
mukul-financial-ops/
├── docs/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/          # env loading, secrets
│   │   ├── models/          # Go structs matching MongoDB schema
│   │   ├── db/              # MongoDB client, collection accessors
│   │   ├── connectors/
│   │   │   ├── schwab/      # Schwab Trader API client
│   │   │   └── coinbase/    # Coinbase API client
│   │   ├── portfolio/       # normalization, metrics computation
│   │   ├── snapshot/        # snapshot creation, diff engine
│   │   ├── alerts/          # rule evaluation engine
│   │   ├── notify/          # email, Telegram, Discord dispatch
│   │   ├── ai/              # Claude SDK wrapper, prompt templates, chat
│   │   └── api/             # Gin route handlers
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/           # Dashboard, Alerts, Chat, Scenarios
│   │   ├── hooks/
│   │   ├── api/
│   │   └── types/
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml
└── .env.example
```

## Data Flow — Sync Workflow

Triggered via `POST /api/sync`:

1. **Fetch** — Schwab connector (OAuth2) + Coinbase connector (API key) fetch holdings concurrently
2. **Normalize** — Map raw responses to unified Position struct: `{ symbol, quantity, mark_price, market_value, asset_type, source, account_id, timestamp }`
3. **Store** — Upsert into `positions_current` (keyed by symbol+source+account_id)
4. **Snapshot** — Create snapshot record + copy positions to `snapshot_positions`
5. **Diff** — Compare vs previous snapshot: net worth change, top movers, new/removed positions. Store in `diffs`
6. **Alert Evaluation** — Load `alert_rules`, evaluate against current state, store triggers in `alert_events`
7. **Notify** — Dispatch triggered alerts to configured channels (dashboard, email, Telegram, Discord) via goroutines
8. **AI Digest** — Send metrics + diff + alerts to Claude, store summary in `ai_digests`

## MongoDB Collections

| Collection | Purpose | Key Fields |
|---|---|---|
| `positions_current` | Latest holdings | symbol, quantity, mark_price, market_value, asset_type, source, account_id, timestamp |
| `snapshots` | Historical portfolio states | snapshot_id, created_at, net_worth |
| `snapshot_positions` | Positions per snapshot | snapshot_id, symbol, quantity, mark_price |
| `diffs` | Changes between snapshots | from_snapshot, to_snapshot, net_worth_change, top_movers |
| `alert_rules` | User-defined monitoring rules | rule_type, threshold, target_asset |
| `alert_events` | Triggered alerts | rule_id, severity, timestamp, evidence |
| `ai_digests` | AI-generated summaries | snapshot_id, summary, risk_notes, timestamp |

## API Endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/api/sync` | Trigger full portfolio sync workflow |
| GET | `/api/portfolio` | Current positions + metrics |
| GET | `/api/snapshots` | List historical snapshots |
| GET | `/api/snapshots/:id/diff` | Diff for a specific snapshot |
| GET | `/api/alerts/rules` | List alert rules |
| POST | `/api/alerts/rules` | Create/update alert rule |
| GET | `/api/alerts/events` | Recent triggered alerts |
| POST | `/api/chat` | Conversational AI with SSE streaming |
| POST | `/api/scenarios` | Run what-if simulation |

## Connectors

### Schwab Trader API
- OAuth2 authentication flow
- Fetch account positions
- Map to unified Position format

### Coinbase API
- API key authentication
- Fetch account holdings
- Map to unified Position format

## Alert System

### Rule Types
- `exposure_threshold` — asset type allocation exceeds X%
- `position_change_percent` — position moved > X% since last snapshot
- `concentration_threshold` — single position > X% of portfolio
- `net_worth_change` — total value changed > X%

### Notification Channels
- **Dashboard** — stored in `alert_events`, polled by frontend
- **Email** — SMTP via `gomail` or `net/smtp`
- **Telegram** — Bot API HTTP POST
- **Discord** — Webhook URL POST

All channels opt-in via env vars. Dispatch is concurrent via goroutines. Common `Notifier` interface for extensibility.

## AI Layer

Direct Anthropic Go SDK. No LangChain.

### System Prompt
```
You are a portfolio analyst for FlowPilot Finance. You have full access
to the user's portfolio data. Help them understand their holdings,
analyze risk, suggest actions, and discuss strategies.
```

### Use Cases
- **Portfolio Digest** — auto-generated after each sync with snapshot diff, top movers, alerts
- **Chat** — conversational interface with SSE streaming, portfolio context injected each turn
- **Scenario Analysis** — hypothetical allocation adjustments, recomputed metrics sent to Claude for comparison

### Conversation Management
- In-memory map of conversation_id to message history
- Capped at last N messages for token limits
- Fresh portfolio context injected as system message each turn

## Frontend

React + TypeScript + Vite. Recharts for data visualization.

### Pages
- **Dashboard** — net worth, allocation chart, positions table, recent changes
- **Alerts** — configure rules + view triggered alert history
- **Chat** — ChatGPT-style interface with SSE streaming and markdown rendering
- **Scenarios** — adjust allocations via sliders, view recomputed metrics + AI commentary

### Key Components
- PortfolioSummary, PositionsTable, AllocationChart, DiffView
- AlertRuleForm, AlertEventList
- ChatWindow (SSE streaming, markdown)
- ScenarioBuilder (sliders, results panel)

### State Management
React hooks + context. No Redux. Each page fetches its own data.

## Configuration

All secrets via environment variables:
- `MONGODB_URI` — Atlas connection string
- `ANTHROPIC_API_KEY` — Claude API key
- `SCHWAB_CLIENT_ID`, `SCHWAB_CLIENT_SECRET`, `SCHWAB_REDIRECT_URI`
- `COINBASE_API_KEY`, `COINBASE_API_SECRET`
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `ALERT_EMAIL_TO`
- `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`
- `DISCORD_WEBHOOK_URL`
