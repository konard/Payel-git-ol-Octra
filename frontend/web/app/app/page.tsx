'use client';

import {
  Activity,
  ChevronDown,
  Copy,
  Layers3,
  LineChart,
  LockKeyhole,
  Pause,
  Plus,
  Search,
  Settings,
  Sparkles,
  Trash2,
  Trophy,
  Workflow,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { EmptyDataPanel } from '../components/EmptyDataPanel';
import { RequestMetricsOverview } from '../components/RequestMetricsOverview';
import { UserBalance } from '../components/UserBalance';
import { WelcomeModal } from '../components/WelcomeModal';
import { WorkflowCanvas, type WorkflowCanvasItem } from '../components/WorkflowCanvas';
import type { Edge } from '@xyflow/react';
import { CreateEnvironmentModal } from '../components/CreateEnvironmentModal';
import { IconButton } from '../components/IconButton';
import { CatalogSearchModal } from '../components/CatalogSearchModal';
import { createDashboardEnvironment, listDashboardEnvironments, patchDashboardEnvironment, deleteDashboardEnvironment, getCanvas, putCanvas, type DashboardEnvironment } from '../server/environments';
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
  const [userId, setUserId] = useState('');
  const [showWelcome, setShowWelcome] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [createError, setCreateError] = useState('');
  const [envs, setEnvs] = useState<DashboardEnvironment[]>([]);
  const [envsLoading, setEnvsLoading] = useState(true);
  const [selectedEnv, setSelectedEnv] = useState(() => getCookie('octra_selected_env'));
  const [searchOpen, setSearchOpen] = useState(false);
  const [canvasItems, setCanvasItems] = useState<WorkflowCanvasItem[]>([]);
  const [canvasEdges, setCanvasEdges] = useState<Edge[]>([]);
  const [copiedEnvId, setCopiedEnvId] = useState<string | null>(null);
  const canvasItemsRef = useRef(canvasItems);
  canvasItemsRef.current = canvasItems;
  const canvasEdgesRef = useRef(canvasEdges);
  canvasEdgesRef.current = canvasEdges;
  const saveTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const selectedEnvRef = useRef(selectedEnv);
  selectedEnvRef.current = selectedEnv;
  const wsRef = useRef<WebSocket | null>(null);
  const wsConnectedRef = useRef(false);
  const backendCanvasSupported = useRef(true);

  const itemsToNodes = useCallback((items: WorkflowCanvasItem[]) =>
    items.map((item, i) => ({
      item_id: item.id,
      kind: item.kind,
      name: item.name,
      detail: item.detail || '',
      description: item.description || '',
      meta: item.meta as Record<string, string | null>,
      position_x: item.positionX ?? 0,
      position_y: item.positionY ?? 0,
      sort_order: i,
    })),
  []);

  const addedKeys = useMemo(() => {
    const keys = new Set<string>();
    for (const item of canvasItems) {
      // canvas item id format: `${type}-${catalogId}-${Date.now()}`
      const lastDash = item.id.lastIndexOf('-');
      if (lastDash > 0) keys.add(item.id.slice(0, lastDash));
    }
    return keys;
  }, [canvasItems]);

  // revealSection makes each secondary-rail button independent: it ensures the
  // side panel is open and then toggles only its own collapsible section, so the
  // metrics, endpoints and deployments buttons each control a distinct panel.
  function revealSection(section: 'metrics' | 'endpoints' | 'deployments') {
    setPanelOpen(true);
    if (section === 'metrics') setMetricsOpen((v) => !v);
    else if (section === 'endpoints') setEndpointsOpen((v) => !v);
    else setDeploymentsOpen((v) => !v);
  }

  function selectEnv(id: string) {
    console.log('[canvas] selectEnv: attempting switch to', id, 'current ref:', selectedEnvRef.current);
    if (id === selectedEnvRef.current) {
      console.log('[canvas] selectEnv: same env, skipping');
      return;
    }
    if (saveTimerRef.current) {
      console.log('[canvas] selectEnv: clearing pending save timer');
      clearTimeout(saveTimerRef.current);
    }
    const prevItems = canvasItemsRef.current;
    const prevId = selectedEnvRef.current;
    console.log('[canvas] selectEnv: prevId=', prevId, 'prevItems count=', prevItems.length);
    if (prevId && prevItems.length > 0) {
      console.log('[canvas] selectEnv: saving prev env before switch');
      saveCanvas(prevItems).catch((err: any) => console.error('save before switch failed', err));
    }
    setSelectedEnv(id);
    document.cookie = `octra_selected_env=${id}; path=/; max-age=31536000; SameSite=Lax`;
  }

  const sendWS = useCallback((msg: object) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      return false;
    }
    ws.send(JSON.stringify(msg));
    return true;
  }, []);

  const saveCanvas = useCallback(async (items: WorkflowCanvasItem[]) => {
    const envId = selectedEnvRef.current;
    if (!envId) return;
    const edges = canvasEdgesRef.current;

    if (wsConnectedRef.current) {
      const nodes = itemsToNodes(items);
      const sent = sendWS({ type: 'save', nodes, edges: edges.map(e => ({ source: e.source, target: e.target, sourceHandle: e.sourceHandle, targetHandle: e.targetHandle })) });
      if (sent) return;
    }

    if (!backendCanvasSupported.current) return;

    try {
      const nodes = itemsToNodes(items);
      const res = await putCanvas(envId, nodes);
      if (res.status === 404) {
        backendCanvasSupported.current = false;
      }
    } catch {}
  }, [itemsToNodes, sendWS]);

  const handleCanvasItemsChange = useCallback((items: WorkflowCanvasItem[]) => {
    console.log('[canvas] handleCanvasItemsChange: items count=', items.length, 'ids:', items.map(i => i.id));
    setCanvasItems(items);
    if (saveTimerRef.current) {
      console.log('[canvas] handleCanvasItemsChange: clearing previous save timer');
      clearTimeout(saveTimerRef.current);
    }
    const envId = selectedEnvRef.current;
    if (!envId) {
      console.log('[canvas] handleCanvasItemsChange: no envId, skipping save');
      return;
    }
    saveTimerRef.current = setTimeout(() => {
      console.log('[canvas] handleCanvasItemsChange: save timer fired, saving', items.length, 'items');
      saveCanvas(items).catch((err: any) => console.error('save canvas failed', err));
    }, 500);
    console.log('[canvas] handleCanvasItemsChange: save timer set for 500ms');
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
    const envId = selectedEnvRef.current;
    console.log('[canvas] handleCatalogSelect: envId=', envId, 'item=', { id: item.id, name: item.name, type: item.type });

    if (item.type === 'mcp_server') {
      setCanvasItems((prev) => {
        const next = [
          ...prev,
          {
            id: `${item.type}-${item.id}-${Date.now()}`,
            kind: 'mcp_server' as const,
            name: item.name,
            detail: item.subtitle || item.description,
            description: item.description,
            meta: {
              command: item.install_cmd || item.subtitle,
              transport: 'stdio',
            },
          },
        ];
        if (envId) {
          saveCanvas(next).catch((err: any) => console.error('save canvas failed', err));
        }
        return next;
      });
      return;
    }

    if (item.type === 'adapter') {
      const protocolMap: Record<string, string> = {
        'adapter-websocket': 'websocket',
        'adapter-grpc': 'grpc',
        'adapter-graphql': 'graphql',
      };
      const protocol = protocolMap[item.id] || 'websocket';
      setCanvasItems((prev) => {
        const next = [
          ...prev,
          {
            id: `${item.type}-${item.id}-${Date.now()}`,
            kind: 'adapter' as const,
            name: item.name,
            detail: item.subtitle || item.description,
            description: item.description,
            meta: {
              protocol,
            },
          },
        ];
        if (envId) {
          saveCanvas(next).catch((err: any) => console.error('save canvas failed', err));
        }
        return next;
      });
      return;
    }

    setCanvasItems((prev) => {
      const next = [
        ...prev,
        {
          id: `${item.type}-${item.id}-${Date.now()}`,
          kind: item.type as Exclude<typeof item.type, 'mcp_server' | 'adapter'>,
          name: item.name,
          detail: item.subtitle || item.description,
          description: item.description,
          meta: {
            provider: item.key,
            base_url: item.base_url,
            model: item.default_model,
            cli: item.name,
            skill: item.skill_id || item.source,
            auth: item.api_key ? 'set' : item.auth_env,
          },
        },
      ];
      console.log('[canvas] handleCatalogSelect: new items count=', next.length, 'ids:', next.map(i => i.id));
      if (envId) {
        console.log('[canvas] handleCatalogSelect: immediate save');
        saveCanvas(next).catch((err: any) => console.error('save canvas failed', err));
      } else {
        console.log('[canvas] handleCatalogSelect: no envId, cannot save');
      }
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
    const cachedUserId = window.localStorage.getItem('octra_user_id');
    if (cachedUserId) setUserId(cachedUserId);
    if (cached) {
      setUsername(cached);
    } else if (hasToken) {
      const token = window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
      if (token) {
        fetchMe(token).then(async (res) => {
          if (!res.ok) return;
          const body = await res.json();
          const data = body?.data ?? body;
          const u = data?.username;
          const id = data?.user_id ?? data?.id;
          if (u) {
            window.localStorage.setItem('octra_username', u);
            setUsername(u);
          }
          if (id) {
            window.localStorage.setItem('octra_user_id', id);
            setUserId(id);
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

  async function loadViaRest(envId: string) {
    console.log('[canvas] loadViaRest:', envId);
    try {
      const res = await getCanvas(envId);
      if (res.status === 404) {
        backendCanvasSupported.current = false;
        return;
      }
      if (!res.ok) {
        console.error('load canvas failed', res.status);
        return;
      }
      const nodes = await res.json();
      console.log('[canvas] loadViaRest: loaded', nodes.length, 'nodes');
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
    } catch (err) {
      console.error('load canvas error', err);
    }
  }

  useEffect(() => {
    const old = wsRef.current;
    if (old) {
      old.close();
      wsRef.current = null;
      wsConnectedRef.current = false;
    }

    setCanvasItems([]);

    if (!selectedEnv) {
      console.log('[canvas] no selectedEnv, skipping ws connect');
      return;
    }

    // Try loading from localStorage first (instant)
    setCanvasEdges([]);
    const local = loadCanvasLocal(selectedEnv);
    if (local) {
      console.log('[canvas] loaded', local.items.length, 'nodes,', local.edges.length, 'edges from localStorage');
      setCanvasItems(local.items);
      setCanvasEdges(local.edges);
    }

    const token = window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
    if (!token) {
      console.warn('[canvas] no auth token, loading via rest');
      if (!local) loadViaRest(selectedEnv);
      return;
    }

    const apiBase = process.env.NEXT_PUBLIC_API_URL || '';
    const wsBase = apiBase ? apiBase.replace(/^http/, 'ws') : '';
    const origin = typeof window !== 'undefined' ? window.location.origin : '';
    const base = wsBase || origin.replace(/^http/, 'ws') || 'ws://localhost';
    const wsUrl = `${base}/ws/canvas/${selectedEnv}?token=${encodeURIComponent(token)}`;

    console.log('[canvas] connecting ws:', wsUrl);
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;
    let didOpen = false;

    ws.onopen = () => {
      console.log('[canvas] ws connected');
      didOpen = true;
      wsConnectedRef.current = true;
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        console.log('[canvas] ws msg:', msg.type, msg);

        switch (msg.type) {
          case 'init': {
            const nodes = msg.nodes || [];
            console.log('[canvas] ws init: loaded', nodes.length, 'nodes');
            const items = nodes.map((n: any) => ({
              id: n.item_id,
              kind: n.kind,
              name: n.name,
              detail: n.detail,
              description: n.description,
              meta: n.meta ?? undefined,
              positionX: n.position_x ?? 0,
              positionY: n.position_y ?? 0,
            }));
            setCanvasItems(items);
            saveCanvasLocal(selectedEnvRef.current, items, canvasEdgesRef.current);
            break;
          }
          case 'saved':
            console.log('[canvas] ws save confirmed');
            break;
          case 'error':
            console.error('[canvas] ws error:', msg.error);
            break;
        }
      } catch (e) {
        console.error('[canvas] ws parse error:', e);
      }
    };

    ws.onclose = () => {
      console.log('[canvas] ws disconnected, didOpen=', didOpen);
      wsConnectedRef.current = false;
      if (!didOpen && !local) {
        console.log('[canvas] ws never opened, falling back to rest');
        loadViaRest(selectedEnv);
      }
    };

    ws.onerror = () => {
      console.error('[canvas] ws error event');
    };

    return () => {
      ws.close();
      wsRef.current = null;
      wsConnectedRef.current = false;
    };
  }, [selectedEnv]);

  // Persist canvas items + edges to localStorage on every change
  useEffect(() => {
    if (!selectedEnv) return;
    saveCanvasLocal(selectedEnv, canvasItems, canvasEdges);
  }, [canvasItems, canvasEdges, selectedEnv]);

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
          <a className="icon-button dark-icon" href={ROUTES.PROFILE_LEADERBOARD} aria-label="User leaderboard" title="User leaderboard">
            <Trophy size={18} />
          </a>
          {isAuthed ? (
            <a className="avatar-button" href={userId ? ROUTES.PROFILE_BY_ID(userId) : ROUTES.PROFILE} aria-label="Open profile">{username ? username[0].toUpperCase() : '?'}</a>
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
            <WorkflowCanvas items={canvasItems} onItemsChange={handleCanvasItemsChange} edges={canvasEdges} onEdgesChange={setCanvasEdges} />
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
                      <IconButton variant="success" onClick={(e) => { e.stopPropagation(); navigator.clipboard.writeText(env.id); setCopiedEnvId(env.id); setTimeout(() => setCopiedEnvId(null), 1200); }} aria-label="Copy environment ID">
                        {copiedEnvId === env.id ? <span className="copy-check-inline">✓</span> : <Copy size={15} />}
                      </IconButton>
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
                <RequestMetricsOverview range="7d" env={selectedEnv || undefined} compact />
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
        <button
          type="button"
          className={panelOpen ? 'active' : ''}
          aria-label="Toggle side panel"
          aria-pressed={panelOpen}
          onClick={() => setPanelOpen((v) => !v)}
        >
          <LineChart size={20} />
        </button>
        <button
          type="button"
          className={panelOpen && metricsOpen ? 'active' : ''}
          aria-label="Runtime metrics"
          aria-pressed={panelOpen && metricsOpen}
          onClick={() => revealSection('metrics')}
        >
          <Activity size={20} />
        </button>
        <button
          type="button"
          className={panelOpen && endpointsOpen ? 'active' : ''}
          aria-label="Endpoints"
          aria-pressed={panelOpen && endpointsOpen}
          onClick={() => revealSection('endpoints')}
        >
          <Layers3 size={20} />
        </button>
        <button
          type="button"
          className={panelOpen && deploymentsOpen ? 'active' : ''}
          aria-label="Deployments"
          aria-pressed={panelOpen && deploymentsOpen}
          onClick={() => revealSection('deployments')}
        >
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
        addedKeys={addedKeys}
      />
    </main>
  );
}

function getCookie(name: string): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : '';
}

const CANVAS_STORAGE_PREFIX = 'octra_canvas_';

function saveCanvasLocal(envId: string, items: WorkflowCanvasItem[], edges: Edge[]) {
  try {
    window.localStorage.setItem(CANVAS_STORAGE_PREFIX + envId, JSON.stringify({ items, edges }));
    console.log('[canvas] saved to localStorage for env', envId, items.length, 'items,', edges.length, 'edges');
  } catch (e) {
    console.error('[canvas] localStorage save failed', e);
  }
}

function loadCanvasLocal(envId: string): { items: WorkflowCanvasItem[]; edges: Edge[] } | null {
  try {
    const raw = window.localStorage.getItem(CANVAS_STORAGE_PREFIX + envId);
    if (!raw) return null;
    const data = JSON.parse(raw);
    const items = data.items ?? data ?? [];
    const edges = data.edges ?? [];
    console.log('[canvas] loaded from localStorage for env', envId, items.length, 'items,', edges.length, 'edges');
    return { items, edges };
  } catch (e) {
    console.error('[canvas] localStorage load failed', e);
    return null;
  }
}
