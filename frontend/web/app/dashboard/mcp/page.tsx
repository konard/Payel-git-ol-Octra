'use client';

import { useEffect, useState, useCallback } from 'react';
import { Cable, Plus, RefreshCw, Trash2, Wifi, WifiOff, Search, Server, Terminal, BookOpen, FileJson } from 'lucide-react';
import { DashboardShell } from '../DashboardShell';
import {
  listMCPServers,
  deleteMCPServer,
  reconnectMCPServer,
  createMCPServer,
  listMCPCatalog,
  type MCPServer,
  type MCPTool,
  type MCPResource,
  type MCPPrompt,
} from '../../server/mcp';
import { IconButton } from '../../components/IconButton';

type Tab = 'servers' | 'catalog';
type CatalogTab = 'tools' | 'resources' | 'prompts';

export default function MCPPage() {
  const [tab, setTab] = useState<Tab>('servers');
  const [catalogTab, setCatalogTab] = useState<CatalogTab>('tools');

  const [servers, setServers] = useState<MCPServer[]>([]);
  const [serversLoading, setServersLoading] = useState(true);

  const [tools, setTools] = useState<MCPTool[]>([]);
  const [resources, setResources] = useState<MCPResource[]>([]);
  const [prompts, setPrompts] = useState<MCPPrompt[]>([]);
  const [catalogLoading, setCatalogLoading] = useState(false);

  const [showCreate, setShowCreate] = useState(false);
  const [createName, setCreateName] = useState('');
  const [createCommand, setCreateCommand] = useState('');
  const [createUrl, setCreateUrl] = useState('');
  const [createTransport, setCreateTransport] = useState<'http' | 'stdio'>('stdio');
  const [createError, setCreateError] = useState('');

  const [searchQuery, setSearchQuery] = useState('');

  const loadServers = useCallback(() => {
    setServersLoading(true);
    listMCPServers().then(async (res) => {
      if (res.ok) {
        const list: MCPServer[] = await res.json();
        setServers(list);
      }
    }).catch(() => {}).finally(() => setServersLoading(false));
  }, []);

  const loadCatalog = useCallback(() => {
    setCatalogLoading(true);
    listMCPCatalog().then(async (res) => {
      if (res.ok) {
        const data: { tools: MCPTool[]; resources: MCPResource[]; prompts: MCPPrompt[] } = await res.json();
        setTools(data.tools ?? []);
        setResources(data.resources ?? []);
        setPrompts(data.prompts ?? []);
      }
    }).catch(() => {}).finally(() => setCatalogLoading(false));
  }, []);

  useEffect(() => { loadServers(); loadCatalog(); }, [loadServers, loadCatalog]);

  async function handleDelete(id: string) {
    if (!confirm('Delete this MCP server?')) return;
    await deleteMCPServer(id);
    setServers((prev) => prev.filter((s) => s.id !== id));
  }

  async function handleReconnect(id: string) {
    const res = await reconnectMCPServer(id);
    if (res.ok) {
      loadServers();
    }
  }

  async function handleCreate() {
    setCreateError('');
    if (!createName.trim()) {
      setCreateError('Name is required');
      return;
    }
    if (createTransport === 'stdio' && !createCommand.trim()) {
      setCreateError('Command is required for stdio transport');
      return;
    }
    if (createTransport === 'http' && !createUrl.trim()) {
      setCreateError('URL is required for HTTP transport');
      return;
    }

    const settings: Record<string, unknown> = { id: createName.trim(), transport: createTransport };
    if (createTransport === 'stdio') {
      const parts = createCommand.trim().split(/\s+/);
      settings.command = parts[0];
      if (parts.length > 1) settings.args = parts.slice(1);
    } else {
      settings.url = createUrl.trim();
    }

    const res = await createMCPServer(settings);
    if (res.ok) {
      setShowCreate(false);
      setCreateName('');
      setCreateCommand('');
      setCreateUrl('');
      setCreateTransport('stdio');
      loadServers();
    } else {
      const text = await res.text();
      setCreateError(text || 'Failed to create MCP server');
    }
  }

  function filteredServers(): MCPServer[] {
    if (!searchQuery) return servers;
    const q = searchQuery.toLowerCase();
    return servers.filter((s) =>
      s.id.toLowerCase().includes(q) ||
      s.command?.toLowerCase().includes(q) ||
      s.url?.toLowerCase().includes(q)
    );
  }

  function filteredTools(): MCPTool[] {
    if (!searchQuery) return tools;
    const q = searchQuery.toLowerCase();
    return tools.filter((t) =>
      t.name.toLowerCase().includes(q) ||
      t.description?.toLowerCase().includes(q)
    );
  }

  function filteredResources(): MCPResource[] {
    if (!searchQuery) return resources;
    const q = searchQuery.toLowerCase();
    return resources.filter((r) =>
      r.name.toLowerCase().includes(q) ||
      r.uri.toLowerCase().includes(q) ||
      r.description?.toLowerCase().includes(q)
    );
  }

  function filteredPrompts(): MCPPrompt[] {
    if (!searchQuery) return prompts;
    const q = searchQuery.toLowerCase();
    return prompts.filter((p) =>
      p.name.toLowerCase().includes(q) ||
      p.description?.toLowerCase().includes(q)
    );
  }

  return (
    <DashboardShell activeSection="mcp" hideTabs showNotifications={false}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="MCP Servers" style={{ padding: 0, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
          <div className="panel-heading" style={{ padding: '16px 24px', borderBottom: '1px solid var(--line)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <Cable size={18} />
              <span>MCP</span>
            </div>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <div style={{ position: 'relative' }}>
                <Search size={14} style={{ position: 'absolute', left: 8, top: '50%', transform: 'translateY(-50%)', color: 'var(--quiet)' }} />
                <input
                  type="text"
                  placeholder="Search…"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  style={{ padding: '6px 8px 6px 28px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--surface)', color: 'var(--text)', fontSize: '0.82rem', width: 180 }}
                />
              </div>
              <button className="small-command ghost-command" onClick={() => { loadServers(); loadCatalog(); }} aria-label="Refresh">
                <RefreshCw size={14} />
              </button>
            </div>
          </div>

          <div className="dashboard-tabs" style={{ padding: '0 24px', borderBottom: '1px solid var(--line)' }} aria-label="MCP views">
            <button className={tab === 'servers' ? 'active' : ''} onClick={() => setTab('servers')} style={{ background: 'none', border: 'none', cursor: 'pointer', font: 'inherit', color: 'inherit', padding: '10px 16px' }}>
              <Server size={14} style={{ marginRight: 6, verticalAlign: 'middle' }} />
              Servers ({servers.length})
            </button>
            <button className={tab === 'catalog' ? 'active' : ''} onClick={() => setTab('catalog')} style={{ background: 'none', border: 'none', cursor: 'pointer', font: 'inherit', color: 'inherit', padding: '10px 16px' }}>
              <BookOpen size={14} style={{ marginRight: 6, verticalAlign: 'middle' }} />
              Catalog ({tools.length + resources.length + prompts.length})
            </button>
          </div>

          <div style={{ padding: 24, overflow: 'auto', flex: 1, minHeight: 0 }}>
            {tab === 'servers' && (
              <>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                  <span style={{ color: 'var(--muted)', fontSize: '0.85rem' }}>Configured MCP servers</span>
                  <button className="small-command accent-command" onClick={() => { setCreateError(''); setShowCreate(true); }}>
                    <Plus size={14} />
                    Add Server
                  </button>
                </div>

                {serversLoading ? (
                  <p style={{ color: 'var(--muted)' }}>Loading servers…</p>
                ) : filteredServers().length === 0 ? (
                  <div style={{ padding: '32px 16px', textAlign: 'center', border: '1px dashed var(--line)', borderRadius: 8 }}>
                    <Cable size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
                    <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>{searchQuery ? 'No servers match your search.' : 'No MCP servers configured.'}</p>
                    <p style={{ margin: '4px 0 0', color: 'var(--quiet)', fontSize: '0.82rem' }}>{searchQuery ? 'Try a different search.' : 'Add one to connect external tools.'}</p>
                  </div>
                ) : (
                  <div style={{ display: 'grid', gap: 6 }}>
                    {filteredServers().map((server) => (
                      <div key={server.id} className="env-list-item" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 14px' }}>
                        <div style={{ flexShrink: 0, color: server.enabled ? 'var(--accent)' : 'var(--quiet)' }}>
                          {server.enabled ? <Wifi size={16} /> : <WifiOff size={16} />}
                        </div>
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{ fontWeight: 500, fontSize: '0.9rem' }}>{server.id}</div>
                          <div style={{ color: 'var(--quiet)', fontSize: '0.78rem' }}>
                            {server.transport === 'stdio' ? server.command : server.url}
                            {server.status ? ` — ${server.status}` : ''}
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: 4, flexShrink: 0 }}>
                          <button className="icon-button" onClick={() => handleReconnect(server.id)} aria-label="Reconnect" style={{ color: 'var(--muted)' }}>
                            <RefreshCw size={13} />
                          </button>
                          <IconButton variant="danger" onClick={() => handleDelete(server.id)} aria-label="Delete">
                            <Trash2 size={13} />
                          </IconButton>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </>
            )}

            {tab === 'catalog' && (
              <>
                <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
                  {(['tools', 'resources', 'prompts'] as CatalogTab[]).map((ct) => (
                    <button
                      key={ct}
                      className={catalogTab === ct ? 'small-command accent-command' : 'small-command ghost-command'}
                      onClick={() => setCatalogTab(ct)}
                      style={{ fontSize: '0.82rem' }}
                    >
                      {ct === 'tools' && <Terminal size={14} style={{ marginRight: 4 }} />}
                      {ct === 'resources' && <FileJson size={14} style={{ marginRight: 4 }} />}
                      {ct === 'prompts' && <BookOpen size={14} style={{ marginRight: 4 }} />}
                      {ct.charAt(0).toUpperCase() + ct.slice(1)} ({ct === 'tools' ? tools.length : ct === 'resources' ? resources.length : prompts.length})
                    </button>
                  ))}
                </div>

                {catalogLoading ? (
                  <p style={{ color: 'var(--muted)' }}>Loading catalog…</p>
                ) : catalogTab === 'tools' && filteredTools().length === 0 ? (
                  <div style={{ padding: '32px 16px', textAlign: 'center', border: '1px dashed var(--line)', borderRadius: 8 }}>
                    <Terminal size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
                    <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>{searchQuery ? 'No tools match your search.' : 'No tools available.'}</p>
                  </div>
                ) : catalogTab === 'resources' && filteredResources().length === 0 ? (
                  <div style={{ padding: '32px 16px', textAlign: 'center', border: '1px dashed var(--line)', borderRadius: 8 }}>
                    <FileJson size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
                    <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>{searchQuery ? 'No resources match your search.' : 'No resources available.'}</p>
                  </div>
                ) : catalogTab === 'prompts' && filteredPrompts().length === 0 ? (
                  <div style={{ padding: '32px 16px', textAlign: 'center', border: '1px dashed var(--line)', borderRadius: 8 }}>
                    <BookOpen size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
                    <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>{searchQuery ? 'No prompts match your search.' : 'No prompts available.'}</p>
                  </div>
                ) : (
                  <div style={{ display: 'grid', gap: 4 }}>
                    {catalogTab === 'tools' && filteredTools().map((tool, i) => (
                      <div key={tool.name + i} className="env-list-item" style={{ padding: '10px 14px' }}>
                        <div style={{ fontWeight: 500, fontSize: '0.85rem', marginBottom: 2 }}>{tool.name}</div>
                        {tool.description && <div style={{ color: 'var(--quiet)', fontSize: '0.78rem' }}>{tool.description}</div>}
                      </div>
                    ))}
                    {catalogTab === 'resources' && filteredResources().map((r, i) => (
                      <div key={r.uri + i} className="env-list-item" style={{ padding: '10px 14px' }}>
                        <div style={{ fontWeight: 500, fontSize: '0.85rem', marginBottom: 2 }}>{r.name}</div>
                        <div style={{ color: 'var(--quiet)', fontSize: '0.78rem' }}>{r.uri}</div>
                        {r.description && <div style={{ color: 'var(--muted)', fontSize: '0.75rem', marginTop: 2 }}>{r.description}</div>}
                      </div>
                    ))}
                    {catalogTab === 'prompts' && filteredPrompts().map((p, i) => (
                      <div key={p.name + i} className="env-list-item" style={{ padding: '10px 14px' }}>
                        <div style={{ fontWeight: 500, fontSize: '0.85rem', marginBottom: 2 }}>{p.name}</div>
                        {p.description && <div style={{ color: 'var(--quiet)', fontSize: '0.78rem' }}>{p.description}</div>}
                      </div>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>
        </article>
      </section>

      {showCreate && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
          <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--line)', padding: 24, width: 420, maxWidth: '90vw' }}>
            <h3 style={{ margin: '0 0 16px', fontSize: '1rem' }}>Add MCP Server</h3>

            <div style={{ marginBottom: 12 }}>
              <label style={{ display: 'block', fontSize: '0.82rem', color: 'var(--muted)', marginBottom: 4 }}>Server ID / Name</label>
              <input
                type="text"
                value={createName}
                onChange={(e) => setCreateName(e.target.value)}
                placeholder="my-database-server"
                style={{ width: '100%', padding: '8px 10px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--bg)', color: 'var(--text)', fontSize: '0.85rem', boxSizing: 'border-box' }}
              />
            </div>

            <div style={{ marginBottom: 12 }}>
              <label style={{ display: 'block', fontSize: '0.82rem', color: 'var(--muted)', marginBottom: 4 }}>Transport</label>
              <div style={{ display: 'flex', gap: 8 }}>
                {(['stdio', 'http'] as const).map((t) => (
                  <button
                    key={t}
                    className={createTransport === t ? 'small-command accent-command' : 'small-command ghost-command'}
                    onClick={() => setCreateTransport(t)}
                    style={{ fontSize: '0.82rem', textTransform: 'capitalize' }}
                  >
                    {t}
                  </button>
                ))}
              </div>
            </div>

            {createTransport === 'stdio' ? (
              <div style={{ marginBottom: 12 }}>
                <label style={{ display: 'block', fontSize: '0.82rem', color: 'var(--muted)', marginBottom: 4 }}>Command + args</label>
                <input
                  type="text"
                  value={createCommand}
                  onChange={(e) => setCreateCommand(e.target.value)}
                  placeholder="npx @modelcontextprotocol/server-filesystem /path"
                  style={{ width: '100%', padding: '8px 10px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--bg)', color: 'var(--text)', fontSize: '0.85rem', boxSizing: 'border-box' }}
                />
              </div>
            ) : (
              <div style={{ marginBottom: 12 }}>
                <label style={{ display: 'block', fontSize: '0.82rem', color: 'var(--muted)', marginBottom: 4 }}>URL</label>
                <input
                  type="text"
                  value={createUrl}
                  onChange={(e) => setCreateUrl(e.target.value)}
                  placeholder="https://mcp.example.com/sse"
                  style={{ width: '100%', padding: '8px 10px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--bg)', color: 'var(--text)', fontSize: '0.85rem', boxSizing: 'border-box' }}
                />
              </div>
            )}

            {createError && <p style={{ color: 'var(--danger)', fontSize: '0.82rem', margin: '0 0 12px' }}>{createError}</p>}

            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="small-command ghost-command" onClick={() => setShowCreate(false)}>Cancel</button>
              <button className="small-command accent-command" onClick={handleCreate}>Create</button>
            </div>
          </div>
        </div>
      )}
    </DashboardShell>
  );
}
