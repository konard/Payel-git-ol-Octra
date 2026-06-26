import {
  ArrowRight,
  Bot,
  Code2,
  Cpu,
  Database,
  GitBranch,
  Globe,
  LockKeyhole,
  PanelRight,
  ShieldCheck,
  Sparkles,
  Workflow,
  Zap,
} from 'lucide-react';

const capabilities = [
  {
    icon: Workflow,
    title: 'MCP environments',
    detail: 'Provision isolated Nix profiles per user with pinned tool versions.',
  },
  {
    icon: Bot,
    title: 'CLI process lifecycle',
    detail: 'Keep Claude Code and other CLI processes warm, route stdin/stdout via Redis state.',
  },
  {
    icon: ShieldCheck,
    title: 'API token auth',
    detail: 'Validate octra-api-token on every request, map to user environment automatically.',
  },
  {
    icon: Cpu,
    title: 'Per-request skills',
    detail: 'Let each prompt enable only the tools it needs — filesystem, GitHub, and more.',
  },
  {
    icon: Database,
    title: 'Redis-backed state',
    detail: 'Track PID, TTL, and port for every running CLI instance in Redis.',
  },
  {
    icon: Code2,
    title: 'Nix + CLI integration',
    detail: 'Install skill packages declaratively with Nix, spawn CLI as a subprocess.',
  },
];

export default function LandingPage() {
  return (
    <main className="landing-shell">
      <header className="top-nav">
        <a className="brand-link" href="/" aria-label="Octra home">
          <img src="/assets/octra-node-logo.svg" alt="" className="brand-mark" />
          <span>Octra</span>
        </a>

        <nav className="nav-links" aria-label="Primary navigation">
          <a href="#capabilities">Capabilities</a>
          <a href="#preview">Preview</a>
          <a href="/auth">Sign in</a>
          <a href="/app">App</a>
        </nav>

        <div className="nav-actions">
          <a className="text-button" href="/auth">
            <LockKeyhole size={17} />
            <span>Sign in</span>
          </a>
          <a className="primary-button" href="/dashboard">
            <Sparkles size={17} />
            <span>Dashboard</span>
            <ArrowRight size={16} />
          </a>
        </div>
      </header>

      <section className="hero-section" id="preview">
        <div className="market-wall" aria-hidden="true">
          <i className="pulse pulse-a" />
          <i className="pulse pulse-b" />
          <i className="pulse pulse-c" />
          <i className="pulse pulse-d" />
        </div>

        <div className="hero-copy">
          <p className="eyebrow">MCP aggregator monolith</p>
          <h1>
            Deploy your agent
            <br />
            on a VPS. Without a VPS!
          </h1>
          <p className="hero-lede">
            Deploy your agents on Octra for 24/7 access and earn money from it.
          </p>
          <div className="hero-actions">
            <a className="primary-button" href="/auth">
              <Zap size={18} />
              <span>Get started</span>
              <ArrowRight size={17} />
            </a>
            <a className="secondary-button" href="/app">
              <PanelRight size={18} />
              <span>Open workspace</span>
            </a>
          </div>
        </div>

        <div className="terminal-preview">
          <div className="terminal-topline">
            <div className="breadcrumb">
              <span>environments</span>
              <span>/</span>
              <span>usr_742</span>
              <span>/</span>
              <span>cli</span>
            </div>
            <label className="search-field">
              <Globe size={16} />
              <input placeholder="Search environments, skills, endpoints..." readOnly />
            </label>
            <div className="small-command">
              <GitBranch size={14} />
              <span>stable</span>
            </div>
          </div>

          <div className="terminal-grid">
            <div className="metric-strip">
              <div className="metric-card">
                <span>Active environments</span>
                <strong>18</strong>
                <div className="spark-bars rise-bars" aria-hidden="true">
                  <i /><i /><i /><i /><i />
                </div>
              </div>
              <div className="metric-card">
                <span>Chat requests</span>
                <strong>124k</strong>
                <div className="spark-bars rise-bars" aria-hidden="true">
                  <i /><i /><i /><i /><i />
                </div>
              </div>
              <div className="metric-card">
                <span>CLI processes</span>
                <strong>7</strong>
                <div className="spark-bars notice-bars" aria-hidden="true">
                  <i /><i /><i /><i /><i />
                </div>
              </div>
            </div>

            <div className="traffic-panel">
              <div className="panel-heading">
                <span>Request flow</span>
                <span className="ghost-command">live</span>
              </div>
              <div className="flow-map">
                <div className="flow-node gateway">POST /api/chat</div>
                <div className="flow-node">Middleware → token check</div>
                <div className="flow-split">
                  <div className="model-node neutral">
                    <span>environment</span>
                    Nix profile
                  </div>
                  <div className="model-node ok">
                    <span>CLI</span>
                    claude code
                  </div>
                  <div className="model-node caution">
                    <span>Redis</span>
                    cli_state
                  </div>
                </div>
              </div>
            </div>

            <div className="architecture-panel">
              <div className="panel-heading">
                <span>Architecture</span>
              </div>
              <div className="architecture-list">
                <div className="rule-row">
                  <Globe size={16} />
                  <span>Runtime</span>
                  <strong>Node.js + Next.js</strong>
                </div>
                <div className="rule-row">
                  <Database size={16} />
                  <span>Cache</span>
                  <strong>Redis</strong>
                </div>
                <div className="rule-row">
                  <Code2 size={16} />
                  <span>Provisioning</span>
                  <strong>Nix</strong>
                </div>
                <div className="rule-row">
                  <LockKeyhole size={16} />
                  <span>Auth</span>
                  <strong>JWT + OAuth</strong>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="capability-band" id="capabilities">
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
          <span>Explore all</span>
          <ArrowRight size={18} />
        </a>
      </section>

      <footer className="site-footer">
        <span>Octra</span>
        <span>MCP aggregator monolith</span>
        <span>MIT License</span>
      </footer>
    </main>
  );
}
