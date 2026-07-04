import { API } from '../config/routes'

export type MeUser = {
  id?: string;
  user_id?: string;
  username: string;
  email: string;
  api_key?: string;
  balance: number;
  subscription?: string;
  created_at: string;
};

export type MeResponse = {
  data?: MeUser & {
    balance_credits?: number;
  };
};

export function fetchMe(token: string): Promise<Response> {
  return fetch(API.ME, {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export type PublicUserSummary = {
  id: string;
  username: string;
  created_at: string;
  public_env_count: number;
  active_public_env_count: number;
  workload_score: number;
};

export type PublicEnvironmentSummary = {
  id: string;
  name: string;
  active: boolean;
  created_at: string;
};

export type PublicWorkloadPoint = {
  start: string;
  label: string;
  value: number;
};

export type PublicWorkloadCandle = {
  start: string;
  label: string;
  open: number;
  high: number;
  low: number;
  close: number;
};

export type PublicWorkloadSummary = {
  range: string;
  bucket: string;
  total: number;
  success: number;
  failed: number;
  avg_latency_ms: number;
  line: PublicWorkloadPoint[];
  candles: PublicWorkloadCandle[];
};

export type PublicProfileResponse = {
  user: PublicUserSummary;
  public_environments: PublicEnvironmentSummary[];
  workload: PublicWorkloadSummary;
};

export type PublicLeaderboardEntry = {
  rank: number;
  user: PublicUserSummary;
  trend: PublicWorkloadPoint[];
};

export type PublicLeaderboardResponse = {
  range: string;
  users: PublicLeaderboardEntry[];
};

export function fetchPublicProfile(id: string, range = '24h'): Promise<Response> {
  return fetch(`${API.USER_PROFILE(id)}?range=${encodeURIComponent(range)}`, { cache: 'no-cache' });
}

export function fetchProfileLeaderboard(range = '7d'): Promise<Response> {
  return fetch(`${API.USER_LEADERBOARD}?range=${encodeURIComponent(range)}`, { cache: 'no-cache' });
}
