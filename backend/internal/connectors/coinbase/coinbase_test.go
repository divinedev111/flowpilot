package coinbase

import (
	"context"
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
			json.NewEncoder(w).Encode(AccountsResponse{
				Data: []Account{
					{
						ID:       "acct-1",
						Currency: Currency{Code: "BTC", Name: "Bitcoin"},
						Balance:  Balance{Amount: "1.5", Currency: "BTC"},
					},
					{
						ID:       "acct-2",
						Currency: Currency{Code: "ETH", Name: "Ethereum"},
						Balance:  Balance{Amount: "10.0", Currency: "ETH"},
					},
				},
			})
			return
		}
		if r.URL.Path == "/v2/prices/BTC-USD/spot" {
			json.NewEncoder(w).Encode(PriceResponse{Data: PriceData{Amount: "60000.00"}})
			return
		}
		if r.URL.Path == "/v2/prices/ETH-USD/spot" {
			json.NewEncoder(w).Encode(PriceResponse{Data: PriceData{Amount: "3000.00"}})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		apiKey:     "test-key",
		apiSecret:  "test-secret",
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
