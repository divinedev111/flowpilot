import { useState, useEffect } from 'react';
import { api } from '../api/client';
import type { AuditEvent } from '../types';

const EVENT_TYPES = [
  { value: '', label: 'All Events' },
  { value: 'api.request', label: 'API Requests' },
  { value: 'auth.initiate', label: 'Auth Initiate' },
  { value: 'auth.callback', label: 'Auth Callback' },
  { value: 'portfolio.sync', label: 'Portfolio Sync' },
];

export default function AuditLog() {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [eventType, setEventType] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchLogs = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await api.getAuditLogs(eventType || undefined, 100);
      setEvents(res.events || []);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchLogs(); }, [eventType]);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h3 style={{ margin: 0 }}>Audit Log</h3>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <select
            value={eventType}
            onChange={e => setEventType(e.target.value)}
            style={{ padding: '0.375rem 0.5rem', borderRadius: '0.375rem', border: '1px solid #d1d5db', fontSize: '0.875rem' }}
          >
            {EVENT_TYPES.map(t => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
          <button
            onClick={fetchLogs}
            disabled={loading}
            style={{
              padding: '0.375rem 0.75rem',
              backgroundColor: '#2563eb',
              color: 'white',
              border: 'none',
              borderRadius: '0.375rem',
              cursor: loading ? 'not-allowed' : 'pointer',
              fontSize: '0.875rem',
            }}
          >
            {loading ? 'Loading...' : 'Refresh'}
          </button>
        </div>
      </div>

      {error && (
        <div style={{ padding: '0.75rem', backgroundColor: '#fee2e2', color: '#dc2626', borderRadius: '0.5rem', marginBottom: '1rem' }}>
          {error}
        </div>
      )}

      {events.length === 0 && !loading ? (
        <p style={{ color: '#9ca3af' }}>No audit events found</p>
      ) : (
        <div style={{ maxHeight: '500px', overflow: 'auto', border: '1px solid #e5e7eb', borderRadius: '0.5rem' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8125rem' }}>
            <thead>
              <tr style={{ backgroundColor: '#f9fafb', position: 'sticky', top: 0 }}>
                <th style={thStyle}>Timestamp</th>
                <th style={thStyle}>Event</th>
                <th style={thStyle}>Action</th>
                <th style={thStyle}>Resource</th>
                <th style={thStyle}>Status</th>
                <th style={thStyle}>IP</th>
              </tr>
            </thead>
            <tbody>
              {events.map((evt, i) => (
                <tr key={evt.id || i} style={{ borderBottom: '1px solid #f3f4f6' }}>
                  <td style={tdStyle}>{new Date(evt.timestamp).toLocaleString()}</td>
                  <td style={tdStyle}>
                    <span style={{
                      padding: '0.125rem 0.375rem',
                      borderRadius: '0.25rem',
                      backgroundColor: eventColor(evt.event_type),
                      color: 'white',
                      fontSize: '0.75rem',
                    }}>
                      {evt.event_type}
                    </span>
                  </td>
                  <td style={tdStyle}>{evt.action}</td>
                  <td style={tdStyle}>{evt.resource}{evt.resource_id ? ` (${evt.resource_id.slice(0, 8)}...)` : ''}</td>
                  <td style={tdStyle}>
                    <span style={{ color: evt.success ? '#059669' : '#dc2626', fontWeight: 'bold' }}>
                      {evt.success ? 'OK' : 'FAIL'}
                    </span>
                    {evt.error_msg && <div style={{ fontSize: '0.75rem', color: '#dc2626' }}>{evt.error_msg}</div>}
                  </td>
                  <td style={tdStyle}>{evt.ip_address || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function eventColor(eventType: string): string {
  if (eventType.startsWith('auth')) return '#7c3aed';
  if (eventType.startsWith('portfolio')) return '#2563eb';
  if (eventType.startsWith('api')) return '#6b7280';
  return '#6b7280';
}

const thStyle: React.CSSProperties = {
  padding: '0.5rem 0.75rem',
  textAlign: 'left',
  borderBottom: '2px solid #e5e7eb',
  fontWeight: 600,
};

const tdStyle: React.CSSProperties = {
  padding: '0.5rem 0.75rem',
};
