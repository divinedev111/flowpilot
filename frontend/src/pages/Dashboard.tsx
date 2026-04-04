import { useState, useEffect } from 'react';
import { api } from '../api/client';

declare global {
  interface Window {
    ethereum?: {
      request: (args: { method: string; params?: unknown[] }) => Promise<string[]>;
    };
  }
}
import type { Position, Metrics, SyncResult } from '../types';
import PortfolioSummary from '../components/PortfolioSummary';
import AllocationChart from '../components/AllocationChart';
import PositionsTable from '../components/PositionsTable';
import DiffView from '../components/DiffView';
import InsightsPanel from '../components/InsightsPanel';

export default function Dashboard() {
  const [positions, setPositions] = useState<Position[]>([]);
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [lastSync, setLastSync] = useState<SyncResult | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState('');
  const [authStatus, setAuthStatus] = useState<{ schwab: boolean; coinbase: boolean; polymarket: boolean; polymarket_wallet: string } | null>(null);
  const [polyWalletInput, setPolyWalletInput] = useState('');
  const [polySaving, setPolySaving] = useState(false);
  const [connSaving, setConnSaving] = useState('');

  const fetchPortfolio = async () => {
    try {
      const data = await api.getPortfolio();
      setPositions(data.positions || []);
      setMetrics(data.metrics);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleSync = async () => {
    setSyncing(true);
    setError('');
    try {
      const result = await api.sync();
      setLastSync(result);
      await Promise.all([fetchPortfolio(), fetchAuthStatus()]);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setSyncing(false);
    }
  };

  const fetchAuthStatus = async () => {
    try {
      const status = await api.getAuthStatus();
      setAuthStatus(status);
    } catch {
      // ignore — auth status is non-critical
    }
  };

  const handleConnectPolymarket = async () => {
    const wallet = polyWalletInput.trim();
    if (!wallet) return;
    setPolySaving(true);
    try {
      await api.savePolymarketWallet(wallet);
      setPolyWalletInput('');
      await fetchAuthStatus();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setPolySaving(false);
    }
  };

  const handleConnectMetaMask = async () => {
    if (!window.ethereum) {
      setError('MetaMask not detected. Please install MetaMask or paste your wallet address.');
      return;
    }
    setPolySaving(true);
    try {
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      if (accounts && accounts.length > 0) {
        await api.savePolymarketWallet(accounts[0]);
        await fetchAuthStatus();
      }
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setPolySaving(false);
    }
  };

  const handleDisconnectPolymarket = async () => {
    setPolySaving(true);
    try {
      await api.savePolymarketWallet('');
      await fetchAuthStatus();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setPolySaving(false);
    }
  };

  const handleDisconnectSchwab = async () => {
    setConnSaving('schwab');
    try {
      await api.disconnectSchwab();
      await fetchAuthStatus();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setConnSaving('');
    }
  };

  const handleDisconnectCoinbase = async () => {
    setConnSaving('coinbase');
    try {
      await api.disconnectCoinbase();
      await fetchAuthStatus();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setConnSaving('');
    }
  };

  useEffect(() => { fetchPortfolio(); fetchAuthStatus(); }, []);

  return (
    <div>
      <div className="page-header animate-in">
        <div>
          <h1 className="page-title">Dashboard</h1>
          <p className="page-subtitle">Portfolio overview and real-time positions</p>
        </div>
        <button className="btn btn-primary" onClick={handleSync} disabled={syncing}>
          {syncing ? <><span className="spinner" /> Syncing...</> : <><SyncIcon /> Sync Portfolio</>}
        </button>
      </div>

      {error && <div className="notice notice-error animate-in">{error}</div>}

      {lastSync?.fetch_errors && lastSync.fetch_errors.filter(e => !e.includes('not authenticated')).length > 0 && (
        <div className="notice notice-warning animate-in">
          <strong>Fetch warnings:</strong>&nbsp;
          {lastSync.fetch_errors.filter(e => !e.includes('not authenticated')).join(', ')}
        </div>
      )}

      {authStatus && (
        <div className="connector-status animate-in">
          <div className="connector-row">
            <div className="connector-item">
              <span className={`connector-dot ${authStatus.coinbase ? 'connected' : 'disconnected'}`} />
              <span className="connector-label">Coinbase</span>
              {authStatus.coinbase ? (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span className="connector-badge badge-connected">Connected</span>
                  <button
                    onClick={handleDisconnectCoinbase}
                    disabled={connSaving === 'coinbase'}
                    className="btn btn-sm btn-outline"
                    style={{ fontSize: '0.7rem', padding: '0.2rem 0.5rem' }}
                  >
                    {connSaving === 'coinbase' ? '...' : 'Disconnect'}
                  </button>
                </div>
              ) : (
                <span className="connector-badge badge-disconnected">Not Connected</span>
              )}
            </div>
            <div className="connector-item">
              <span className={`connector-dot ${authStatus.schwab ? 'connected' : 'disconnected'}`} />
              <span className="connector-label">Schwab</span>
              {authStatus.schwab ? (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span className="connector-badge badge-connected">Connected</span>
                  <button
                    onClick={handleDisconnectSchwab}
                    disabled={connSaving === 'schwab'}
                    className="btn btn-sm btn-outline"
                    style={{ fontSize: '0.7rem', padding: '0.2rem 0.5rem' }}
                  >
                    {connSaving === 'schwab' ? '...' : 'Disconnect'}
                  </button>
                </div>
              ) : (
                <a href="/api/auth/schwab" target="_blank" rel="noopener noreferrer" className="btn btn-sm btn-outline">Connect Schwab</a>
              )}
            </div>
            <div className="connector-item">
              <span className={`connector-dot ${authStatus.polymarket ? 'connected' : 'disconnected'}`} />
              <span className="connector-label">Polymarket</span>
              {authStatus.polymarket ? (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span className="connector-badge badge-connected" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.7rem' }}>
                    {authStatus.polymarket_wallet.slice(0, 6)}...{authStatus.polymarket_wallet.slice(-4)}
                  </span>
                  <button
                    onClick={handleDisconnectPolymarket}
                    disabled={polySaving}
                    className="btn btn-sm btn-outline"
                    style={{ fontSize: '0.7rem', padding: '0.2rem 0.5rem' }}
                  >
                    Disconnect
                  </button>
                </div>
              ) : (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                  <button onClick={handleConnectMetaMask} disabled={polySaving} className="btn btn-sm btn-outline" style={{ fontSize: '0.7rem', background: 'linear-gradient(135deg, #f6851b, #e2761b)', color: 'white', border: 'none' }}>
                    MetaMask
                  </button>
                  <span style={{ fontSize: '0.7rem', color: 'var(--text-dim)' }}>or</span>
                  <input
                    type="text"
                    value={polyWalletInput}
                    onChange={e => setPolyWalletInput(e.target.value)}
                    placeholder="0x..."
                    style={{
                      width: '140px',
                      padding: '0.2rem 0.4rem',
                      fontSize: '0.7rem',
                      borderRadius: 'var(--radius-sm)',
                      border: '1px solid var(--border-default)',
                      backgroundColor: 'var(--bg-surface)',
                      color: 'var(--text-primary)',
                      fontFamily: 'var(--font-mono)',
                    }}
                    onKeyDown={e => e.key === 'Enter' && handleConnectPolymarket()}
                  />
                  <button
                    onClick={handleConnectPolymarket}
                    disabled={polySaving || !polyWalletInput.trim()}
                    className="btn btn-sm btn-outline"
                    style={{ fontSize: '0.7rem', padding: '0.2rem 0.5rem' }}
                  >
                    {polySaving ? '...' : 'Save'}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {metrics && metrics.position_count > 0 && (
        <div className="animate-in animate-in-1">
          <PortfolioSummary metrics={metrics} alertsTriggered={lastSync?.alerts_triggered} />
        </div>
      )}

      <div className={lastSync?.diff ? 'grid-2' : ''} style={{ marginBottom: '1.5rem' }}>
        {metrics && metrics.position_count > 0 && (
          <div className="animate-in animate-in-2">
            <AllocationChart exposure={metrics.exposure_by_type} />
          </div>
        )}
        {lastSync?.diff && (
          <div className="animate-in animate-in-3">
            <DiffView diff={lastSync.diff} />
          </div>
        )}
      </div>

      {metrics && metrics.position_count > 0 && (
        <div className="animate-in animate-in-4">
          <InsightsPanel />
        </div>
      )}

      {positions.length > 0 && (
        <div className="animate-in animate-in-5">
          <PositionsTable positions={positions} />
        </div>
      )}

      {(!metrics || metrics.position_count === 0) && positions.length === 0 && !error && (
        <div className="card animate-in">
          <div className="empty-state">
            <div className="empty-state-icon">&#x25C8;</div>
            <p style={{ marginBottom: '0.5rem' }}>No portfolio data yet</p>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
              Click <strong style={{ color: 'var(--accent-green)' }}>Sync Portfolio</strong> to pull your positions from Schwab & Coinbase
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

function SyncIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 2v6h-6" /><path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
      <path d="M3 22v-6h6" /><path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
    </svg>
  );
}
