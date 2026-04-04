import AuditLog from '../components/AuditLog';

export default function Settings() {
  return (
    <div>
      <h1>Settings</h1>

      <div style={{
        padding: '1.25rem',
        borderRadius: '0.75rem',
        border: '1px solid #e5e7eb',
        backgroundColor: 'white',
      }}>
        <AuditLog />
      </div>
    </div>
  );
}
