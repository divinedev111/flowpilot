import { useState, useEffect } from 'react';
import Markdown from 'react-markdown';
import { api } from '../api/client';
import type { Metrics } from '../types';
import { formatAssetType } from '../utils/format';

interface Adjustment {
  asset_type: string;
  target_pct: number;
}

const ALLOC_COLORS = ['#2563eb', '#16a34a', '#d97706', '#7c3aed', '#dc2626', '#0891b2', '#e11d48'];

export default function Scenarios() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [adjustments, setAdjustments] = useState<Adjustment[]>([]);
  const [result, setResult] = useState<{ current: Metrics; hypothetical: Metrics; commentary: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchPortfolio = async () => {
      try {
        const data = await api.getPortfolio();
        const m = data.metrics as Metrics;
        setMetrics(m);
        const adjs = Object.entries(m.exposure_by_type).map(([asset_type, pct]) => ({
          asset_type,
          target_pct: Math.round(pct * 10) / 10,
        }));
        setAdjustments(adjs);
      } catch (err: unknown) {
        if (err instanceof Error) setError(err.message);
      }
    };
    fetchPortfolio();
  }, []);

  const handleSliderChange = (index: number, value: number) => {
    setAdjustments(prev => {
      const updated = [...prev];
      updated[index] = { ...updated[index], target_pct: value };
      return updated;
    });
  };

  const handleRun = async () => {
    setLoading(true);
    setError('');
    setResult(null);
    try {
      const data = await api.runScenario(adjustments);
      setResult(data);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    if (!metrics) return;
    const adjs = Object.entries(metrics.exposure_by_type).map(([asset_type, pct]) => ({
      asset_type,
      target_pct: Math.round(pct * 10) / 10,
    }));
    setAdjustments(adjs);
    setResult(null);
  };

  const hasChanges = metrics && adjustments.some((adj) => {
    const currentPct = metrics.exposure_by_type[adj.asset_type] || 0;
    return Math.abs(adj.target_pct - currentPct) > 0.1;
  });

  const totalPct = adjustments.reduce((s, a) => s + a.target_pct, 0);
  const totalValid = Math.abs(totalPct - 100) < 1;

  return (
    <div>
      <div className="page-header animate-in">
        <div>
          <h1 className="page-title">Scenario Analysis</h1>
          <p className="page-subtitle">Simulate allocation changes and get AI-powered insights</p>
        </div>
        {metrics && (
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {hasChanges && (
              <button className="btn btn-secondary" onClick={handleReset}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 2v6h-6" /><path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
                  <path d="M3 22v-6h6" /><path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
                </svg>
                Reset
              </button>
            )}
            <button
              className="btn btn-primary"
              onClick={handleRun}
              disabled={loading || !hasChanges || !totalValid}
            >
              {loading ? (
                <><span className="spinner" /> Analyzing...</>
              ) : (
                <>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <polygon points="5 3 19 12 5 21 5 3" />
                  </svg>
                  Run Scenario
                </>
              )}
            </button>
          </div>
        )}
      </div>

      {error && <div className="notice notice-error animate-in">{error}</div>}

      {!metrics ? (
        <div className="card animate-in">
          <div className="empty-state">
            <div className="empty-state-icon">&#x25C8;</div>
            <p>Loading portfolio data...</p>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>Sync your portfolio first to use scenarios</p>
          </div>
        </div>
      ) : (
        <>
          <div className="scenario-overview animate-in animate-in-1">
            <div className="card scenario-current-card">
              <div className="card-header">
                <span className="card-title">Current Allocation</span>
                <span className="badge" style={{ background: 'var(--accent-green-dim)', color: 'var(--accent-green)' }}>Live</span>
              </div>
              <div className="scenario-alloc-visual">
                {Object.entries(metrics.exposure_by_type).map(([type, pct], i) => (
                  <div key={type} className="scenario-alloc-item">
                    <div className="scenario-alloc-ring-wrapper">
                      <svg viewBox="0 0 36 36" className="scenario-alloc-ring">
                        <circle cx="18" cy="18" r="15.9" fill="none" stroke="var(--border-default)" strokeWidth="3" />
                        <circle
                          cx="18" cy="18" r="15.9" fill="none"
                          stroke={ALLOC_COLORS[i % ALLOC_COLORS.length]}
                          strokeWidth="3"
                          strokeDasharray={`${pct} ${100 - pct}`}
                          strokeDashoffset="25"
                          strokeLinecap="round"
                        />
                      </svg>
                      <span className="scenario-alloc-ring-pct mono">{pct.toFixed(0)}%</span>
                    </div>
                    <div className="scenario-alloc-label">{formatAssetType(type)}</div>
                    <div className="scenario-alloc-value mono">${((metrics.net_worth * pct) / 100).toLocaleString(undefined, { maximumFractionDigits: 0 })}</div>
                  </div>
                ))}
              </div>
              <div className="scenario-net-worth">
                <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>Total Portfolio</span>
                <span className="mono" style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)' }}>
                  ${metrics.net_worth.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </span>
              </div>
            </div>

            <div className="card scenario-target-card">
              <div className="card-header">
                <span className="card-title">Target Allocation</span>
                <span className={`mono scenario-total-badge ${totalValid ? '' : 'scenario-total-invalid'}`}>
                  {totalPct.toFixed(1)}%
                </span>
              </div>

              {adjustments.map((adj, i) => {
                const currentPct = metrics.exposure_by_type[adj.asset_type] || 0;
                const changed = Math.abs(adj.target_pct - currentPct) > 0.1;
                const diff = adj.target_pct - currentPct;
                return (
                  <div key={adj.asset_type} className="scenario-slider">
                    <div className="scenario-slider-header">
                      <div className="scenario-slider-name">
                        <span className="scenario-slider-dot" style={{ background: ALLOC_COLORS[i % ALLOC_COLORS.length] }} />
                        <span style={{ fontWeight: 600, fontSize: '0.85rem' }}>{formatAssetType(adj.asset_type)}</span>
                      </div>
                      <div className="scenario-slider-values">
                        <span className="mono" style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>{currentPct.toFixed(1)}%</span>
                        {changed && (
                          <>
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--accent-green)" strokeWidth="2"><polyline points="9 18 15 12 9 6" /></svg>
                            <span className="mono" style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--accent-green)' }}>{adj.target_pct.toFixed(1)}%</span>
                            <span className={`change-badge ${diff > 0 ? 'positive' : 'negative'}`} style={{ fontSize: '0.65rem' }}>
                              {diff > 0 ? '+' : ''}{diff.toFixed(1)}
                            </span>
                          </>
                        )}
                        {!changed && (
                          <span className="mono" style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--text-primary)' }}>{adj.target_pct.toFixed(1)}%</span>
                        )}
                      </div>
                    </div>
                    <div className="scenario-slider-track-wrapper">
                      <div className="scenario-slider-bar-bg">
                        <div
                          className="scenario-slider-bar-fill"
                          style={{
                            width: `${adj.target_pct}%`,
                            background: changed ? ALLOC_COLORS[i % ALLOC_COLORS.length] : 'var(--border-strong)',
                          }}
                        />
                      </div>
                      <input
                        type="range"
                        min={0}
                        max={100}
                        step={0.5}
                        value={adj.target_pct}
                        onChange={e => handleSliderChange(i, parseFloat(e.target.value))}
                        className="scenario-range-input"
                      />
                    </div>
                  </div>
                );
              })}

              {!totalValid && (
                <div className="notice notice-warning" style={{ marginTop: '1rem', marginBottom: 0 }}>
                  Total allocation is {totalPct.toFixed(1)}% — should be close to 100%
                </div>
              )}
            </div>
          </div>

          {result && (
            <div className="animate-in" style={{ marginTop: '1.5rem' }}>
              <div className="card scenario-comparison-card animate-in animate-in-2">
                <div className="card-header">
                  <span className="card-title">Allocation Comparison</span>
                </div>
                <div className="scenario-comparison-bars">
                  <div className="scenario-comparison-row">
                    <span className="scenario-comparison-label">Current</span>
                    <div className="scenario-comparison-bar">
                      {Object.entries(result.current.exposure_by_type).map(([type, pct], i) => (
                        <div
                          key={type}
                          className="scenario-comparison-segment"
                          style={{ width: `${pct}%`, background: ALLOC_COLORS[i % ALLOC_COLORS.length] }}
                          title={`${type}: ${pct.toFixed(1)}%`}
                        />
                      ))}
                    </div>
                  </div>
                  <div className="scenario-comparison-row">
                    <span className="scenario-comparison-label">Target</span>
                    <div className="scenario-comparison-bar">
                      {Object.entries(result.hypothetical.exposure_by_type).map(([type, pct], i) => (
                        <div
                          key={type}
                          className="scenario-comparison-segment"
                          style={{ width: `${pct}%`, background: ALLOC_COLORS[i % ALLOC_COLORS.length] }}
                          title={`${type}: ${pct.toFixed(1)}%`}
                        />
                      ))}
                    </div>
                  </div>
                </div>
                <div className="scenario-comparison-legend">
                  {Object.keys(result.hypothetical.exposure_by_type).map((type, i) => (
                    <span key={type} className="alloc-legend-item">
                      <span className="alloc-legend-dot" style={{ background: ALLOC_COLORS[i % ALLOC_COLORS.length] }} />
                      <span style={{ textTransform: 'capitalize' }}>{type}</span>
                    </span>
                  ))}
                </div>
              </div>

              <div className="grid-2" style={{ marginTop: '1rem' }}>
                <div className="card animate-in animate-in-3">
                  <div className="card-header">
                    <span className="card-title">Detailed Breakdown</span>
                  </div>
                  <div className="table-wrapper">
                    <table className="data-table">
                      <thead>
                        <tr>
                          <th>Asset Type</th>
                          <th>Current</th>
                          <th>Target</th>
                          <th>Change</th>
                          <th>Value Impact</th>
                        </tr>
                      </thead>
                      <tbody>
                        {Object.keys(result.hypothetical.exposure_by_type).map((type, i) => {
                          const curr = result.current.exposure_by_type[type] || 0;
                          const hyp = result.hypothetical.exposure_by_type[type] || 0;
                          const diff = hyp - curr;
                          const valueDiff = (diff / 100) * metrics.net_worth;
                          return (
                            <tr key={type}>
                              <td>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                  <span style={{ width: 8, height: 8, borderRadius: '50%', background: ALLOC_COLORS[i % ALLOC_COLORS.length], flexShrink: 0 }} />
                                  <span style={{ textTransform: 'capitalize', fontWeight: 600 }}>{type}</span>
                                </div>
                              </td>
                              <td className="mono">{curr.toFixed(1)}%</td>
                              <td className="mono" style={{ fontWeight: 600 }}>{hyp.toFixed(1)}%</td>
                              <td>
                                <span className={`change-badge ${diff > 0 ? 'positive' : diff < 0 ? 'negative' : ''}`}>
                                  {diff > 0 ? '+' : ''}{diff.toFixed(1)}%
                                </span>
                              </td>
                              <td className="mono" style={{
                                color: valueDiff > 0 ? 'var(--accent-green)' : valueDiff < 0 ? 'var(--accent-red)' : 'var(--text-muted)',
                                fontWeight: 600,
                              }}>
                                {valueDiff > 0 ? '+' : ''}${valueDiff.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>

                <div className="card scenario-ai-card animate-in animate-in-4">
                  <div className="card-header">
                    <span className="card-title">AI Analysis</span>
                    <span className="scenario-ai-badge">
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M12 2L2 7l10 5 10-5-10-5z" />
                        <path d="M2 17l10 5 10-5" />
                        <path d="M2 12l10 5 10-5" />
                      </svg>
                      Claude
                    </span>
                  </div>
                  {result.commentary ? (
                    <div className="scenario-ai-content">
                      <Markdown>{result.commentary}</Markdown>
                    </div>
                  ) : (
                    <p className="text-muted">No AI commentary available</p>
                  )}
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
