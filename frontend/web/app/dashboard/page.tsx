'use client';

import { Settings } from 'lucide-react';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { RequestMetricsOverview } from '../components/RequestMetricsOverview';
import { DashboardShell } from './DashboardShell';
import { ROUTES } from '../config/routes';

export default function DashboardPage() {
  return (
    <DashboardShell activeSection="overview">
      <section className="dashboard-metrics" aria-label="Pipeline metrics">
        <RequestMetricsOverview range="7d" compact />
      </section>

      <section className="dashboard-grid">
        <article className="architecture-panel large-panel" aria-label="Workspace policy data">
          <div className="panel-heading">
            <span>Workspace policy</span>
          </div>
          <EmptyDataPanel
            icon={Settings}
            title="No policy graph yet"
            detail="Workspace routing and billing policy data will appear here after configuration is saved."
            actionHref={ROUTES.SETTINGS}
            actionLabel="Open settings"
          />
        </article>
      </section>
    </DashboardShell>
  );
}
