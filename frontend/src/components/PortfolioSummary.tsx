import type { Metrics } from '../types';
import { formatAssetType } from '../utils/format';

const ALLOC_COLORS = ['#3b82f6', '#22c55e', '#f59e0b', '#8b5cf6', '#ef4444', '#06b6d4', '#e11d48'];

interface Props {
  metrics: Metrics;
  alertsTriggered?: number;
}

export default function PortfolioSummary({ metrics, alertsTriggered }: Props) {
  const topHolding = getTopHolding(metrics.concentration_by_symbol);
  const topPct = getTopHoldingPct(metrics.concentration_by_symbol);
  const assetTypes = Object.keys(metrics.exposure_by_type);

  return (
    <div>
      <div className="hero-card" style={{ marginBottom: '1rem' }}>
        <div className="hero-top">
          <div>
            <div className="hero-label">Total Portfolio Value</div>
            <div className="hero-value">
              ${metrics.net_worth.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </div>
            {metrics.day_pnl !== 0 && (
              <div className={`hero-change ${metrics.day_pnl >= 0 ? 'positive' : 'negative'}`}>
                {metrics.day_pnl >= 0 ? '+' : ''}
                ${Math.abs(metrics.day_pnl).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                {' '}today
              </div>
            )}
          </div>
        </div>

        <div className="hero-meta">
          <span>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>
            {metrics.position_count} positions
          </span>
          <span>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21.21 15.89A10 10 0 1 1 8 2.83"/><path d="M22 12A10 10 0 0 0 12 2v10z"/></svg>
            {assetTypes.length} asset type{assetTypes.length !== 1 ? 's' : ''}
          </span>
          {topHolding !== '—' && (
            <span>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
              {topHolding} {topPct}%
            </span>
          )}
        </div>

        <div className="hero-divider" />

        <div className="alloc-bar">
          {Object.entries(metrics.exposure_by_type)
            .filter(([, pct]) => pct >= 0.1)
            .map(([type, pct], i) => (
            <div
              key={type}
              className="alloc-bar-segment"
              style={{ width: `${pct}%`, background: ALLOC_COLORS[i % ALLOC_COLORS.length] }}
              title={`${formatAssetType(type)}: ${pct.toFixed(1)}%`}
            />
          ))}
        </div>
        <div className="alloc-legend">
          {Object.entries(metrics.exposure_by_type)
            .filter(([, pct]) => pct >= 0.1)
            .map(([type, pct], i) => (
            <span key={type} className="alloc-legend-item">
              <span className="alloc-legend-dot" style={{ background: ALLOC_COLORS[i % ALLOC_COLORS.length] }} />
              {formatAssetType(type)} {pct.toFixed(0)}%
            </span>
          ))}
        </div>
      </div>

      <div className="grid-3" style={{ marginBottom: '1.5rem' }}>
        <div className="stat-card">
          <div className="stat-card-row">
            <div className="stat-icon blue">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="3" width="7" height="7" rx="1" />
                <rect x="14" y="3" width="7" height="7" rx="1" />
                <rect x="3" y="14" width="7" height="7" rx="1" />
                <rect x="14" y="14" width="7" height="7" rx="1" />
              </svg>
            </div>
            <div>
              <div className="stat-label">Positions</div>
              <div className="stat-value" style={{ color: '#2563eb' }}>{metrics.position_count}</div>
              <div className="stat-sub">Active holdings</div>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-card-row">
            <div className="stat-icon purple">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21.21 15.89A10 10 0 1 1 8 2.83" />
                <path d="M22 12A10 10 0 0 0 12 2v10z" />
              </svg>
            </div>
            <div>
              <div className="stat-label">Asset Types</div>
              <div className="stat-value" style={{ color: '#7c3aed' }}>{assetTypes.length}</div>
              <div className="stat-sub">{assetTypes.map(formatAssetType).join(', ') || 'None'}</div>
            </div>
          </div>
        </div>

        {metrics.day_pnl !== undefined && metrics.day_pnl !== 0 ? (
          <div className="stat-card">
            <div className="stat-card-row">
              <div className={`stat-icon ${metrics.day_pnl >= 0 ? 'green' : 'red'}`}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points={metrics.day_pnl >= 0 ? '23 6 13.5 15.5 8.5 10.5 1 18' : '23 18 13.5 8.5 8.5 13.5 1 6'} />
                  <polyline points={metrics.day_pnl >= 0 ? '17 6 23 6 23 12' : '17 18 23 18 23 12'} />
                </svg>
              </div>
              <div>
                <div className="stat-label">Day P&amp;L</div>
                <div className="stat-value" style={{ fontSize: '1.25rem', color: metrics.day_pnl >= 0 ? 'var(--accent-green)' : 'var(--accent-red)' }}>
                  {metrics.day_pnl >= 0 ? '+' : ''}{metrics.day_pnl.toLocaleString(undefined, { style: 'currency', currency: 'USD' })}
                </div>
                <div className="stat-sub">Since last sync</div>
              </div>
            </div>
          </div>
        ) : alertsTriggered !== undefined && alertsTriggered > 0 ? (
          <div className="stat-card">
            <div className="stat-card-row">
              <div className="stat-icon red">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
                  <path d="M13.73 21a2 2 0 0 1-3.46 0" />
                </svg>
              </div>
              <div>
                <div className="stat-label">Alerts Fired</div>
                <div className="stat-value" style={{ color: '#dc2626' }}>{alertsTriggered}</div>
                <div className="stat-sub">This sync</div>
              </div>
            </div>
          </div>
        ) : (
          <div className="stat-card">
            <div className="stat-card-row">
              <div className="stat-icon amber">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
              </div>
              <div>
                <div className="stat-label">Top Holding</div>
                <div className="stat-value" style={{ fontSize: '1.25rem', color: '#d97706' }}>{topHolding}</div>
                <div className="stat-sub">{topPct}% of portfolio</div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function getTopHolding(conc: Record<string, number>): string {
  let max = 0;
  let sym = '—';
  for (const [k, v] of Object.entries(conc)) {
    if (v > max) { max = v; sym = k; }
  }
  return sym;
}

function getTopHoldingPct(conc: Record<string, number>): string {
  const max = Math.max(0, ...Object.values(conc));
  return max.toFixed(1);
}
