import { Globe } from 'lucide-react';
import { ApiTabs } from '../ApiTabs';

export default function PublicApiPage() {
  return (
    <div style={{ padding: 14 }}>
      <ApiTabs />
      <div className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Public API" style={{ padding: 24, minHeight: 0 }}>
          <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16 }}>
            <Globe size={20} />
            <h2 style={{ margin: 0, fontSize: '1rem' }}>Public API</h2>
          </div>
          <p style={{ color: 'var(--muted)', lineHeight: 1.6, margin: 0 }}>
            Public API endpoints and documentation will be available here.
          </p>
        </article>
      </div>
    </div>
  );
}
