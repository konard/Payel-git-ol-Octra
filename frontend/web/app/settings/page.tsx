import { Settings } from 'lucide-react';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { DashboardShell } from '../dashboard/DashboardShell';
import { ROUTES } from '../config/routes';

export default function SettingsPage() {
  return (
    <DashboardShell activeSection="settings" hideSidebarItems={['models', 'files', 'security', 'overview', 'environments']} showNotifications={false} hideNewFlow={true}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Settings">
          <EmptyDataPanel
            icon={Settings}
            title="Settings"
            detail="Account, billing, and workspace settings will be managed from this screen."
            actionHref={ROUTES.WORKSPACE}
            actionLabel="Open workspace"
          />
        </article>
      </section>
    </DashboardShell>
  );
}
