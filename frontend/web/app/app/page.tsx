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
import { ASSETS } from '../config/images';
import { ROUTES } from '../config/routes';
import { fetchMe } from '../server/user';

const toolItems = [
  { label: 'Workflows', icon: Workflow, href: '/dashboard/flows' },
  { label: 'Streams', icon: Activity, href: '/dashboard/metrics' },
];

export default function HomePage() {
  const [panelOpen, setPanelOpen] = useState(false);
  const [metricsOpen, setMetricsOpen] = useState(false);
  const [endpointsOpen, setEndpointsOpen] = useState(false);
  const [deploymentsOpen, setDeploymentsOpen] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);
  const [username, setUsername] = useState('');

  useEffect(() => {
    const hasToken = ['octra_access_token', 'access_token'].some((key) => window.localStorage.getItem(key));
    setIsAuthed(hasToken);

    const cached = window.localStorage.getItem('octra_username');
    if (cached) {
      setUsername(cached);
    } else if (hasToken) {
      const token = window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
      if (token) {
        fetchMe(token).then(async (res) => {
          if (!res.ok) return;
          const body = await res.json();
          const u = body?.data?.username ?? body?.username;
          if (u) {
            window.localStorage.setItem('octra_username', u);
            setUsername(u);
          }
        }).catch(() => {});
      }
    }
  }, []);

  return (
    <main className="site-shell workspace-home">
      <header className="tv-header">
        <a className="tv-brand" href="/" aria-label="Octra home">
          <img src={ASSETS.LOGO} alt="" />
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
            <a className="avatar-button" href={ROUTES.PROFILE} aria-label="Open profile">{username ? username[0].toUpperCase() : '?'}</a>
          ) : (
            <a className="upgrade-button" href={ROUTES.LOGIN}>
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
        <div className="rail-spacer" />
        <a className="tool-button" href={ROUTES.SETTINGS} aria-label="Settings">
          <Settings size={20} />
        </a>
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
              <a className="terminal-button" href={ROUTES.DASHBOARD_FLOWS}>
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
          <div className={`collapse-wrap${metricsOpen ? '' : ' collapsed'}`}>
            <div className="collapse-inner">
              <div className="quote-stack">
                <EmptyDataPanel
                  compact
                  icon={Activity}
                  title="No live metrics yet"
                  detail="Runtime counters will appear here when backend telemetry is available."
                  actionHref={ROUTES.DASHBOARD_METRICS}
                  actionLabel="Open metrics"
                />
              </div>
            </div>
          </div>

          <div className="market-heading" role="button" tabIndex={0} onClick={() => setEndpointsOpen((v) => !v)} onKeyDown={(e) => e.key === 'Enter' && setEndpointsOpen((v) => !v)}>
            <span>Endpoints</span>
            <ChevronDown size={16} className={`chevron${endpointsOpen ? '' : ' collapsed'}`} />
          </div>
          <div className={`collapse-wrap${endpointsOpen ? '' : ' collapsed'}`}>
            <div className="collapse-inner">
              <div className="quote-stack">
                <EmptyDataPanel
                  compact
                  icon={Activity}
                  title="No endpoints configured"
                  detail="Configured API endpoints will appear here."
                  actionHref={ROUTES.DASHBOARD_FLOWS}
                  actionLabel="Configure"
                />
              </div>
            </div>
          </div>

          <div className="market-heading" role="button" tabIndex={0} onClick={() => setDeploymentsOpen((v) => !v)} onKeyDown={(e) => e.key === 'Enter' && setDeploymentsOpen((v) => !v)}>
            <span>Deployments</span>
            <ChevronDown size={16} className={`chevron${deploymentsOpen ? '' : ' collapsed'}`} />
          </div>
          <div className={`collapse-wrap${deploymentsOpen ? '' : ' collapsed'}`}>
            <div className="collapse-inner">
              <div className="quote-stack">
                <EmptyDataPanel
                  compact
                  icon={Activity}
                  title="No active deployments"
                  detail="Environment deployments will appear here."
                  actionHref={ROUTES.DASHBOARD}
                  actionLabel="View"
                />
              </div>
            </div>
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
