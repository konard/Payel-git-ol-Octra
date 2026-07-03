'use client';

import { useEffect, useState, useCallback } from 'react';
import { Play, Square, RotateCcw, RefreshCw, Search, Workflow, Clock } from 'lucide-react';
import { DashboardShell } from '../DashboardShell';
import { listRuns, getRun, startRun, cancelRun, restartRun, type RunSummary, type RunSnapshot } from '../../server/runs';
import { listDashboardEnvironments, type DashboardEnvironment } from '../../server/environments';
import { IconButton } from '../../components/IconButton';

const STATUS_COLORS: Record<string, string> = {
  completed: 'var(--success)',
  running: 'var(--accent)',
  suspended: 'var(--warning)',
  failed: 'var(--danger)',
  cancelled: 'var(--quiet)',
};

export default function RunsPage() {
  const [envs, setEnvs] = useState<DashboardEnvironment[]>([]);
  const [selectedEnv, setSelectedEnv] = useState<string>('');
  const [runs, setRuns] = useState<RunSummary[]>([]);
  const [selectedRun, setSelectedRun] = useState<RunSnapshot | null>(null);
  const [loading, setLoading] = useState(true);
  const [runLoading, setRunLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const loadEnvs = useCallback(() => {
    listDashboardEnvironments().then(async (res) => {
      if (res.ok) {
        const list: DashboardEnvironment[] = await res.json();
        setEnvs(list);
        if (list.length > 0 && !selectedEnv) {
          setSelectedEnv(list[0].id);
        }
      }
    }).catch(() => {});
  }, [selectedEnv]);

  const loadRuns = useCallback(() => {
    if (!selectedEnv) return;
    setLoading(true);
    listRuns(selectedEnv).then(async (res) => {
      if (res.ok) {
        const data: { runs: RunSummary[] } = await res.json();
        setRuns(data.runs ?? []);
      }
    }).catch(() => {}).finally(() => setLoading(false));
  }, [selectedEnv]);

  useEffect(() => { loadEnvs(); }, [loadEnvs]);
  useEffect(() => { loadRuns(); }, [loadRuns]);

  async function handleStartRun() {
    if (!selectedEnv) return;
    setRunLoading(true);
    const res = await startRun(selectedEnv);
    if (res.ok) {
      loadRuns();
    }
    setRunLoading(false);
  }

  async function handleCancel(runId: string) {
    if (!selectedEnv) return;
    await cancelRun(selectedEnv, runId);
    loadRuns();
    if (selectedRun?.run_id === runId) setSelectedRun(null);
  }

  async function handleRestart(runId: string) {
    if (!selectedEnv) return;
    await restartRun(selectedEnv, runId);
    loadRuns();
  }

  async function handleSelectRun(runId: string) {
    if (!selectedEnv) return;
    setRunLoading(true);
    const snapRes = await getRun(selectedEnv, runId);
    if (snapRes.ok) {
      setSelectedRun(await snapRes.json());
    }
    setRunLoading(false);
  }

  function filteredRuns(): RunSummary[] {
    if (!searchQuery) return runs;
    const q = searchQuery.toLowerCase();
    return runs.filter((r) =>
      r.run_id.toLowerCase().includes(q) ||
      r.status.toLowerCase().includes(q) ||
      r.workflow_id.toLowerCase().includes(q)
    );
  }

  return (
    <DashboardShell activeSection="runs" hideTabs showNotifications={false}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Workflow Runs" style={{ padding: 0, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
          <div className="panel-heading" style={{ padding: '16px 24px', borderBottom: '1px solid var(--line)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <Play size={18} />
              <span>Workflow Runs</span>
            </div>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <select
                value={selectedEnv}
                onChange={(e) => setSelectedEnv(e.target.value)}
                style={{ padding: '6px 10px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--surface)', color: 'var(--text)', fontSize: '0.82rem' }}
              >
                {envs.length === 0 && <option value="">No environments</option>}
                {envs.map((env) => (
                  <option key={env.id} value={env.id}>{env.name}</option>
                ))}
              </select>
              <div style={{ position: 'relative' }}>
                <Search size={14} style={{ position: 'absolute', left: 8, top: '50%', transform: 'translateY(-50%)', color: 'var(--quiet)' }} />
                <input
                  type="text"
                  placeholder="Search runs…"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  style={{ padding: '6px 8px 6px 28px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--surface)', color: 'var(--text)', fontSize: '0.82rem', width: 160 }}
                />
              </div>
              <button className="small-command ghost-command" onClick={loadRuns} aria-label="Refresh">
                <RefreshCw size={14} />
              </button>
              <button className="small-command accent-command" onClick={handleStartRun} disabled={!selectedEnv || runLoading}>
                <Play size={14} />
                Run
              </button>
            </div>
          </div>

          <div style={{ display: 'flex', flex: 1, minHeight: 0 }}>
            <div style={{ flex: 1, padding: 16, overflow: 'auto', borderRight: '1px solid var(--line)' }}>
              {loading ? (
                <p style={{ color: 'var(--muted)', padding: 16 }}>Loading runs…</p>
              ) : filteredRuns().length === 0 ? (
                <div style={{ padding: '32px 16px', textAlign: 'center' }}>
                  <Workflow size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
                  <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>
                    {searchQuery ? 'No runs match your search.' : 'No runs yet.'}
                  </p>
                  <p style={{ margin: '4px 0 0', color: 'var(--quiet)', fontSize: '0.82rem' }}>
                    {searchQuery ? 'Try a different search.' : 'Click "Run" to start one.'}
                  </p>
                </div>
              ) : (
                <div style={{ display: 'grid', gap: 4 }}>
                  {filteredRuns().map((run) => (
                    <div
                      key={run.run_id}
                      className={`env-list-item${selectedRun?.run_id === run.run_id ? ' active' : ''}`}
                      onClick={() => handleSelectRun(run.run_id)}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => e.key === 'Enter' && handleSelectRun(run.run_id)}
                      style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 14px' }}
                    >
                      <div style={{ width: 8, height: 8, borderRadius: '50%', background: STATUS_COLORS[run.status] || 'var(--quiet)', flexShrink: 0 }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontWeight: 500, fontSize: '0.85rem', fontFamily: 'monospace' }}>{run.run_id.slice(0, 12)}…</div>
                        <div style={{ color: 'var(--quiet)', fontSize: '0.78rem', textTransform: 'capitalize' }}>{run.status}</div>
                      </div>
                      <div style={{ color: 'var(--quiet)', fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: 4 }}>
                        <Clock size={12} />
                        {new Date(run.updated_at).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                      </div>
                      <div style={{ display: 'flex', gap: 4, flexShrink: 0 }}>
                        {(run.status === 'running' || run.status === 'suspended') && (
                          <IconButton variant="warning" onClick={(e) => { e.stopPropagation(); handleCancel(run.run_id); }} aria-label="Cancel">
                            <Square size={12} />
                          </IconButton>
                        )}
                        {run.status === 'failed' || run.status === 'cancelled' ? (
                          <IconButton variant="default" onClick={(e) => { e.stopPropagation(); handleRestart(run.run_id); }} aria-label="Restart">
                            <RotateCcw size={12} />
                          </IconButton>
                        ) : null}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div style={{ flex: 1, padding: 16, overflow: 'auto' }}>
              {!selectedRun ? (
                <div style={{ padding: '32px 16px', textAlign: 'center' }}>
                  <p style={{ color: 'var(--quiet)', fontSize: '0.9rem' }}>Select a run to view details</p>
                </div>
              ) : runLoading ? (
                <p style={{ color: 'var(--muted)' }}>Loading run details…</p>
              ) : (
                <div>
                  <h3 style={{ margin: '0 0 4px', fontSize: '0.95rem', fontFamily: 'monospace' }}>{selectedRun.run_id}</h3>
                  <p style={{ margin: '0 0 16px', color: 'var(--muted)', fontSize: '0.82rem', textTransform: 'capitalize' }}>
                    Status: <span style={{ color: STATUS_COLORS[selectedRun.status] || 'var(--text)', fontWeight: 500 }}>{selectedRun.status}</span>
                    &middot; Updated: {new Date(selectedRun.updated_at).toLocaleString()}
                  </p>

                  {selectedRun.error && (
                    <div style={{ padding: '10px 14px', background: 'color-mix(in srgb, var(--danger) 10%, transparent)', borderRadius: 8, marginBottom: 16, border: '1px solid color-mix(in srgb, var(--danger) 30%, transparent)' }}>
                      <p style={{ margin: '0 0 4px', fontWeight: 600, fontSize: '0.82rem', color: 'var(--danger)' }}>{selectedRun.error.type}</p>
                      <p style={{ margin: 0, fontSize: '0.82rem', color: 'var(--muted)' }}>{selectedRun.error.message}</p>
                    </div>
                  )}

                  {selectedRun.output && (
                    <div style={{ marginBottom: 16 }}>
                      <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)' }}>Output</h4>
                      <pre style={{ margin: 0, padding: 12, background: 'var(--bg)', borderRadius: 8, fontSize: '0.78rem', overflow: 'auto', maxHeight: 200, border: '1px solid var(--line)' }}>
                        {JSON.stringify(selectedRun.output, null, 2)}
                      </pre>
                    </div>
                  )}

                  {selectedRun.state && (
                    <div style={{ marginBottom: 16 }}>
                      <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)' }}>State</h4>
                      <pre style={{ margin: 0, padding: 12, background: 'var(--bg)', borderRadius: 8, fontSize: '0.78rem', overflow: 'auto', maxHeight: 200, border: '1px solid var(--line)' }}>
                        {JSON.stringify(selectedRun.state, null, 2)}
                      </pre>
                    </div>
                  )}

                  {selectedRun.resume_labels && selectedRun.resume_labels.length > 0 && (
                    <div>
                      <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)' }}>Resume Labels</h4>
                      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                        {selectedRun.resume_labels.map((label) => (
                          <span key={label} style={{ padding: '2px 8px', background: 'var(--surface)', borderRadius: 4, fontSize: '0.78rem', border: '1px solid var(--line)' }}>{label}</span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </article>
      </section>
    </DashboardShell>
  );
}
