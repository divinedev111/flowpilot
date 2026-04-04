package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"flowpilot/internal/models"
)

const (
	baseURL       = "https://gamma-api.polymarket.com"
	requestTimeout = 15 * time.Second
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

type gammaEvent struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	EndDate     string        `json:"endDate"`
	Active      bool          `json:"active"`
	Closed      bool          `json:"closed"`
	Markets     []gammaMarket `json:"markets"`
}

type gammaMarket struct {
	ID              string `json:"id"`
	Question        string `json:"question"`
	Description     string `json:"description"`
	OutcomePrices   string `json:"outcomePrices"` // JSON string: "[\"0.72\",\"0.28\"]"
	Volume          string `json:"volume"`
	VolumeNum       float64 `json:"volumeNum"`
	EndDate         string `json:"endDate"`
	Active          bool   `json:"active"`
	Closed          bool   `json:"closed"`
	Slug            string `json:"slug"`
}

// symbolKeywords maps ticker symbols to search terms used when querying
// the Polymarket API for relevant prediction markets.
var symbolKeywords = map[string][]string{
	// Equities
	"NVDA": {"Nvidia", "NVDA"},
	"AAPL": {"Apple", "AAPL"},
	"MSFT": {"Microsoft", "MSFT"},
	"GOOGL": {"Google", "Alphabet", "GOOGL"},
	"GOOG": {"Google", "Alphabet", "GOOG"},
	"AMZN": {"Amazon", "AMZN"},
	"META": {"Meta", "Facebook", "META"},
	"TSLA": {"Tesla", "TSLA"},
	"AMD":  {"AMD"},
	"INTC": {"Intel", "INTC"},

	// Crypto
	"BTC":  {"Bitcoin", "BTC"},
	"ETH":  {"Ethereum", "ETH"},
	"SOL":  {"Solana", "SOL"},
	"DOGE": {"Dogecoin", "DOGE"},
	"XRP":  {"Ripple", "XRP"},
	"ADA":  {"Cardano", "ADA"},
	"AVAX": {"Avalanche", "AVAX"},
	"MATIC": {"Polygon", "MATIC"},

	// Indices / broad
	"SPY":  {"S&P 500", "S&P", "stock market"},
	"QQQ":  {"Nasdaq", "tech stocks"},
	"DIA":  {"Dow Jones", "Dow"},
	"IWM":  {"Russell 2000"},
	"VTI":  {"stock market"},
}

// generalKeywords are always searched to capture broad macro markets.
var generalKeywords = []string{
	"Fed rate",
	"interest rate",
	"inflation",
	"recession",
	"GDP",
	"S&P 500",
	"stock market",
	"crypto",
	"Bitcoin",
}

func (c *Client) FetchEvents(ctx context.Context) ([]gammaEvent, error) {
	u, _ := url.Parse(baseURL + "/events")
	q := u.Query()
	q.Set("closed", "false")
	q.Set("limit", "100")
	q.Set("active", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("polymarket: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("polymarket: fetch events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("polymarket: events API status %d: %s", resp.StatusCode, string(body))
	}

	var events []gammaEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("polymarket: decode events: %w", err)
	}
	return events, nil
}

func (c *Client) FetchMarkets(ctx context.Context) ([]gammaMarket, error) {
	u, _ := url.Parse(baseURL + "/markets")
	q := u.Query()
	q.Set("closed", "false")
	q.Set("active", "true")
	q.Set("limit", "100")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("polymarket: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("polymarket: fetch markets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("polymarket: markets API status %d: %s", resp.StatusCode, string(body))
	}

	var markets []gammaMarket
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, fmt.Errorf("polymarket: decode markets: %w", err)
	}
	return markets, nil
}

// MatchMarkets filters fetched markets against the given portfolio symbols,
// returning PredictionMarket models for markets whose title or description
// matches any keyword associated with those symbols.
func MatchMarkets(markets []gammaMarket, symbols []string) []models.PredictionMarket {
	// Build keyword → symbol(s) mapping.
	keywordToSymbols := make(map[string][]string)
	for _, sym := range symbols {
		upper := strings.ToUpper(sym)
		if kws, ok := symbolKeywords[upper]; ok {
			for _, kw := range kws {
				keywordToSymbols[strings.ToLower(kw)] = appendUnique(keywordToSymbols[strings.ToLower(kw)], upper)
			}
		}
	}
	// Always include general macro keywords (attributed to "MACRO").
	for _, kw := range generalKeywords {
		keywordToSymbols[strings.ToLower(kw)] = appendUnique(keywordToSymbols[strings.ToLower(kw)], "MACRO")
	}

	seen := make(map[string]bool)
	var results []models.PredictionMarket

	for _, m := range markets {
		titleLower := strings.ToLower(m.Question)
		descLower := strings.ToLower(m.Description)

		var relatedSymbols []string
		for kw, syms := range keywordToSymbols {
			if strings.Contains(titleLower, kw) || strings.Contains(descLower, kw) {
				for _, s := range syms {
					relatedSymbols = appendUnique(relatedSymbols, s)
				}
			}
		}

		if len(relatedSymbols) == 0 {
			continue
		}
		if seen[m.ID] {
			continue
		}
		seen[m.ID] = true

		yesPrice, noPrice := parseOutcomePrices(m.OutcomePrices)
		volume := m.VolumeNum
		if volume == 0 {
			volume, _ = strconv.ParseFloat(m.Volume, 64)
		}

		category := categorize(relatedSymbols)

		pm := models.PredictionMarket{
			ID:             m.ID,
			Source:         "polymarket",
			Title:          m.Question,
			Description:    m.Description,
			OutcomeYes:     yesPrice,
			OutcomeNo:      noPrice,
			Volume:         volume,
			EndDate:        m.EndDate,
			Category:       category,
			RelatedSymbols: relatedSymbols,
			URL:            fmt.Sprintf("https://polymarket.com/event/%s", m.Slug),
			LastSynced:     time.Now(),
		}
		results = append(results, pm)
	}

	return results
}

// parseOutcomePrices extracts the Yes and No probabilities from the Gamma API's
// outcomePrices field, which is a JSON string like `"[\"0.72\",\"0.28\"]"`.
func parseOutcomePrices(raw string) (yes, no float64) {
	if raw == "" {
		return 0, 0
	}

	var prices []string
	if err := json.Unmarshal([]byte(raw), &prices); err != nil {
		return 0, 0
	}

	if len(prices) >= 1 {
		yes, _ = strconv.ParseFloat(prices[0], 64)
	}
	if len(prices) >= 2 {
		no, _ = strconv.ParseFloat(prices[1], 64)
	}
	return yes, no
}

func categorize(symbols []string) string {
	cryptoSyms := map[string]bool{
		"BTC": true, "ETH": true, "SOL": true, "DOGE": true,
		"XRP": true, "ADA": true, "AVAX": true, "MATIC": true,
	}
	techSyms := map[string]bool{
		"NVDA": true, "AAPL": true, "MSFT": true, "GOOGL": true,
		"GOOG": true, "AMZN": true, "META": true, "TSLA": true,
		"AMD": true, "INTC": true, "QQQ": true,
	}

	hasCrypto, hasTech, hasMacro := false, false, false
	for _, s := range symbols {
		if cryptoSyms[s] {
			hasCrypto = true
		} else if techSyms[s] {
			hasTech = true
		} else {
			hasMacro = true
		}
	}

	switch {
	case hasCrypto && !hasTech && !hasMacro:
		return "crypto"
	case hasTech && !hasCrypto && !hasMacro:
		return "tech"
	case hasMacro && !hasCrypto && !hasTech:
		return "macro"
	default:
		return "macro"
	}
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
