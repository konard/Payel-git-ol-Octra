import {
  Activity,
  ArrowRight,
  BarChart3,
  Bell,
  Bot,
  ChevronDown,
  CircleDollarSign,
  Code2,
  Cpu,
  FileText,
  Github,
  Globe2,
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

const quotes = [
  { pair: 'OCTRA', price: '78.0725', move: '+2.20%', tone: 'up' },
  { pair: 'AGENTS', price: '124k', move: '+15%', tone: 'up' },
  { pair: 'RISK', price: '0.85', move: '-4%', tone: 'down' },
  { pair: 'LATENCY', price: '850ms', move: '+45ms', tone: 'warn' },
  { pair: 'MERGE', price: '21', move: '+6', tone: 'up' },
  { pair: 'GUARD', price: '2.4k', move: '-4%', tone: 'down' },
  { pair: 'SPEND', price: '$452', move: '-$120', tone: 'up' },
];

const workflowNodes = [
  {
    title: 'Prompt ingress',
    meta: 'Slack, GitHub, API',
    metric: '124k events',
    state: 'Market open',
    icon: Bot,
    tone: 'blue',
    position: 'node-ingress',
  },
  {
    title: 'Policy guard',
    meta: 'PII, injection, schema',
    metric: '2.4k blocked',
    state: 'Strict',
    icon: ShieldCheck,
    tone: 'teal',
    position: 'node-guard',
  },
  {
    title: 'Dynamic router',
    meta: 'Intent and budget matrix',
    metric: '68% GPT-4o',
    state: 'Live',
    icon: Workflow,
    tone: 'amber',
    position: 'node-router',
  },
  {
    title: 'Model desk',
    meta: 'GPT-4o, Claude, Gemini',
    metric: '850ms avg',
    state: 'Balanced',
    icon: Cpu,
    tone: 'violet',
    position: 'node-models',
  },
  {
    title: 'Review gate',
    meta: 'Diffs, screenshots, PR notes',
    metric: '21 merges',
    state: 'Queued',
    icon: Github,
    tone: 'green',
    position: 'node-review',
  },
  {
    title: 'Release output',
    meta: 'Docs, pull request, deploy',
    metric: '$452 cost',
    state: 'Ready',
    icon: FileText,
    tone: 'red',
    position: 'node-output',
  },
];

const toolItems = [
  { label: 'Workflows', icon: Workflow },
  { label: 'Signals', icon: BarChart3 },
  { label: 'Models', icon: Bot },
  { label: 'Security', icon: ShieldCheck },
  { label: 'Code', icon: Code2 },
  { label: 'Settings', icon: Settings },
];

const timelineRows = [
  ['10:42', 'Prompt injection attempt', 'Blocked', 'down'],
  ['10:31', 'Billing refund flow', 'GPT-4o', 'up'],
  ['10:15', 'Schema repair', 'Retried', 'warn'],
  ['09:57', 'Release notes generated', 'Ready', 'up'],
];

export default function HomePage() {
  return (
    <main className="site-shell workspace-home">
      <header className="tv-header">
        <a className="tv-brand" href="/" aria-label="Octra home">
          <img src="/assets/icon.png" alt="" />
          <span>Octra</span>
        </a>

        <label className="tv-search">
          <Search size={18} />
          <span className="sr-only">Search Octra</span>
          <input placeholder="Search flows, agents, pull requests..." />
        </label>

        <nav className="tv-nav" aria-label="Primary navigation">
          <a href="/dashboard">Products</a>
          <a href="#node-canvas">Community</a>
          <a href="#quote-board">Markets</a>
          <a href="/auth">Brokers</a>
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

      <section className="workspace-frame" aria-label="Octra workflow terminal">
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
              <span>Pipelines</span>
              <span>Autonomous support desk</span>
              <span>Live graph</span>
            </div>
            <div className="canvas-actions">
              <button className="terminal-button" type="button">
                <Plus size={15} />
                New flow
              </button>
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
              <span>Ask Octra</span>
              <strong>Build a visible agent workflow from prompt to pull request</strong>
            </div>
            <button type="button" aria-label="Run workflow request">
              <ArrowRight size={22} />
            </button>
          </section>

          <section className="node-canvas" id="node-canvas" aria-label="Visible pipeline nodes">
            <svg className="node-links" viewBox="0 0 1000 520" preserveAspectRatio="none" aria-hidden="true">
              <path d="M132 138 C260 138 262 214 390 214" />
              <path d="M132 138 C290 138 332 88 502 88" />
              <path d="M392 214 C498 214 506 342 604 342" />
              <path d="M502 88 C614 92 670 180 732 244" />
              <path d="M604 342 C712 342 720 266 842 266" />
              <path d="M732 244 C784 244 790 266 842 266" />
            </svg>

            {workflowNodes.map((node) => (
              <article className={`workflow-node ${node.position} node-${node.tone}`} key={node.title}>
                <div className="node-title">
                  <node.icon size={18} />
                  <span>{node.title}</span>
                </div>
                <p>{node.meta}</p>
                <div className="node-foot">
                  <strong>{node.metric}</strong>
                  <span>{node.state}</span>
                </div>
              </article>
            ))}
          </section>

          <section className="execution-strip" aria-label="Live execution log">
            <div className="strip-heading">
              <Activity size={16} />
              <span>Live execution tape</span>
            </div>
            <div className="execution-rows">
              {timelineRows.map(([time, event, status, tone]) => (
                <div className="execution-row" key={`${time}-${event}`}>
                  <span>{time}</span>
                  <strong>{event}</strong>
                  <em className={tone}>{status}</em>
                </div>
              ))}
            </div>
          </section>
        </section>

        <aside className="home-market-panel" id="quote-board" aria-label="Octra quote board">
          <div className="market-heading">
            <span>Quote board</span>
            <ChevronDown size={16} />
          </div>
          <div className="market-columns">
            <span>Instrument</span>
            <span>Last</span>
            <span>Move</span>
          </div>
          <div className="quote-stack">
            {quotes.map((quote) => (
              <div className="market-row" key={quote.pair}>
                <span>{quote.pair}</span>
                <strong>{quote.price}</strong>
                <em className={quote.tone}>{quote.move}</em>
              </div>
            ))}
          </div>

          <section className="selected-node" aria-label="Selected node details">
            <div className="selected-topline">
              <div>
                <span>Selected node</span>
                <strong>Dynamic router</strong>
              </div>
              <PanelRight size={18} />
            </div>
            <div className="price-line">
              <strong>68.0</strong>
              <span>% GPT-4o route</span>
              <em className="up">+15%</em>
            </div>
            <div className="market-bars" aria-hidden="true">
              <i /><i /><i /><i /><i /><i /><i /><i /><i /><i /><i /><i />
            </div>
          </section>

          <section className="guard-stack" aria-label="Guardrail controls">
            <div className="guard-row">
              <LockKeyhole size={16} />
              <span>Auth protocol</span>
              <strong>Bearer strict</strong>
            </div>
            <div className="guard-row">
              <CircleDollarSign size={16} />
              <span>Budget guard</span>
              <strong>$12 / run</strong>
            </div>
            <div className="guard-row">
              <Globe2 size={16} />
              <span>Region</span>
              <strong>Global</strong>
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
