import {
  Activity,
  Bell,
  Bot,
  CircleDollarSign,
  Code2,
  Cpu,
  FileText,
  Gauge,
  Github,
  Layers3,
  LineChart,
  LockKeyhole,
  PanelLeft,
  Plus,
  Search,
  Settings,
  ShieldCheck,
  Sparkles,
  Workflow,
  Zap,
} from 'lucide-react';
import { WorkflowCanvas } from '../components/WorkflowCanvas';

const metrics = [
  { label: 'Processed prompts', value: '124k', delta: '+15%', tone: 'up' },
  { label: 'Avg. latency', value: '850ms', delta: '+45ms', tone: 'warn' },
  { label: 'Total cost', value: '$452.10', delta: '-$120', tone: 'up' },
  { label: 'Blocked requests', value: '2.4k', delta: '-4%', tone: 'down' },
];

const navItems = [
  { label: 'Overview', icon: Gauge },
  { label: 'Flows', icon: Workflow },
  { label: 'Models', icon: Bot },
  { label: 'Files', icon: FileText },
  { label: 'Security', icon: ShieldCheck },
  { label: 'Settings', icon: Settings },
];

export default function DashboardPage() {
  return (
    <main className="dashboard-page">
      <aside className="app-sidebar" aria-label="Octra sections">
        <a className="square-brand" href="/app" aria-label="Octra home">
          <img src="/assets/octra-node-logo.svg" alt="" />
        </a>
        <nav>
          {navItems.map((item, index) => (
            <a className={index === 0 ? 'side-icon active' : 'side-icon'} href="#" key={item.label} aria-label={item.label}>
              <item.icon size={18} />
            </a>
          ))}
        </nav>
        <a className="side-icon" href="/auth" aria-label="Account">
          <Github size={18} />
        </a>
      </aside>

      <section className="dashboard-scene" aria-labelledby="dashboard-title">
        <header className="dashboard-topbar">
          <div className="crumbs">
            <button className="icon-button dark-icon" type="button" aria-label="Toggle navigation">
              <PanelLeft size={18} />
            </button>
            <a href="/">Octra</a>
            <span>Pipeline Command</span>
          </div>
          <label className="search-field dashboard-search">
            <Search size={15} />
            <span className="sr-only">Search dashboard</span>
            <input placeholder="Search..." />
          </label>
          <div className="topbar-actions">
            <button className="icon-button dark-icon" type="button" aria-label="Notifications">
              <Bell size={18} />
            </button>
            <button className="small-command accent-command" type="button">
              <Plus size={15} />
              New flow
            </button>
          </div>
        </header>

        <div className="dashboard-tabs" aria-label="Dashboard views">
          <a className="active" href="#">Overview</a>
          <a href="#">Metrics</a>
          <a href="#">Evaluations</a>
          <a href="#">Deployments</a>
        </div>

        <section className="dashboard-metrics" aria-label="Pipeline metrics">
          {metrics.map((metric) => (
            <article className="metric-card dashboard-metric" key={metric.label}>
              <div>
                <span>{metric.label}</span>
                <strong>{metric.value}</strong>
              </div>
              <span className={`metric-delta ${metric.tone}`}>{metric.delta}</span>
              <div className={`spark-bars ${metric.tone === 'down' ? 'low-bars' : metric.tone === 'warn' ? 'soft-bars' : 'rise-bars'}`} aria-hidden="true">
                <i /><i /><i /><i /><i />
              </div>
            </article>
          ))}
        </section>

        <section className="dashboard-grid">
          <article className="traffic-panel large-panel" aria-label="Environment nodes canvas">
            <div className="panel-heading">
              <span>Environment nodes canvas</span>
              <button className="ghost-command" type="button">
                <Code2 size={15} />
                Edit
              </button>
            </div>
            <section className="dashboard-canvas" id="node-canvas">
              <WorkflowCanvas />
            </section>
          </article>
          <article className="architecture-panel large-panel" aria-label="Active pipeline architecture">
            <div className="panel-heading">
              <span>Active pipeline architecture</span>
              <button className="ghost-command" type="button">
                <Code2 size={15} />
                Edit
              </button>
            </div>
            <div className="architecture-list">
              <section>
                <h2>1. Ingress</h2>
                <div className="rule-row">
                  <Cpu size={16} />
                  <span>Endpoint</span>
                  <strong>/v1/chat/completions</strong>
                </div>
                <div className="rule-row">
                  <LockKeyhole size={16} />
                  <span>Auth protocol</span>
                  <strong>Bearer strict</strong>
                </div>
              </section>
              <section>
                <h2>2. Pre-processing</h2>
                <div className="policy-grid">
                  <div>
                    <span>PII redact</span>
                    <strong>Mask</strong>
                  </div>
                  <div>
                    <span>Injection</span>
                    <strong>Strict</strong>
                  </div>
                  <div>
                    <span>Toxicity</span>
                    <strong>Drop</strong>
                  </div>
                </div>
              </section>
              <section>
                <h2>3. Dynamic router</h2>
                <div className="router-rule">
                  <div className="rule-condition">
                    <span>IF</span>
                    <strong>intent = billing</strong>
                    <span>THEN</span>
                  </div>
                  <div className="rule-row">
                    <Activity size={16} />
                    <span>Route to</span>
                    <strong>GPT-4o</strong>
                  </div>
                  <div className="rule-row">
                    <CircleDollarSign size={16} />
                    <span>Max cost</span>
                    <strong>$0.20</strong>
                  </div>
                </div>
              </section>
            </div>
          </article>
        </section>
      </section>
    </main>
  );
}
