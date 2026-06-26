import {
  Activity,
  ArrowRight,
  Bell,
  Bot,
  ChevronDown,
  CircleDollarSign,
  Code2,
  Database,
  FileText,
  Layers3,
  LineChart,
  LockKeyhole,
  PanelRight,
  Plus,
  Search,
  Settings,
  ShieldCheck,
  Sparkles,
  Workflow,
  Zap,
} from 'lucide-react';
import { WorkflowCanvas } from '../components/WorkflowCanvas';

const backendEndpoints = [
  { label: 'Environment', value: 'POST /environment', detail: 'Nix profile + skills' },
  { label: 'Chat API', value: 'POST /api/chat', detail: 'octra-api-token prompt' },
  { label: 'CLI state', value: 'Redis cli_state', detail: 'PID, port, TTL' },
];

const backendMetrics = [
  { label: 'ACTIVE ENVS', value: '18', detail: 'Nix profiles', tone: 'success' },
  { label: 'CHAT REQS', value: '124k', detail: '/api/chat', tone: 'success' },
  { label: 'CLI PROCS', value: '7', detail: 'stdin/stdout live', tone: 'warning' },
  { label: 'SKILLS', value: '42', detail: 'installed packages', tone: 'success' },
  { label: 'REDIS TTL', value: '38m', detail: 'cli_state cache', tone: 'warning' },
  { label: 'PROXY MODE', value: '3', detail: 'LLM fallback', tone: 'danger' },
];

const activeEnvironments = [
  {
    name: 'Claude Code CLI',
    environment_id: 'env:claude-code:01',
    user_id: 'usr_742',
    status: 'running',
    ttl: '38m',
    endpoint: '/api/chat',
  },
  {
    name: 'OpenCode CLI',
    environment_id: 'env:opencode:04',
    user_id: 'usr_318',
    status: 'warm',
    ttl: '21m',
    endpoint: 'stdin pipe',
  },
  {
    name: 'Codex CLI',
    environment_id: 'env:codex:17',
    user_id: 'usr_904',
    status: 'installing',
    ttl: '9m',
    endpoint: 'Nix profile',
  },
  {
    name: 'Direct LLM proxy',
    environment_id: 'env:proxy:09',
    user_id: 'usr_556',
    status: 'ready',
    ttl: 'no CLI',
    endpoint: 'LLM base_url',
  },
];

const toolItems = [
  { label: 'Workflows', icon: Workflow },
  { label: 'Streams', icon: Activity },
  { label: 'Environments', icon: Bot },
  { label: 'Security', icon: ShieldCheck },
  { label: 'Code', icon: Code2 },
  { label: 'Settings', icon: Settings },
];

export default function HomePage() {
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
          <a className="upgrade-button" href="/auth">
            <Sparkles size={16} />
            <span>Sign in</span>
          </a>
          <a className="avatar-button" href="/dashboard" aria-label="Open dashboard">
            O
          </a>
        </div>
      </header>

      <section className="workspace-frame" aria-label="Octra backend workflow terminal">
        <aside className="workspace-tools" aria-label="Workspace tools">
          {toolItems.map((item) => (
            <a className="tool-button" href="/dashboard" aria-label={item.label} key={item.label}>
              <item.icon size={19} />
            </a>
          ))}
        </aside>

        <section className="home-canvas" aria-labelledby="workspace-title">
          <div className="canvas-header">
            <div className="canvas-crumbs">
              <span>Environments</span>
              <span>MCP endpoint</span>
              <span>React Flow graph</span>
            </div>
            <div className="canvas-actions">
              <button className="icon-button dark-icon" type="button" aria-label="Notifications">
                <Bell size={17} />
              </button>
            </div>
          </div>

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
              <button className="terminal-button" type="button">
                <Plus size={15} />
                New flow
              </button>
            </div>
            <div className="active-environment-list">
              {activeEnvironments.map((environment) => (
                <article className="active-environment-row" key={environment.environment_id}>
                  <div>
                    <span>{environment.name}</span>
                    <strong>{environment.environment_id}</strong>
                  </div>
                  <div>
                    <span>Status</span>
                    <strong>{environment.status}</strong>
                  </div>
                  <div>
                    <span>TTL</span>
                    <strong>{environment.ttl}</strong>
                  </div>
                  <div>
                    <span>Endpoint</span>
                    <strong>{environment.endpoint}</strong>
                  </div>
                  <div>
                    <span>user_id</span>
                    <strong>{environment.user_id}</strong>
                  </div>
                </article>
              ))}
            </div>
          </section>
        </section>

        <aside className="home-market-panel" id="runtime-metrics" aria-label="Backend runtime metrics">
          <div className="market-heading">
            <span>Runtime metrics</span>
            <ChevronDown size={16} />
          </div>
          <div className="market-columns">
            <span>Signal</span>
            <span>Value</span>
            <span>Source</span>
          </div>
          <div className="quote-stack">
            {backendMetrics.map((metric) => (
              <div className={`market-row metric-tone-${metric.tone}`} key={metric.label}>
                <span>{metric.label}</span>
                <strong>{metric.value}</strong>
                <em>{metric.detail}</em>
              </div>
            ))}
          </div>

          <section className="selected-node" aria-label="Selected node details">
            <div className="selected-topline">
              <div>
                <span>Selected node</span>
                <strong>Claude Code CLI</strong>
              </div>
              <PanelRight size={18} />
            </div>
            <div className="price-line">
              <strong>38</strong>
              <span>minutes TTL in Redis cli_state</span>
              <em>running</em>
            </div>
            <div className="market-bars" aria-hidden="true">
              <i /><i /><i /><i /><i /><i /><i /><i /><i /><i /><i /><i />
            </div>
          </section>

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

        <aside className="home-rail" aria-label="Secondary tools">
          <button type="button" aria-label="Chart">
            <LineChart size={20} />
          </button>
          <button type="button" aria-label="Activity">
            <Activity size={20} />
          </button>
          <button type="button" aria-label="Layers">
            <Layers3 size={20} />
          </button>
          <button type="button" aria-label="Automation">
            <Zap size={20} />
          </button>
        </aside>
      </section>
    </main>
  );
}
