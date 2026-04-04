package schwab

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
		if r.URL.Path == "/accounts" {
			json.NewEncoder(w).Encode([]AccountResponse{
				{SecuritiesAccount: SecuritiesAccount{
					AccountNumber: "12345",
					Positions: []SchwabPosition{
						{
							Instrument:   Instrument{Symbol: "AAPL", AssetType: "EQUITY"},
							LongQuantity: 10,
							MarketValue:  1500.0,
							AveragePrice: 150.0,
						},
						{
							Instrument:   Instrument{Symbol: "MSFT", AssetType: "EQUITY"},
							LongQuantity: 5,
							MarketValue:  2000.0,
							AveragePrice: 400.0,
						},
					},
				}},
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	c := &Client{
		baseURL:     server.URL,
		accessToken: "test-token",
		httpClient:  server.Client(),
	}

	positions, err := c.FetchPositions(context.Background())
	require.NoError(t, err)
	assert.Len(t, positions, 2)
	assert.Equal(t, "AAPL", positions[0].Symbol)
	assert.Equal(t, float64(10), positions[0].Quantity)
	assert.Equal(t, 1500.0, positions[0].MarketValue)
	assert.Equal(t, "schwab", positions[0].Source)
	assert.Equal(t, "equity", positions[0].AssetType)
}
