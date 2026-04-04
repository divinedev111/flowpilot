package correlation

import (
	"math"
	"sort"
	"time"

	"gonum.org/v1/gonum/stat"
)

type CorrelationResult struct {
	Symbols    []string    `json:"symbols"`
	Matrix     [][]float64 `json:"matrix"`
	PeriodDays int         `json:"period_days"`
	ComputedAt time.Time   `json:"computed_at"`
}

type CorrelatedPair struct {
	A           string  `json:"a"`
	B           string  `json:"b"`
	Correlation float64 `json:"correlation"`
}

// DailyReturns computes log returns from daily close prices.
// Returns (dates, returns) where returns[i] = ln(close[i+1]/close[i]).
func DailyReturns(prices []DayPrice) []float64 {
	if len(prices) < 2 {
		return nil
	}
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1].Close > 0 {
			returns[i-1] = math.Log(prices[i].Close / prices[i-1].Close)
		}
	}
	return returns
}

// AlignReturns takes a map of symbol->DayPrice slices and produces aligned return
// series (same dates) for all symbols. Returns (symbols, returns[][]float64, periodDays).
func AlignReturns(priceMap map[string][]DayPrice) ([]string, [][]float64, int) {
	if len(priceMap) == 0 {
		return nil, nil, 0
	}

	// Build a set of dates present in ALL symbols
	symbols := make([]string, 0, len(priceMap))
	for s := range priceMap {
		symbols = append(symbols, s)
	}
	sort.Strings(symbols)

	// Build date->close maps for each symbol
	dateMaps := make([]map[string]float64, len(symbols))
	for i, sym := range symbols {
		dm := make(map[string]float64)
		for _, dp := range priceMap[sym] {
			dm[dp.Date] = dp.Close
		}
		dateMaps[i] = dm
	}

	// Find common dates
	commonDates := make([]string, 0)
	for date := range dateMaps[0] {
		allHave := true
		for _, dm := range dateMaps[1:] {
			if _, ok := dm[date]; !ok {
				allHave = false
				break
			}
		}
		if allHave {
			commonDates = append(commonDates, date)
		}
	}
	sort.Strings(commonDates)

	if len(commonDates) < 2 {
		return symbols, nil, 0
	}

	// Build aligned return series for each symbol
	allReturns := make([][]float64, len(symbols))
	for i, sym := range symbols {
		dm := dateMaps[i]
		_ = sym
		aligned := make([]DayPrice, len(commonDates))
		for j, d := range commonDates {
			aligned[j] = DayPrice{Date: d, Close: dm[d]}
		}
		allReturns[i] = DailyReturns(aligned)
	}

	return symbols, allReturns, len(commonDates)
}

// BuildCorrelationMatrix computes the NxN Pearson correlation matrix from
// aligned return series.
func BuildCorrelationMatrix(symbols []string, returns [][]float64) *CorrelationResult {
	n := len(symbols)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		for j := range matrix[i] {
			if i == j {
				matrix[i][j] = 1.0
			} else if returns[i] != nil && returns[j] != nil && len(returns[i]) > 0 && len(returns[i]) == len(returns[j]) {
				matrix[i][j] = stat.Correlation(returns[i], returns[j], nil)
				// Clamp to [-1, 1] for floating-point edge cases
				if matrix[i][j] > 1.0 {
					matrix[i][j] = 1.0
				}
				if matrix[i][j] < -1.0 {
					matrix[i][j] = -1.0
				}
			}
		}
	}

	return &CorrelationResult{
		Symbols:    symbols,
		Matrix:     matrix,
		ComputedAt: time.Now(),
	}
}

// TopPairs extracts the most correlated pairs sorted by absolute correlation descending.
func TopPairs(result *CorrelationResult, limit int) []CorrelatedPair {
	var pairs []CorrelatedPair
	n := len(result.Symbols)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			pairs = append(pairs, CorrelatedPair{
				A:           result.Symbols[i],
				B:           result.Symbols[j],
				Correlation: math.Round(result.Matrix[i][j]*1000) / 1000,
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return math.Abs(pairs[i].Correlation) > math.Abs(pairs[j].Correlation)
	})

	if limit > 0 && len(pairs) > limit {
		pairs = pairs[:limit]
	}
	return pairs
}
