import { Settings, ExternalLink } from 'lucide-react';
import { ROUTES } from '../../config/routes';

export default function ProfileSettingsPage() {
  return (
    <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
      <article className="large-panel" aria-label="Settings" style={{ padding: 32 }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16 }}>
          <Settings size={20} />
          <h1 style={{ margin: 0, fontSize: '1.2rem' }}>Profile settings</h1>
        </div>
        <p style={{ color: 'var(--muted)', lineHeight: 1.6, margin: '0 0 24px' }}>
          Account and workspace settings will be available here.
        </p>
        <a className="text-button" href={ROUTES.SETTINGS} style={{ gap: 8 }}>
          <span>Go to workspace settings</span>
          <ExternalLink size={15} />
        </a>
      </article>
    </div>
  );
}
