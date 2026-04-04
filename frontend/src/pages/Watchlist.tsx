import { useState, useEffect } from 'react';
import { api } from '../api/client';
import type { WatchlistItem } from '../types';

const ASSET_TYPES = [
  { value: 'stock', label: 'Stock' },
  { value: 'crypto', label: 'Crypto' },
  { value: 'prediction', label: 'Prediction Market' },
];

function assetTypeBadgeClass(type: string): string {
  switch (type) {
    case 'crypto': return 'badge badge-crypto';
    case 'stock': return 'badge badge-equity';
    case 'prediction': return 'badge badge-high';
    default: return 'badge';
  }
}

export default function Watchlist() {
  const [items, setItems] = useState<WatchlistItem[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [symbol, setSymbol] = useState('');
  const [name, setName] = useState('');
  const [assetType, setAssetType] = useState('stock');
  const [notes, setNotes] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const fetchWatchlist = async () => {
    try {
      const res = await api.getWatchlist();
      setItems(res.items || []);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await api.addToWatchlist({
        symbol: symbol.toUpperCase(),
        name,
        asset_type: assetType,
        notes,
      });
      setSymbol('');
      setName('');
      setAssetType('stock');
      setNotes('');
      setShowForm(false);
      fetchWatchlist();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemove = async (id: string) => {
    try {
      await api.removeFromWatchlist(id);
      setItems(prev => prev.filter(item => item.id !== id));
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  useEffect(() => { fetchWatchlist(); }, []);

  return (
    <div>
      <div className="page-header animate-in">
        <div>
          <h1 className="page-title">Watchlist</h1>
          <p className="page-subtitle">Track symbols you are interested in</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => setShowForm(f => !f)}
        >
          {showForm ? 'Cancel' : '+ Add Symbol'}
        </button>
      </div>

      {error && <div className="notice notice-error animate-in">{error}</div>}

      {showForm && (
        <div className="card animate-in" style={{ marginBottom: '1.5rem' }}>
          <div className="card-header">
            <span className="card-title">Add to Watchlist</span>
          </div>
          <form onSubmit={handleAdd} style={{ display: 'flex', gap: '1rem', alignItems: 'flex-end', flexWrap: 'wrap' }}>
            <div className="form-group" style={{ minWidth: '120px' }}>
              <label className="form-label">Symbol</label>
              <input
                className="form-input"
                type="text"
                value={symbol}
                onChange={e => setSymbol(e.target.value)}
                required
                placeholder="NVDA"
              />
            </div>
            <div className="form-group" style={{ minWidth: '180px', flex: 1 }}>
              <label className="form-label">Name</label>
              <input
                className="form-input"
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                required
                placeholder="NVIDIA Corporation"
              />
            </div>
            <div className="form-group" style={{ minWidth: '160px' }}>
              <label className="form-label">Asset Type</label>
              <select
                className="form-select"
                value={assetType}
                onChange={e => setAssetType(e.target.value)}
              >
                {ASSET_TYPES.map(at => (
                  <option key={at.value} value={at.value}>{at.label}</option>
                ))}
              </select>
            </div>
            <div className="form-group" style={{ minWidth: '200px', flex: 1 }}>
              <label className="form-label">Notes (optional)</label>
              <input
                className="form-input"
                type="text"
                value={notes}
                onChange={e => setNotes(e.target.value)}
                placeholder="Potential buy at $800"
              />
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button type="submit" className="btn btn-primary" disabled={submitting}>
                {submitting ? 'Saving...' : 'Save'}
              </button>
              <button type="button" className="btn btn-secondary" onClick={() => setShowForm(false)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="empty-state">
          <div className="spinner" style={{ margin: '0 auto' }} />
        </div>
      ) : items.length === 0 ? (
        <div className="card animate-in">
          <div className="empty-state">
            <div className="empty-state-icon">*</div>
            <p className="text-muted">No items in your watchlist. Add symbols to track.</p>
          </div>
        </div>
      ) : (
        <div className="grid-3 animate-in animate-in-1">
          {items.map(item => (
            <div key={item.id} className="card watchlist-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <div style={{ fontWeight: 700, fontSize: '1.1rem', letterSpacing: '-0.01em' }}>
                    {item.symbol}
                  </div>
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '2px' }}>
                    {item.name}
                  </div>
                </div>
                <button
                  className="btn btn-ghost"
                  onClick={() => handleRemove(item.id)}
                  title="Remove from watchlist"
                  style={{ padding: '0.25rem 0.4rem', fontSize: '0.9rem', lineHeight: 1 }}
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                </button>
              </div>

              <div style={{ marginTop: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <span className={assetTypeBadgeClass(item.asset_type)}>
                  {ASSET_TYPES.find(at => at.value === item.asset_type)?.label || item.asset_type}
                </span>
              </div>

              {item.notes && (
                <div style={{
                  marginTop: '0.75rem',
                  fontSize: '0.8rem',
                  color: 'var(--text-secondary)',
                  lineHeight: 1.5,
                }}>
                  {item.notes}
                </div>
              )}

              <div style={{
                marginTop: '0.75rem',
                paddingTop: '0.6rem',
                borderTop: '1px solid var(--border-subtle)',
                fontSize: '0.7rem',
                color: 'var(--text-muted)',
                fontFamily: 'var(--font-mono)',
              }}>
                Added {new Date(item.added_at).toLocaleDateString()}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
