import { useState } from 'react';
import type { Position } from '../types';
import { formatAssetType } from '../utils/format';

interface Props {
  positions: Position[];
}

type SortKey = 'symbol' | 'market_value' | 'quantity' | 'mark_price' | 'asset_type' | 'source' | 'total_pnl' | 'day_change_pct';

const HEADERS: { key: SortKey; label: string }[] = [
  { key: 'symbol', label: 'Symbol' },
  { key: 'quantity', label: 'Quantity' },
  { key: 'mark_price', label: 'Price' },
  { key: 'market_value', label: 'Value' },
  { key: 'total_pnl', label: 'P&L' },
  { key: 'day_change_pct', label: 'Day' },
  { key: 'asset_type', label: 'Type' },
  { key: 'source', label: 'Source' },
];

const TICKER_COLORS = [
  '#2563eb', '#16a34a', '#d97706', '#7c3aed', '#dc2626',
  '#0891b2', '#e11d48', '#059669', '#7c2d12', '#4338ca',
];

const SYMBOL_NAMES: Record<string, string> = {
  SOXL: 'Direxion Semiconductor Bull 3X',
  SPXL: 'Direxion Daily S&P 500 Bull 3X',
  TQQQ: 'ProShares UltraPro QQQ',
  NVDA: 'NVIDIA Corporation',
  AAPL: 'Apple Inc.',
  MSFT: 'Microsoft Corporation',
  GOOGL: 'Alphabet Inc.',
  GOOG: 'Alphabet Inc.',
  AMZN: 'Amazon.com Inc.',
  META: 'Meta Platforms Inc.',
  TSLA: 'Tesla Inc.',
  AMD: 'Advanced Micro Devices',
  INTC: 'Intel Corporation',
  PFE: 'Pfizer Inc.',
  RNA: 'Avidity Biosciences',
  SPY: 'SPDR S&P 500 ETF',
  QQQ: 'Invesco QQQ Trust',
  VOO: 'Vanguard S&P 500 ETF',
  VTI: 'Vanguard Total Stock Market',
  ARKK: 'ARK Innovation ETF',
  SOFI: 'SoFi Technologies',
  PLTR: 'Palantir Technologies',
  COIN: 'Coinbase Global',
  SQ: 'Block Inc.',
  PYPL: 'PayPal Holdings',
  V: 'Visa Inc.',
  MA: 'Mastercard Inc.',
  JPM: 'JPMorgan Chase',
  BAC: 'Bank of America',
  GS: 'Goldman Sachs',
  WFC: 'Wells Fargo',
  DIS: 'Walt Disney Co.',
  NFLX: 'Netflix Inc.',
  CRM: 'Salesforce Inc.',
  UBER: 'Uber Technologies',
  ABNB: 'Airbnb Inc.',
  SNAP: 'Snap Inc.',
  SHOP: 'Shopify Inc.',
  ROKU: 'Roku Inc.',
  RBLX: 'Roblox Corporation',
  U: 'Unity Software',
  NET: 'Cloudflare Inc.',
  SNOW: 'Snowflake Inc.',
  DDOG: 'Datadog Inc.',
  ZS: 'Zscaler Inc.',
  PANW: 'Palo Alto Networks',
  CRWD: 'CrowdStrike Holdings',
  BTC: 'Bitcoin',
  ETH: 'Ethereum',
  SOL: 'Solana',
  DOGE: 'Dogecoin',
  ADA: 'Cardano',
  DOT: 'Polkadot',
  AVAX: 'Avalanche',
  MATIC: 'Polygon',
  LINK: 'Chainlink',
  UNI: 'Uniswap',
  ATOM: 'Cosmos',
  XRP: 'Ripple',
  LTC: 'Litecoin',
  SHIB: 'Shiba Inu',
  QTUM: 'Qtum',
  DIMO: 'DIMO Network',
  CORECHAIN: 'CoreChain',
  VARA: 'Vara Network',
  AAVE: 'Aave',
  COMP: 'Compound',
  MKR: 'Maker',
  SNX: 'Synthetix',
  FIL: 'Filecoin',
  NEAR: 'NEAR Protocol',
  APT: 'Aptos',
  SUI: 'Sui',
  ARB: 'Arbitrum',
  OP: 'Optimism',
  PEPE: 'Pepe',
  XLM: 'Stellar',
};

function getSymbolName(symbol: string): string {
  // Check direct match
  const upper = symbol.toUpperCase();
  if (SYMBOL_NAMES[upper]) return SYMBOL_NAMES[upper];
  if (SYMBOL_NAMES[symbol]) return SYMBOL_NAMES[symbol];
  // CUSIP-like IDs (e.g. 05370A108) — just show generic label
  if (/^\d/.test(symbol)) return 'Collective Investment';
  return '';
}

function tickerColor(symbol: string): string {
  let hash = 0;
  for (let i = 0; i < symbol.length; i++) {
    hash = symbol.charCodeAt(i) + ((hash << 5) - hash);
  }
  return TICKER_COLORS[Math.abs(hash) % TICKER_COLORS.length];
}

export default function PositionsTable({ positions }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('market_value');
  const [sortAsc, setSortAsc] = useState(false);

  const sorted = [...positions].sort((a, b) => {
    const aVal = a[sortKey];
    const bVal = b[sortKey];
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      return sortAsc ? aVal - bVal : bVal - aVal;
    }
    return sortAsc ? String(aVal).localeCompare(String(bVal)) : String(bVal).localeCompare(String(aVal));
  });

  const maxValue = Math.max(...positions.map(p => p.market_value), 1);

  const handleSort = (key: SortKey) => {
    if (key === sortKey) setSortAsc(!sortAsc);
    else { setSortKey(key); setSortAsc(false); }
  };

  return (
    <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
      <div className="card-header" style={{ padding: '1.25rem 1.5rem 0' }}>
        <span className="card-title">Holdings</span>
        <span className="mono" style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          {positions.length} positions
        </span>
      </div>
      <div className="table-wrapper">
        <table className="data-table">
          <thead>
            <tr>
              {HEADERS.map(h => (
                <th key={h.key} onClick={() => handleSort(h.key)}>
                  {h.label}
                  {sortKey === h.key ? (sortAsc ? ' \u2191' : ' \u2193') : ''}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sorted.map((p, i) => {
              const name = getSymbolName(p.symbol);
              return (
                <tr key={i}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                      <span className="ticker-badge" style={{ background: tickerColor(p.symbol) }}>
                        {p.symbol.slice(0, 2).toUpperCase()}
                      </span>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.05rem' }}>
                        <span className="mono" style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '0.875rem' }}>
                          {p.symbol}
                        </span>
                        {name && (
                          <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)', lineHeight: 1.2 }}>
                            {name}
                          </span>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="mono" style={{ color: 'var(--text-secondary)' }}>
                    {p.quantity.toLocaleString(undefined, { maximumFractionDigits: 6 })}
                  </td>
                  <td className="mono" style={{ color: 'var(--text-secondary)' }}>
                    ${p.mark_price.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </td>
                  <td>
                    <div className="value-cell">
                      <span className="mono" style={{ fontWeight: 600, color: 'var(--accent-green)' }}>
                        ${p.market_value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </span>
                      <div className="value-bar-track">
                        <div
                          className="value-bar-fill"
                          style={{
                            width: `${(p.market_value / maxValue) * 100}%`,
                            background: 'var(--accent-green)',
                          }}
                        />
                      </div>
                    </div>
                  </td>
                  <td>
                    <span className={`change-badge ${p.total_pnl >= 0 ? 'positive' : 'negative'}`}>
                      {p.total_pnl >= 0 ? '+' : ''}
                      ${Math.abs(p.total_pnl).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      <span style={{ marginLeft: '0.25rem', opacity: 0.8 }}>
                        ({p.total_pnl_pct >= 0 ? '+' : ''}{p.total_pnl_pct.toFixed(1)}%)
                      </span>
                    </span>
                  </td>
                  <td>
                    {p.day_change_pct !== 0 ? (
                      <span className={`change-badge ${p.day_change_pct >= 0 ? 'positive' : 'negative'}`}>
                        {p.day_change_pct >= 0 ? '+' : ''}{p.day_change_pct.toFixed(2)}%
                      </span>
                    ) : (
                      <span className="mono" style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>--</span>
                    )}
                  </td>
                  <td>
                    <span className={`badge ${p.asset_type === 'crypto' ? 'badge-crypto' : 'badge-equity'}`}>
                      {formatAssetType(p.asset_type)}
                    </span>
                  </td>
                  <td>
                    <span className={`source-tag ${p.source}`}>
                      {p.source === 'coinbase' && (
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><path d="M14.5 9h-5v6h5"/></svg>
                      )}
                      {p.source === 'schwab' && (
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 12h6"/></svg>
                      )}
                      {p.source}
                    </span>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
