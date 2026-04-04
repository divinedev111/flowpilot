package portfolio

import (
	"flowpilot/internal/models"
)

type Metrics struct {
	NetWorth              float64            `json:"net_worth"`
	ExposureByType        map[string]float64 `json:"exposure_by_type"`
	ConcentrationBySymbol map[string]float64 `json:"concentration_by_symbol"`
	PositionCount         int                `json:"position_count"`
	TotalPnL              float64            `json:"total_pnl"`
	DayPnL                float64            `json:"day_pnl"`
	TotalPnLPct           float64            `json:"total_pnl_pct"`
}

func ComputeMetrics(positions []models.Position) Metrics {
	if len(positions) == 0 {
		return Metrics{
			ExposureByType:        make(map[string]float64),
			ConcentrationBySymbol: make(map[string]float64),
		}
	}

	var netWorth float64
	var totalPnL float64
	var dayPnL float64
	var totalCostBasis float64
	typeValue := make(map[string]float64)
	symbolValue := make(map[string]float64)

	for _, p := range positions {
		netWorth += p.MarketValue
		typeValue[p.AssetType] += p.MarketValue
		symbolValue[p.Symbol] += p.MarketValue
		totalPnL += p.TotalPnL
		dayPnL += p.DayChange
		totalCostBasis += p.CostBasis * p.Quantity
	}

	exposure := make(map[string]float64)
	for t, v := range typeValue {
		exposure[t] = (v / netWorth) * 100
	}

	concentration := make(map[string]float64)
	for s, v := range symbolValue {
		concentration[s] = (v / netWorth) * 100
	}

	var totalPnLPct float64
	if totalCostBasis > 0 {
		totalPnLPct = (totalPnL / totalCostBasis) * 100
	}

	return Metrics{
		NetWorth:              netWorth,
		ExposureByType:        exposure,
		ConcentrationBySymbol: concentration,
		PositionCount:         len(positions),
		TotalPnL:              totalPnL,
		DayPnL:                dayPnL,
		TotalPnLPct:           totalPnLPct,
	}
}
