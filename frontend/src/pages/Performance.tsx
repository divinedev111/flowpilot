import { useState, useEffect } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer,
  LineChart, Line, CartesianGrid, Area, AreaChart, Cell,
} from 'recharts';
import { api } from '../api/client';
import { formatAssetType } from '../utils/format';

interface AttributionEntry {
  label: string;
  value: number;
  pct: number;
  change: number;
}

interface HistoryPoint {
  date: string;
  net_worth: number;
}

interface PerformanceData {
  by_asset: AttributionEntry[];
  by_type: AttributionEntry[];
  by_source: AttributionEntry[];
  history: HistoryPoint[];
}

const COLORS = ['#2563eb', '#16a34a', '#d97706', '#7c3aed', '#dc2626', '#0891b2', '#e11d48'];

type ViewMode = 'asset' | 'type' | 'source';

export default function Performance() {
  const [data, setData] = useState<PerformanceData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [view, setView] = useState<ViewMode>('asset');

  useEffect(() => {
    const fetch = async () => {
      try {
        const result = await api.getPerformance();
        setData(result);
      } catch (err: unknown) {
        if (err instanceof Error) setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetch();
  }, []);

  if (loading) {
    return (
      <div>
        <div className="page-header animate-in">
          <div>
            <h1 className="page-title">Performance</h1>
            <p className="page-subtitle">Portfolio attribution and historical analysis</p>
          </div>
        </div>
        <div className="card animate-in">
          <div className="empty-state">
            <p className="text-muted">Loading performance data...</p>
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
            <h1 className="page-title">Performance</h1>
            <p className="page-subtitle">Portfolio attribution and historical analysis</p>
          </div>
        </div>
        <div className="notice notice-error animate-in">{error}</div>
      </div>
    );
  }

  if (!data || (data.by_asset.length === 0 && data.history.length === 0)) {
    return (
      <div>
        <div className="page-header animate-in">
          <div>
            <h1 className="page-title">Performance</h1>
            <p className="page-subtitle">Portfolio attribution and historical analysis</p>
          </div>
        </div>
        <div className="card animate-in">
          <div className="empty-state">
            <div className="empty-state-icon">&#x25C8;</div>
            <p>No performance data yet</p>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
              Sync your portfolio to start tracking performance
            </p>
          </div>
        </div>
      </div>
    );
  }

  const rawData = view === 'asset' ? data.by_asset.slice(0, 10)
    : view === 'type' ? data.by_type
    : data.by_source;
  const activeData = view === 'type'
    ? rawData.map(d => ({ ...d, label: formatAssetType(d.label) }))
    : rawData;

  const totalValue = activeData.reduce((s, d) => s + d.value, 0);

  return (
    <div>
      <div className="page-header animate-in">
        <div>
          <h1 className="page-title">Performance</h1>
          <p className="page-subtitle">Portfolio attribution and historical analysis</p>
        </div>
      </div>

      {data.history.length > 1 && (
        <div className="card animate-in animate-in-1" style={{ marginBottom: '1.5rem' }}>
          <div className="card-header">
            <span className="card-title">Portfolio Value History</span>
            <span className="mono" style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              {data.history.length} snapshots
            </span>
          </div>
          <ResponsiveContainer width="100%" height={240}>
            <AreaChart data={data.history} margin={{ top: 5, right: 10, left: 10, bottom: 0 }}>
              <defs>
                <linearGradient id="netWorthGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--accent-green)" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="var(--accent-green)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border-default)" />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 11, fill: 'var(--text-muted)' }}
                axisLine={{ stroke: 'var(--border-default)' }}
                tickLine={false}
              />
              <YAxis
                tick={{ fontSize: 11, fill: 'var(--text-muted)' }}
                axisLine={false}
                tickLine={false}
                tickFormatter={v => `$${(v / 1000).toFixed(0)}k`}
              />
              <Tooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null;
                  const d = payload[0].payload as HistoryPoint;
                  return (
                    <div style={{
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border-strong)',
                      borderRadius: 'var(--radius-sm)',
                      padding: '0.5rem 0.75rem',
                      fontFamily: 'var(--font-mono)',
                      fontSize: '0.8rem',
                    }}>
                      <div style={{ color: 'var(--text-muted)', marginBottom: 2 }}>{d.date}</div>
                      <div style={{ color: 'var(--accent-green)', fontWeight: 600 }}>
                        ${d.net_worth.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                      </div>
                    </div>
                  );
                }}
              />
              <Area
                type="monotone"
                dataKey="net_worth"
                stroke="var(--accent-green)"
                strokeWidth={2}
                fill="url(#netWorthGrad)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}

      <div className="card animate-in animate-in-2" style={{ marginBottom: '1.5rem' }}>
        <div className="card-header">
          <span className="card-title">Attribution Breakdown</span>
          <div className="perf-tabs">
            {(['asset', 'type', 'source'] as ViewMode[]).map(v => (
              <button
                key={v}
                className={`perf-tab ${view === v ? 'perf-tab-active' : ''}`}
                onClick={() => setView(v)}
              >
                {v === 'asset' ? 'By Position' : v === 'type' ? 'By Type' : 'By Source'}
              </button>
            ))}
          </div>
        </div>

        <ResponsiveContainer width="100%" height={Math.max(200, activeData.length * 44)}>
          <BarChart data={activeData} layout="vertical" margin={{ top: 5, right: 40, left: 10, bottom: 5 }}>
            <XAxis
              type="number"
              tick={{ fontSize: 11, fill: 'var(--text-muted)' }}
              axisLine={{ stroke: 'var(--border-default)' }}
              tickLine={false}
              tickFormatter={v => `$${(v / 1000).toFixed(1)}k`}
            />
            <YAxis
              type="category"
              dataKey="label"
              tick={{ fontSize: 12, fill: 'var(--text-primary)', fontWeight: 500 }}
              axisLine={false}
              tickLine={false}
              width={120}
            />
            <Tooltip
              content={({ active, payload }) => {
                if (!active || !payload?.length) return null;
                const d = payload[0].payload as AttributionEntry;
                return (
                  <div style={{
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-strong)',
                    borderRadius: 'var(--radius-sm)',
                    padding: '0.5rem 0.75rem',
                    fontFamily: 'var(--font-mono)',
                    fontSize: '0.8rem',
                  }}>
                    <div style={{ fontWeight: 600, marginBottom: 2 }}>{d.label}</div>
                    <div style={{ color: 'var(--accent-green)' }}>
                      ${d.value.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                    </div>
                    <div style={{ color: 'var(--text-muted)' }}>{d.pct.toFixed(1)}% of portfolio</div>
                    {d.change !== 0 && (
                      <div style={{ color: d.change >= 0 ? 'var(--accent-green)' : 'var(--accent-red)' }}>
                        {d.change >= 0 ? '+' : ''}{d.change.toFixed(1)}% change
                      </div>
                    )}
                  </div>
                );
              }}
            />
            <Bar dataKey="value" radius={[0, 4, 4, 0]} maxBarSize={28}>
              {activeData.map((_, i) => (
                <Cell key={i} fill={COLORS[i % COLORS.length]} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="card animate-in animate-in-3" style={{ padding: 0, overflow: 'hidden' }}>
        <div className="card-header" style={{ padding: '1.25rem 1.5rem 0' }}>
          <span className="card-title">
            {view === 'asset' ? 'Position' : view === 'type' ? 'Asset Type' : 'Source'} Details
          </span>
        </div>
        <div className="table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>{view === 'asset' ? 'Position' : view === 'type' ? 'Asset Type' : 'Source'}</th>
                <th>Market Value</th>
                <th>Portfolio %</th>
                <th>Change</th>
              </tr>
            </thead>
            <tbody>
              {activeData.map((d, i) => (
                <tr key={d.label}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <span style={{
                        width: 8, height: 8, borderRadius: '50%',
                        background: COLORS[i % COLORS.length], flexShrink: 0,
                      }} />
                      <span style={{ fontWeight: 600, textTransform: 'capitalize' }}>{d.label}</span>
                    </div>
                  </td>
                  <td className="mono" style={{ fontWeight: 600, color: 'var(--accent-green)' }}>
                    ${d.value.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                  </td>
                  <td className="mono">{d.pct.toFixed(1)}%</td>
                  <td>
                    {d.change !== 0 ? (
                      <span className={`change-badge ${d.change >= 0 ? 'positive' : 'negative'}`}>
                        {d.change >= 0 ? '+' : ''}{d.change.toFixed(1)}%
                      </span>
                    ) : (
                      <span style={{ color: 'var(--text-dim)' }}>—</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
