'use client';

import { useEffect, useMemo, useState } from 'react';
import { Activity, Calendar, Clock3, ExternalLink, LogOut, Mail, Server, Trophy, User } from 'lucide-react';
import { ROUTES } from '../config/routes';
import { fetchMe, fetchPublicProfile, type MeUser, type PublicProfileResponse, type PublicWorkloadCandle } from '../server/user';

type ProfileState =
  | { status: 'loading' }
  | { status: 'signed-out' }
  | { status: 'ready'; profile: PublicProfileResponse; me: MeUser | null }
  | { status: 'error'; message: string };

const reservedProfileSegments = new Set(['api', 'leaderboard', 'settings']);

function readToken(): string | null {
  return window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
}

function readProfileIdFromPath(): string | null {
  const parts = window.location.pathname.split('/').filter(Boolean);
  if (parts[0] !== 'profile' || !parts[1] || reservedProfileSegments.has(parts[1])) return null;
  return decodeURIComponent(parts[1]);
}

function userIdFromMe(user: MeUser | null): string {
  return user?.user_id ?? user?.id ?? '';
}

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
}

function successRate(profile: PublicProfileResponse): string {
  const total = profile.workload.total;
  if (!total) return '0%';
  return `${Math.round((profile.workload.success / total) * 100)}%`;
}

function handleLogout() {
  const keys = [
    'octra_access_token', 'access_token',
    'octra_refresh_token', 'refresh_token',
    'octra_username', 'octra_user_id', 'octra_show_welcome',
  ];
  keys.forEach((k) => window.localStorage.removeItem(k));
  window.location.href = ROUTES.HOME;
}

export default function ProfilePage() {
  const [state, setState] = useState<ProfileState>({ status: 'loading' });

  useEffect(() => {
    let cancelled = false;

    async function loadProfile() {
      const pathProfileId = readProfileIdFromPath();
      const token = readToken();
      let me: MeUser | null = null;

      if (token) {
        try {
          const meResponse = await fetchMe(token);
          if (meResponse.ok) {
            const body = await meResponse.json();
            me = body?.data ?? body;
            if (me?.username) window.localStorage.setItem('octra_username', me.username);
            const currentUserId = userIdFromMe(me);
            if (currentUserId) window.localStorage.setItem('octra_user_id', currentUserId);
          }
        } catch {
          me = null;
        }
      }

      const targetId = pathProfileId ?? userIdFromMe(me);
      if (!targetId) {
        if (!cancelled) setState({ status: 'signed-out' });
        return;
      }

      if (!pathProfileId) {
        window.history.replaceState({}, '', ROUTES.PROFILE_BY_ID(targetId));
      }

      try {
        const profileResponse = await fetchPublicProfile(targetId, '24h');
        if (!profileResponse.ok) {
          throw new Error(`Failed to load profile (${profileResponse.status})`);
        }
        const profile = await profileResponse.json() as PublicProfileResponse;
        if (!cancelled) setState({ status: 'ready', profile, me });
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Network error loading profile';
        if (!cancelled) setState({ status: 'error', message });
      }
    }

    loadProfile();
    return () => {
      cancelled = true;
    };
  }, []);

  if (state.status === 'loading') {
    return (
      <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
        <article className="large-panel profile-loading" aria-label="Profile" style={{ color: 'var(--muted)' }}>
          Loading profile...
        </article>
      </div>
    );
  }

  if (state.status === 'signed-out') {
    return (
      <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
        <article className="large-panel" aria-label="Profile" style={{ padding: 32, minHeight: 0 }}>
          <p style={{ margin: '0 0 12px', color: 'var(--muted)' }}>You are not signed in.</p>
          <a className="primary-button" href={ROUTES.LOGIN}>Sign in</a>
        </article>
      </div>
    );
  }

  if (state.status === 'error') {
    return (
      <div className="dashboard-grid dashboard-grid-single" style={{ padding: 14 }}>
        <article className="large-panel" aria-label="Profile" style={{ padding: 32, minHeight: 0, color: 'var(--danger)' }}>
          {state.message}
        </article>
      </div>
    );
  }

  const { profile, me } = state;
  const ownUserId = userIdFromMe(me);
  const isOwnProfile = ownUserId === profile.user.id;
  const memberSince = formatDate(profile.user.created_at);

  return (
    <div className="profile-page" aria-label="Profile">
      <aside className="profile-identity" aria-label="User summary">
        <div className="profile-large-avatar">{profile.user.username[0]?.toUpperCase() ?? '?'}</div>
        <div className="profile-name-block">
          <h1 id="profile-title">{profile.user.username}</h1>
          <code>{profile.user.id}</code>
        </div>
        <div className="profile-meta">
          <span><User size={15} /> {profile.user.username}</span>
          {isOwnProfile && me?.email && <span><Mail size={15} /> {me.email}</span>}
          <span><Calendar size={15} /> Member since {memberSince}</span>
          <span><Server size={15} /> {profile.user.public_env_count} public envs</span>
        </div>
        <a className="profile-rank-link" href={ROUTES.PROFILE_LEADERBOARD}>
          <Trophy size={16} />
          <span>Open leaderboard</span>
        </a>
        {isOwnProfile && (
          <button className="logout-button profile-logout" onClick={handleLogout} type="button">
            <LogOut size={16} />
            <span>Sign out</span>
          </button>
        )}
      </aside>

      <section className="profile-main" aria-label="Public profile content">
        <section className="profile-section" aria-label="Public environments">
          <div className="profile-section-header">
            <h2>Public environments</h2>
            <span>{profile.user.active_public_env_count} active</span>
          </div>
          <div className="public-env-grid">
            {profile.public_environments.length > 0 ? (
              profile.public_environments.map((env) => (
                <article className="public-env-card" key={env.id}>
                  <div className="public-env-title">
                    <Server size={16} />
                    <strong>{env.name}</strong>
                    <span>Public</span>
                  </div>
                  <p>{env.active ? 'Running environment' : 'Paused environment'}</p>
                  <div className="public-env-foot">
                    <span className={env.active ? 'status-dot active' : 'status-dot'} />
                    <span>{formatDate(env.created_at)}</span>
                  </div>
                </article>
              ))
            ) : (
              <div className="profile-empty">No public environments</div>
            )}
          </div>
        </section>

        <section className="profile-section" aria-label="Workload chart">
          <div className="profile-section-header">
            <h2>Hourly workload</h2>
            <span>{profile.workload.total.toLocaleString()} requests</span>
          </div>
          <div className="workload-panel">
            <CandlestickChart candles={profile.workload.candles} />
          </div>
        </section>

        <section className="profile-stat-grid" aria-label="Workload summary">
          <div>
            <Activity size={16} />
            <span>Total load</span>
            <strong>{profile.workload.total.toLocaleString()}</strong>
          </div>
          <div>
            <Trophy size={16} />
            <span>Success rate</span>
            <strong>{successRate(profile)}</strong>
          </div>
          <div>
            <Clock3 size={16} />
            <span>Average latency</span>
            <strong>{profile.workload.avg_latency_ms ? `${profile.workload.avg_latency_ms} ms` : '0 ms'}</strong>
          </div>
          <a href={ROUTES.PROFILE_LEADERBOARD}>
            <ExternalLink size={16} />
            <span>Leaderboard</span>
            <strong>View</strong>
          </a>
        </section>
      </section>
    </div>
  );
}

function CandlestickChart({ candles }: { candles: PublicWorkloadCandle[] }) {
  const geometry = useMemo(() => {
    const width = 760;
    const height = 280;
    const padding = { top: 18, right: 18, bottom: 28, left: 38 };
    const plotWidth = width - padding.left - padding.right;
    const plotHeight = height - padding.top - padding.bottom;
    const maxValue = Math.max(1, ...candles.flatMap((candle) => [candle.open, candle.high, candle.low, candle.close]));
    const step = plotWidth / Math.max(candles.length, 1);
    const candleWidth = Math.max(5, Math.min(13, step * 0.48));
    const y = (value: number) => padding.top + plotHeight - (value / maxValue) * plotHeight;
    const items = candles.map((candle, index) => {
      const x = padding.left + index * step + step / 2;
      const openY = y(candle.open);
      const closeY = y(candle.close);
      return {
        ...candle,
        x,
        openY,
        closeY,
        highY: y(candle.high),
        lowY: y(candle.low),
        bodyY: Math.min(openY, closeY),
        bodyHeight: Math.max(3, Math.abs(openY - closeY)),
        width: candleWidth,
        rising: candle.close >= candle.open,
      };
    });
    return { width, height, padding, plotWidth, plotHeight, maxValue, items };
  }, [candles]);

  return (
    <svg className="candlestick-chart" viewBox={`0 0 ${geometry.width} ${geometry.height}`} role="img" aria-label="Hourly workload candlestick chart">
      {[0, 1, 2, 3].map((line) => {
        const y = geometry.padding.top + (geometry.plotHeight / 3) * line;
        return <line className="chart-grid-line" x1={geometry.padding.left} x2={geometry.width - geometry.padding.right} y1={y} y2={y} key={line} />;
      })}
      {geometry.items.map((item, index) => (
        <g className={item.rising ? 'candle candle-up' : 'candle candle-down'} key={`${item.start}-${index}`}>
          <line x1={item.x} x2={item.x} y1={item.highY} y2={item.lowY} />
          <rect x={item.x - item.width / 2} y={item.bodyY} width={item.width} height={item.bodyHeight} rx={2} />
        </g>
      ))}
      <text x={geometry.padding.left} y={geometry.height - 8}>{candles[0]?.label ?? ''}</text>
      <text x={geometry.width - geometry.padding.right} y={geometry.height - 8} textAnchor="end">{candles[candles.length - 1]?.label ?? ''}</text>
      <text x={geometry.padding.left - 8} y={geometry.padding.top + 4} textAnchor="end">{geometry.maxValue}</text>
    </svg>
  );
}
