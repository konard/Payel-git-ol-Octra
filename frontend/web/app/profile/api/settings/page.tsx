import { Settings } from 'lucide-react';
import { ApiTabs } from '../ApiTabs';

export default function ApiSettingsPage() {
  return (
    <div style={{ padding: 14 }}>
      <ApiTabs />
      <div className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="API settings" style={{ padding: 24, minHeight: 0 }}>
          <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16 }}>
            <Settings size={20} />
            <h2 style={{ margin: 0, fontSize: '1rem' }}>API settings</h2>
          </div>
          <p style={{ color: 'var(--muted)', lineHeight: 1.6, margin: 0 }}>
            Rate limits, IP whitelisting, and other API configuration options will be managed here.
          </p>
        </article>
      </div>
    </div>
  );
}
