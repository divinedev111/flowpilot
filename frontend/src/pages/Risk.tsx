import { useState, useEffect } from 'react';
import {
  XAxis, YAxis, Tooltip, ResponsiveContainer, Area, AreaChart,
} from 'recharts';
import { api } from '../api/client';
import { formatAssetType } from '../utils/format';

interface TopPosition {
  symbol: string;
  pct: number;
}

interface Concentration {
  top_position: TopPosition;
  top_positions: TopPosition[];
  top_3_pct: number;
  top_5_pct: number;
  hhi: number;
}

interface Exposure {
  by_type: Record<string, number>;
  by_source: Record<string, number>;
}

interface Drawdown {
  max_drawdown_pct: number;
  current_drawdown_pct: number;
  peak_value: number;
  current_value: number;
}

interface HistoryPoint {
  date: string;
  net_worth: number;
}

interface RiskData {
  concentration: Concentration;
  exposure: Exposure;
  drawdown: Drawdown;
  diversification_score: number;
  risk_level: string;
  warnings: string[];
  history: HistoryPoint[];
}

const RISK_COLORS: Record<string, string> = {
  low: 'var(--accent-green)',
  moderate: 'var(--accent-amber)',
  high: 'var(--accent-red)',
  critical: 'var(--accent-red)',
};

const TYPE_COLORS = ['#2563eb', '#16a34a', '#d97706', '#7c3aed', '#dc2626', '#0891b2'];
const SOURCE_COLORS = ['#16a34a', '#2563eb', '#d97706', '#7c3aed'];

export default function Risk() {
  const [data, setData] = useState<RiskData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchRisk = async () => {
      try {
        const result = await api.getRisk();
        setData(result);
      } catch (err: unknown) {
        if (err instanceof Error) setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetchRisk();
  }, []);

  if (loading) {
    return (
      <div>
        <div className="page-header animate-in">
          <div>
            <h1 className="page-title">Risk Dashboard</h1>
            <p className="page-subtitle">Portfolio risk analysis and diversification metrics</p>
          </div>
        </div>
        <div className="card animate-in">
          <div className="empty-state">
            <span className="spinner" style={{ width: 24, height: 24, margin: '0 auto 1rem' }} />
            <p>Loading risk data...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div>
        <div className="page-header animate-in">
          <div>
            <h1 className="page-title">Risk Dashboard</h1>
            <p className="page-subtitle">Portfolio risk analysis and diversification metrics</p>
          </div>
        </div>
        <div className="notice notice-error animate-in">{error}</div>
      </div>
    );
  }

  if (!data) return null;

  const riskColor = RISK_COLORS[data.risk_level] || 'var(--text-muted)';

  const gaugeRadius = 54;
  const gaugeCircumference = 2 * Math.PI * gaugeRadius;
  const gaugeProgress = (data.diversification_score / 100) * gaugeCircumference;

  const concentrationEntries = data.concentration.top_positions || [];

  return (
    <div>
      <div className="page-header animate-in">
        <div>
          <h1 className="page-title">Risk Dashboard</h1>
          <p className="page-subtitle">Portfolio risk analysis and diversification metrics</p>
        </div>
      </div>

      <div className="hero-card animate-in animate-in-1">
        <div className="risk-hero-layout">
          <div className="risk-hero-gauge">
            <svg viewBox="0 0 128 128" className="risk-gauge-svg">
              <circle
                cx="64" cy="64" r={gaugeRadius}
                fill="none"
                stroke="rgba(255,255,255,0.1)"
                strokeWidth="8"
              />
              <circle
                cx="64" cy="64" r={gaugeRadius}
                fill="none"
                stroke={riskColor}
                strokeWidth="8"
                strokeDasharray={`${gaugeProgress} ${gaugeCircumference - gaugeProgress}`}
                strokeDashoffset={gaugeCircumference * 0.25}
                strokeLinecap="round"
                style={{ transition: 'stroke-dasharray 0.8s ease' }}
              />
            </svg>
            <div className="risk-gauge-center">
              <span className="risk-gauge-value">{data.diversification_score}</span>
              <span className="risk-gauge-label">SCORE</span>
            </div>
          </div>
          <div className="risk-hero-info">
            <div className="hero-label">Risk Assessment</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.75rem' }}>
              <span
                className="risk-level-badge"
                style={{ background: riskColor }}
              >
                {data.risk_level.charAt(0).toUpperCase() + data.risk_level.slice(1)}
              </span>
              <span style={{ fontSize: '0.85rem', opacity: 0.7 }}>
                Diversification Score: {data.diversification_score}/100
              </span>
            </div>
            {data.warnings.length > 0 && (
              <div className="risk-warnings">
                {data.warnings.map((w, i) => (
                  <div key={i} className="risk-warning-item">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                      <line x1="12" y1="9" x2="12" y2="13" />
                      <line x1="12" y1="17" x2="12.01" y2="17" />
                    </svg>
                    {w}
                  </div>
                ))}
              </div>
            )}
            {data.warnings.length === 0 && (
              <div style={{ fontSize: '0.85rem', opacity: 0.65 }}>
                No active risk warnings. Your portfolio looks well-balanced.
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="grid-4 animate-in animate-in-2" style={{ margin: '1rem 0' }}>
        <div className="stat-card">
          <div className="stat-card-row">
            <div className="stat-icon red">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="23 18 13.5 8.5 8.5 13.5 1 6" />
                <polyline points="17 18 23 18 23 12" />
              </svg>
            </div>
            <div>
              <div className="stat-label">Max Drawdown</div>
              <div className="stat-value" style={{ color: 'var(--accent-red)' }}>
                {data.drawdown.max_drawdown_pct.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card-row">
            <div className="stat-icon amber">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 2v20M2 12h20" />
              </svg>
            </div>
            <div>
              <div className="stat-label">Current Drawdown</div>
              <div className="stat-value" style={{ color: data.drawdown.current_drawdown_pct < 0 ? 'var(--accent-red)' : 'var(--accent-green)' }}>
                {data.drawdown.current_drawdown_pct.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card-row">
            <div className="stat-icon green">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
                <polyline points="17 6 23 6 23 12" />
              </svg>
            </div>
            <div>
              <div className="stat-label">Peak Value</div>
              <div className="stat-value">
                ${data.drawdown.peak_value.toLocaleString(undefined, { maximumFractionDigits: 0 })}
              </div>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card-row">
            <div className="stat-icon purple">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="12" y1="1" x2="12" y2="23" />
                <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
            </div>
            <div>
              <div className="stat-label">Current Value</div>
              <div className="stat-value">
                ${data.drawdown.current_value.toLocaleString(undefined, { maximumFractionDigits: 0 })}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="grid-2 animate-in animate-in-3" style={{ marginBottom: '1rem' }}>
        <div className="card">
          <div className="card-header">
            <span className="card-title">Concentration Analysis</span>
            <span className="badge" style={{ background: 'var(--accent-purple-dim)', color: 'var(--accent-purple)' }}>
              HHI {data.concentration.hhi.toFixed(0)}
            </span>
          </div>

          {concentrationEntries.length > 0 ? (
            <div className="risk-bars">
              {concentrationEntries.map((entry, i) => (
                <div key={entry.symbol} className="risk-bar-row">
                  <span className="risk-bar-label mono">{entry.symbol}</span>
                  <div className="risk-bar-track">
                    <div
                      className="risk-bar-fill"
                      style={{
                        width: `${Math.min(entry.pct, 100)}%`,
                        background: TYPE_COLORS[i % TYPE_COLORS.length],
                      }}
                    />
                  </div>
                  <span className="risk-bar-value mono">{entry.pct.toFixed(1)}%</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-dim" style={{ fontSize: '0.85rem' }}>No concentration data available</p>
          )}

          <div className="risk-concentration-stats">
            <div className="risk-conc-stat">
              <span className="risk-conc-stat-label">Top Holding</span>
              <span className="risk-conc-stat-value mono">
                {data.concentration.top_position.symbol || '--'} ({data.concentration.top_position.pct || 0}%)
              </span>
            </div>
            <div className="risk-conc-stat">
              <span className="risk-conc-stat-label">Top 3</span>
              <span className="risk-conc-stat-value mono">{data.concentration.top_3_pct}%</span>
            </div>
            <div className="risk-conc-stat">
              <span className="risk-conc-stat-label">Top 5</span>
              <span className="risk-conc-stat-value mono">{data.concentration.top_5_pct}%</span>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="card-header">
            <span className="card-title">Exposure Breakdown</span>
          </div>

          <div style={{ marginBottom: '1.25rem' }}>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.5rem', fontWeight: 500 }}>By Asset Type</div>
            <div className="risk-bars">
              {Object.entries(data.exposure.by_type).map(([type, pct], i) => (
                <div key={type} className="risk-bar-row">
                  <span className="risk-bar-label">{formatAssetType(type)}</span>
                  <div className="risk-bar-track">
                    <div
                      className="risk-bar-fill"
                      style={{
                        width: `${Math.min(pct, 100)}%`,
                        background: TYPE_COLORS[i % TYPE_COLORS.length],
                      }}
                    />
                  </div>
                  <span className="risk-bar-value mono">{pct.toFixed(1)}%</span>
                </div>
              ))}
            </div>
          </div>

          <div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.5rem', fontWeight: 500 }}>By Source</div>
            <div className="risk-bars">
              {Object.entries(data.exposure.by_source).map(([source, pct], i) => (
                <div key={source} className="risk-bar-row">
                  <span className="risk-bar-label" style={{ textTransform: 'capitalize' }}>{source}</span>
                  <div className="risk-bar-track">
                    <div
                      className="risk-bar-fill"
                      style={{
                        width: `${Math.min(pct, 100)}%`,
                        background: SOURCE_COLORS[i % SOURCE_COLORS.length],
                      }}
                    />
                  </div>
                  <span className="risk-bar-value mono">{pct.toFixed(1)}%</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="card animate-in animate-in-4">
        <div className="card-header">
          <span className="card-title">Portfolio Value History</span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Last {data.history.length} snapshots</span>
        </div>
        {data.history.length > 1 ? (
          <ResponsiveContainer width="100%" height={280}>
            <AreaChart data={data.history} margin={{ top: 8, right: 8, left: 8, bottom: 0 }}>
              <defs>
                <linearGradient id="historyGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--accent-green)" stopOpacity={0.2} />
                  <stop offset="95%" stopColor="var(--accent-green)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis
                dataKey="date"
                tick={{ fontSize: 11, fill: 'var(--text-muted)' }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v: string) => {
                  const parts = v.split('-');
                  return `${parts[1]}/${parts[2]}`;
                }}
              />
              <YAxis
                tick={{ fontSize: 11, fill: 'var(--text-muted)' }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v: number) => `$${(v / 1000).toFixed(1)}k`}
                width={55}
              />
              <Tooltip
                formatter={(value: number) => [`$${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, 'Net Worth']}
                labelFormatter={(label: string) => label}
              />
              <Area
                type="monotone"
                dataKey="net_worth"
                stroke="var(--accent-green)"
                strokeWidth={2}
                fill="url(#historyGrad)"
                dot={false}
                activeDot={{ r: 4, strokeWidth: 2 }}
              />
            </AreaChart>
          </ResponsiveContainer>
        ) : (
          <div className="empty-state" style={{ padding: '2rem' }}>
            <p style={{ fontSize: '0.85rem' }}>Not enough snapshots to display chart. Sync your portfolio to start tracking history.</p>
          </div>
        )}
      </div>
    </div>
  );
}

