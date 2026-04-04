import type { AIDigest, InsightsResult, NewsFeed, WatchlistItem, Note, Strategy, PolicyRule, BacktestResult, CorrelationResult, AuditEvent, PredictionMarket, WalletPosition, WalletSummary } from '../types';

const API_BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  return res.json();
}

export const api = {
  getPortfolio: () => request<{ positions: any[]; metrics: any }>('/portfolio'),
  sync: () => request<any>('/sync', { method: 'POST' }),

  getSnapshots: () => request<{ snapshots: any[] }>('/snapshots'),
  getSnapshotDiff: (id: string) => request<{ diff: any }>(`/snapshots/${id}/diff`),

  getAlertRules: () => request<{ rules: any[] }>('/alerts/rules'),
  createAlertRule: (rule: any) =>
    request<{ rule: any }>('/alerts/rules', {
      method: 'POST',
      body: JSON.stringify(rule),
    }),
  getAlertEvents: () => request<{ events: any[] }>('/alerts/events'),

  chat: async (message: string, conversationId: string) => {
    const res = await fetch(`${API_BASE}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message, conversation_id: conversationId }),
    });
    return res;
  },

  runScenario: (adjustments: { asset_type: string; target_pct: number }[]) =>
    request<any>('/scenarios', {
      method: 'POST',
      body: JSON.stringify({ adjustments }),
    }),

  getRisk: () => request<any>('/risk'),

  getPerformance: () => request<any>('/performance'),

  getWatchlist: () => request<{ items: WatchlistItem[] }>('/watchlist'),
  addToWatchlist: (item: Omit<WatchlistItem, 'id' | 'added_at'>) =>
    request<{ item: WatchlistItem }>('/watchlist', {
      method: 'POST',
      body: JSON.stringify(item),
    }),
  removeFromWatchlist: (id: string) =>
    request<void>(`/watchlist/${id}`, { method: 'DELETE' }),

  getLatestDigest: () => request<{ digest: AIDigest }>('/digest/latest'),
  getDigestHistory: () => request<{ digests: AIDigest[] }>('/digest/history'),
  getInsights: () => request<InsightsResult>('/insights', { method: 'POST' }),

  getNews: () => request<NewsFeed>('/news'),
  getSymbolNews: (symbol: string) => request<NewsFeed>(`/news/symbol/${symbol}`),

  getAuthStatus: () => request<{ schwab: boolean; coinbase: boolean; polymarket: boolean; polymarket_wallet: string }>('/auth/status'),

  getNotes: () => request<{ notes: Note[] }>('/notes'),
  getNote: (id: string) => request<{ note: Note }>(`/notes/${id}`),
  createNote: (note: Partial<Note>) =>
    request<{ note: Note }>('/notes', { method: 'POST', body: JSON.stringify(note) }),
  updateNote: (id: string, note: Partial<Note>) =>
    request<{ note: Note }>(`/notes/${id}`, { method: 'PUT', body: JSON.stringify(note) }),
  deleteNote: (id: string) =>
    request<void>(`/notes/${id}`, { method: 'DELETE' }),

  getStrategies: () => request<{ strategies: Strategy[] }>('/strategies'),
  getStrategy: (id: string) => request<Strategy>(`/strategies/${id}`),
  createStrategy: (s: Partial<Strategy>) =>
    request<{ strategy: Strategy }>('/strategies', { method: 'POST', body: JSON.stringify(s) }),
  updateStrategy: (id: string, s: Partial<Strategy>) =>
    request<{ strategy: Strategy }>(`/strategies/${id}`, { method: 'PUT', body: JSON.stringify(s) }),
  deleteStrategy: (id: string) =>
    request<void>(`/strategies/${id}`, { method: 'DELETE' }),
  assignStrategy: (id: string, symbol: string) =>
    request<void>(`/strategies/${id}/assign`, { method: 'POST', body: JSON.stringify({ symbol }) }),
  unassignStrategy: (id: string, symbol: string) =>
    request<void>(`/strategies/${id}/assign/${symbol}`, { method: 'DELETE' }),

  getPolicies: () => request<{ rules: PolicyRule[]; violations: any[] }>('/policies'),
  createPolicy: (rule: Partial<PolicyRule>) =>
    request<{ rule: PolicyRule }>('/policies', { method: 'POST', body: JSON.stringify(rule) }),
  updatePolicy: (id: string, rule: Partial<PolicyRule>) =>
    request<{ status: string }>(`/policies/${id}`, { method: 'PUT', body: JSON.stringify(rule) }),
  deletePolicy: (id: string) =>
    request<void>(`/policies/${id}`, { method: 'DELETE' }),
  checkPolicies: () => request<{ status: string; violations: any[]; metrics: any }>('/policies/check', { method: 'POST' }),

  runBacktest: (rules: any[], from_date: string, to_date: string) =>
    request<BacktestResult>('/backtest', { method: 'POST', body: JSON.stringify({ rules, from_date, to_date }) }),

  getCorrelation: () => request<CorrelationResult>('/correlation'),

  getAuditLogs: (eventType?: string, limit?: number) => {
    const params = new URLSearchParams();
    if (eventType) params.set('event_type', eventType);
    if (limit) params.set('limit', String(limit));
    const qs = params.toString();
    return request<{ events: AuditEvent[] }>(`/audit${qs ? '?' + qs : ''}`);
  },

  getPredictions: () => request<{ markets: PredictionMarket[] }>('/predictions'),
  getWalletPositions: (wallet: string) =>
    request<{ positions: WalletPosition[]; summary: WalletSummary }>(`/predictions/positions?wallet=${wallet}`),
  savePolymarketWallet: (wallet: string) =>
    request<{ status: string; wallet?: string }>('/auth/polymarket', {
      method: 'POST',
      body: JSON.stringify({ wallet }),
    }),
  disconnectSchwab: () =>
    request<{ status: string }>('/auth/schwab/disconnect', { method: 'POST' }),
  disconnectCoinbase: () =>
    request<{ status: string }>('/auth/coinbase/disconnect', { method: 'POST' }),
};
