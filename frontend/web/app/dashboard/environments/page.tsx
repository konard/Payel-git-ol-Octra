'use client';

import { useEffect, useState, useCallback } from 'react';
import { Lock, LockOpen, Pause, Play, Trash2, Workflow } from 'lucide-react';
import { DashboardShell } from '../DashboardShell';
import { listDashboardEnvironments, patchDashboardEnvironment, deleteDashboardEnvironment, type DashboardEnvironment } from '../../server/environments';
import { IconButton } from '../../components/IconButton';

const ENV_COOKIE = 'octra_selected_env';

export default function EnvironmentsPage() {
  const [envs, setEnvs] = useState<DashboardEnvironment[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState(() => getCookie(ENV_COOKIE));

  const load = useCallback(() => {
    setLoading(true);
    listDashboardEnvironments().then(async (res) => {
      if (res.ok) {
        const list: DashboardEnvironment[] = await res.json();
        setEnvs(list);
      }
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  useEffect(() => { load(); }, [load]);

  function selectEnv(id: string) {
    setSelected(id);
    document.cookie = `${ENV_COOKIE}=${id}; path=/; max-age=31536000; SameSite=Lax`;
  }

  async function handleResume(id: string) {
    await patchDashboardEnvironment(id, { active: true });
    setEnvs((prev) => prev.map((e) => e.id === id ? { ...e, active: true } : e));
  }

  async function handlePause(id: string) {
    await patchDashboardEnvironment(id, { active: false });
    setEnvs((prev) => prev.map((e) => e.id === id ? { ...e, active: false } : e));
  }

  async function handleToggleVisibility(id: string, current: string) {
    const next = current === 'private' ? 'public' : 'private';
    await patchDashboardEnvironment(id, { visibility: next });
    setEnvs((prev) => prev.map((e) => e.id === id ? { ...e, visibility: next } : e));
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this environment?')) return;
    await deleteDashboardEnvironment(id);
    setEnvs((prev) => prev.filter((e) => e.id !== id));
    if (selected === id) {
      document.cookie = `${ENV_COOKIE}=; path=/; max-age=0`;
      setSelected('');
    }
  }

  return (
    <DashboardShell activeSection="environments" hideTabs showNotifications={false}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Environments" style={{ padding: 24, minHeight: 0 }}>
          {loading ? (
            <p style={{ color: 'var(--muted)' }}>Loading environments…</p>
          ) : envs.length === 0 ? (
            <div style={{ padding: '32px 16px', textAlign: 'center', border: '1px dashed var(--line)', borderRadius: 8 }}>
              <Workflow size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
              <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>No environments yet.</p>
              <p style={{ margin: '4px 0 0', color: 'var(--quiet)', fontSize: '0.82rem' }}>Create one to get started.</p>
            </div>
          ) : (
            <div style={{ display: 'grid', gap: 6 }}>
              {envs.map((env) => (
                <div
                  key={env.id}
                  className={`env-list-item${selected === env.id ? ' active' : ''}${!env.active ? ' paused' : ''}`}
                  onClick={() => selectEnv(env.id)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => e.key === 'Enter' && selectEnv(env.id)}
                >
                  <button className="env-list-icon-btn" onClick={(e) => { e.stopPropagation(); handleToggleVisibility(env.id, env.visibility); }} aria-label="Toggle visibility">
                    {env.visibility === 'private' ? <Lock size={14} /> : <LockOpen size={14} />}
                  </button>
                  <span className="env-list-name">{env.name}</span>
                  <span className="env-list-actions">
                    <IconButton
                      variant={env.active ? 'warning' : 'success'}
                      onClick={(e) => { e.stopPropagation(); env.active ? handlePause(env.id) : handleResume(env.id); }}
                      aria-label={env.active ? 'Pause' : 'Resume'}
                    >
                      {env.active ? <Pause size={13} /> : <Play size={13} />}
                    </IconButton>
                    <IconButton variant="danger" onClick={(e) => { e.stopPropagation(); handleDelete(env.id); }} aria-label="Delete">
                      <Trash2 size={13} />
                    </IconButton>
                  </span>
                  <span className="env-list-date">{new Date(env.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
                </div>
              ))}
            </div>
          )}
        </article>
      </section>
    </DashboardShell>
  );
}

function getCookie(name: string): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : '';
}
