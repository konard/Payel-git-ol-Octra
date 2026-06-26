import {
  Activity,
  ArrowRight,
  BarChart3,
  Bot,
  ChevronDown,
  CircleDollarSign,
  Cpu,
  FileText,
  Github,
  Globe2,
  LockKeyhole,
  Plus,
  Search,
  ShieldCheck,
  Sparkles,
  Workflow,
  Zap,
} from 'lucide-react';

const quotes = [
  { pair: 'OCTRA', price: '78.0725', move: '+2.20%', tone: 'up' },
  { pair: 'BUILD', price: '124k', move: '+15%', tone: 'up' },
  { pair: 'RISK', price: '0.85', move: '-4%', tone: 'down' },
  { pair: 'LATENCY', price: '850ms', move: '+45ms', tone: 'warn' },
  { pair: 'MERGE', price: '21', move: '+6', tone: 'up' },
];

const capabilities = [
  {
    icon: Workflow,
    title: 'Pipeline view',
    detail: 'Every prompt, agent, model, and delivery step lands in one visible flow.',
  },
  {
    icon: ShieldCheck,
    title: 'Guarded execution',
    detail: 'Policy checks, provider routing, and review gates sit beside the work queue.',
  },
  {
    icon: FileText,
    title: 'Readable outcomes',
    detail: 'Plans, diffs, documents, screenshots, and pull requests stay connected.',
  },
];

const modelBars = [
  { label: 'GPT-4o', value: 68, color: 'blue' },
  { label: 'Claude', value: 22, color: 'green' },
  { label: 'Gemini', value: 10, color: 'amber' },
];

export default function LandingPage() {
  return (
    <main className="site-shell landing-shell">
      <header className="top-nav">
        <a className="brand-link" href="/" aria-label="Octra home">
          <img src="/assets/icon.png" alt="" className="brand-mark" />
          <span>Octra</span>
        </a>
        <nav className="nav-links" aria-label="Primary navigation">
          <a href="#signals">Signals</a>
          <a href="#pipelines">Pipelines</a>
          <a href="#models">Models</a>
          <a href="/dashboard">Dashboard</a>
        </nav>
        <div className="nav-actions">
          <a className="icon-button" href="/dashboard" aria-label="Open dashboard">
            <BarChart3 size={18} />
          </a>
          <a className="text-button" href="/auth">
            <Sparkles size={17} />
            <span>Sign in</span>
          </a>
        </div>
      </header>

      <section className="hero-section" aria-labelledby="hero-title">
        <div className="market-wall" aria-hidden="true">
          <div className="pulse pulse-a" />
          <div className="pulse pulse-b" />
          <div className="pulse pulse-c" />
          <div className="pulse pulse-d" />
        </div>

        <aside className="quote-rail" aria-label="Live Octra signals">
          <div className="rail-title">
            <span>Quote board</span>
            <ChevronDown size={15} />
          </div>
          <div className="quote-head">
            <span>Instrument</span>
            <span>Last</span>
            <span>Move</span>
          </div>
          {quotes.map((quote) => (
            <div className="quote-row" key={quote.pair}>
              <span>{quote.pair}</span>
              <strong>{quote.price}</strong>
              <span className={`quote-${quote.tone}`}>{quote.move}</span>
            </div>
          ))}
          <div className="mini-card">
            <div>
              <span>Market is open</span>
              <strong>Autonomy index +18%</strong>
            </div>
            <Activity size={20} />
          </div>
        </aside>

        <div className="hero-copy">
          <p className="eyebrow">AI delivery cockpit</p>
          <h1 id="hero-title">Octra</h1>
          <p className="hero-lede">
            A dense command surface for teams that want to plan, route, review, and ship AI-assisted work with the same clarity as a market terminal.
          </p>
          <div className="hero-actions">
            <a className="primary-button" href="/dashboard">
              <BarChart3 size={18} />
              <span>Open dashboard</span>
            </a>
            <a className="secondary-button" href="/auth">
              <Github size={18} />
              <span>Create account</span>
            </a>
          </div>
        </div>

        <div className="terminal-preview" aria-label="Octra dashboard preview">
          <div className="terminal-topline">
            <div className="breadcrumb">
              <Bot size={16} />
              <span>Pipelines</span>
              <span>Customer Support Bot</span>
            </div>
            <label className="search-field">
              <Search size={15} />
              <span className="sr-only">Search pipelines</span>
              <input placeholder="Search..." />
            </label>
            <button className="small-command" type="button">
              <Plus size={15} />
              New flow
            </button>
          </div>

          <div className="terminal-grid">
            <section className="metric-strip" id="signals">
              <article className="metric-card">
                <span>Processed prompts</span>
                <strong>124k</strong>
                <div className="spark-bars blue-bars" aria-hidden="true">
                  <i /><i /><i /><i /><i />
                </div>
              </article>
              <article className="metric-card">
                <span>Avg. latency</span>
                <strong>850ms</strong>
                <div className="spark-bars violet-bars" aria-hidden="true">
                  <i /><i /><i /><i /><i />
                </div>
              </article>
              <article className="metric-card">
                <span>Total cost</span>
                <strong>$452.10</strong>
                <div className="spark-bars amber-bars" aria-hidden="true">
                  <i /><i /><i /><i /><i />
                </div>
              </article>
            </section>

            <section className="traffic-panel" id="pipelines" aria-label="Endpoint traffic and health">
              <div className="panel-heading">
                <span>Endpoint traffic and health</span>
                <button className="ghost-command" type="button">View details</button>
              </div>
              <div className="flow-map">
                <div className="flow-node gateway">API Gateway</div>
                <div className="flow-split">
                  {modelBars.map((model) => (
                    <div className={`flow-node model-node ${model.color}`} key={model.label}>
                      <strong>{model.label}</strong>
                      <span>{model.value}% traffic</span>
                    </div>
                  ))}
                </div>
              </div>
            </section>

            <section className="architecture-panel" id="models" aria-label="Active pipeline architecture">
              <div className="panel-heading">
                <span>Active pipeline architecture</span>
                <button className="ghost-command" type="button">Edit</button>
              </div>
              <div className="rule-stack">
                <div className="rule-row">
                  <Cpu size={16} />
                  <span>Ingress endpoint</span>
                  <strong>/v1/chat/completions</strong>
                </div>
                <div className="rule-row">
                  <LockKeyhole size={16} />
                  <span>Auth protocol</span>
                  <strong>Bearer strict</strong>
                </div>
                <div className="rule-row">
                  <Zap size={16} />
                  <span>Dynamic router</span>
                  <strong>GPT-4o primary</strong>
                </div>
                <div className="rule-row">
                  <CircleDollarSign size={16} />
                  <span>Budget guard</span>
                  <strong>$12.00 / run</strong>
                </div>
              </div>
            </section>
          </div>
        </div>
      </section>

      <section className="capability-band" aria-label="Octra capabilities">
        {capabilities.map((item) => (
          <article className="capability-card" key={item.title}>
            <item.icon size={22} />
            <div>
              <h2>{item.title}</h2>
              <p>{item.detail}</p>
            </div>
          </article>
        ))}
        <a className="capability-link" href="/auth">
          <Globe2 size={20} />
          <span>Join Octra</span>
          <ArrowRight size={18} />
        </a>
      </section>

      <footer className="site-footer">
        <span>Octra</span>
        <span>Signal-first automation for software teams</span>
      </footer>
    </main>
  );
}
