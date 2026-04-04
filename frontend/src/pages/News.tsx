import { useState, useEffect } from 'react';
import { api } from '../api/client';
import type { NewsItem, NewsFeed } from '../types';

type FilterCategory = 'all' | 'high-impact' | 'earnings' | 'macro' | 'sector' | 'crypto';

const FILTER_OPTIONS: { value: FilterCategory; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'high-impact', label: 'High Impact' },
  { value: 'earnings', label: 'Earnings' },
  { value: 'macro', label: 'Macro' },
  { value: 'sector', label: 'Sector' },
  { value: 'crypto', label: 'Crypto' },
];

function sentimentColor(sentiment: string): string {
  switch (sentiment) {
    case 'positive': return 'var(--accent-green)';
    case 'negative': return 'var(--accent-red)';
    default: return 'var(--text-muted)';
  }
}

function impactClass(impact: string): string {
  switch (impact) {
    case 'high': return 'impact-badge impact-high';
    case 'medium': return 'impact-badge impact-medium';
    case 'low': return 'impact-badge impact-low';
    default: return 'impact-badge';
  }
}

function categoryClass(category: string): string {
  switch (category) {
    case 'earnings': return 'badge badge-equity';
    case 'macro': return 'badge badge-high';
    case 'sector': return 'badge badge-crypto';
    case 'crypto': return 'badge badge-crypto';
    default: return 'badge';
  }
}

function timeAgo(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);

  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;

  const diffHours = Math.floor(diffMin / 60);
  if (diffHours < 24) return `${diffHours}h ago`;

  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}

export default function News() {
  const [feed, setFeed] = useState<NewsFeed | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<FilterCategory>('all');

  useEffect(() => {
    const fetchNews = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await api.getNews();
        setFeed(data);
      } catch (err: unknown) {
        if (err instanceof Error) setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetchNews();
  }, []);

  const filteredItems: NewsItem[] = feed?.items
    ? feed.items.filter(item => {
        if (filter === 'all') return true;
        if (filter === 'high-impact') return item.impact === 'high';
        return item.category === filter;
      })
    : [];

  const handleRefresh = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await api.getNews();
      setFeed(data);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div className="page-header animate-in">
        <div>
          <h1 className="page-title">Market News</h1>
          <p className="page-subtitle">Real-time news for your portfolio holdings</p>
        </div>
        <button
          className="btn btn-secondary"
          onClick={handleRefresh}
          disabled={loading}
        >
          {loading ? (
            <><span className="spinner" /> Refreshing...</>
          ) : (
            <>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 2v6h-6" /><path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
                <path d="M3 22v-6h6" /><path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
              </svg>
              Refresh
            </>
          )}
        </button>
      </div>

      {error && <div className="notice notice-error animate-in">{error}</div>}

      <div className="news-filter-tabs animate-in animate-in-1">
        {FILTER_OPTIONS.map(opt => (
          <button
            key={opt.value}
            className={`news-filter-tab ${filter === opt.value ? 'active' : ''}`}
            onClick={() => setFilter(opt.value)}
          >
            {opt.label}
          </button>
        ))}
        {feed?.holdings_analyzed && feed.holdings_analyzed.length > 0 && (
          <span className="news-holdings-count">
            Tracking {feed.holdings_analyzed.length} holding{feed.holdings_analyzed.length !== 1 ? 's' : ''}
          </span>
        )}
      </div>

      {loading && !feed && (
        <div className="card animate-in animate-in-2">
          <div className="empty-state">
            <span className="spinner" style={{ width: 24, height: 24, margin: '0 auto 1rem' }} />
            <p>Fetching latest news...</p>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)', marginTop: '0.35rem' }}>
              Loading articles for your portfolio holdings
            </p>
          </div>
        </div>
      )}

      {!loading && feed && feed.items.length === 0 && (
        <div className="card animate-in animate-in-2">
          <div className="empty-state">
            <div className="empty-state-icon">
              <NewspaperIcon />
            </div>
            <p style={{ fontWeight: 600, marginBottom: '0.35rem' }}>No portfolio data</p>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
              Sync your portfolio first to get market news for your holdings
            </p>
          </div>
        </div>
      )}

      {!loading && feed && feed.items.length > 0 && filteredItems.length === 0 && (
        <div className="card animate-in animate-in-2">
          <div className="empty-state">
            <p style={{ fontWeight: 600 }}>No items match this filter</p>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)', marginTop: '0.35rem' }}>
              Try a different category or view all news
            </p>
          </div>
        </div>
      )}

      <div className="news-grid">
        {filteredItems.map((item, index) => (
          <div
            key={item.id}
            className={`news-card animate-in animate-in-${Math.min(index + 2, 5)}`}
            style={{ borderLeftColor: sentimentColor(item.sentiment) }}
          >
            <div className="news-card-header">
              <div className="news-card-badges">
                <span className={impactClass(item.impact)}>
                  {item.impact.toUpperCase()}
                </span>
                <span className={categoryClass(item.category)}>
                  {item.category}
                </span>
              </div>
              <span className="news-card-time">{timeAgo(item.published_at)}</span>
            </div>

            {item.url ? (
              <a
                href={item.url}
                target="_blank"
                rel="noopener noreferrer"
                className="news-card-title news-card-link"
              >
                {item.title}
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginLeft: '0.35rem', flexShrink: 0, opacity: 0.5 }}>
                  <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                  <polyline points="15 3 21 3 21 9" />
                  <line x1="10" y1="14" x2="21" y2="3" />
                </svg>
              </a>
            ) : (
              <h3 className="news-card-title">{item.title}</h3>
            )}

            {item.summary && <p className="news-card-summary">{item.summary}</p>}

            <div className="news-card-footer">
              <div className="news-symbols">
                {item.symbols.map(sym => (
                  <span key={sym} className="news-symbol-pill">{sym}</span>
                ))}
              </div>
              <span className="news-card-source">{item.source}</span>
            </div>
          </div>
        ))}
      </div>

      {feed?.generated_at && (
        <div className="news-generated-at animate-in">
          Last updated: {new Date(feed.generated_at).toLocaleString()}
        </div>
      )}
    </div>
  );
}

function NewspaperIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2Zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2" />
      <path d="M18 14h-8" />
      <path d="M15 18h-5" />
      <path d="M10 6h8v4h-8V6Z" />
    </svg>
  );
}
