'use client';

import {
  Activity,
  ChevronDown,
  Layers3,
  LineChart,
  LockKeyhole,
  Pause,
  Plus,
  Search,
  Settings,
  Sparkles,
  Trash2,
  Workflow,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { UserBalance } from '../components/UserBalance';
import { WelcomeModal } from '../components/WelcomeModal';
import { WorkflowCanvas, type WorkflowCanvasItem } from '../components/WorkflowCanvas';
import { CreateEnvironmentModal } from '../components/CreateEnvironmentModal';
import { IconButton } from '../components/IconButton';
import { CatalogSearchModal } from '../components/CatalogSearchModal';
import { createDashboardEnvironment, listDashboardEnvironments, patchDashboardEnvironment, deleteDashboardEnvironment, getCanvas, putCanvas, type DashboardEnvironment, type CanvasNodeRequest } from '../server/environments';
import type { CatalogItem } from '../server/catalog';
import { ASSETS } from '../config/images';
import { ROUTES } from '../config/routes';
import { fetchMe } from '../server/user';

const toolItems = [
  { label: 'Environments', icon: Workflow, href: ROUTES.DASHBOARD_ENVIRONMENTS },
  { label: 'Streams', icon: Activity, href: '/dashboard/metrics' },
];

export default function HomePage() {
  const router = useRouter();
  const [panelOpen, setPanelOpen] = useState(false);
  const [metricsOpen, setMetricsOpen] = useState(false);
  const [endpointsOpen, setEndpointsOpen] = useState(false);
  const [deploymentsOpen, setDeploymentsOpen] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);
  const [username, setUsername] = useState('');
  const [showWelcome, setShowWelcome] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [createError, setCreateError] = useState('');
  const [envs, setEnvs] = useState<DashboardEnvironment[]>([]);
  const [envsLoading, setEnvsLoading] = useState(true);
  const [selectedEnv, setSelectedEnv] = useState(() => getCookie('octra_selected_env'));
  const [searchOpen, setSearchOpen] = useState(false);
  const [canvasItems, setCanvasItems] = useState<WorkflowCanvasItem[]>([]);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout>>();

  function selectEnv(id: string) {
    setSelectedEnv(id);
    document.cookie = `octra_selected_env=${id}; path=/; max-age=31536000; SameSite=Lax`;
  }

  const saveCanvas = useCallback(async (envId: string, items: WorkflowCanvasItem[]) => {
    if (!envId) return;
    const nodes: CanvasNodeRequest[] = items.map((item, i) => ({
      item_id: item.id,
      kind: item.kind,
      name: item.name,
      detail: item.detail,
      description: item.description,
      meta: item.meta as Record<string, string | null>,
      position_x: item.positionX ?? 0,
      position_y: item.positionY ?? 0,
      sort_order: i,
    }));
    await putCanvas(envId, nodes);
  }, []);

  const handleCanvasItemsChange = useCallback((items: WorkflowCanvasItem[]) => {
    setCanvasItems(items);
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    saveTimerRef.current = setTimeout(() => {
      const envId = getCookie('octra_selected_env');
      if (envId) saveCanvas(envId, items);
    }, 500);
  }, [saveCanvas]);

  async function handlePause(id: string) {
    const res = await patchDashboardEnvironment(id, { active: false });
    if (!res.ok) return;
    setEnvs((prev) => prev.filter((e) => e.id !== id));
    if (selectedEnv === id) {
      const next = envs.find((e) => e.id !== id);
      if (next) selectEnv(next.id);
      else document.cookie = `octra_selected_env=; path=/; max-age=0`;
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this environment?')) return;
    await deleteDashboardEnvironment(id);
    setEnvs((prev) => prev.filter((e) => e.id !== id));
    if (selectedEnv === id) {
      const next = envs.find((e) => e.id !== id);
      if (next) selectEnv(next.id);
      else document.cookie = `octra_selected_env=; path=/; max-age=0`;
    }
  }

  const activeEnvs = envs.filter((e) => e.active);

  async function handleCreate(name: string, visibility: 'private' | 'public') {
    setCreateError('');
    const res = await createDashboardEnvironment(name, visibility);
    if (!res.ok) {
      const text = await res.text();
      setCreateError(text || 'Failed to create environment');
      return;
    }
    const env = await res.json();
    document.cookie = `octra_selected_env=${env.id}; path=/; max-age=31536000; SameSite=Lax`;
    setEnvs((prev) => [env, ...prev]);
    setSelectedEnv(env.id);
    setShowCreate(false);
  }

  function handleCatalogSelect(item: CatalogItem) {
    setCanvasItems((prev) => {
      const next = [
        ...prev,
        {
          id: `${item.type}-${item.id}-${Date.now()}`,
          kind: item.type,
          name: item.name,
          detail: item.subtitle || item.description,
          description: item.description,
          meta: {
            provider: item.key,
            base_url: item.base_url,
            model: item.default_model,
            cli: item.nix_attr || item.install_cmd,
            skill: item.skill_id || item.source,
            auth: item.api_key ? 'set' : item.auth_env,
          },
        },
      ];
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      saveTimerRef.current = setTimeout(() => {
        const envId = getCookie('octra_selected_env');
        if (envId) saveCanvas(envId, next);
      }, 500);
      return next;
    });
  }

  function fetchEnvs() {
    return listDashboardEnvironments().then(async (res) => {
      if (res.ok) setEnvs(await res.json());
    }).catch(() => {});
  }

  useEffect(() => {
    const hasToken = ['octra_access_token', 'access_token'].some((key) => window.localStorage.getItem(key));
    setIsAuthed(hasToken);

    if (hasToken) {
      setEnvsLoading(true);
      fetchEnvs().finally(() => setEnvsLoading(false));
    } else {
      setEnvsLoading(false);
    }

    if (window.localStorage.getItem('octra_show_welcome')) {
      setShowWelcome(true);
      window.localStorage.removeItem('octra_show_welcome');
    }

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

    function onVisibility() {
      if (document.visibilityState === 'visible') fetchEnvs();
    }
    document.addEventListener('visibilitychange', onVisibility);
    return () => document.removeEventListener('visibilitychange', onVisibility);
  }, []);

  useEffect(() => {
    if (!selectedEnv) {
      setCanvasItems([]);
      return;
    }
    getCanvas(selectedEnv).then(async (res) => {
      if (!res.ok) return;
      const nodes = await res.json();
      setCanvasItems(
        nodes.map((n: any) => ({
          id: n.item_id,
          kind: n.kind,
          name: n.name,
          detail: n.detail,
          description: n.description,
          meta: n.meta ?? undefined,
          positionX: n.position_x,
          positionY: n.position_y,
        })),
      );
    }).catch(() => {});
  }, [selectedEnv]);

  return (
    <main className="site-shell workspace-home">
      <header className="tv-header">
        <a className="tv-brand" href="/" aria-label="Octra home">
          <img src={ASSETS.LOGO} alt="" />
          <span>Octra</span>
        </a>

        <button type="button" className="tv-search" onClick={() => setSearchOpen(true)}>
          <Search size={18} />
          <span className="tv-search-placeholder">Search environments, skills, endpoints...</span>
        </button>

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
            <WorkflowCanvas items={canvasItems} onItemsChange={handleCanvasItemsChange} />
          </section>

          <section className="active-environments" aria-label="Active environments list">
            <div className="active-environments-header">
              <div className="active-environments-title">
                <span>Active environments</span>
              </div>
              <button className="terminal-button" onClick={() => setShowCreate(true)}>
                <Plus size={15} />
                New
              </button>
            </div>
            {envsLoading ? (
              <p className="empty-flows-message">Loading environments…</p>
            ) : activeEnvs.length === 0 ? (
              <p className="empty-flows-message">You don't have any active environments.</p>
            ) : (
              <div className="active-environment-list">
                {activeEnvs.map((env) => (
                  <article
                    className={`active-environment-row${selectedEnv === env.id ? ' selected' : ''}`}
                    key={env.id}
                    onClick={() => selectEnv(env.id)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => e.key === 'Enter' && selectEnv(env.id)}
                  >
                    <div>
                      <span>Active</span>
                      <strong className="environment-name-line">
                        <LockKeyhole size={14} aria-hidden="true" />
                        {env.name}
                      </strong>
                    </div>
                    <div>
                      <span>environment_id</span>
                      <strong>{env.id}</strong>
                    </div>
                    <div className="environment-actions">
                      <IconButton variant="warning" onClick={(e) => { e.stopPropagation(); handlePause(env.id); }} aria-label={`Pause ${env.name}`}>
                        <Pause size={15} />
                      </IconButton>
                      <IconButton variant="danger" onClick={(e) => { e.stopPropagation(); handleDelete(env.id); }} aria-label={`Delete ${env.name}`}>
                        <Trash2 size={15} />
                      </IconButton>
                    </div>
                  </article>
                ))}
              </div>
            )}
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
                  actionHref={ROUTES.DASHBOARD_ENVIRONMENTS}
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

      {showCreate && (
        <CreateEnvironmentModal
          onClose={() => setShowCreate(false)}
          onCreate={handleCreate}
          error={createError}
        />
      )}

      {showWelcome && (
        <WelcomeModal
          username={username}
          onClose={() => setShowWelcome(false)}
        />
      )}

      <CatalogSearchModal
        open={searchOpen}
        onClose={() => setSearchOpen(false)}
        onSelect={handleCatalogSelect}
      />
    </main>
  );
}

function getCookie(name: string): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : '';
}
