import { useState, useEffect } from 'react';
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { api } from '../api/client';
import type { Strategy, Position } from '../types';
import { formatAssetType } from '../utils/format';

const PRESET_COLORS = ['#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316'];

const inputStyle: React.CSSProperties = {
  padding: '0.5rem',
  borderRadius: '0.375rem',
  border: '1px solid #d1d5db',
  fontSize: '0.875rem',
  width: '100%',
  boxSizing: 'border-box',
};

const cardStyle: React.CSSProperties = {
  padding: '1.25rem',
  borderRadius: '0.75rem',
  border: '1px solid #e5e7eb',
  backgroundColor: 'white',
};

export default function Strategies() {
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [expandedPositions, setExpandedPositions] = useState<Position[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formName, setFormName] = useState('');
  const [formDesc, setFormDesc] = useState('');
  const [formColor, setFormColor] = useState(PRESET_COLORS[0]);
  const [assignDialogId, setAssignDialogId] = useState<string | null>(null);
  const [assignSymbol, setAssignSymbol] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const fetchData = async () => {
    try {
      const [stratRes, portRes] = await Promise.all([
        api.getStrategies(),
        api.getPortfolio(),
      ]);
      setStrategies(stratRes.strategies || []);
      setPositions(portRes.positions || []);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await api.createStrategy({ name: formName, description: formDesc, color: formColor });
      setShowCreate(false);
      setFormName('');
      setFormDesc('');
      setFormColor(PRESET_COLORS[0]);
      await fetchData();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingId) return;
    setError('');
    setLoading(true);
    try {
      await api.updateStrategy(editingId, { name: formName, description: formDesc, color: formColor });
      setEditingId(null);
      setFormName('');
      setFormDesc('');
      setFormColor(PRESET_COLORS[0]);
      await fetchData();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    setError('');
    try {
      await api.deleteStrategy(id);
      if (expandedId === id) setExpandedId(null);
      await fetchData();
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleExpand = async (id: string) => {
    if (expandedId === id) {
      setExpandedId(null);
      setExpandedPositions([]);
      return;
    }
    try {
      const detail = await api.getStrategy(id);
      setExpandedId(id);
      setExpandedPositions(detail.positions || []);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleAssign = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!assignDialogId || !assignSymbol) return;
    setError('');
    try {
      await api.assignStrategy(assignDialogId, assignSymbol);
      setAssignDialogId(null);
      setAssignSymbol('');
      await fetchData();
      // Re-expand if needed
      if (expandedId === assignDialogId) {
        const detail = await api.getStrategy(assignDialogId);
        setExpandedPositions(detail.positions || []);
      }
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleUnassign = async (strategyId: string, symbol: string) => {
    setError('');
    try {
      await api.unassignStrategy(strategyId, symbol);
      await fetchData();
      if (expandedId === strategyId) {
        const detail = await api.getStrategy(strategyId);
        setExpandedPositions(detail.positions || []);
      }
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const startEdit = (s: Strategy) => {
    setEditingId(s.id);
    setFormName(s.name);
    setFormDesc(s.description);
    setFormColor(s.color);
    setShowCreate(false);
  };

  const cancelForm = () => {
    setShowCreate(false);
    setEditingId(null);
    setFormName('');
    setFormDesc('');
    setFormColor(PRESET_COLORS[0]);
  };

  const assignedSymbols = new Set(strategies.flatMap(s => s.symbols || []));
  const availableSymbols = positions
    .map(p => p.symbol)
    .filter((sym, i, arr) => arr.indexOf(sym) === i)
    .filter(sym => {
      if (!assignDialogId) return true;
      const strategy = strategies.find(s => s.id === assignDialogId);
      return !strategy?.symbols?.includes(sym);
    });

  const pieData = strategies
    .filter(s => (s.total_value ?? 0) > 0)
    .map(s => ({
      name: s.name,
      value: Math.round((s.allocation_pct ?? 0) * 10) / 10,
      color: s.color,
    }));

  const assignedPct = strategies.reduce((sum, s) => sum + (s.allocation_pct ?? 0), 0);
  if (assignedPct < 100) {
    pieData.push({
      name: 'Unassigned',
      value: Math.round((100 - assignedPct) * 10) / 10,
      color: '#d1d5db',
    });
  }

  const fmt = (n: number) =>
    n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0, maximumFractionDigits: 0 });

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h1 style={{ margin: 0 }}>Strategies</h1>
        <button
          onClick={() => { setShowCreate(true); setEditingId(null); setFormName(''); setFormDesc(''); setFormColor(PRESET_COLORS[0]); }}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: '#2563eb',
            color: 'white',
            border: 'none',
            borderRadius: '0.5rem',
            cursor: 'pointer',
            fontSize: '1rem',
          }}
        >
          + New Strategy
        </button>
      </div>

      {error && (
        <div style={{ padding: '0.75rem', backgroundColor: '#fee2e2', color: '#dc2626', borderRadius: '0.5rem', marginBottom: '1rem' }}>
          {error}
        </div>
      )}

      {(showCreate || editingId) && (
        <div style={{ ...cardStyle, marginBottom: '1.5rem' }}>
          <h3 style={{ marginTop: 0 }}>{editingId ? 'Edit Strategy' : 'Create Strategy'}</h3>
          <form onSubmit={editingId ? handleUpdate : handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Name</label>
              <input
                type="text"
                value={formName}
                onChange={e => setFormName(e.target.value)}
                required
                style={inputStyle}
                placeholder="e.g. Growth, Value, Dividend"
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Description</label>
              <input
                type="text"
                value={formDesc}
                onChange={e => setFormDesc(e.target.value)}
                style={inputStyle}
                placeholder="Brief description of this strategy"
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Color</label>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                {PRESET_COLORS.map(color => (
                  <button
                    key={color}
                    type="button"
                    onClick={() => setFormColor(color)}
                    style={{
                      width: '2rem',
                      height: '2rem',
                      borderRadius: '50%',
                      backgroundColor: color,
                      border: formColor === color ? '3px solid #1e293b' : '2px solid transparent',
                      cursor: 'pointer',
                      padding: 0,
                    }}
                  />
                ))}
              </div>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                type="submit"
                disabled={loading}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: loading ? '#ccc' : '#2563eb',
                  color: 'white',
                  border: 'none',
                  borderRadius: '0.375rem',
                  cursor: loading ? 'not-allowed' : 'pointer',
                }}
              >
                {editingId ? 'Save Changes' : 'Create'}
              </button>
              <button
                type="button"
                onClick={cancelForm}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: 'transparent',
                  color: '#6b7280',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  cursor: 'pointer',
                }}
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {assignDialogId && (
        <div style={{ ...cardStyle, marginBottom: '1.5rem' }}>
          <h3 style={{ marginTop: 0 }}>
            Assign Symbol to {strategies.find(s => s.id === assignDialogId)?.name}
          </h3>
          <form onSubmit={handleAssign} style={{ display: 'flex', gap: '1rem', alignItems: 'flex-end' }}>
            <div style={{ flex: 1 }}>
              <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Symbol</label>
              <select
                value={assignSymbol}
                onChange={e => setAssignSymbol(e.target.value)}
                required
                style={inputStyle}
              >
                <option value="">Select a symbol...</option>
                {availableSymbols.map(sym => (
                  <option key={sym} value={sym}>
                    {sym} {assignedSymbols.has(sym) ? '(assigned elsewhere)' : ''}
                  </option>
                ))}
              </select>
            </div>
            <button
              type="submit"
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: '#2563eb',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                cursor: 'pointer',
              }}
            >
              Assign
            </button>
            <button
              type="button"
              onClick={() => { setAssignDialogId(null); setAssignSymbol(''); }}
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: 'transparent',
                color: '#6b7280',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
          </form>
        </div>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1.5rem' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {strategies.length === 0 ? (
            <div style={{ ...cardStyle, color: '#9ca3af', textAlign: 'center', padding: '3rem' }}>
              No strategies yet. Create one to get started.
            </div>
          ) : (
            strategies.map(s => (
              <div
                key={s.id}
                style={{
                  ...cardStyle,
                  borderLeft: `4px solid ${s.color}`,
                  cursor: 'pointer',
                }}
              >
                <div onClick={() => handleExpand(s.id)}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                      <h3 style={{ margin: 0 }}>{s.name}</h3>
                      {s.description && (
                        <p style={{ margin: '0.25rem 0 0', fontSize: '0.875rem', color: '#6b7280' }}>
                          {s.description}
                        </p>
                      )}
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <div style={{ fontWeight: 'bold', fontSize: '1.25rem' }}>{fmt(s.total_value ?? 0)}</div>
                      <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                        {(s.allocation_pct ?? 0).toFixed(1)}% of portfolio
                      </div>
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: '1rem', marginTop: '0.75rem', alignItems: 'center' }}>
                    <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                      {s.position_count ?? 0} position{(s.position_count ?? 0) !== 1 ? 's' : ''}
                    </span>
                    <div style={{ display: 'flex', gap: '0.375rem', flexWrap: 'wrap' }}>
                      {(s.symbols || []).map(sym => (
                        <span
                          key={sym}
                          className="strategy-pill"
                          style={{ backgroundColor: s.color + '1a', color: s.color, borderColor: s.color + '40' }}
                        >
                          {sym}
                          <button
                            className="strategy-pill-remove"
                            onClick={e => { e.stopPropagation(); handleUnassign(s.id, sym); }}
                          >
                            x
                          </button>
                        </span>
                      ))}
                    </div>
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.75rem', borderTop: '1px solid #f3f4f6', paddingTop: '0.75rem' }}>
                  <button
                    onClick={e => { e.stopPropagation(); setAssignDialogId(s.id); setAssignSymbol(''); }}
                    className="strategy-action-btn"
                  >
                    + Assign Symbol
                  </button>
                  <button
                    onClick={e => { e.stopPropagation(); startEdit(s); }}
                    className="strategy-action-btn"
                  >
                    Edit
                  </button>
                  <button
                    onClick={e => { e.stopPropagation(); handleDelete(s.id); }}
                    className="strategy-action-btn strategy-action-btn-danger"
                  >
                    Delete
                  </button>
                </div>

                {expandedId === s.id && (
                  <div style={{ marginTop: '1rem', borderTop: '1px solid #f3f4f6', paddingTop: '1rem' }}>
                    <h4 style={{ margin: '0 0 0.5rem' }}>Positions</h4>
                    {expandedPositions.length === 0 ? (
                      <p style={{ color: '#9ca3af', fontSize: '0.875rem' }}>No positions assigned yet.</p>
                    ) : (
                      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead>
                          <tr>
                            <th style={thStyle}>Symbol</th>
                            <th style={thStyle}>Quantity</th>
                            <th style={thStyle}>Price</th>
                            <th style={thStyle}>Market Value</th>
                            <th style={thStyle}>Type</th>
                          </tr>
                        </thead>
                        <tbody>
                          {expandedPositions.map((p, i) => (
                            <tr key={i}>
                              <td style={tdStyle}>{p.symbol}</td>
                              <td style={tdStyle}>{p.quantity}</td>
                              <td style={tdStyle}>{fmt(p.mark_price)}</td>
                              <td style={tdStyle}>{fmt(p.market_value)}</td>
                              <td style={tdStyle}>{formatAssetType(p.asset_type)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    )}
                  </div>
                )}
              </div>
            ))
          )}
        </div>

        <div style={cardStyle}>
          <h3 style={{ marginTop: 0 }}>Allocation Breakdown</h3>
          {pieData.length === 0 ? (
            <p style={{ color: '#9ca3af', textAlign: 'center' }}>No allocation data</p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  dataKey="value"
                  nameKey="name"
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  label={({ name, value }) => `${name} ${value}%`}
                >
                  {pieData.map((entry, i) => (
                    <Cell key={i} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip formatter={(value) => `${value}%`} />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  );
}

const thStyle: React.CSSProperties = {
  padding: '0.5rem 0.75rem',
  textAlign: 'left',
  borderBottom: '2px solid #e5e7eb',
  fontSize: '0.75rem',
  color: '#6b7280',
  textTransform: 'uppercase',
};

const tdStyle: React.CSSProperties = {
  padding: '0.5rem 0.75rem',
  borderBottom: '1px solid #f3f4f6',
  fontSize: '0.875rem',
};
