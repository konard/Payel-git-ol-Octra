'use client';

import { useEffect, useState, useCallback } from 'react';
import { UserCheck, RefreshCw, Play, Square, Clock, MessageSquare, FileJson } from 'lucide-react';
import { DashboardShell } from '../DashboardShell';
import { listHITLRuns, getHITLRun, resumeHITLRun, cancelHITLRun, type HITLRun, type HITLRunDetail } from '../../server/hitl';
import { IconButton } from '../../components/IconButton';

export default function HITLPage() {
  const [runs, setRuns] = useState<HITLRun[]>([]);
  const [selectedRun, setSelectedRun] = useState<HITLRunDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);
  const [resumeInput, setResumeInput] = useState('');

  const loadRuns = useCallback(() => {
    setLoading(true);
    listHITLRuns().then(async (res) => {
      if (res.ok) {
        const data: { runs: HITLRun[] } = await res.json();
        setRuns(data.runs ?? []);
      }
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  useEffect(() => { loadRuns(); }, [loadRuns]);

  async function handleSelectRun(workflowId: string, runId: string) {
    setDetailLoading(true);
    setResumeInput('');
    const res = await getHITLRun(workflowId, runId);
    if (res.ok) {
      setSelectedRun(await res.json());
    }
    setDetailLoading(false);
  }

  async function handleResume(run: HITLRun) {
    const data = resumeInput.trim() ? { input: resumeInput.trim() } : undefined;
    const res = await resumeHITLRun(run.workflow_id, run.run_id, data);
    if (res.ok) {
      setSelectedRun(null);
      loadRuns();
    }
  }

  async function handleCancel(run: HITLRun) {
    if (!confirm('Cancel this suspended run?')) return;
    await cancelHITLRun(run.workflow_id, run.run_id);
    setSelectedRun(null);
    loadRuns();
  }

  return (
    <DashboardShell activeSection="hitl" hideTabs showNotifications={false}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Human-in-the-Loop" style={{ padding: 0, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
          <div className="panel-heading" style={{ padding: '16px 24px', borderBottom: '1px solid var(--line)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <UserCheck size={18} />
              <span>Human-in-the-Loop</span>
            </div>
            <button className="small-command ghost-command" onClick={loadRuns} aria-label="Refresh">
              <RefreshCw size={14} />
            </button>
          </div>

          <div style={{ display: 'flex', flex: 1, minHeight: 0 }}>
            <div style={{ flex: 1, padding: 16, overflow: 'auto', borderRight: '1px solid var(--line)' }}>
              <p style={{ margin: '0 0 12px', color: 'var(--muted)', fontSize: '0.82rem' }}>
                {runs.length} run{runs.length !== 1 ? 's' : ''} waiting for human input
              </p>

              {loading ? (
                <p style={{ color: 'var(--muted)' }}>Loading…</p>
              ) : runs.length === 0 ? (
                <div style={{ padding: '32px 16px', textAlign: 'center' }}>
                  <UserCheck size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
                  <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>No runs waiting for input.</p>
                  <p style={{ margin: '4px 0 0', color: 'var(--quiet)', fontSize: '0.82rem' }}>Suspended workflow runs will appear here.</p>
                </div>
              ) : (
                <div style={{ display: 'grid', gap: 4 }}>
                  {runs.map((run) => (
                    <div
                      key={`${run.workflow_id}/${run.run_id}`}
                      className={`env-list-item${selectedRun?.run_id === run.run_id ? ' active' : ''}`}
                      onClick={() => handleSelectRun(run.workflow_id, run.run_id)}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => e.key === 'Enter' && handleSelectRun(run.workflow_id, run.run_id)}
                      style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 14px' }}
                    >
                      <div style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--warning)', flexShrink: 0 }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontWeight: 500, fontSize: '0.85rem', fontFamily: 'monospace' }}>{run.run_id.slice(0, 12)}…</div>
                        <div style={{ color: 'var(--quiet)', fontSize: '0.78rem' }}>
                          Workflow: <span style={{ fontFamily: 'monospace' }}>{run.workflow_id.slice(0, 12)}…</span>
                        </div>
                      </div>
                      <div style={{ color: 'var(--quiet)', fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: 4 }}>
                        <Clock size={12} />
                        {new Date(run.updated_at).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
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
              ) : detailLoading ? (
                <p style={{ color: 'var(--muted)' }}>Loading run details…</p>
              ) : (
                <div>
                  <h3 style={{ margin: '0 0 4px', fontSize: '0.95rem', fontFamily: 'monospace' }}>{selectedRun.run_id}</h3>
                  <p style={{ margin: '0 0 16px', color: 'var(--muted)', fontSize: '0.82rem' }}>
                    Status: <span style={{ color: 'var(--warning)', fontWeight: 500 }}>suspended</span>
                    &middot; Updated: {new Date(selectedRun.updated_at).toLocaleString()}
                  </p>

                  {selectedRun.resume_labels && selectedRun.resume_labels.length > 0 && (
                    <div style={{ marginBottom: 16 }}>
                      <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)', display: 'flex', alignItems: 'center', gap: 6 }}>
                        <MessageSquare size={14} /> Resume Labels
                      </h4>
                      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                        {selectedRun.resume_labels.map((label) => (
                          <span key={label} style={{ padding: '2px 8px', background: 'var(--surface)', borderRadius: 4, fontSize: '0.78rem', border: '1px solid var(--line)' }}>{label}</span>
                        ))}
                      </div>
                    </div>
                  )}

                  {selectedRun.suspend_payload && (
                    <div style={{ marginBottom: 16 }}>
                      <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)', display: 'flex', alignItems: 'center', gap: 6 }}>
                        <FileJson size={14} /> Suspend Payload
                      </h4>
                      <pre style={{ margin: 0, padding: 12, background: 'var(--bg)', borderRadius: 8, fontSize: '0.78rem', overflow: 'auto', maxHeight: 200, border: '1px solid var(--line)' }}>
                        {JSON.stringify(selectedRun.suspend_payload, null, 2)}
                      </pre>
                    </div>
                  )}

                  {selectedRun.state && (
                    <div style={{ marginBottom: 16 }}>
                      <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)' }}>State</h4>
                      <pre style={{ margin: 0, padding: 12, background: 'var(--bg)', borderRadius: 8, fontSize: '0.78rem', overflow: 'auto', maxHeight: 150, border: '1px solid var(--line)' }}>
                        {JSON.stringify(selectedRun.state, null, 2)}
                      </pre>
                    </div>
                  )}

                  <div style={{ marginTop: 24, paddingTop: 16, borderTop: '1px solid var(--line)' }}>
                    <h4 style={{ margin: '0 0 8px', fontSize: '0.82rem', color: 'var(--muted)' }}>Respond</h4>
                    <textarea
                      value={resumeInput}
                      onChange={(e) => setResumeInput(e.target.value)}
                      placeholder="Optional input to resume with…"
                      rows={3}
                      style={{ width: '100%', padding: '8px 10px', borderRadius: 6, border: '1px solid var(--line)', background: 'var(--bg)', color: 'var(--text)', fontSize: '0.85rem', boxSizing: 'border-box', resize: 'vertical', fontFamily: 'inherit' }}
                    />
                    <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 8 }}>
                      <button className="small-command ghost-command" onClick={() => handleCancel(selectedRun)} style={{ color: 'var(--danger)' }}>
                        <Square size={14} /> Cancel Run
                      </button>
                      <button className="small-command accent-command" onClick={() => handleResume(selectedRun)}>
                        <Play size={14} /> Resume
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </article>
      </section>
    </DashboardShell>
  );
}
