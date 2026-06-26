'use client';

import {
  Activity,
  ChevronDown,
  Layers3,
  LineChart,
  Plus,
  Search,
  Settings,
  Sparkles,
  Workflow,
  Zap,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { UserBalance } from '../components/UserBalance';
import { WorkflowCanvas } from '../components/WorkflowCanvas';

const toolItems = [
  { label: 'Workflows', icon: Workflow, href: '/dashboard/flows' },
  { label: 'Streams', icon: Activity, href: '/dashboard/metrics' },
  { label: 'Settings', icon: Settings, href: '/settings' },
];

export default function HomePage() {
  const [panelOpen, setPanelOpen] = useState(false);
  const [metricsOpen, setMetricsOpen] = useState(true);
  const [isAuthed, setIsAuthed] = useState(false);

  useEffect(() => {
    const hasToken = ['octra_access_token', 'access_token'].some((key) => window.localStorage.getItem(key));
    setIsAuthed(hasToken);
  }, []);

  return (
    <main className="site-shell workspace-home">
      <header className="tv-header">
        <a className="tv-brand" href="/" aria-label="Octra home">
          <img src="/assets/octra-node-logo.svg" alt="" />
          <span>Octra</span>
        </a>

        <label className="tv-search">
          <Search size={18} />
          <span className="sr-only">Search Octra</span>
          <input placeholder="Search environments, skills, endpoints..." />
        </label>

        <div className="tv-actions">
          <UserBalance />
          {isAuthed ? (
            <a className="avatar-button" href="/dashboard" aria-label="Open dashboard">O</a>
          ) : (
            <a className="upgrade-button" href="/auth">
              <Sparkles size={16} />
              <span>Sign in</span>
            </a>
          )}
        </div>
      </header>

      <aside className="workspace-tools" aria-label="Workspace tools">
        {toolItems.map((item) => {
          const Icon = item.icon;
          return (
            <a className="tool-button" href={item.href} aria-label={item.label} key={item.label}>
              <Icon size={20} />
            </a>
          );
        })}
      </aside>

      <section className={`workspace-frame${panelOpen ? '' : ' panel-collapsed'}`} aria-label="Octra backend workflow terminal">
        <section className="home-canvas" aria-labelledby="workspace-title">
          <h1 id="workspace-title" className="workspace-word">
            Octra
          </h1>

          <section className="node-canvas" id="node-canvas" aria-label="React Flow environment nodes">
            <WorkflowCanvas />
          </section>

          <section className="active-environments" aria-label="Active environments list">
            <div className="active-environments-header">
              <div className="active-environments-title">
                <span>Active environments</span>
              </div>
              <a className="terminal-button" href="/dashboard/flows">
                <Plus size={15} />
                New flow
              </a>
            </div>
            <p className="empty-flows-message">You don't have any flows yet.</p>
          </section>
        </section>

        <aside className={`home-market-panel${panelOpen ? '' : ' panel-hidden'}`} id="runtime-metrics" aria-label="Backend runtime metrics">
          <div className="market-heading" role="button" tabIndex={0} onClick={() => setMetricsOpen((v) => !v)} onKeyDown={(e) => e.key === 'Enter' && setMetricsOpen((v) => !v)}>
            <span>Runtime metrics</span>
            <ChevronDown size={16} className={`chevron${metricsOpen ? '' : ' collapsed'}`} />
          </div>
          {metricsOpen && (
            <div className="quote-stack">
              <EmptyDataPanel
                compact
                icon={Activity}
                title="No live metrics yet"
                detail="Runtime counters will appear here when backend telemetry is available."
                actionHref="/dashboard/metrics"
                actionLabel="Open metrics"
              />
            </div>
          )}

          <div className="market-heading" role="button" tabIndex={0}>
            <span>Endpoints</span>
          </div>

          <div className="market-heading" role="button" tabIndex={0}>
            <span>Deployments</span>
          </div>
        </aside>
      </section>

      <aside className="home-rail" aria-label="Secondary tools">
        <button type="button" className={panelOpen ? 'active' : ''} aria-label="Chart" onClick={() => setPanelOpen((v) => !v)}>
          <LineChart size={20} />
        </button>
        <button type="button" className={panelOpen ? 'active' : ''} aria-label="Activity" onClick={() => setPanelOpen((v) => !v)}>
          <Activity size={20} />
        </button>
        <button type="button" className={panelOpen ? 'active' : ''} aria-label="Layers" onClick={() => setPanelOpen((v) => !v)}>
          <Layers3 size={20} />
        </button>
        <button type="button" className={panelOpen ? 'active' : ''} aria-label="Automation" onClick={() => setPanelOpen((v) => !v)}>
          <Zap size={20} />
        </button>
      </aside>
    </main>
  );
}
