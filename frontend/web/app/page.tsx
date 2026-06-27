import {
  Activity,
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
import { ROUTES } from './config/routes';
import { ASSETS } from './config/images';
import { FAKE_METRICS } from './config/fake-metrics';

const capabilities = [
  {
    icon: Workflow,
    title: 'Deploy AI agents 24/7',
    detail: 'Run Claude Code and other AI agents on a remote server that stays online even when you close your laptop.',
  },
  {
    icon: Bot,
    title: 'Personal MCP endpoints',
    detail: 'Each user gets their own MCP endpoint with isolated environments. Share tools with your team without sharing credentials.',
  },
  {
    icon: ShieldCheck,
    title: 'Token-based access',
    detail: 'Secure every request with API tokens. Each token maps to a specific user environment automatically.',
  },
  {
    icon: Cpu,
    title: 'Per-request tool control',
    detail: 'Choose exactly which tools each prompt can use — filesystem, GitHub, browsing, or your own custom skills.',
  },
  {
    icon: Sparkles,
    title: 'Usage analytics & billing',
    detail: 'Track token usage, monitor active sessions, and manage billing in real time from a single dashboard.',
  },
  {
    icon: Zap,
    title: 'Earn from your agents',
    detail: 'Monetise your deployed agents. Octra handles routing, auth, and provisioning so you focus on building.',
  },
];

export default function LandingPage() {
  return (
    <main className="landing-shell">
      <header className="top-nav">
        <a className="brand-link" href={ROUTES.HOME} aria-label="Octra home">
          <img src={ASSETS.LOGO} alt="" className="brand-mark" />
          <span>Octra</span>
        </a>

        <nav className="nav-links" aria-label="Primary navigation">
          <a href="#capabilities">Capabilities</a>
          <a href="#preview">Preview</a>
          <a href={ROUTES.LOGIN}>Sign in</a>
          <a href={ROUTES.WORKSPACE}>App</a>
        </nav>

        <div className="nav-actions">
          <a className="text-button" href={ROUTES.LOGIN}>
            <LockKeyhole size={17} />
            <span>Sign in</span>
          </a>
          <a className="primary-button" href={ROUTES.DASHBOARD}>
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
          <p className="eyebrow">VPS aggregator</p>
          <h1>
            Deploy your agent
            <br />
            on a VPS. Without a VPS!
          </h1>
          <p className="hero-lede">
            Deploy your agents on Octra for 24/7 access and earn money from it.
          </p>
          <div className="hero-actions">
            <a className="primary-button" href={ROUTES.LOGIN}>
              <Zap size={18} />
              <span>Get started</span>
              <ArrowRight size={17} />
            </a>
            <a className="secondary-button" href={ROUTES.WORKSPACE}>
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
              {FAKE_METRICS.map((m) => (
                <div className={`metric-card metric-tone-${m.tone}`} key={m.label}>
                  <span>{m.label}</span>
                  <strong>{m.value}</strong>
                  <em className="metric-delta">{m.delta}</em>
                  <div className="spark-bars">
                    {m.bars.map((h, i) => (
                      <i key={i} style={{ height: `${h}%` }} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="traffic-panel">
              <div className="panel-heading">
                <span>Request flow</span>
                <span className="ghost-command">live</span>
              </div>
              <div className="flow-map">
                <div className="flow-node gateway">POST /api/chat</div>
                <div className="flow-node">Middleware → token check</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="capability-band" id="capabilities">
        {capabilities.map((item) => {
          const Icon = item.icon;
          return (
            <article className="capability-card" key={item.title}>
              <Icon size={22} />
              <div>
                <h2>{item.title}</h2>
                <p>{item.detail}</p>
              </div>
            </article>
          );
        })}
        <a className="capability-link" href={ROUTES.LOGIN}>
          <span>Explore all</span>
          <ArrowRight size={18} />
        </a>
      </section>

      <footer className="site-footer">
        <span>Octra</span>
        <span>VPS aggregator</span>
        <a href="https://github.com/Payel-git-ol/Octra/blob/main/LICENSE">MIT License</a>
      </footer>
    </main>
  );
}
