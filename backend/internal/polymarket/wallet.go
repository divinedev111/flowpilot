package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const dataAPIBase = "https://data-api.polymarket.com"

type WalletPosition struct {
	ConditionID    string  `json:"condition_id"`
	Title          string  `json:"title"`
	Outcome        string  `json:"outcome"`
	Size           float64 `json:"size"`
	AvgPrice       float64 `json:"avg_price"`
	CurrentPrice   float64 `json:"current_price"`
	CurrentValue   float64 `json:"current_value"`
	InitialValue   float64 `json:"initial_value"`
	UnrealizedPnl  float64 `json:"unrealized_pnl"`
	CashPnl        float64 `json:"cash_pnl"`
	PercentPnl     float64 `json:"percent_pnl"`
	Redeemable     bool    `json:"redeemable"`
	Resolved       bool    `json:"resolved"`
}

type WalletSummary struct {
	TotalPositions    int     `json:"total_positions"`
	ActivePositions   int     `json:"active_positions"`
	ResolvedPositions int     `json:"resolved_positions"`
	ClaimableCount    int     `json:"claimable_count"`
	TotalValue        float64 `json:"total_value"`
	TotalCostBasis    float64 `json:"total_cost_basis"`
	TotalUnrealizedPnl float64 `json:"total_unrealized_pnl"`
	TotalCashPnl      float64 `json:"total_cash_pnl"`
}

type dataAPIPosition struct {
	Asset         string  `json:"asset"`
	ConditionID   string  `json:"conditionId"`
	Title         string  `json:"title"`
	Outcome       string  `json:"outcome"`
	Size          float64 `json:"size"`
	AvgPrice      float64 `json:"avgPrice"`
	CurPrice      float64 `json:"curPrice"`
	InitialValue  float64 `json:"initialValue"`
	CurrentValue  float64 `json:"currentValue"`
	CashPnl       float64 `json:"cashPnl"`
	PercentPnl    float64 `json:"percentPnl"`
	Redeemable    bool    `json:"redeemable"`
	Resolved      bool    `json:"resolved"`
}

func FetchWalletPositions(ctx context.Context, walletAddress string) ([]WalletPosition, *WalletSummary, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	url := fmt.Sprintf("%s/positions?user=%s&limit=500&sizeThreshold=0", dataAPIBase, walletAddress)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("data api error %d: %s", resp.StatusCode, string(body))
	}

	var rawPositions []dataAPIPosition
	if err := json.Unmarshal(body, &rawPositions); err != nil {
		return nil, nil, fmt.Errorf("unmarshal: %w", err)
	}

	positions := make([]WalletPosition, 0, len(rawPositions))
	summary := &WalletSummary{}

	for _, raw := range rawPositions {
		if raw.Size == 0 {
			continue
		}

		pos := WalletPosition{
			ConditionID:   raw.ConditionID,
			Title:         raw.Title,
			Outcome:       raw.Outcome,
			Size:          raw.Size,
			AvgPrice:      raw.AvgPrice,
			CurrentPrice:  raw.CurPrice,
			CurrentValue:  raw.CurrentValue,
			InitialValue:  raw.InitialValue,
			UnrealizedPnl: raw.CurrentValue - raw.InitialValue,
			CashPnl:       raw.CashPnl,
			PercentPnl:    raw.PercentPnl,
			Redeemable:    raw.Redeemable,
			Resolved:      raw.Resolved,
		}

		positions = append(positions, pos)

		summary.TotalPositions++
		if raw.Resolved {
			summary.ResolvedPositions++
		} else {
			summary.ActivePositions++
			summary.TotalValue += raw.CurrentValue
			summary.TotalCostBasis += raw.InitialValue
			summary.TotalUnrealizedPnl += pos.UnrealizedPnl
		}
		if raw.Redeemable {
			summary.ClaimableCount++
		}
		summary.TotalCashPnl += raw.CashPnl
	}

	return positions, summary, nil
}
