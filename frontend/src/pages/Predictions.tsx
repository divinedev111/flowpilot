import { useState, useEffect, useMemo } from 'react';
import { api } from '../api/client';
import type { PredictionMarket, WalletPosition, WalletSummary } from '../types';

type MainTab = 'markets' | 'positions';
type FilterTab = 'all' | 'crypto' | 'macro' | 'tech';

declare global {
  interface Window {
    ethereum?: {
      request: (args: { method: string; params?: unknown[] }) => Promise<string[]>;
      on?: (event: string, cb: (...args: unknown[]) => void) => void;
    };
  }
}

function formatVolume(v: number): string {
  if (v >= 1_000_000) return `$${(v / 1_000_000).toFixed(1)}M`;
  if (v >= 1_000) return `$${(v / 1_000).toFixed(1)}K`;
  return `$${v.toFixed(0)}`;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

function formatUsd(v: number): string {
  return v.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });
}

function probabilityColor(p: number): string {
  if (p >= 0.7) return '#10b981';
  if (p >= 0.4) return '#f59e0b';
  return '#ef4444';
}

function probabilityBarGradient(p: number): string {
  if (p >= 0.6) return 'linear-gradient(90deg, #10b981, #34d399)';
  if (p >= 0.4) return 'linear-gradient(90deg, #f59e0b, #fbbf24)';
  return 'linear-gradient(90deg, #ef4444, #f87171)';
}

function pnlColor(v: number): string {
  if (v > 0) return 'var(--accent-green)';
  if (v < 0) return 'var(--accent-red)';
  return 'var(--text-secondary)';
}

function truncateAddr(addr: string): string {
  return addr.slice(0, 6) + '...' + addr.slice(-4);
}

const filterTabs: { key: FilterTab; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'crypto', label: 'Crypto' },
  { key: 'macro', label: 'Macro' },
  { key: 'tech', label: 'Tech' },
];

export default function Predictions() {
  const [mainTab, setMainTab] = useState<MainTab>('markets');
  const [markets, setMarkets] = useState<PredictionMarket[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<FilterTab>('all');

  // Wallet state
  const [walletAddr, setWalletAddr] = useState<string>(() =>
    localStorage.getItem('fp-polymarket-wallet') || ''
  );
  const [walletPositions, setWalletPositions] = useState<WalletPosition[]>([]);
  const [walletSummary, setWalletSummary] = useState<WalletSummary | null>(null);
  const [walletLoading, setWalletLoading] = useState(false);
  const [walletError, setWalletError] = useState('');

  const fetchPredictions = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await api.getPredictions();
      setMarkets(data.markets || []);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchWalletPositions = async (addr?: string) => {
    const wallet = addr || walletAddr;
    if (!wallet) return;
    setWalletLoading(true);
    setWalletError('');
    try {
      const data = await api.getWalletPositions(wallet);
      setWalletPositions(data.positions || []);
      setWalletSummary(data.summary || null);
    } catch (err: unknown) {
      if (err instanceof Error) setWalletError(err.message);
    } finally {
      setWalletLoading(false);
    }
  };

  const connectMetaMask = async () => {
    if (!window.ethereum) {
      setWalletError('MetaMask not detected. Please install MetaMask.');
      return;
    }
    try {
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      if (accounts && accounts.length > 0) {
        const addr = accounts[0];
        setWalletAddr(addr);
        localStorage.setItem('fp-polymarket-wallet', addr);
        fetchWalletPositions(addr);
      }
    } catch (err: unknown) {
      if (err instanceof Error) setWalletError(err.message);
    }
  };

  const disconnectWallet = () => {
    setWalletAddr('');
    setWalletPositions([]);
    setWalletSummary(null);
    localStorage.removeItem('fp-polymarket-wallet');
  };

  useEffect(() => {
    fetchPredictions();
  }, []);

  useEffect(() => {
    if (walletAddr && mainTab === 'positions') {
      fetchWalletPositions();
    }
  }, [mainTab]);

  const filtered = useMemo(() => {
    if (filter === 'all') return markets;
    return markets.filter(m => m.category === filter);
  }, [markets, filter]);

  const grouped = useMemo(() => {
    const groups: Record<string, PredictionMarket[]> = {};
    for (const m of filtered) {
      const key = m.category || 'other';
      if (!groups[key]) groups[key] = [];
      groups[key].push(m);
    }
    return groups;
  }, [filtered]);

  const categoryOrder = ['crypto', 'tech', 'macro', 'other'];

  const activePositions = useMemo(() =>
    walletPositions.filter(p => !p.resolved), [walletPositions]);
  const resolvedPositions = useMemo(() =>
    walletPositions.filter(p => p.resolved), [walletPositions]);

  const tabStyle = (tab: MainTab): React.CSSProperties => ({
    padding: '0.6rem 1.5rem',
    cursor: 'pointer',
    border: 'none',
    borderBottom: mainTab === tab ? '3px solid var(--accent-cyan)' : '3px solid transparent',
    backgroundColor: 'transparent',
    fontWeight: mainTab === tab ? 700 : 400,
    color: mainTab === tab ? 'var(--accent-cyan)' : 'var(--text-muted)',
    fontSize: '0.95rem',
    fontFamily: 'var(--font-display)',
  });

  return (
    <div className="predictions-page">
      <div className="predictions-header">
        <div>
          <h1 className="predictions-title">Prediction Markets</h1>
          <p className="predictions-subtitle">Polymarket odds & your positions</p>
        </div>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          {walletAddr ? (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{
                padding: '0.4rem 0.8rem',
                borderRadius: 'var(--radius-md)',
                backgroundColor: 'var(--accent-green-dim)',
                color: 'var(--accent-green)',
                fontSize: '0.8rem',
                fontWeight: 600,
                fontFamily: 'var(--font-mono)',
              }}>
                {truncateAddr(walletAddr)}
              </span>
              <button
                onClick={disconnectWallet}
                style={{
                  padding: '0.35rem 0.6rem',
                  border: '1px solid var(--border-default)',
                  borderRadius: 'var(--radius-sm)',
                  backgroundColor: 'transparent',
                  color: 'var(--text-muted)',
                  cursor: 'pointer',
                  fontSize: '0.75rem',
                }}
              >
                Disconnect
              </button>
            </div>
          ) : (
            <button
              onClick={connectMetaMask}
              style={{
                padding: '0.5rem 1rem',
                background: 'linear-gradient(135deg, #f6851b, #e2761b)',
                color: 'white',
                border: 'none',
                borderRadius: 'var(--radius-md)',
                cursor: 'pointer',
                fontSize: '0.85rem',
                fontWeight: 600,
              }}
            >
              Connect MetaMask
            </button>
          )}
          {mainTab === 'markets' && (
            <button
              className="predictions-refresh-btn"
              onClick={fetchPredictions}
              disabled={loading}
            >
              {loading ? 'Refreshing...' : 'Refresh'}
            </button>
          )}
          {mainTab === 'positions' && walletAddr && (
            <button
              className="predictions-refresh-btn"
              onClick={() => fetchWalletPositions()}
              disabled={walletLoading}
            >
              {walletLoading ? 'Loading...' : 'Refresh'}
            </button>
          )}
        </div>
      </div>

      <div style={{ display: 'flex', gap: 0, borderBottom: '1px solid var(--border-default)', marginBottom: '1.5rem' }}>
        <button style={tabStyle('markets')} onClick={() => setMainTab('markets')}>Markets</button>
        <button style={tabStyle('positions')} onClick={() => setMainTab('positions')}>My Positions</button>
      </div>

      {error && <div className="predictions-error">{error}</div>}
      {walletError && <div className="predictions-error">{walletError}</div>}

      {mainTab === 'markets' && (
        <>
          <div className="predictions-tabs">
            {filterTabs.map(tab => (
              <button
                key={tab.key}
                className={`predictions-tab ${filter === tab.key ? 'predictions-tab--active' : ''}`}
                onClick={() => setFilter(tab.key)}
              >
                {tab.label}
                <span className="predictions-tab-count">
                  {tab.key === 'all' ? markets.length : markets.filter(m => m.category === tab.key).length}
                </span>
              </button>
            ))}
          </div>

          {loading && (
            <div className="predictions-loading">
              <div className="predictions-spinner" />
              <p>Scanning prediction markets...</p>
            </div>
          )}

          {!loading && filtered.length === 0 && (
            <div className="predictions-empty">
              <div className="predictions-empty-icon">&#9878;</div>
              <h3>No relevant markets found</h3>
              <p>Sync your portfolio to discover prediction markets related to your holdings.</p>
            </div>
          )}

          {!loading && categoryOrder.map(cat => {
            const items = grouped[cat];
            if (!items || items.length === 0) return null;
            return (
              <div key={cat} className="predictions-group">
                <h2 className="predictions-group-title">{cat.charAt(0).toUpperCase() + cat.slice(1)}</h2>
                <div className="predictions-grid">
                  {items.map(market => (
                    <MarketCard key={market.id} market={market} />
                  ))}
                </div>
              </div>
            );
          })}
        </>
      )}

      {mainTab === 'positions' && (
        <>
          {!walletAddr ? (
            <div className="predictions-empty">
              <div className="predictions-empty-icon" style={{ fontSize: '2.5rem' }}>&#129418;</div>
              <h3>Connect your wallet</h3>
              <p>Connect MetaMask to view your Polymarket positions and P&L.</p>
            </div>
          ) : walletLoading ? (
            <div className="predictions-loading">
              <div className="predictions-spinner" />
              <p>Loading positions...</p>
            </div>
          ) : (
            <>
              {walletSummary && (
                <div style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
                  gap: '1rem',
                  marginBottom: '1.5rem',
                }}>
                  <SummaryCard label="Active Positions" value={walletSummary.active_positions.toString()} />
                  <SummaryCard label="Portfolio Value" value={formatUsd(walletSummary.total_value)} />
                  <SummaryCard label="Cost Basis" value={formatUsd(walletSummary.total_cost_basis)} />
                  <SummaryCard
                    label="Unrealized P&L"
                    value={formatUsd(walletSummary.total_unrealized_pnl)}
                    color={pnlColor(walletSummary.total_unrealized_pnl)}
                  />
                  <SummaryCard
                    label="Total Cash P&L"
                    value={formatUsd(walletSummary.total_cash_pnl)}
                    color={pnlColor(walletSummary.total_cash_pnl)}
                  />
                  <SummaryCard label="Claimable" value={walletSummary.claimable_count.toString()} />
                </div>
              )}

              {activePositions.length > 0 && (
                <div style={{ marginBottom: '2rem' }}>
                  <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem', color: 'var(--text-primary)' }}>
                    Active ({activePositions.length})
                  </h2>
                  <div className="predictions-grid">
                    {activePositions.map((pos, i) => (
                      <PositionCard key={pos.condition_id + '-' + i} position={pos} />
                    ))}
                  </div>
                </div>
              )}

              {resolvedPositions.length > 0 && (
                <div>
                  <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem', color: 'var(--text-muted)' }}>
                    Resolved ({resolvedPositions.length})
                  </h2>
                  <div className="predictions-grid">
                    {resolvedPositions.map((pos, i) => (
                      <PositionCard key={pos.condition_id + '-resolved-' + i} position={pos} />
                    ))}
                  </div>
                </div>
              )}

              {walletPositions.length === 0 && (
                <div className="predictions-empty">
                  <h3>No positions found</h3>
                  <p>This wallet has no Polymarket positions. Try a different address or start trading on Polymarket.</p>
                </div>
              )}
            </>
          )}
        </>
      )}
    </div>
  );
}

function SummaryCard({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={{
      padding: '1rem',
      borderRadius: 'var(--radius-lg)',
      border: '1px solid var(--border-default)',
      backgroundColor: 'var(--bg-card)',
      boxShadow: 'var(--shadow-card)',
    }}>
      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem', textTransform: 'uppercase', letterSpacing: '0.04em' }}>
        {label}
      </div>
      <div style={{ fontSize: '1.25rem', fontWeight: 700, color: color || 'var(--text-primary)' }}>
        {value}
      </div>
    </div>
  );
}

function PositionCard({ position }: { position: WalletPosition }) {
  const pnl = position.unrealized_pnl;
  const pnlPct = position.initial_value > 0
    ? ((pnl / position.initial_value) * 100).toFixed(1)
    : '0.0';
  const isWinning = position.current_price >= 0.5;

  return (
    <div className="prediction-card" style={{
      opacity: position.resolved ? 0.7 : 1,
    }}>
      <div className="prediction-card-header">
        <span style={{
          padding: '0.15rem 0.6rem',
          borderRadius: '9999px',
          fontSize: '0.7rem',
          fontWeight: 700,
          textTransform: 'uppercase',
          backgroundColor: position.outcome === 'Yes'
            ? 'var(--accent-green-dim)'
            : 'var(--accent-red-dim)',
          color: position.outcome === 'Yes'
            ? 'var(--accent-green)'
            : 'var(--accent-red)',
        }}>
          {position.outcome}
        </span>
        {position.redeemable && (
          <span style={{
            padding: '0.15rem 0.6rem',
            borderRadius: '9999px',
            fontSize: '0.7rem',
            fontWeight: 700,
            backgroundColor: 'var(--accent-amber-dim)',
            color: 'var(--accent-amber)',
          }}>
            CLAIMABLE
          </span>
        )}
        {position.resolved && !position.redeemable && (
          <span style={{
            padding: '0.15rem 0.6rem',
            borderRadius: '9999px',
            fontSize: '0.7rem',
            fontWeight: 700,
            backgroundColor: 'var(--bg-surface)',
            color: 'var(--text-muted)',
          }}>
            RESOLVED
          </span>
        )}
      </div>

      <h3 className="prediction-card-title">{position.title}</h3>

      <div className="prediction-probability">
        <div className="prediction-probability-label">
          <span className="prediction-probability-value" style={{
            color: isWinning ? 'var(--accent-green)' : 'var(--accent-red)',
            fontSize: '1.25rem',
          }}>
            {(position.current_price * 100).toFixed(0)}%
          </span>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Entry: {(position.avg_price * 100).toFixed(1)}%
          </span>
        </div>
        <div className="prediction-bar-track">
          <div
            className="prediction-bar-fill"
            style={{
              width: `${position.current_price * 100}%`,
              background: probabilityBarGradient(position.current_price),
            }}
          />
        </div>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: '0.5rem',
        fontSize: '0.8rem',
      }}>
        <div>
          <span style={{ color: 'var(--text-muted)' }}>Size: </span>
          <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{position.size.toFixed(2)} shares</span>
        </div>
        <div>
          <span style={{ color: 'var(--text-muted)' }}>Value: </span>
          <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{formatUsd(position.current_value)}</span>
        </div>
        <div>
          <span style={{ color: 'var(--text-muted)' }}>Cost: </span>
          <span style={{ color: 'var(--text-secondary)' }}>{formatUsd(position.initial_value)}</span>
        </div>
        <div>
          <span style={{ color: 'var(--text-muted)' }}>P&L: </span>
          <span style={{ fontWeight: 700, color: pnlColor(pnl) }}>
            {pnl >= 0 ? '+' : ''}{formatUsd(pnl)} ({pnlPct}%)
          </span>
        </div>
      </div>

      {position.cash_pnl !== 0 && (
        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          Realized P&L: <span style={{ color: pnlColor(position.cash_pnl), fontWeight: 600 }}>
            {position.cash_pnl >= 0 ? '+' : ''}{formatUsd(position.cash_pnl)}
          </span>
        </div>
      )}
    </div>
  );
}

function MarketCard({ market }: { market: PredictionMarket }) {
  const yesPct = Math.round(market.outcome_yes * 100);

  return (
    <div className="prediction-card">
      <div className="prediction-card-header">
        <span className={`prediction-category-tag prediction-category-tag--${market.category}`}>
          {market.category}
        </span>
        {market.end_date && (
          <span className="prediction-end-date">
            Ends {formatDate(market.end_date)}
          </span>
        )}
      </div>

      <h3 className="prediction-card-title">{market.title}</h3>

      <div className="prediction-probability">
        <div className="prediction-probability-label">
          <span
            className="prediction-probability-value"
            style={{ color: probabilityColor(market.outcome_yes) }}
          >
            {yesPct}% YES
          </span>
          <span className="prediction-probability-no">
            {Math.round(market.outcome_no * 100)}% NO
          </span>
        </div>
        <div className="prediction-bar-track">
          <div
            className="prediction-bar-fill"
            style={{
              width: `${yesPct}%`,
              background: probabilityBarGradient(market.outcome_yes),
            }}
          />
        </div>
      </div>

      <div className="prediction-meta">
        <span className="prediction-volume-badge">
          {formatVolume(market.volume)} volume
        </span>
      </div>

      {market.related_symbols && market.related_symbols.length > 0 && (
        <div className="prediction-symbols">
          {market.related_symbols.map(sym => (
            <span key={sym} className="prediction-symbol-pill">{sym}</span>
          ))}
        </div>
      )}

      {market.url && (
        <a
          href={market.url}
          target="_blank"
          rel="noopener noreferrer"
          className="prediction-link"
        >
          View on Polymarket &rarr;
        </a>
      )}
    </div>
  );
}
