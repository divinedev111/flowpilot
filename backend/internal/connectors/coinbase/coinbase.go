package coinbase

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flowpilot/internal/connectors"
	"flowpilot/internal/models"
)

const defaultBaseURL = "https://api.coinbase.com"

var _ connectors.Connector = (*Client)(nil)

type Client struct {
	baseURL    string
	apiKeyID   string
	privateKey ed25519.PrivateKey
	httpClient *http.Client
}

func NewClient(apiKeyID, privateKeyB64 string) *Client {
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		log.Printf("coinbase: failed to decode private key: %v", err)
		return &Client{baseURL: defaultBaseURL, apiKeyID: apiKeyID, httpClient: &http.Client{Timeout: 30 * time.Second}}
	}

	var privKey ed25519.PrivateKey
	if len(keyBytes) == ed25519.PrivateKeySize {
		privKey = ed25519.PrivateKey(keyBytes)
	} else if len(keyBytes) == ed25519.SeedSize {
		privKey = ed25519.NewKeyFromSeed(keyBytes)
	} else {
		log.Printf("coinbase: unexpected key size %d (expected %d or %d)", len(keyBytes), ed25519.SeedSize, ed25519.PrivateKeySize)
	}

	return &Client{
		baseURL:    defaultBaseURL,
		apiKeyID:   apiKeyID,
		privateKey: privKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "coinbase" }

type AccountsResponse struct {
	Data []struct {
		ID       string `json:"id"`
		Currency struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"currency"`
		Balance struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"balance"`
	} `json:"data"`
}

type PriceResponse struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

func (c *Client) buildJWT(method, host, path string) (string, error) {
	if c.privateKey == nil {
		return "", fmt.Errorf("no private key configured")
	}

	now := time.Now().Unix()

	// Random nonce
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)

	uri := method + " " + host + path

	header := map[string]string{
		"alg":   "EdDSA",
		"kid":   c.apiKeyID,
		"typ":   "JWT",
		"nonce": nonce,
	}

	payload := map[string]interface{}{
		"sub":  c.apiKeyID,
		"iss":  "cdp",
		"aud":  []string{"cdp_service"},
		"nbf":  now,
		"exp":  now + 120,
		"uris": []string{uri},
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64URLEncode(headerJSON)
	payloadB64 := base64URLEncode(payloadJSON)

	signingInput := headerB64 + "." + payloadB64
	sig := ed25519.Sign(c.privateKey, []byte(signingInput))
	sigB64 := base64URLEncode(sig)

	return signingInput + "." + sigB64, nil
}

func base64URLEncode(data []byte) string {
	s := base64.StdEncoding.EncodeToString(data)
	s = strings.TrimRight(s, "=")
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func (c *Client) doGet(ctx context.Context, path string, out interface{}) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	jwt, err := c.buildJWT("GET", "api.coinbase.com", path)
	if err != nil {
		return fmt.Errorf("coinbase jwt: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("coinbase %s %d: %s", path, resp.StatusCode, string(body))
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
