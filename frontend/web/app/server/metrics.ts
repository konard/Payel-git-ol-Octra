import { API } from '../config/routes';

function authHeaders(): HeadersInit {
  const token = window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
  return { 'octra-api-token': token ?? '' };
}

export type MetricsRange = '24h' | '7d' | '30d';

export type RequestMetricsBucket = {
  start: string;
  label: string;
  count: number;
  success: number;
  failed: number;
};

export type RequestMetricsEnvBreakdown = {
  id: string;
  name: string;
  count: number;
  active: boolean;
};

export type RequestMetricsResult = {
  range: string;
  bucket: string;
  total: number;
  success: number;
  failed: number;
  avg_latency_ms: number;
  series: RequestMetricsBucket[];
  environments: RequestMetricsEnvBreakdown[];
};

/**
 * Fetches aggregated request metrics from the backend. `env` scopes the result
 * to a single dashboard environment; omit it for the global (all-environments)
 * view. Throws on a non-2xx response so callers can surface an error state.
 */
export async function fetchRequestMetrics(range: MetricsRange = '7d', env?: string): Promise<RequestMetricsResult> {
  const params = new URLSearchParams({ range });
  if (env) params.set('env', env);
  const res = await fetch(`${API.METRICS_REQUESTS}?${params.toString()}`, {
    headers: authHeaders(),
    cache: 'no-cache',
  });
  if (!res.ok) {
    throw new Error(`metrics request failed (${res.status})`);
  }
  return res.json();
}
