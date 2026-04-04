package schwab

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"flowpilot/internal/connectors"
	"flowpilot/internal/models"
)

var _ connectors.Connector = (*Client)(nil)

const defaultBaseURL = "https://api.schwabapi.com/trader/v1"
const authBaseURL = "https://api.schwabapi.com/v1"

type TokenSaver func(access, refresh string)

type Client struct {
	baseURL      string
	accessToken  string
	refreshToken string
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
	mu           sync.RWMutex
	onTokenSave  TokenSaver
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

func (c *Client) SetTokenSaver(fn TokenSaver) {
	c.onTokenSave = fn
}

func (c *Client) Name() string { return "schwab" }

func (c *Client) SetTokens(access, refresh string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = access
	c.refreshToken = refresh
}

func (c *Client) HasTokens() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken != ""
}

func (c *Client) AuthURL() string {
	return fmt.Sprintf("https://api.schwabapi.com/v1/oauth/authorize?client_id=%s&redirect_uri=%s",
		url.QueryEscape(c.clientID),
		url.QueryEscape(c.redirectURI))
}

func (c *Client) ExchangeCode(ctx context.Context, code string) error {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.redirectURI},
	}

	return c.doTokenRequest(ctx, data)
}

func (c *Client) RefreshAccessToken(ctx context.Context) error {
	c.mu.RLock()
	rt := c.refreshToken
	c.mu.RUnlock()

	if rt == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {rt},
	}

	return c.doTokenRequest(ctx, data)
}

func (c *Client) doTokenRequest(ctx context.Context, data url.Values) error {
	req, err := http.NewRequestWithContext(ctx, "POST", authBaseURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Schwab uses Basic auth with clientID:clientSecret
	creds := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+creds)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("schwab token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("schwab token exchange returned %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("schwab token decode: %w", err)
	}

	c.mu.Lock()
	c.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken
	}
	access := c.accessToken
	refresh := c.refreshToken
	c.mu.Unlock()

	if c.onTokenSave != nil {
		c.onTokenSave(access, refresh)
	}

	return nil
}

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
	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	if token == "" {
		return nil, fmt.Errorf("schwab not authenticated — visit /api/auth/schwab to connect")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/accounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.URL.RawQuery = "fields=positions"

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("schwab fetch accounts: %w", err)
	}
	defer resp.Body.Close()

	// If 401, try refresh
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.RefreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf("schwab token expired, re-authorize at /api/auth/schwab: %w", err)
		}
		// Retry with new token
		return c.FetchPositions(ctx)
	}

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
