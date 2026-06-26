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

        <section className="node-canvas dashboard-canvas" id="node-canvas">
          <WorkflowCanvas />
        </section>
      </section>
    </main>
  );
}
