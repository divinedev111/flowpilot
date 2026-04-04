import { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { api } from '../api/client';
import type { Insight, InsightsResult, AIDigest } from '../types';

export default function InsightsPanel() {
  const [insights, setInsights] = useState<InsightsResult | null>(null);
  const [digest, setDigest] = useState<AIDigest | null>(null);
  const [loading, setLoading] = useState(false);
  const [digestExpanded, setDigestExpanded] = useState(false);
  const [insightsExpanded, setInsightsExpanded] = useState(false);
  const [error, setError] = useState('');

  const fetchDigest = async () => {
    try {
      const data = await api.getLatestDigest();
      setDigest(data.digest);
    } catch {
      // digest is non-critical
    }
  };

  const fetchInsights = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await api.getInsights();
      setInsights(data);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDigest();
    fetchInsights();
  }, []);

  const handleRefresh = () => {
    fetchDigest();
    fetchInsights();
  };

  const getDigestPreview = (summary: string) => {
    const lines = summary.split('\n').filter(l => l.trim() && !l.trim().startsWith('#'));
    const preview = lines.slice(0, 3).join(' ').replace(/\*\*/g, '').replace(/\*/g, '');
    return preview.length > 200 ? preview.slice(0, 200) + '...' : preview;
  };

  return (
    <div className="insights-panel">
      {digest && (
        <div className="card digest-card-v2">
          <div className="digest-v2-header">
            <div className="digest-v2-title-row">
              <div className="digest-v2-icon-wrap">
                <DigestIcon />
              </div>
              <div>
                <div className="digest-v2-title">Portfolio Analysis</div>
                <div className="digest-v2-meta">
                  AI-generated · {new Date(digest.timestamp).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
                  {' · '}
                  {new Date(digest.timestamp).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })}
                </div>
              </div>
            </div>
            <button
              className="btn btn-ghost digest-v2-toggle"
              onClick={() => setDigestExpanded(!digestExpanded)}
            >
              {digestExpanded ? 'Collapse' : 'Read Full Report'}
              <ChevronIcon expanded={digestExpanded} />
            </button>
          </div>

          {!digestExpanded && (
            <div className="digest-v2-preview">
              {getDigestPreview(digest.summary)}
            </div>
          )}

          {digestExpanded && (
            <div className="digest-v2-content">
              <ReactMarkdown>{digest.summary}</ReactMarkdown>
              {digest.risk_notes && (
                <div className="digest-v2-risk-section">
                  <div className="digest-v2-risk-label">
                    <WarningIcon /> Risk Notes
                  </div>
                  <ReactMarkdown>{digest.risk_notes}</ReactMarkdown>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      <div className="card">
        <div className="card-header">
          <span className="card-title" style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <InsightsIcon />
            AI Insights
            {insights && (
              <HealthBadge health={insights.overall_health} />
            )}
            {!insightsExpanded && insights && insights.insights.length > 0 && (
              <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', fontWeight: 400 }}>
                — {insights.insights.length} finding{insights.insights.length !== 1 ? 's' : ''}
              </span>
            )}
          </span>
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <button className="btn btn-ghost" onClick={handleRefresh} disabled={loading}>
              {loading ? <><span className="spinner" /> Analyzing...</> : <><RefreshIcon /> Refresh</>}
            </button>
            {insights && insights.insights.length > 0 && (
              <button
                className="btn btn-ghost digest-v2-toggle"
                onClick={() => setInsightsExpanded(!insightsExpanded)}
              >
                {insightsExpanded ? 'Collapse' : 'Expand'}
                <ChevronIcon expanded={insightsExpanded} />
              </button>
            )}
          </div>
        </div>

        {error && <div className="notice notice-error">{error}</div>}

        {loading && !insights && (
          <div className="empty-state" style={{ padding: '2rem' }}>
            <span className="spinner" style={{ width: '24px', height: '24px', margin: '0 auto 0.75rem' }} />
            <p>Analyzing your portfolio...</p>
          </div>
        )}

        {!loading && insights && insights.insights.length === 0 && (
          <div className="empty-state" style={{ padding: '2rem' }}>
            <div className="empty-state-icon">&#x2713;</div>
            <p>No issues detected. Your portfolio looks healthy.</p>
          </div>
        )}

        {insightsExpanded && insights && insights.insights.length > 0 && (
          <div className="insights-list">
            {insights.insights.map((insight, i) => (
              <InsightCard key={`${insight.type}-${i}`} insight={insight} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function InsightCard({ insight }: { insight: Insight }) {
  return (
    <div className={`insight-card insight-severity-${insight.severity}`}>
      <div className="insight-card-header">
        <SeverityIcon severity={insight.severity} />
        <div className="insight-card-title">{insight.title}</div>
        <span className={`badge badge-${insight.severity === 'high' ? 'critical' : insight.severity}`}>
          {insight.severity}
        </span>
      </div>
      <div className="insight-card-description">{insight.description}</div>
      <div className="insight-card-recommendation">
        <LightbulbIcon />
        {insight.recommendation}
      </div>
    </div>
  );
}

function HealthBadge({ health }: { health: string }) {
  const colorMap: Record<string, string> = {
    good: 'var(--accent-green)',
    fair: 'var(--accent-amber)',
    poor: 'var(--accent-red)',
    unknown: 'var(--text-muted)',
  };
  const bgMap: Record<string, string> = {
    good: 'var(--accent-green-dim)',
    fair: 'var(--accent-amber-dim)',
    poor: 'var(--accent-red-dim)',
    unknown: 'var(--bg-surface)',
  };
  return (
    <span
      className="badge"
      style={{
        background: bgMap[health] || bgMap.unknown,
        color: colorMap[health] || colorMap.unknown,
        marginLeft: '0.35rem',
      }}
    >
      {health}
    </span>
  );
}

function ChevronIcon({ expanded }: { expanded: boolean }) {
  return (
    <svg
      width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
      strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
      style={{ transition: 'transform 0.2s', transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)' }}
    >
      <polyline points="6 9 12 15 18 9" />
    </svg>
  );
}

function WarningIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--accent-amber)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
      <line x1="12" y1="9" x2="12" y2="13" />
      <line x1="12" y1="17" x2="12.01" y2="17" />
    </svg>
  );
}

function SeverityIcon({ severity }: { severity: string }) {
  if (severity === 'high') {
    return (
      <svg className="insight-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--accent-red)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
    );
  }
  if (severity === 'medium') {
    return (
      <svg className="insight-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--accent-amber)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
    );
  }
  return (
    <svg className="insight-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--accent-cyan)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <line x1="12" y1="16" x2="12" y2="12" />
      <line x1="12" y1="8" x2="12.01" y2="8" />
    </svg>
  );
}

function DigestIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <polyline points="14 2 14 8 20 8" />
      <line x1="16" y1="13" x2="8" y2="13" />
      <line x1="16" y1="17" x2="8" y2="17" />
      <polyline points="10 9 9 9 8 9" />
    </svg>
  );
}

function InsightsIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
    </svg>
  );
}

function RefreshIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 2v6h-6" /><path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
      <path d="M3 22v-6h6" /><path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
    </svg>
  );
}

function LightbulbIcon() {
  return (
    <svg className="insight-rec-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="9" y1="18" x2="15" y2="18" />
      <line x1="10" y1="22" x2="14" y2="22" />
      <path d="M15.09 14c.18-.98.65-1.74 1.41-2.5A4.65 4.65 0 0 0 18 8 6 6 0 0 0 6 8c0 1 .23 2.23 1.5 3.5A4.61 4.61 0 0 1 8.91 14" />
    </svg>
  );
}
