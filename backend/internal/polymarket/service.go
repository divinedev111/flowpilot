package polymarket

import (
	"context"
	"log"
	"sort"
	"time"

	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const cacheTTL = 30 * time.Minute

type Service struct {
	client *Client
	db     *db.MongoDB
}

func NewService(db *db.MongoDB) *Service {
	return &Service{
		client: NewClient(),
		db:     db,
	}
}

// FetchRelevantMarkets returns prediction markets relevant to the given
// portfolio symbols. Results are cached in MongoDB for 30 minutes.
func (s *Service) FetchRelevantMarkets(ctx context.Context, symbols []string) ([]models.PredictionMarket, error) {
	// Try cached results first.
	cached, err := s.getCached(ctx)
	if err == nil && len(cached) > 0 {
		return filterBySymbols(cached, symbols), nil
	}

	// Fetch fresh data from Polymarket.
	markets, err := s.client.FetchMarkets(ctx)
	if err != nil {
		log.Printf("polymarket: failed to fetch markets: %v", err)
	}

	// Also fetch events (which embed markets).
	events, err := s.client.FetchEvents(ctx)
	if err != nil {
		log.Printf("polymarket: failed to fetch events: %v", err)
	}

	// Collect all markets from both endpoints.
	var allMarkets []gammaMarket
	allMarkets = append(allMarkets, markets...)
	for _, ev := range events {
		allMarkets = append(allMarkets, ev.Markets...)
	}

	if len(allMarkets) == 0 {
		return nil, nil
	}

	// Match markets against ALL known keywords (not just current symbols)
	// so the cache is broadly useful.
	allSymbols := allKnownSymbols()
	matched := MatchMarkets(allMarkets, allSymbols)

	// Cache results.
	if err := s.cacheResults(ctx, matched); err != nil {
		log.Printf("polymarket: failed to cache results: %v", err)
	}

	// Filter for the requested symbols.
	relevant := filterBySymbols(matched, symbols)

	// Sort by volume descending.
	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].Volume > relevant[j].Volume
	})

	return relevant, nil
}

func (s *Service) getCached(ctx context.Context) ([]models.PredictionMarket, error) {
	cutoff := time.Now().Add(-cacheTTL)

	filter := bson.D{{Key: "last_synced", Value: bson.D{{Key: "$gte", Value: cutoff}}}}
	cursor, err := s.db.PredictionMarkets().Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var markets []models.PredictionMarket
	if err := cursor.All(ctx, &markets); err != nil {
		return nil, err
	}

	if len(markets) == 0 {
		return nil, nil
	}

	return markets, nil
}

func (s *Service) cacheResults(ctx context.Context, markets []models.PredictionMarket) error {
	coll := s.db.PredictionMarkets()

	for _, m := range markets {
		filter := bson.D{{Key: "_id", Value: m.ID}}
		update := bson.D{{Key: "$set", Value: m}}
		opts := options.UpdateOne().SetUpsert(true)

		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return err
		}
	}
	return nil
}

// filterBySymbols returns only markets whose RelatedSymbols overlap with the
// requested symbols, or that are categorized as "macro" (always relevant).
func filterBySymbols(markets []models.PredictionMarket, symbols []string) []models.PredictionMarket {
	symbolSet := make(map[string]bool)
	for _, s := range symbols {
		symbolSet[s] = true
	}

	var result []models.PredictionMarket
	for _, m := range markets {
		// Always include macro markets.
		if m.Category == "macro" {
			result = append(result, m)
			continue
		}
		for _, rs := range m.RelatedSymbols {
			if symbolSet[rs] {
				result = append(result, m)
				break
			}
		}
	}

	// Sort by volume descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Volume > result[j].Volume
	})

	return result
}

func allKnownSymbols() []string {
	var syms []string
	for s := range symbolKeywords {
		syms = append(syms, s)
	}
	return syms
}
