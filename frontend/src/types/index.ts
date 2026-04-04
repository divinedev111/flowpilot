export interface Position {
  id: string;
  symbol: string;
  quantity: number;
  mark_price: number;
  market_value: number;
  asset_type: string;
  source: string;
  account_id: string;
  timestamp: string;
  cost_basis: number;
  day_change: number;
  day_change_pct: number;
  total_pnl: number;
  total_pnl_pct: number;
}

export interface Metrics {
  net_worth: number;
  exposure_by_type: Record<string, number>;
  concentration_by_symbol: Record<string, number>;
  position_count: number;
  total_pnl: number;
  day_pnl: number;
  total_pnl_pct: number;
}

export interface Snapshot {
  id: string;
  created_at: string;
  net_worth: number;
}

export interface Mover {
  symbol: string;
  change_percent: number;
  old_value: number;
  new_value: number;
}

export interface Diff {
  from_snapshot: string;
  to_snapshot: string;
  net_worth_change: number;
  top_movers: Mover[];
  added_symbols: string[];
  removed_symbols: string[];
}

export interface AlertRule {
  id?: string;
  rule_type: string;
  threshold: number;
  target_asset: string;
  enabled: boolean;
}

export interface AlertEvent {
  id: string;
  rule_id: string;
  severity: string;
  message: string;
  evidence: string;
  timestamp: string;
}

export interface SyncResult {
  snapshot: Snapshot;
  metrics: Metrics;
  diff: Diff | null;
  positions: number;
  fetch_errors: string[];
  alerts_triggered: number;
}

export interface AIDigest {
  id: string;
  snapshot_id: string;
  summary: string;
  risk_notes: string;
  timestamp: string;
}

export interface Insight {
  type: string;
  severity: 'high' | 'medium' | 'low';
  title: string;
  description: string;
  recommendation: string;
}

export interface InsightsResult {
  insights: Insight[];
  overall_health: 'good' | 'fair' | 'poor' | 'unknown';
  generated_at: string;
}

export interface WatchlistItem {
  id: string;
  symbol: string;
  name: string;
  asset_type: string;
  notes: string;
  added_at: string;
}

export interface NewsItem {
  id: string;
  title: string;
  summary: string;
  source: string;
  url: string;
  symbols: string[];
  sentiment: 'positive' | 'negative' | 'neutral';
  impact: 'high' | 'medium' | 'low';
  relevance: number;
  published_at: string;
  category: 'earnings' | 'macro' | 'sector' | 'crypto';
}

export interface NewsFeed {
  items: NewsItem[];
  generated_at: string;
  holdings_analyzed: string[];
}

export interface Note {
  id: string;
  symbol: string;
  title: string;
  content: string;
  tags: string[];
  pinned: boolean;
  created_at: string;
  updated_at: string;
}

export interface Strategy {
  id: string;
  name: string;
  description: string;
  color: string;
  symbols: string[];
  total_value: number;
  allocation_pct: number;
  position_count: number;
  positions: Position[];
  created_at: string;
  updated_at: string;
}

export interface PolicyRule {
  id?: string;
  name: string;
  rule_type: string;
  threshold_pct: number;
  asset_type: string;
  target_symbol: string;
  severity: string;
  enabled: boolean;
  created_at: string;
}

export interface PolicyViolation {
  id: string;
  rule_id: string;
  rule_name: string;
  message: string;
  severity: string;
  current_value: number;
  threshold: number;
  symbols: string[];
  resolved: boolean;
  timestamp: string;
}

export interface BacktestEvent {
  rule_type: string;
  severity: string;
  message: string;
  evidence: string;
  snapshot_date: string;
}

export interface BacktestResult {
  events: BacktestEvent[];
  total_triggers: number;
  snapshots_analyzed: number;
}

export interface CorrelationResult {
  symbols: string[];
  matrix: number[][];
  period_days: number;
  top_pairs: { a: string; b: string; correlation: number }[];
  warning?: string;
}

export interface AuditEvent {
  id: string;
  timestamp: string;
  event_type: string;
  action: string;
  resource: string;
  resource_id: string;
  details: Record<string, any>;
  ip_address: string;
  success: boolean;
  error_msg: string;
}

export interface PredictionMarket {
  id: string;
  source: string;
  title: string;
  description: string;
  outcome_yes: number;
  outcome_no: number;
  volume: number;
  end_date: string;
  category: string;
  related_symbols: string[];
  url: string;
  last_synced: string;
}

export interface WalletPosition {
  condition_id: string;
  title: string;
  outcome: string;
  size: number;
  avg_price: number;
  current_price: number;
  current_value: number;
  initial_value: number;
  unrealized_pnl: number;
  cash_pnl: number;
  percent_pnl: number;
  redeemable: boolean;
  resolved: boolean;
}

export interface WalletSummary {
  total_positions: number;
  active_positions: number;
  resolved_positions: number;
  claimable_count: number;
  total_value: number;
  total_cost_basis: number;
  total_unrealized_pnl: number;
  total_cash_pnl: number;
}
