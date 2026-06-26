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
import { WorkflowCanvas } from './components/WorkflowCanvas';

const backendEndpoints = [
  { label: 'Task stream', value: 'GET /task/create', detail: 'WebSocket CreateTaskRequest' },
  { label: 'Task status', value: 'GET /task/status', detail: 'task_id progress lookup' },
  { label: 'Saved workflow', value: 'POST /workflows', detail: 'nodes and edges JSON' },
];

const backendMetrics = [
  { label: 'ACTIVE TASKS', value: '18', detail: 'redis streams' },
  { label: 'PROGRESS', value: '72%', detail: 'latest TaskUpdate' },
  { label: 'MANAGERS', value: '6', detail: 'role + priority' },
  { label: 'WORKERS', value: '24', detail: 'predefined workers' },
  { label: 'HISTORY', value: '256', detail: 'stored updates' },
  { label: 'STOP QUEUE', value: '0', detail: '/task/:taskId/stop' },
];

const activeAgents = [
  {
    role: 'Boss planner',
    agent_id: 'boss:planner:01',
    task_id: 'usr_742:6e9b',
    status: 'boss_planning',
    progress: '32%',
    endpoint: 'CreateTaskStream',
  },
  {
    role: 'Frontend manager',
    agent_id: 'mgr:frontend:04',
    task_id: 'usr_742:6e9b',
    status: 'managers_assigned',
    progress: '58%',
    endpoint: 'ManagerConfig.workers',
  },
  {
    role: 'Worker code',
    agent_id: 'worker:code:17',
    task_id: 'usr_742:6e9b',
    status: 'processing',
    progress: '74%',
    endpoint: 'TaskUpdate.data',
  },
  {
    role: 'Reviewer',
    agent_id: 'worker:review:09',
    task_id: 'usr_742:6e9b',
    status: 'queued',
    progress: '12%',
    endpoint: 'ResumeTaskStream',
  },
];

const toolItems = [
  { label: 'Workflows', icon: Workflow },
  { label: 'Streams', icon: Activity },
  { label: 'Agents', icon: Bot },
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
          <input placeholder="Search tasks, agents, workflows..." />
        </label>

        <nav className="tv-nav" aria-label="Primary navigation">
          <a href="/dashboard">Products</a>
          <a href="#node-canvas">Agents</a>
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
          <div className="canvas-grid" aria-hidden="true" />
          <div className="canvas-header">
            <div className="canvas-crumbs">
              <span>Tasks</span>
              <span>Backend orchestration</span>
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

          <section className="command-bar" aria-label="Create workflow request">
            <div>
              <span>Prompt to backend task</span>
              <strong>Start a CreateTaskRequest and watch managers, workers, and status updates</strong>
            </div>
            <button type="button" aria-label="Run workflow request">
              <ArrowRight size={22} />
            </button>
          </section>

          <section className="node-canvas" id="node-canvas" aria-label="React Flow backend nodes">
            <WorkflowCanvas />
          </section>

          <section className="active-agents" aria-label="Active agents list">
            <div className="active-agents-header">
              <div className="active-agents-title">
                <Activity size={16} />
                <span>Active agents</span>
              </div>
              <button className="terminal-button" type="button">
                <Plus size={15} />
                New flow
              </button>
            </div>
            <div className="active-agent-list">
              {activeAgents.map((agent) => (
                <article className="active-agent-row" key={agent.agent_id}>
                  <div>
                    <span>{agent.role}</span>
                    <strong>{agent.agent_id}</strong>
                  </div>
                  <div>
                    <span>Status</span>
                    <strong>{agent.status}</strong>
                  </div>
                  <div>
                    <span>Progress</span>
                    <strong>{agent.progress}</strong>
                  </div>
                  <div>
                    <span>Endpoint</span>
                    <strong>{agent.endpoint}</strong>
                  </div>
                  <div>
                    <span>task_id</span>
                    <strong>{agent.task_id}</strong>
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
              <div className="market-row" key={metric.label}>
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
                <strong>Worker code</strong>
              </div>
              <PanelRight size={18} />
            </div>
            <div className="price-line">
              <strong>74</strong>
              <span>% progress from TaskUpdate</span>
              <em>processing</em>
            </div>
            <div className="market-bars" aria-hidden="true">
              <i /><i /><i /><i /><i /><i /><i /><i /><i /><i /><i /><i />
            </div>
          </section>

          <section className="guard-stack" aria-label="Backend endpoints">
            {backendEndpoints.map((endpoint) => (
              <div className="guard-row" key={endpoint.label}>
                {endpoint.label === 'Task stream' ? (
                  <Zap size={16} />
                ) : endpoint.label === 'Task status' ? (
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
              <strong>task_create</strong>
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
