'use client';

import { useEffect, useState } from 'react';
import { User, Mail, Calendar } from 'lucide-react';
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

  if (state.status === 'loading') {
    return (
      <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
        <article className="large-panel profile-card" aria-label="Profile" style={{ color: 'var(--muted)' }}>
          Loading profile…
        </article>
      </div>
    );
  }

  if (state.status === 'signed-out') {
    return (
      <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
        <article className="large-panel" aria-label="Profile" style={{ padding: 32 }}>
          <p style={{ margin: '0 0 12px', color: 'var(--muted)' }}>You are not signed in.</p>
          <a className="primary-button" href={ROUTES.LOGIN}>Sign in</a>
        </article>
      </div>
    );
  }

  if (state.status === 'error') {
    return (
      <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
        <article className="large-panel" aria-label="Profile" style={{ padding: 32, color: 'var(--danger)' }}>
          {state.message}
        </article>
      </div>
    );
  }

  const { user } = state;
  const memberSince = new Date(user.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });

  return (
    <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
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
    </div>
  );
}
