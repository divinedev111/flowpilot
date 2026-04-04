import { useState, useEffect } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import { api } from '../api/client';
import type { AlertRule, AlertEvent, BacktestEvent } from '../types';

const RULE_TYPES = [
  { value: 'exposure_threshold', label: 'Exposure Threshold', desc: 'Alert when asset type allocation exceeds %' },
  { value: 'position_change_percent', label: 'Position Change %', desc: 'Alert when position value changes by more than %' },
  { value: 'concentration_threshold', label: 'Concentration Threshold', desc: 'Alert when single position exceeds % of portfolio' },
  { value: 'net_worth_change', label: 'Net Worth Change', desc: 'Alert when total portfolio value changes by more than %' },
];

type TabType = 'alerts' | 'backtest';

export default function Alerts() {
  const [activeTab, setActiveTab] = useState<TabType>('alerts');
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [events, setEvents] = useState<AlertEvent[]>([]);
  const [ruleType, setRuleType] = useState('exposure_threshold');
  const [threshold, setThreshold] = useState('');
  const [targetAsset, setTargetAsset] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [btSelectedRules, setBtSelectedRules] = useState<Set<number>>(new Set());
  const [btFromDate, setBtFromDate] = useState('');
  const [btToDate, setBtToDate] = useState('');
  const [btLoading, setBtLoading] = useState(false);
  const [btEvents, setBtEvents] = useState<BacktestEvent[]>([]);
  const [btTotalTriggers, setBtTotalTriggers] = useState<number | null>(null);
  const [btSnapshotsAnalyzed, setBtSnapshotsAnalyzed] = useState<number | null>(null);

  const fetchData = async () => {
    try {
      const [rulesRes, eventsRes] = await Promise.all([
        api.getAlertRules(),
        api.getAlertEvents(),
      ]);
      setRules(rulesRes.rules || []);
      setEvents(eventsRes.events || []);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    try {
      await api.createAlertRule({
        rule_type: ruleType,
        threshold: parseFloat(threshold),
        target_asset: targetAsset,
      });
      setSuccess('Rule created');
      setThreshold('');
      setTargetAsset('');
      fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const severityColor: Record<string, string> = {
    critical: '#dc2626',
    high: '#f59e0b',
    medium: '#3b82f6',
  };

  const handleToggleBtRule = (idx: number) => {
    setBtSelectedRules(prev => {
      const next = new Set(prev);
      if (next.has(idx)) {
        next.delete(idx);
      } else {
        next.add(idx);
      }
      return next;
    });
  };

  const handleRunBacktest = async () => {
    setError('');
    setSuccess('');

    if (btSelectedRules.size === 0) {
      setError('Select at least one rule for backtesting');
      return;
    }
    if (!btFromDate || !btToDate) {
      setError('Please select both from and to dates');
      return;
    }

    setBtLoading(true);
    try {
      const selectedRulesList = Array.from(btSelectedRules).map(idx => rules[idx]);
      const res = await api.runBacktest(selectedRulesList, btFromDate, btToDate);
      setBtEvents(res.events || []);
      setBtTotalTriggers(res.total_triggers);
      setBtSnapshotsAnalyzed(res.snapshots_analyzed);
      setSuccess(`Backtest complete: ${res.total_triggers} triggers across ${res.snapshots_analyzed} snapshots`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    } finally {
      setBtLoading(false);
    }
  };

  const chartData = (() => {
    const dateMap: Record<string, number> = {};
    for (const evt of btEvents) {
      const date = evt.snapshot_date.slice(0, 10);
      dateMap[date] = (dateMap[date] || 0) + 1;
    }
    return Object.entries(dateMap)
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([date, count]) => ({ date, triggers: count }));
  })();

  const tabStyle = (tab: TabType): React.CSSProperties => ({
    padding: '0.5rem 1.5rem',
    cursor: 'pointer',
    border: 'none',
    borderBottom: activeTab === tab ? '3px solid #2563eb' : '3px solid transparent',
    backgroundColor: 'transparent',
    fontWeight: activeTab === tab ? 'bold' : 'normal',
    color: activeTab === tab ? '#2563eb' : '#6b7280',
    fontSize: '1rem',
  });

  return (
    <div>
      <h1>Alerts</h1>

      <div style={{ display: 'flex', gap: '0', borderBottom: '1px solid #e5e7eb', marginBottom: '1.5rem' }}>
        <button style={tabStyle('alerts')} onClick={() => setActiveTab('alerts')}>Alert Rules</button>
        <button style={tabStyle('backtest')} onClick={() => setActiveTab('backtest')}>Backtest</button>
      </div>

      {error && <div style={{ padding: '0.75rem', backgroundColor: '#fee2e2', color: '#dc2626', borderRadius: '0.5rem', marginBottom: '1rem' }}>{error}</div>}
      {success && <div style={{ padding: '0.75rem', backgroundColor: '#d1fae5', color: '#059669', borderRadius: '0.5rem', marginBottom: '1rem' }}>{success}</div>}

      {activeTab === 'alerts' && (
        <>
          <div style={{ padding: '1.25rem', borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: 'white', marginBottom: '1.5rem' }}>
            <h3 style={{ marginTop: 0 }}>Create Alert Rule</h3>
            <form onSubmit={handleCreate} style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'flex-end' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Rule Type</label>
                <select value={ruleType} onChange={e => setRuleType(e.target.value)} style={inputStyle}>
                  {RULE_TYPES.map(rt => <option key={rt.value} value={rt.value}>{rt.label}</option>)}
                </select>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Threshold (%)</label>
                <input type="number" step="0.1" value={threshold} onChange={e => setThreshold(e.target.value)} required style={inputStyle} placeholder="e.g. 30" />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Target Asset</label>
                <input type="text" value={targetAsset} onChange={e => setTargetAsset(e.target.value)} style={inputStyle} placeholder="e.g. crypto, AAPL" />
              </div>
              <button type="submit" style={{ padding: '0.5rem 1rem', backgroundColor: '#2563eb', color: 'white', border: 'none', borderRadius: '0.375rem', cursor: 'pointer' }}>
                Add Rule
              </button>
            </form>
            <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.5rem' }}>
              {RULE_TYPES.find(r => r.value === ruleType)?.desc}
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
            <div style={{ padding: '1.25rem', borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: 'white' }}>
              <h3 style={{ marginTop: 0 }}>Active Rules ({rules.length})</h3>
              {rules.length === 0 ? <p style={{ color: '#9ca3af' }}>No rules configured</p> : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {rules.map((rule, i) => (
                    <div key={i} style={{ padding: '0.75rem', borderRadius: '0.5rem', backgroundColor: '#f9fafb', border: '1px solid #f3f4f6' }}>
                      <div style={{ fontWeight: 'bold' }}>{RULE_TYPES.find(r => r.value === rule.rule_type)?.label || rule.rule_type}</div>
                      <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                        Threshold: {rule.threshold}% {rule.target_asset && `| Target: ${rule.target_asset}`}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div style={{ padding: '1.25rem', borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: 'white' }}>
              <h3 style={{ marginTop: 0 }}>Triggered Alerts ({events.length})</h3>
              {events.length === 0 ? <p style={{ color: '#9ca3af' }}>No alerts triggered yet</p> : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', maxHeight: '400px', overflow: 'auto' }}>
                  {events.map((evt, i) => (
                    <div key={i} style={{ padding: '0.75rem', borderRadius: '0.5rem', backgroundColor: '#f9fafb', borderLeft: `4px solid ${severityColor[evt.severity] || '#9ca3af'}` }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <span style={{ fontWeight: 'bold' }}>{evt.message}</span>
                        <span style={{
                          padding: '0.125rem 0.5rem',
                          borderRadius: '0.25rem',
                          fontSize: '0.75rem',
                          color: 'white',
                          backgroundColor: severityColor[evt.severity] || '#9ca3af',
                        }}>
                          {evt.severity}
                        </span>
                      </div>
                      <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.25rem' }}>
                        {evt.evidence} | {new Date(evt.timestamp).toLocaleString()}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </>
      )}

      {activeTab === 'backtest' && (
        <>
          <div style={{ padding: '1.25rem', borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: 'white', marginBottom: '1.5rem' }}>
            <h3 style={{ marginTop: 0 }}>Alert Backtesting</h3>
            <p style={{ fontSize: '0.875rem', color: '#6b7280', marginTop: 0 }}>
              Select alert rules and a date range to see when they would have triggered historically.
            </p>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.5rem', fontWeight: 'bold' }}>Select Rules</label>
              {rules.length === 0 ? (
                <p style={{ color: '#9ca3af', fontSize: '0.875rem' }}>No alert rules available. Create some in the Alert Rules tab first.</p>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {rules.map((rule, idx) => (
                    <label key={idx} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer', fontSize: '0.875rem' }}>
                      <input
                        type="checkbox"
                        checked={btSelectedRules.has(idx)}
                        onChange={() => handleToggleBtRule(idx)}
                      />
                      <span style={{ fontWeight: 'bold' }}>{RULE_TYPES.find(r => r.value === rule.rule_type)?.label || rule.rule_type}</span>
                      <span style={{ color: '#6b7280' }}>
                        — Threshold: {rule.threshold}% {rule.target_asset && `| Target: ${rule.target_asset}`}
                      </span>
                    </label>
                  ))}
                </div>
              )}
            </div>

            <div style={{ display: 'flex', gap: '1rem', alignItems: 'flex-end', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>From Date</label>
                <input type="date" value={btFromDate} onChange={e => setBtFromDate(e.target.value)} style={inputStyle} />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>To Date</label>
                <input type="date" value={btToDate} onChange={e => setBtToDate(e.target.value)} style={inputStyle} />
              </div>
              <button
                onClick={handleRunBacktest}
                disabled={btLoading}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: '#2563eb',
                  color: 'white',
                  border: 'none',
                  borderRadius: '0.375rem',
                  cursor: btLoading ? 'not-allowed' : 'pointer',
                  opacity: btLoading ? 0.7 : 1,
                }}
              >
                {btLoading ? 'Running...' : 'Run Backtest'}
              </button>
            </div>
          </div>

          {btTotalTriggers !== null && (
            <div style={{ padding: '1.25rem', borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: 'white', marginBottom: '1.5rem' }}>
              <h3 style={{ marginTop: 0 }}>Backtest Results</h3>

              <div style={{ display: 'flex', gap: '2rem', marginBottom: '1.5rem' }}>
                <div style={{ padding: '1rem', borderRadius: '0.5rem', backgroundColor: '#f0f9ff', border: '1px solid #bae6fd', flex: 1, textAlign: 'center' }}>
                  <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#0369a1' }}>{btTotalTriggers}</div>
                  <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>Total Triggers</div>
                </div>
                <div style={{ padding: '1rem', borderRadius: '0.5rem', backgroundColor: '#f0f9ff', border: '1px solid #bae6fd', flex: 1, textAlign: 'center' }}>
                  <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#0369a1' }}>{btSnapshotsAnalyzed}</div>
                  <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>Snapshots Analyzed</div>
                </div>
              </div>

              {chartData.length > 0 && (
                <div style={{ marginBottom: '1.5rem' }}>
                  <h4 style={{ marginBottom: '0.5rem' }}>Alert Timeline</h4>
                  <ResponsiveContainer width="100%" height={250}>
                    <BarChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" fontSize={12} />
                      <YAxis allowDecimals={false} fontSize={12} />
                      <Tooltip />
                      <Bar dataKey="triggers" fill="#2563eb" name="Triggers" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}

              {btEvents.length > 0 && (
                <div>
                  <h4 style={{ marginBottom: '0.5rem' }}>Event Details</h4>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', maxHeight: '400px', overflow: 'auto' }}>
                    {btEvents.map((evt, i) => (
                      <div key={i} style={{
                        padding: '0.75rem',
                        borderRadius: '0.5rem',
                        backgroundColor: '#f9fafb',
                        borderLeft: `4px solid ${severityColor[evt.severity] || '#3b82f6'}`,
                      }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <span style={{ fontWeight: 'bold' }}>{evt.message}</span>
                          <span style={{
                            padding: '0.125rem 0.5rem',
                            borderRadius: '0.25rem',
                            fontSize: '0.75rem',
                            color: 'white',
                            backgroundColor: severityColor[evt.severity] || '#3b82f6',
                          }}>
                            {evt.severity}
                          </span>
                        </div>
                        <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.25rem' }}>
                          {evt.rule_type} | {evt.evidence} | {new Date(evt.snapshot_date).toLocaleDateString()}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {btEvents.length === 0 && (
                <p style={{ color: '#9ca3af', textAlign: 'center' }}>No alerts would have triggered in this date range with the selected rules.</p>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: '0.5rem',
  borderRadius: '0.375rem',
  border: '1px solid #d1d5db',
  fontSize: '0.875rem',
};
