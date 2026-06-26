import { Activity, Code2, Settings } from 'lucide-react';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { WorkflowCanvas } from '../components/WorkflowCanvas';
import { DashboardShell } from './DashboardShell';

export default function DashboardPage() {
  return (
    <DashboardShell activeSection="overview">
      <section className="dashboard-metrics" aria-label="Pipeline metrics">
        <EmptyDataPanel
          compact
          icon={Activity}
          title="No live metrics yet"
          detail="Runtime counters will appear here when environments begin reporting telemetry."
          actionHref="/dashboard/metrics"
          actionLabel="Open metrics"
        />
      </section>

      <section className="dashboard-grid">
        <article className="traffic-panel large-panel" aria-label="Environment nodes canvas">
          <div className="panel-heading">
            <span>Environment nodes canvas</span>
            <a className="ghost-command" href="/dashboard/flows">
              <Code2 size={15} />
              Edit
            </a>
          </div>
          <section className="dashboard-canvas" id="node-canvas">
            <WorkflowCanvas />
          </section>
        </article>
        <article className="architecture-panel large-panel" aria-label="Workspace policy data">
          <div className="panel-heading">
            <span>Workspace policy</span>
          </div>
          <EmptyDataPanel
            icon={Settings}
            title="No policy graph yet"
            detail="Workspace routing and billing policy data will appear here after configuration is saved."
            actionHref="/settings"
            actionLabel="Open settings"
          />
        </article>
      </section>
    </DashboardShell>
  );
}
