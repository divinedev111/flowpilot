import { useState, useEffect } from 'react';
import { api } from '../api/client';
import type { PolicyRule, PolicyViolation } from '../types';
import { formatAssetType } from '../utils/format';

const RULE_TYPES = [
  { value: 'max_concentration', label: 'Max Concentration', desc: 'No single position may exceed this % of portfolio value' },
  { value: 'min_positions', label: 'Min Positions', desc: 'Portfolio must have at least this many positions' },
  { value: 'max_exposure', label: 'Max Exposure', desc: 'No asset type may exceed this % of portfolio value' },
  { value: 'max_position_count', label: 'Max Position Count', desc: 'Portfolio must not exceed this many positions' },
];

const SEVERITIES = [
  { value: 'warning', label: 'Warning' },
  { value: 'critical', label: 'Critical' },
];

export default function Policies() {
  const [rules, setRules] = useState<PolicyRule[]>([]);
  const [violations, setViolations] = useState<PolicyViolation[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingRule, setEditingRule] = useState<PolicyRule | null>(null);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [complianceStatus, setComplianceStatus] = useState<string | null>(null);
  const [checkLoading, setCheckLoading] = useState(false);

  const [formName, setFormName] = useState('');
  const [formRuleType, setFormRuleType] = useState('max_concentration');
  const [formThreshold, setFormThreshold] = useState('');
  const [formAssetType, setFormAssetType] = useState('');
  const [formTargetSymbol, setFormTargetSymbol] = useState('');
  const [formSeverity, setFormSeverity] = useState('warning');

  const fetchData = async () => {
    try {
      const res = await api.getPolicies();
      setRules(res.rules || []);
      setViolations(res.violations || []);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const resetForm = () => {
    setFormName('');
    setFormRuleType('max_concentration');
    setFormThreshold('');
    setFormAssetType('');
    setFormTargetSymbol('');
    setFormSeverity('warning');
    setEditingRule(null);
  };

  const openCreate = () => {
    resetForm();
    setShowModal(true);
  };

  const openEdit = (rule: PolicyRule) => {
    setEditingRule(rule);
    setFormName(rule.name);
    setFormRuleType(rule.rule_type);
    setFormThreshold(String(rule.threshold_pct));
    setFormAssetType(rule.asset_type || '');
    setFormTargetSymbol(rule.target_symbol || '');
    setFormSeverity(rule.severity);
    setShowModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    const payload: PolicyRule = {
      name: formName,
      rule_type: formRuleType,
      threshold_pct: parseFloat(formThreshold),
      asset_type: formAssetType || undefined,
      target_symbol: formTargetSymbol || undefined,
      severity: formSeverity,
      enabled: true,
    };

    try {
      if (editingRule?.id) {
        await api.updatePolicy(editingRule.id, { ...payload, enabled: editingRule.enabled });
        setSuccess('Policy rule updated');
      } else {
        await api.createPolicy(payload);
        setSuccess('Policy rule created');
      }
      setShowModal(false);
      resetForm();
      fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    }
  };

  const handleToggle = async (rule: PolicyRule) => {
    if (!rule.id) return;
    try {
      await api.updatePolicy(rule.id, { ...rule, enabled: !rule.enabled });
      fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deletePolicy(id);
      setSuccess('Policy rule deleted');
      fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    }
  };

  const handleCheckCompliance = async () => {
    setCheckLoading(true);
    setError('');
    try {
      const res = await api.checkPolicies();
      setViolations(res.violations || []);
      setComplianceStatus(res.status);
      setSuccess(res.status === 'passing' ? 'All policies passing' : `${(res.violations || []).length} violation(s) detected`);
      fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
    } finally {
      setCheckLoading(false);
    }
  };

  const severityColor: Record<string, string> = {
    critical: '#dc2626',
    warning: '#f59e0b',
  };

  const activeViolations = violations.filter(v => !v.resolved);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h1 style={{ margin: 0 }}>Portfolio Policies</h1>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          {complianceStatus !== null && (
            <span style={{
              padding: '0.375rem 0.75rem',
              borderRadius: '999px',
              fontSize: '0.875rem',
              fontWeight: 'bold',
              color: 'white',
              backgroundColor: complianceStatus === 'passing' ? '#059669' : '#dc2626',
            }}>
              {complianceStatus === 'passing' ? 'Passing' : 'Violations Detected'}
            </span>
          )}
          <button
            onClick={handleCheckCompliance}
            disabled={checkLoading}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: '#059669',
              color: 'white',
              border: 'none',
              borderRadius: '0.375rem',
              cursor: checkLoading ? 'not-allowed' : 'pointer',
              opacity: checkLoading ? 0.7 : 1,
            }}
          >
            {checkLoading ? 'Checking...' : 'Check Compliance'}
          </button>
          <button
            onClick={openCreate}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: '#2563eb',
              color: 'white',
              border: 'none',
              borderRadius: '0.375rem',
              cursor: 'pointer',
            }}
          >
            + Add Rule
          </button>
        </div>
      </div>

      {error && <div style={{ padding: '0.75rem', backgroundColor: '#fee2e2', color: '#dc2626', borderRadius: '0.5rem', marginBottom: '1rem' }}>{error}</div>}
      {success && <div style={{ padding: '0.75rem', backgroundColor: '#d1fae5', color: '#059669', borderRadius: '0.5rem', marginBottom: '1rem' }}>{success}</div>}

      {activeViolations.length > 0 && (
        <div style={{ marginBottom: '1.5rem' }}>
          <h3>Current Violations ({activeViolations.length})</h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '1rem' }}>
            {activeViolations.map((v, i) => (
              <div key={i} style={{
                padding: '1rem',
                borderRadius: '0.75rem',
                backgroundColor: 'white',
                border: '1px solid #e5e7eb',
                borderLeft: `4px solid ${severityColor[v.severity] || '#9ca3af'}`,
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                  <span style={{ fontWeight: 'bold' }}>{v.rule_name}</span>
                  <span style={{
                    padding: '0.125rem 0.5rem',
                    borderRadius: '0.25rem',
                    fontSize: '0.75rem',
                    color: 'white',
                    backgroundColor: severityColor[v.severity] || '#9ca3af',
                  }}>
                    {v.severity}
                  </span>
                </div>
                <div style={{ fontSize: '0.875rem', color: '#374151' }}>{v.message}</div>
                <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.5rem' }}>
                  Current: {v.current_value.toFixed(1)}% | Threshold: {v.threshold.toFixed(1)}%
                  {v.symbols && v.symbols.length > 0 && ` | Symbols: ${v.symbols.join(', ')}`}
                </div>
                <div style={{ fontSize: '0.75rem', color: '#9ca3af' }}>
                  {new Date(v.timestamp).toLocaleString()}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div style={{ padding: '1.25rem', borderRadius: '0.75rem', border: '1px solid #e5e7eb', backgroundColor: 'white' }}>
        <h3 style={{ marginTop: 0 }}>Policy Rules ({rules.length})</h3>
        {rules.length === 0 ? <p style={{ color: '#9ca3af' }}>No policy rules configured</p> : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {rules.map((rule, i) => (
              <div key={i} style={{
                padding: '0.75rem',
                borderRadius: '0.5rem',
                backgroundColor: rule.enabled ? '#f9fafb' : '#f3f4f6',
                border: '1px solid #f3f4f6',
                opacity: rule.enabled ? 1 : 0.6,
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 'bold' }}>{rule.name || RULE_TYPES.find(r => r.value === rule.rule_type)?.label || rule.rule_type}</div>
                  <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                    {RULE_TYPES.find(r => r.value === rule.rule_type)?.label} | Threshold: {rule.threshold_pct}%
                    {rule.asset_type && ` | Asset Type: ${formatAssetType(rule.asset_type)}`}
                    {rule.target_symbol && ` | Symbol: ${rule.target_symbol}`}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: '#9ca3af' }}>
                    Severity: {rule.severity}
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                  <label className="policy-toggle" style={{ position: 'relative', display: 'inline-block', width: '44px', height: '24px', cursor: 'pointer' }}>
                    <input
                      type="checkbox"
                      checked={rule.enabled}
                      onChange={() => handleToggle(rule)}
                      style={{ opacity: 0, width: 0, height: 0 }}
                    />
                    <span style={{
                      position: 'absolute',
                      top: 0, left: 0, right: 0, bottom: 0,
                      backgroundColor: rule.enabled ? '#059669' : '#d1d5db',
                      borderRadius: '12px',
                      transition: 'background-color 0.2s',
                    }}>
                      <span style={{
                        position: 'absolute',
                        content: '""',
                        height: '18px',
                        width: '18px',
                        left: rule.enabled ? '22px' : '3px',
                        bottom: '3px',
                        backgroundColor: 'white',
                        borderRadius: '50%',
                        transition: 'left 0.2s',
                      }} />
                    </span>
                  </label>
                  <button
                    onClick={() => openEdit(rule)}
                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem', backgroundColor: '#f3f4f6', border: '1px solid #d1d5db', borderRadius: '0.25rem', cursor: 'pointer' }}
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => rule.id && handleDelete(rule.id)}
                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem', backgroundColor: '#fee2e2', color: '#dc2626', border: '1px solid #fecaca', borderRadius: '0.25rem', cursor: 'pointer' }}
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {showModal && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1000,
        }}>
          <div style={{
            backgroundColor: 'white', borderRadius: '0.75rem', padding: '1.5rem',
            width: '480px', maxWidth: '90vw', color: '#1f2937',
          }}>
            <h3 style={{ marginTop: 0 }}>{editingRule ? 'Edit Policy Rule' : 'Create Policy Rule'}</h3>
            <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Name</label>
                <input type="text" value={formName} onChange={e => setFormName(e.target.value)} required style={inputStyle} placeholder="e.g. Max Crypto Exposure" />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Rule Type</label>
                <select value={formRuleType} onChange={e => setFormRuleType(e.target.value)} style={inputStyle}>
                  {RULE_TYPES.map(rt => <option key={rt.value} value={rt.value}>{rt.label}</option>)}
                </select>
                <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.25rem' }}>
                  {RULE_TYPES.find(r => r.value === formRuleType)?.desc}
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>
                  {formRuleType === 'min_positions' || formRuleType === 'max_position_count' ? 'Count' : 'Threshold (%)'}
                </label>
                <input type="number" step="0.1" value={formThreshold} onChange={e => setFormThreshold(e.target.value)} required style={inputStyle} placeholder="e.g. 30" />
              </div>
              {formRuleType === 'max_exposure' && (
                <div>
                  <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Asset Type</label>
                  <input type="text" value={formAssetType} onChange={e => setFormAssetType(e.target.value)} style={inputStyle} placeholder="e.g. crypto, stock" />
                </div>
              )}
              {formRuleType === 'max_concentration' && (
                <div>
                  <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Target Symbol (optional)</label>
                  <input type="text" value={formTargetSymbol} onChange={e => setFormTargetSymbol(e.target.value)} style={inputStyle} placeholder="e.g. AAPL (blank for all)" />
                </div>
              )}
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', color: '#6b7280', marginBottom: '0.25rem' }}>Severity</label>
                <select value={formSeverity} onChange={e => setFormSeverity(e.target.value)} style={inputStyle}>
                  {SEVERITIES.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
                </select>
              </div>
              <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'flex-end' }}>
                <button
                  type="button"
                  onClick={() => { setShowModal(false); resetForm(); }}
                  style={{ padding: '0.5rem 1rem', backgroundColor: '#f3f4f6', border: '1px solid #d1d5db', borderRadius: '0.375rem', cursor: 'pointer', color: '#374151' }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  style={{ padding: '0.5rem 1rem', backgroundColor: '#2563eb', color: 'white', border: 'none', borderRadius: '0.375rem', cursor: 'pointer' }}
                >
                  {editingRule ? 'Update' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: '0.5rem',
  borderRadius: '0.375rem',
  border: '1px solid #d1d5db',
  fontSize: '0.875rem',
  width: '100%',
  boxSizing: 'border-box',
};
