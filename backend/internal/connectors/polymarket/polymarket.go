package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"flowpilot/internal/connectors"
	"flowpilot/internal/models"
)

var _ connectors.Connector = (*Client)(nil)

const dataAPIBase = "https://data-api.polymarket.com"

type Client struct {
	walletAddress string
	client        *http.Client
}

func NewClient(walletAddress string) *Client {
	return &Client{
		walletAddress: walletAddress,
		client:        &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "polymarket" }

func (c *Client) SetWalletAddress(addr string) {
	c.walletAddress = addr
}

func (c *Client) WalletAddress() string {
	return c.walletAddress
}

func (c *Client) HasWallet() bool {
	return c.walletAddress != ""
}

type rawPosition struct {
	Asset        string  `json:"asset"`
	ConditionID  string  `json:"conditionId"`
	Title        string  `json:"title"`
	Outcome      string  `json:"outcome"`
	Size         float64 `json:"size"`
	AvgPrice     float64 `json:"avgPrice"`
	CurPrice     float64 `json:"curPrice"`
	InitialValue float64 `json:"initialValue"`
	CurrentValue float64 `json:"currentValue"`
	CashPnl      float64 `json:"cashPnl"`
	PercentPnl   float64 `json:"percentPnl"`
	Redeemable   bool    `json:"redeemable"`
	Resolved     bool    `json:"resolved"`
}

func (c *Client) FetchPositions(ctx context.Context) ([]models.Position, error) {
	if c.walletAddress == "" {
		return nil, fmt.Errorf("not authenticated: no wallet address configured")
	}

	url := fmt.Sprintf("%s/positions?user=%s&limit=500&sizeThreshold=0", dataAPIBase, c.walletAddress)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("data api error %d: %s", resp.StatusCode, string(body))
	}

	var raw []rawPosition
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	now := time.Now()
	var positions []models.Position
	for _, r := range raw {
		// Skip dust positions (under $0.05 value)
		if r.CurrentValue < 0.05 {
			continue
		}

		// Build a readable symbol from the title
		symbol := buildSymbol(r.Title, r.Outcome)

		pnl := r.CurrentValue - r.InitialValue
		pnlPct := 0.0
		if r.InitialValue > 0 {
			pnlPct = (pnl / r.InitialValue) * 100
		}

		positions = append(positions, models.Position{
			Symbol:      symbol,
			Quantity:    r.Size,
			MarkPrice:   r.CurPrice,
			MarketValue: r.CurrentValue,
			AssetType:   "prediction_market",
			Source:       "polymarket",
			AccountID:   c.walletAddress,
			Timestamp:   now,
			CostBasis:   r.InitialValue,
			TotalPnL:    pnl,
			TotalPnLPct: pnlPct,
		})
	}

	return positions, nil
}

// buildSymbol creates a short identifier from the market title.
func buildSymbol(title, outcome string) string {
	// Truncate title to something reasonable
	t := title
	if len(t) > 40 {
		t = t[:40] + "..."
	}
	// Remove special characters that could break display
	t = strings.ReplaceAll(t, "|", "-")
	return fmt.Sprintf("%s [%s]", t, outcome)
}
