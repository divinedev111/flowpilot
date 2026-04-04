import { useState, useEffect, useCallback } from 'react';
import { ResponsiveHeatMap } from '@nivo/heatmap';
import { api } from '../api/client.ts';
import type { CorrelationResult } from '../types/index.ts';

interface HeatMapDatum {
  x: string;
  y: number | null;
}

interface HeatMapSerie {
  id: string;
  data: HeatMapDatum[];
}

function transformToNivoData(result: CorrelationResult): HeatMapSerie[] {
  return result.symbols.map((rowSymbol, i) => ({
    id: rowSymbol,
    data: result.symbols.map((colSymbol, j) => ({
      x: colSymbol,
      y: result.matrix[i]?.[j] ?? null,
    })),
  }));
}

function getStrengthBadge(corr: number): { label: string; color: string; bg: string } {
  const abs = Math.abs(corr);
  if (abs >= 0.7) {
    return corr > 0
      ? { label: 'Strong +', color: '#1d4ed8', bg: '#dbeafe' }
      : { label: 'Strong -', color: '#dc2626', bg: '#fee2e2' };
  }
  if (abs >= 0.4) {
    return corr > 0
      ? { label: 'Moderate +', color: '#2563eb', bg: '#eff6ff' }
      : { label: 'Moderate -', color: '#ef4444', bg: '#fef2f2' };
  }
  return { label: 'Weak', color: '#6b7280', bg: '#f3f4f6' };
}

export default function Correlation() {
  const [data, setData] = useState<CorrelationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await api.getCorrelation();
      setData(result);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const nivoData = data && data.symbols.length > 0 ? transformToNivoData(data) : null;

  return (
    <div>
      <div className="correlation-header">
        <div>
          <h1 style={{ margin: 0 }}>Correlation Analysis</h1>
          <p style={{ color: '#6b7280', marginTop: '0.25rem' }}>
            Pearson correlation matrix across your equity, ETF, and crypto holdings
            {data && data.period_days > 0 && ` (${data.period_days} trading days)`}
          </p>
        </div>
        <button
          onClick={fetchData}
          disabled={loading}
          className="correlation-refresh-btn"
        >
          {loading ? 'Computing...' : 'Refresh'}
        </button>
      </div>

      {error && (
        <div className="correlation-error">{error}</div>
      )}

      {loading && !data && (
        <div className="correlation-loading">
          <p>Computing correlations...</p>
          <p style={{ fontSize: '0.875rem', color: '#9ca3af' }}>
            This may take a minute due to API rate limits.
          </p>
        </div>
      )}

      {!loading && !error && data && data.symbols.length === 0 && (
        <div className="correlation-empty">
          <p>No equity or ETF positions found.</p>
          <p style={{ fontSize: '0.875rem', color: '#9ca3af' }}>
            Add at least two equity/ETF positions to see correlation analysis.
          </p>
        </div>
      )}

      {data?.warning && (
        <div className="notice notice-warning animate-in" style={{ marginBottom: '1rem' }}>
          {data.warning}
        </div>
      )}

      {nivoData && (
        <>
          <div className="correlation-heatmap-card">
            <h3 style={{ marginTop: 0 }}>Correlation Matrix</h3>
            <div style={{ height: Math.max(400, data.symbols.length * 60 + 80) }}>
              <ResponsiveHeatMap
                data={nivoData}
                margin={{ top: 60, right: 90, bottom: 60, left: 90 }}
                valueFormat=".2f"
                axisTop={{
                  tickSize: 5,
                  tickPadding: 5,
                  tickRotation: -45,
                }}
                axisLeft={{
                  tickSize: 5,
                  tickPadding: 5,
                  tickRotation: 0,
                }}
                colors={{
                  type: 'diverging',
                  scheme: 'red_yellow_blue',
                  minValue: -1,
                  maxValue: 1,
                }}
                emptyColor="#f3f4f6"
                borderWidth={1}
                borderColor={{ from: 'color', modifiers: [['darker', 0.4]] }}
                enableLabels={true}
                labelTextColor={{ from: 'color', modifiers: [['darker', 3]] }}
                legends={[
                  {
                    anchor: 'right',
                    translateX: 40,
                    translateY: 0,
                    length: 200,
                    thickness: 10,
                    direction: 'column',
                    tickPosition: 'after',
                    tickSize: 3,
                    tickSpacing: 4,
                    tickOverlap: false,
                    title: 'Correlation',
                    titleAlign: 'start',
                    titleOffset: 4,
                  },
                ]}
                animate={true}
                hoverTarget="cell"
              />
            </div>
          </div>

          {data.top_pairs && data.top_pairs.length > 0 && (
            <div className="correlation-pairs-card">
              <h3 style={{ marginTop: 0 }}>Top Correlated Pairs</h3>
              <table className="correlation-table">
                <thead>
                  <tr>
                    <th>Pair</th>
                    <th>Correlation</th>
                    <th>Strength</th>
                  </tr>
                </thead>
                <tbody>
                  {data.top_pairs.map((pair) => {
                    const badge = getStrengthBadge(pair.correlation);
                    return (
                      <tr key={`${pair.a}-${pair.b}`}>
                        <td style={{ fontWeight: 500 }}>
                          {pair.a} / {pair.b}
                        </td>
                        <td>
                          <span
                            style={{
                              color: pair.correlation > 0.7 ? '#1d4ed8'
                                : pair.correlation < -0.3 ? '#dc2626'
                                : '#6b7280',
                              fontWeight: 600,
                              fontFamily: 'monospace',
                            }}
                          >
                            {pair.correlation > 0 ? '+' : ''}
                            {pair.correlation.toFixed(3)}
                          </span>
                        </td>
                        <td>
                          <span
                            className="correlation-badge"
                            style={{
                              color: badge.color,
                              backgroundColor: badge.bg,
                            }}
                          >
                            {badge.label}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
