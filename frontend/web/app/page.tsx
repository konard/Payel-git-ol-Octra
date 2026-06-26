import {
  ArrowRight,
  Bot,
  Code2,
  Cpu,
  Database,
  Layers,
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

        <div className="deployment-map">
          <div className="deployment-node root-node">
            <Layers size={17} />
            API Gateway
            <span>124k reqs</span>
          </div>
          <div className="deployment-branches" aria-hidden="true" />
          <div className="deployment-leaves">
            <div className="deployment-node healthy">
              <Bot size={17} />
              GPT-4o
              <span>68% traffic</span>
            </div>
            <div className="deployment-node healthy">
              <Sparkles size={17} />
              Claude 3.5
              <span>22% traffic</span>
            </div>
            <div className="deployment-node warning">
              <Zap size={17} />
              Gemini Flash
              <span>10% traffic</span>
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
