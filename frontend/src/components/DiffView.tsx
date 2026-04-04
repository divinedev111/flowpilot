import type { Diff } from '../types';

interface Props {
  diff: Diff;
}

export default function DiffView({ diff }: Props) {
  const positive = diff.net_worth_change >= 0;

  return (
    <div className="card" style={{ height: '100%' }}>
      <div className="card-header">
        <span className="card-title">Recent Changes</span>
      </div>

      <div style={{
        padding: '1rem 1.25rem',
        borderRadius: 'var(--radius-md)',
        background: positive ? 'var(--accent-green-dim)' : 'var(--accent-red-dim)',
        marginBottom: '1.25rem',
      }}>
        <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '0.25rem', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
          Net Worth Change
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <span className="mono" style={{
            fontSize: '1.5rem',
            fontWeight: 700,
            color: positive ? 'var(--accent-green)' : 'var(--accent-red)',
          }}>
            {positive ? '+' : ''}${diff.net_worth_change.toLocaleString(undefined, { minimumFractionDigits: 2 })}
          </span>
          <span className={`change-badge ${positive ? 'positive' : 'negative'}`}>
            {positive ? '\u25B2' : '\u25BC'} {positive ? '+' : ''}{diff.net_worth_change !== 0 ? ((diff.net_worth_change / Math.abs(diff.net_worth_change)) * 100).toFixed(1) : '0.0'}%
          </span>
        </div>
      </div>

      {diff.top_movers && diff.top_movers.length > 0 && (
        <div style={{ marginBottom: '1rem' }}>
          <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '0.5rem', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 600 }}>
            Top Movers
          </div>
          {diff.top_movers.map((m, i) => (
            <div key={i} style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              padding: '0.5rem 0',
              borderBottom: i < diff.top_movers.length - 1 ? '1px solid var(--border-subtle)' : 'none',
            }}>
              <span className="mono" style={{ fontWeight: 600, fontSize: '0.85rem' }}>{m.symbol}</span>
              <span className={`change-badge ${m.change_percent >= 0 ? 'positive' : 'negative'}`}>
                {m.change_percent >= 0 ? '\u25B2' : '\u25BC'} {m.change_percent >= 0 ? '+' : ''}{m.change_percent.toFixed(1)}%
              </span>
            </div>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: '1rem' }}>
        {diff.added_symbols && diff.added_symbols.length > 0 && (
          <div>
            <div style={{ fontSize: '0.65rem', color: 'var(--accent-green)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: '0.35rem', fontWeight: 600 }}>
              + Added
            </div>
            <div style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
              {diff.added_symbols.map(s => (
                <span key={s} className="change-badge positive">{s}</span>
              ))}
            </div>
          </div>
        )}
        {diff.removed_symbols && diff.removed_symbols.length > 0 && (
          <div>
            <div style={{ fontSize: '0.65rem', color: 'var(--accent-red)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: '0.35rem', fontWeight: 600 }}>
              - Removed
            </div>
            <div style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
              {diff.removed_symbols.map(s => (
                <span key={s} className="change-badge negative">{s}</span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
