'use client';

import {
  Activity,
  ChevronDown,
  CircleDollarSign,
  Database,
  FileText,
  Layers3,
  LineChart,
  LockKeyhole,
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

const backendEndpoints = [
  { label: 'Environment', value: 'POST /environment', detail: 'environment_id' },
  { label: 'Chat API', value: 'POST /api/chat', detail: 'octra-api-token prompt' },
  { label: 'CLI state', value: 'Redis cli_state', detail: 'PID, port, TTL' },
];

const toolItems = [
  { label: 'Workflows', icon: Workflow, href: '/dashboard/flows' },
  { label: 'Streams', icon: Activity, href: '/dashboard/metrics' },
  { label: 'Settings', icon: Settings, href: '/settings' },
];

export default function HomePage() {
  const [panelOpen, setPanelOpen] = useState(false);
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

        <nav className="tv-nav" aria-label="Primary navigation">
          <a href="/dashboard">Products</a>
          <a href="#node-canvas">Environments</a>
          <a href="#runtime-metrics">Metrics</a>
          <a href="/auth">Auth</a>
          <a href="/dashboard">More</a>
        </nav>

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
                <Activity size={16} />
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
          <div className="market-heading">
            <span>Runtime metrics</span>
            <ChevronDown size={16} />
          </div>
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

          <section className="guard-stack" aria-label="Backend endpoints">
            {backendEndpoints.map((endpoint) => (
              <div className="guard-row" key={endpoint.label}>
                {endpoint.label === 'Environment' ? (
                  <Zap size={16} />
                ) : endpoint.label === 'CLI state' ? (
                  <Database size={16} />
                ) : (
                  <FileText size={16} />
                )}
                <span>{endpoint.label}</span>
                <strong>{endpoint.value}</strong>
              </div>
            ))}
            <div className="guard-row">
              <LockKeyhole size={16} />
              <span>Auth</span>
              <strong>Bearer cookie</strong>
            </div>
            <div className="guard-row">
              <CircleDollarSign size={16} />
              <span>Rate limit</span>
              <strong>api_chat</strong>
            </div>
          </section>
        </aside>
      </section>

      <aside className="home-rail" aria-label="Secondary tools">
        <button type="button" aria-label="Chart" onClick={() => setPanelOpen((v) => !v)}>
          <LineChart size={20} />
        </button>
        <button type="button" aria-label="Activity" onClick={() => setPanelOpen((v) => !v)}>
          <Activity size={20} />
        </button>
        <button type="button" aria-label="Layers" onClick={() => setPanelOpen((v) => !v)}>
          <Layers3 size={20} />
        </button>
        <button type="button" aria-label="Automation" onClick={() => setPanelOpen((v) => !v)}>
          <Zap size={20} />
        </button>
      </aside>
    </main>
  );
}
