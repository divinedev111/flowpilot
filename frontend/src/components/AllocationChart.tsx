import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts';
import { formatAssetType } from '../utils/format';

const COLORS = ['#2563eb', '#16a34a', '#d97706', '#7c3aed', '#dc2626', '#0891b2', '#e11d48'];

interface Props {
  exposure: Record<string, number>;
}

export default function AllocationChart({ exposure }: Props) {
  const data = Object.entries(exposure)
    .filter(([, value]) => value >= 0.1)
    .map(([name, value]) => ({
      name: formatAssetType(name),
      value: Math.round(value * 10) / 10,
    }));

  if (data.length === 0) return null;

  return (
    <div className="card" style={{ height: '100%' }}>
      <div className="card-header">
        <span className="card-title">Asset Allocation</span>
      </div>
      <div style={{ position: 'relative', width: '100%', maxWidth: 300, margin: '0 auto' }}>
        <ResponsiveContainer width="100%" height={220}>
          <PieChart>
            <Pie
              data={data}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              innerRadius={55}
              outerRadius={85}
              strokeWidth={0}
              paddingAngle={2}
            >
              {data.map((_, i) => (
                <Cell key={i} fill={COLORS[i % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip
              content={({ active, payload }) => {
                if (!active || !payload?.length) return null;
                const d = payload[0];
                return (
                  <div style={{
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-strong)',
                    borderRadius: 'var(--radius-sm)',
                    padding: '0.5rem 0.75rem',
                    fontFamily: 'var(--font-mono)',
                    fontSize: '0.8rem',
                    zIndex: 20,
                    position: 'relative',
                  }}>
                    <div style={{ color: 'var(--text-secondary)', marginBottom: '2px', textTransform: 'capitalize' }}>
                      {String(d.name)}
                    </div>
                    <div style={{ color: d.payload?.fill || 'var(--text-primary)', fontWeight: 600 }}>
                      {String(d.value)}%
                    </div>
                  </div>
                );
              }}
            />
          </PieChart>
        </ResponsiveContainer>
        <div style={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          textAlign: 'center',
          pointerEvents: 'none',
          zIndex: 1,
        }}>
          <div className="chart-center-value">{data.length}</div>
          <div className="chart-center-label">Asset Types</div>
        </div>
      </div>
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem', justifyContent: 'center', marginTop: '0.25rem' }}>
        {data.map((d, i) => (
          <div key={d.name} style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <div style={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              background: COLORS[i % COLORS.length],
            }} />
            <span className="mono" style={{
              fontSize: '0.7rem',
              color: 'var(--text-secondary)',
              textTransform: 'capitalize',
            }}>
              {d.name} {d.value}%
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
