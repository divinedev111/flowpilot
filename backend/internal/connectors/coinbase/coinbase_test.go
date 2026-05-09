package coinbase

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchPositions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/accounts" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id": "acct-1",
						"currency": map[string]interface{}{
							"code": "BTC",
							"name": "Bitcoin",
						},
						"balance": map[string]interface{}{
							"amount":   "1.5",
							"currency": "BTC",
						},
					},
					{
						"id": "acct-2",
						"currency": map[string]interface{}{
							"code": "ETH",
							"name": "Ethereum",
						},
						"balance": map[string]interface{}{
							"amount":   "10.0",
							"currency": "ETH",
						},
					},
				},
			})
			return
		}
		if r.URL.Path == "/v2/prices/BTC-USD/spot" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"amount": "60000.00",
				},
			})
			return
		}
		if r.URL.Path == "/v2/prices/ETH-USD/spot" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"amount": "3000.00",
				},
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	c := &Client{
		baseURL:    server.URL,
		apiKeyID:   "test-key-id",
		privateKey: privKey,
		httpClient: server.Client(),
	}

	positions, err := c.FetchPositions(context.Background())
	require.NoError(t, err)
	assert.Len(t, positions, 2)
	assert.Equal(t, "BTC", positions[0].Symbol)
	assert.Equal(t, 1.5, positions[0].Quantity)
	assert.Equal(t, 60000.0, positions[0].MarkPrice)
	assert.Equal(t, 90000.0, positions[0].MarketValue)
	assert.Equal(t, "crypto", positions[0].AssetType)
	assert.Equal(t, "coinbase", positions[0].Source)
}
