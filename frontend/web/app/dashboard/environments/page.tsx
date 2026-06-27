'use client';

import { useEffect, useState, useCallback } from 'react';
import { Lock, LockOpen, Workflow } from 'lucide-react';
import { DashboardShell } from '../DashboardShell';
import { listDashboardEnvironments, type DashboardEnvironment } from '../../server/environments';

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
                <button
                  key={env.id}
                  className={`env-list-item${selected === env.id ? ' active' : ''}`}
                  onClick={() => selectEnv(env.id)}
                >
                  <span className="env-list-icon">
                    {env.visibility === 'private' ? <Lock size={14} /> : <LockOpen size={14} />}
                  </span>
                  <span className="env-list-name">{env.name}</span>
                  <span className="env-list-date">{new Date(env.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
                </button>
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
