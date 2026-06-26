import { Settings } from 'lucide-react';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { DashboardShell } from '../dashboard/DashboardShell';

export default function SettingsPage() {
  return (
    <DashboardShell activeSection="settings" hideSidebarItems={['models', 'files', 'security', 'overview', 'flows']} showNotifications={false}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Settings">
          <EmptyDataPanel
            icon={Settings}
            title="Settings"
            detail="Account, billing, and workspace settings will be managed from this screen."
            actionHref="/app"
            actionLabel="Open workspace"
          />
        </article>
      </section>
    </DashboardShell>
  );
}
