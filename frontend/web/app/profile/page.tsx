'use client';

import { useEffect, useState } from 'react';
import { Copy, Eye, EyeOff, User, Mail, KeyRound, Calendar } from 'lucide-react';
import { DashboardShell } from '../dashboard/DashboardShell';
import { ROUTES } from '../config/routes';
import { fetchMe } from '../server/user';

type UserData = {
  id: string;
  username: string;
  email: string;
  api_key: string;
  balance: number;
  subscription: string;
  created_at: string;
};

type ProfileState =
  | { status: 'loading' }
  | { status: 'signed-out' }
  | { status: 'ready'; user: UserData }
  | { status: 'error'; message: string };

function readToken(): string | null {
  return window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
}

export default function ProfilePage() {
  const [state, setState] = useState<ProfileState>({ status: 'loading' });
  const [showKey, setShowKey] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const token = readToken();
    if (!token) {
      setState({ status: 'signed-out' });
      return;
    }

    fetchMe(token).then(async (res) => {
      if (!res.ok) {
        setState({ status: 'error', message: `Failed to load profile (${res.status})` });
        return;
      }
      const body = await res.json();
      const user = body?.data ?? body;
      if (user?.username) {
        setState({ status: 'ready', user });
      } else {
        setState({ status: 'error', message: 'Unexpected response format' });
      }
    }).catch(() => {
      setState({ status: 'error', message: 'Network error loading profile' });
    });
  }, []);

  async function copyKey() {
    if (state.status !== 'ready') return;
    try {
      await navigator.clipboard.writeText(state.user.api_key);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {}
  }

  if (state.status === 'loading') {
    return (
      <DashboardShell activeSection="" hideSidebarItems={['models', 'files', 'security', 'overview', 'flows', 'settings']} showNotifications={false} hideNewFlow={true} hideTabs={true}>
        <section className="dashboard-grid dashboard-grid-single">
          <article className="large-panel" aria-label="Profile" style={{ padding: 32, color: 'var(--muted)' }}>
            Loading profile…
          </article>
        </section>
      </DashboardShell>
    );
  }

  if (state.status === 'signed-out') {
    return (
      <DashboardShell activeSection="" hideSidebarItems={['models', 'files', 'security', 'overview', 'flows', 'settings']} showNotifications={false} hideNewFlow={true} hideTabs={true}>
        <section className="dashboard-grid dashboard-grid-single">
          <article className="large-panel" aria-label="Profile" style={{ padding: 32 }}>
            <p style={{ margin: '0 0 12px', color: 'var(--muted)' }}>You are not signed in.</p>
            <a className="primary-button" href={ROUTES.LOGIN}>Sign in</a>
          </article>
        </section>
      </DashboardShell>
    );
  }

  if (state.status === 'error') {
    return (
      <DashboardShell activeSection="" hideSidebarItems={['models', 'files', 'security', 'overview', 'flows', 'settings']} showNotifications={false} hideNewFlow={true} hideTabs={true}>
        <section className="dashboard-grid dashboard-grid-single">
          <article className="large-panel" aria-label="Profile" style={{ padding: 32, color: 'var(--danger)' }}>
            {state.message}
          </article>
        </section>
      </DashboardShell>
    );
  }

  const { user } = state;
  const memberSince = new Date(user.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });

  return (
    <DashboardShell activeSection="" hideSidebarItems={['models', 'files', 'security', 'overview', 'flows', 'settings']} showNotifications={false} hideNewFlow={true} hideTabs={true}>
      <section className="profile-grid">
        <article className="large-panel profile-card" aria-label="Profile">
          <div className="profile-avatar">{user.username[0].toUpperCase()}</div>
          <div className="profile-info">
            <div className="profile-row">
              <User size={16} />
              <span>{user.username}</span>
            </div>
            <div className="profile-row">
              <Mail size={16} />
              <span>{user.email}</span>
            </div>
            <div className="profile-row">
              <Calendar size={16} />
              <span>Member since {memberSince}</span>
            </div>
          </div>
        </article>

        <article className="large-panel" aria-label="API key" style={{ padding: 24 }}>
          <h2 style={{ margin: '0 0 16px', fontSize: '1rem' }}>API key</h2>
          <div className="api-key-shell">
            <code className="api-key-value">
              {showKey ? user.api_key : user.api_key.slice(0, 8) + '••••••••' + user.api_key.slice(-4)}
            </code>
            <div className="api-key-actions">
              <button className="icon-button" onClick={() => setShowKey(!showKey)} aria-label={showKey ? 'Hide' : 'Show'}>
                {showKey ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
              <button className="icon-button" onClick={copyKey} aria-label="Copy">
                <Copy size={16} />
              </button>
            </div>
          </div>
          {copied && <span style={{ fontSize: '0.82rem', color: 'var(--metric-success)' }}>Copied!</span>}
          <p style={{ margin: '16px 0 0', fontSize: '0.82rem', color: 'var(--muted)', lineHeight: 1.55 }}>
            Use this key to authenticate API requests from your CLI or MCP client.
            Treat it like a password — do not share it publicly.
          </p>
        </article>
      </section>
    </DashboardShell>
  );
}
