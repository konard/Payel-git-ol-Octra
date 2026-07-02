import { DashboardShell } from '../DashboardShell';
import { MetricsView } from './MetricsView';

export default function DashboardMetricsPage() {
  return (
    <DashboardShell activeSection="metrics">
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Request metrics">
          <div className="panel-heading">
            <span>Request metrics</span>
          </div>
          <MetricsView />
        </article>
      </section>
    </DashboardShell>
  );
}
