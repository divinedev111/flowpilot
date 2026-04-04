package correlation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"flowpilot/internal/db"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PriceCacheEntry struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Symbol    string        `bson:"symbol"`
	Prices    []DayPrice    `bson:"prices"`
	FetchedAt time.Time     `bson:"fetched_at"`
}

type DayPrice struct {
	Date  string  `bson:"date"`
	Close float64 `bson:"close"`
}

type PriceService struct {
	db     *db.MongoDB
	client *http.Client
}

func NewPriceService(database *db.MongoDB) *PriceService {
	return &PriceService{
		db:     database,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetDailyPrices returns up to 120 trading days of close prices.
// isCrypto determines whether to use Binance (true) or Yahoo Finance (false).
func (s *PriceService) GetDailyPrices(ctx context.Context, symbol string, isCrypto bool) ([]DayPrice, error) {
	cached, err := s.getCached(ctx, symbol)
	if err == nil && cached != nil {
		return cached, nil
	}

	var prices []DayPrice
	if isCrypto {
		prices, err = s.fetchFromBinance(ctx, symbol)
	} else {
		prices, err = s.fetchFromYahoo(ctx, symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("fetch prices for %s: %w", symbol, err)
	}

	if cacheErr := s.cacheResult(ctx, symbol, prices); cacheErr != nil {
		_ = cacheErr
	}

	return prices, nil
}

func (s *PriceService) getCached(ctx context.Context, symbol string) ([]DayPrice, error) {
	var entry PriceCacheEntry
	filter := bson.D{
		{Key: "symbol", Value: symbol},
		{Key: "fetched_at", Value: bson.D{{Key: "$gt", Value: time.Now().Add(-24 * time.Hour)}}},
	}
	err := s.db.PriceCache().FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		return nil, err
	}
	return entry.Prices, nil
}

func (s *PriceService) cacheResult(ctx context.Context, symbol string, prices []DayPrice) error {
	filter := bson.D{{Key: "symbol", Value: symbol}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "symbol", Value: symbol},
		{Key: "prices", Value: prices},
		{Key: "fetched_at", Value: time.Now()},
	}}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := s.db.PriceCache().UpdateOne(ctx, filter, update, opts)
	return err
}

// fetchFromBinance fetches daily klines from the Binance public API.
// Returns up to 120 days of daily close prices. No auth required.
func (s *PriceService) fetchFromBinance(_ context.Context, symbol string) ([]DayPrice, error) {
	// Binance uses pairs like BTCUSDT, ETHUSDT
	pair := symbol + "USDT"
	url := fmt.Sprintf(
		"https://api.binance.com/api/v3/klines?symbol=%s&interval=1d&limit=120",
		pair,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance error %d: %s", resp.StatusCode, string(body))
	}

	// Response is array of arrays: [[openTime, open, high, low, close, volume, closeTime, ...], ...]
	var klines [][]json.RawMessage
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	prices := make([]DayPrice, 0, len(klines))
	for _, k := range klines {
		if len(k) < 5 {
			continue
		}
		// Index 0: open time (ms), Index 4: close price (string)
		var openTimeMs int64
		if err := json.Unmarshal(k[0], &openTimeMs); err != nil {
			continue
		}
		var closeStr string
		if err := json.Unmarshal(k[4], &closeStr); err != nil {
			continue
		}
		closeVal, err := strconv.ParseFloat(closeStr, 64)
		if err != nil {
			continue
		}
		date := time.UnixMilli(openTimeMs).UTC().Format("2006-01-02")
		prices = append(prices, DayPrice{Date: date, Close: closeVal})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date < prices[j].Date
	})

	return prices, nil
}

// yahooChartResponse represents the nested JSON structure from Yahoo Finance chart endpoint.
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// fetchFromYahoo fetches daily prices from Yahoo Finance chart endpoint.
// Returns up to 6 months of daily close prices. No auth required.
func (s *PriceService) fetchFromYahoo(_ context.Context, symbol string) ([]DayPrice, error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?range=6mo&interval=1d",
		symbol,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo error %d: %s", resp.StatusCode, string(body))
	}

	var yResp yahooChartResponse
	if err := json.Unmarshal(body, &yResp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if yResp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error: %s", yResp.Chart.Error.Description)
	}

	if len(yResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	result := yResp.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data for %s", symbol)
	}

	timestamps := result.Timestamp
	closes := result.Indicators.Quote[0].Close

	n := len(timestamps)
	if len(closes) < n {
		n = len(closes)
	}

	prices := make([]DayPrice, 0, n)
	for i := 0; i < n; i++ {
		if closes[i] == 0 {
			continue // skip null/zero entries
		}
		date := time.Unix(timestamps[i], 0).UTC().Format("2006-01-02")
		prices = append(prices, DayPrice{Date: date, Close: closes[i]})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date < prices[j].Date
	})

	if len(prices) > 120 {
		prices = prices[len(prices)-120:]
	}

	return prices, nil
}
