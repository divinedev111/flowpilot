package news

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"flowpilot/internal/ai"
	"flowpilot/internal/db"
	"flowpilot/internal/models"

	"github.com/anthropics/anthropic-sdk-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	cacheTTL         = 30 * time.Minute
	maxSymbols       = 5
	articlesPerSymbol = 3
	httpTimeout      = 10 * time.Second
)

// Known crypto symbols for query classification.
var cryptoSymbols = map[string]bool{
	"BTC": true, "ETH": true, "SOL": true, "ADA": true,
	"DOGE": true, "DOT": true, "AVAX": true, "MATIC": true,
	"LINK": true, "UNI": true, "XRP": true, "LTC": true,
	"SHIB": true, "ATOM": true, "NEAR": true, "APT": true,
}

// RSS XML structures for Google News feed parsing.
type rssResponse struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title   string    `xml:"title"`
	Link    string    `xml:"link"`
	PubDate string    `xml:"pubDate"`
	Source  rssSource `xml:"source"`
}

type rssSource struct {
	Name string `xml:",chardata"`
	URL  string `xml:"url,attr"`
}

// Service fetches real news from Google News RSS and optionally scores
// sentiment via AI. Results are cached in MongoDB.
type Service struct {
	ai   *ai.Client
	db   *db.MongoDB
	http *http.Client
}

func NewService(aiClient *ai.Client, database *db.MongoDB) *Service {
	return &Service{
		ai: aiClient,
		db: database,
		http: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// GetFeed returns a news feed for the given symbols by fetching real articles
// from Google News RSS. Results are cached for 30 minutes.
func (s *Service) GetFeed(ctx context.Context, symbols []string) (*models.NewsFeed, error) {
	// Check cache first
	cached, err := s.getCached(ctx, symbols)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Fetch real news from RSS
	feed, err := s.fetchRSSFeed(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("fetch news feed: %w", err)
	}

	// Optionally score sentiment/impact via AI
	if s.ai != nil && len(feed.Items) > 0 {
		s.scoreSentiment(ctx, feed)
	}

	// Cache the result
	_ = s.cacheFeed(ctx, feed)

	return feed, nil
}

func (s *Service) GetSymbolNews(ctx context.Context, symbol string) (*models.NewsFeed, error) {
	return s.GetFeed(ctx, []string{symbol})
}

// fetchRSSFeed fetches real articles from Google News RSS for each symbol,
// deduplicates by title, and sorts by publish date descending.
func (s *Service) fetchRSSFeed(ctx context.Context, symbols []string) (*models.NewsFeed, error) {
	// Limit symbols to avoid too many requests
	if len(symbols) > maxSymbols {
		symbols = symbols[:maxSymbols]
	}

	seen := make(map[string]bool)     // deduplicate by normalized title
	var allItems []models.NewsItem
	counter := 0

	for _, symbol := range symbols {
		items, err := s.fetchSymbolRSS(ctx, symbol)
		if err != nil {
			// Log and continue - don't fail the whole feed for one symbol
			continue
		}

		for _, item := range items {
			normalizedTitle := strings.ToLower(strings.TrimSpace(item.Title))
			if seen[normalizedTitle] {
				// Already have this article from another symbol; add this symbol to it
				for i := range allItems {
					if strings.ToLower(strings.TrimSpace(allItems[i].Title)) == normalizedTitle {
						allItems[i].Symbols = appendUnique(allItems[i].Symbols, formatSymbolTag(symbol))
						break
					}
				}
				continue
			}
			seen[normalizedTitle] = true

			counter++
			item.ID = fmt.Sprintf("news-%d", counter)
			item.Symbols = []string{formatSymbolTag(symbol)}
			item.Sentiment = "neutral"
			item.Impact = "medium"
			item.Relevance = 0.5
			item.Category = classifyCategory(symbol)

			allItems = append(allItems, item)
		}
	}

	// Sort by publish date descending
	sortByDate(allItems)

	now := time.Now()
	feed := &models.NewsFeed{
		Items:            allItems,
		GeneratedAt:      now,
		HoldingsAnalyzed: symbols,
	}

	return feed, nil
}

// fetchSymbolRSS fetches up to articlesPerSymbol articles from Google News RSS
// for a single symbol or topic query.
func (s *Service) fetchSymbolRSS(ctx context.Context, symbol string) ([]models.NewsItem, error) {
	upper := strings.ToUpper(symbol)
	var query string
	if len(symbol) > 6 || strings.Contains(symbol, " ") {
		// Looks like a topic/phrase (from prediction markets), use as-is
		query = strings.ReplaceAll(symbol, " ", "+")
	} else if isCrypto(upper) {
		query = upper + "+crypto"
	} else {
		query = upper + "+stock"
	}

	url := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "FlowPilot/1.0")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch RSS for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS returned status %d for %s", resp.StatusCode, symbol)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read RSS body: %w", err)
	}

	var rss rssResponse
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("parse RSS XML: %w", err)
	}

	var items []models.NewsItem
	for i, rssItem := range rss.Channel.Items {
		if i >= articlesPerSymbol {
			break
		}

		title, source := parseRSSTitle(rssItem.Title)
		if rssItem.Source.Name != "" {
			source = rssItem.Source.Name
		}

		pubTime := parseRSSDate(rssItem.PubDate)

		items = append(items, models.NewsItem{
			ID:          hashID(rssItem.Link),
			Title:       title,
			Summary:     "", // No summary from RSS; could be enriched later
			Source:      source,
			URL:         rssItem.Link,
			PublishedAt: pubTime,
		})
	}

	return items, nil
}

// scoreSentiment uses AI to batch-score sentiment and impact for all articles.
func (s *Service) scoreSentiment(ctx context.Context, feed *models.NewsFeed) {
	if s.ai == nil || len(feed.Items) == 0 {
		return
	}

	// Build a compact list for the AI to score
	type articleRef struct {
		Index int    `json:"index"`
		Title string `json:"title"`
	}
	refs := make([]articleRef, len(feed.Items))
	for i, item := range feed.Items {
		refs[i] = articleRef{Index: i, Title: item.Title}
	}

	refsJSON, err := json.Marshal(refs)
	if err != nil {
		return
	}

	prompt := fmt.Sprintf(`Score the sentiment and financial impact of each news headline below.

For each article, respond with a JSON array where each element has:
- "index": the article index number
- "sentiment": one of "positive", "negative", "neutral"
- "impact": one of "high", "medium", "low"

Respond with ONLY the JSON array, no other text.

Articles:
%s`, string(refsJSON))

	message, err := s.ai.GetClient().Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return // silently fail - defaults are already set
	}

	var rawJSON string
	for _, block := range message.Content {
		if block.Type == "text" {
			rawJSON += block.Text
		}
	}

	// Parse AI scores
	type scoreResult struct {
		Index     int    `json:"index"`
		Sentiment string `json:"sentiment"`
		Impact    string `json:"impact"`
	}

	rawJSON = strings.TrimSpace(rawJSON)
	// Strip markdown code fences if present
	if strings.HasPrefix(rawJSON, "```") {
		lines := strings.Split(rawJSON, "\n")
		if len(lines) > 2 {
			lines = lines[1 : len(lines)-1]
			rawJSON = strings.Join(lines, "\n")
		}
	}
	start := strings.Index(rawJSON, "[")
	end := strings.LastIndex(rawJSON, "]")
	if start >= 0 && end > start {
		rawJSON = rawJSON[start : end+1]
	}

	var scores []scoreResult
	if err := json.Unmarshal([]byte(rawJSON), &scores); err != nil {
		return
	}

	for _, sc := range scores {
		if sc.Index >= 0 && sc.Index < len(feed.Items) {
			if sc.Sentiment == "positive" || sc.Sentiment == "negative" || sc.Sentiment == "neutral" {
				feed.Items[sc.Index].Sentiment = sc.Sentiment
			}
			if sc.Impact == "high" || sc.Impact == "medium" || sc.Impact == "low" {
				feed.Items[sc.Index].Impact = sc.Impact
			}
		}
	}
}

func (s *Service) getCached(ctx context.Context, symbols []string) (*models.NewsFeed, error) {
	if s.db == nil {
		return nil, fmt.Errorf("no database")
	}

	cutoff := time.Now().Add(-cacheTTL)

	// Sort symbols for consistent cache keys
	sorted := make([]string, len(symbols))
	copy(sorted, symbols)
	sortStrings(sorted)

	filter := bson.D{
		{Key: "holdings_analyzed", Value: sorted},
		{Key: "generated_at", Value: bson.D{{Key: "$gte", Value: cutoff}}},
	}

	var feed models.NewsFeed
	err := s.db.NewsCache().FindOne(ctx, filter).Decode(&feed)
	if err != nil {
		return nil, err
	}
	// Reject old AI-generated cache entries that have no URLs
	if len(feed.Items) > 0 && feed.Items[0].URL == "" {
		return nil, fmt.Errorf("stale cache: missing URLs")
	}
	return &feed, nil
}

func (s *Service) cacheFeed(ctx context.Context, feed *models.NewsFeed) error {
	if s.db == nil {
		return nil
	}

	sorted := make([]string, len(feed.HoldingsAnalyzed))
	copy(sorted, feed.HoldingsAnalyzed)
	sortStrings(sorted)
	feed.HoldingsAnalyzed = sorted

	filter := bson.D{
		{Key: "holdings_analyzed", Value: sorted},
	}
	update := bson.D{{Key: "$set", Value: feed}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := s.db.NewsCache().UpdateOne(ctx, filter, update, opts)
	return err
}

// --- Helpers ---

// parseRSSTitle splits "Article Title - Source Name" into (title, source).
func parseRSSTitle(raw string) (title, source string) {
	idx := strings.LastIndex(raw, " - ")
	if idx > 0 {
		return strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+3:])
	}
	return raw, "Unknown"
}

// parseRSSDate parses the RFC1123-style date used in RSS pubDate fields.
func parseRSSDate(dateStr string) time.Time {
	// Google News RSS uses RFC1123 / RFC2822 style dates
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t
		}
	}
	return time.Now()
}

func isCrypto(symbol string) bool {
	upper := strings.ToUpper(symbol)
	if cryptoSymbols[upper] {
		return true
	}
	return strings.HasSuffix(upper, "-USD")
}

func classifyCategory(symbol string) string {
	if isCrypto(strings.ToUpper(symbol)) {
		return "crypto"
	}
	// If it looks like a topic phrase (from prediction markets), classify appropriately
	if len(symbol) > 6 || strings.Contains(symbol, " ") {
		lower := strings.ToLower(symbol)
		if strings.Contains(lower, "earn") || strings.Contains(lower, "revenue") || strings.Contains(lower, "beat") {
			return "earnings"
		}
		if strings.Contains(lower, "senate") || strings.Contains(lower, "democrat") || strings.Contains(lower, "republican") ||
			strings.Contains(lower, "trump") || strings.Contains(lower, "election") || strings.Contains(lower, "vote") {
			return "macro"
		}
		return "sector"
	}
	return "sector"
}

func hashID(url string) string {
	h := sha256.Sum256([]byte(url))
	return fmt.Sprintf("news-%x", h[:6])
}

// formatSymbolTag formats a symbol for display. Short tickers stay uppercase,
// but long topic phrases (from prediction markets) get truncated and tagged.
func formatSymbolTag(symbol string) string {
	if len(symbol) <= 6 && !strings.Contains(symbol, " ") {
		return strings.ToUpper(symbol)
	}
	// It's a prediction market topic — extract a short label
	s := symbol
	if len(s) > 25 {
		s = s[:25] + "..."
	}
	return s
}

func appendUnique(slice []string, val string) []string {
	upper := strings.ToUpper(val)
	for _, s := range slice {
		if strings.ToUpper(s) == upper {
			return slice
		}
	}
	return append(slice, strings.ToUpper(val))
}

func sortByDate(items []models.NewsItem) {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 && items[j].PublishedAt.Before(key.PublishedAt) {
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
