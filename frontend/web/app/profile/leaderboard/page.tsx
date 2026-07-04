'use client';

import { useEffect, useMemo, useState } from 'react';
import { ArrowRight, Activity, Server, Trophy } from 'lucide-react';
import { ROUTES } from '../../config/routes';
import { fetchProfileLeaderboard, type PublicLeaderboardEntry, type PublicLeaderboardResponse, type PublicWorkloadPoint } from '../../server/user';

type LeaderboardState =
  | { status: 'loading' }
  | { status: 'ready'; users: PublicLeaderboardEntry[] }
  | { status: 'error'; message: string };

export default function ProfileLeaderboardPage() {
  const [state, setState] = useState<LeaderboardState>({ status: 'loading' });

  useEffect(() => {
    let cancelled = false;
    fetchProfileLeaderboard('7d')
      .then(async (response) => {
        if (!response.ok) throw new Error(`Failed to load leaderboard (${response.status})`);
        return response.json() as Promise<PublicLeaderboardResponse>;
      })
      .then((payload) => {
        if (!cancelled) setState({ status: 'ready', users: payload.users });
      })
      .catch((err) => {
        if (!cancelled) setState({ status: 'error', message: err instanceof Error ? err.message : 'Network error loading leaderboard' });
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (state.status === 'loading') {
    return (
      <div className="leaderboard-page">
        <h1 id="profile-title">Leaderboard</h1>
        <div className="profile-empty">Loading leaderboard...</div>
      </div>
    );
  }

  if (state.status === 'error') {
    return (
      <div className="leaderboard-page">
        <h1 id="profile-title">Leaderboard</h1>
        <div className="profile-empty" style={{ color: 'var(--danger)' }}>{state.message}</div>
      </div>
    );
  }

  return (
    <div className="leaderboard-page">
      <header className="leaderboard-header">
        <div>
          <h1 id="profile-title">Leaderboard</h1>
          <p>Users ranked by recent environment workload</p>
        </div>
        <Trophy size={28} />
      </header>

      <section className="leaderboard-list" aria-label="User leaderboard">
        {state.users.length > 0 ? (
          state.users.map((entry) => (
            <article className="leaderboard-row" key={entry.user.id}>
              <div className="leaderboard-rank">{entry.rank}</div>
              <div className="leaderboard-user">
                <div className="leaderboard-avatar">{entry.user.username[0]?.toUpperCase() ?? '?'}</div>
                <div>
                  <strong>{entry.user.username}</strong>
                  <span>{entry.user.id}</span>
                </div>
              </div>
              <div className="leaderboard-chart-wrap">
                <MiniLineChart points={entry.trend} />
              </div>
              <div className="leaderboard-metrics">
                <span><Activity size={14} /> {entry.user.workload_score.toLocaleString()}</span>
                <span><Server size={14} /> {entry.user.public_env_count}</span>
              </div>
              <a className="leaderboard-profile-link" href={ROUTES.PROFILE_BY_ID(entry.user.id)} aria-label={`Open ${entry.user.username} profile`}>
                <ArrowRight size={18} />
              </a>
            </article>
          ))
        ) : (
          <div className="profile-empty">No users yet</div>
        )}
      </section>
    </div>
  );
}

function MiniLineChart({ points }: { points: PublicWorkloadPoint[] }) {
  const pathData = useMemo(() => {
    const width = 220;
    const height = 58;
    const padding = 6;
    const maxValue = Math.max(1, ...points.map((point) => point.value));
    const step = (width - padding * 2) / Math.max(points.length - 1, 1);
    const coords = points.map((point, index) => {
      const x = padding + index * step;
      const y = height - padding - (point.value / maxValue) * (height - padding * 2);
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    });
    return {
      width,
      height,
      d: coords.length ? `M ${coords.join(' L ')}` : '',
    };
  }, [points]);

  return (
    <svg className="leaderboard-line-chart" viewBox={`0 0 ${pathData.width} ${pathData.height}`} role="img" aria-label="Seven day workload trend">
      <path d={pathData.d} />
    </svg>
  );
}
