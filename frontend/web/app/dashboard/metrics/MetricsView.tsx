'use client';

import { useEffect, useMemo, useState } from 'react';
import { Activity, BarChart3, LineChart } from 'lucide-react';
import { EmptyDataPanel } from '../../components/EmptyDataPanel';
import { RequestAreaChart, RequestBarChart } from '../../components/RequestMetricsCharts';
import { fetchRequestMetrics, type MetricsRange, type RequestMetricsResult } from '../../server/metrics';

const RANGES: { value: MetricsRange; label: string }[] = [
  { value: '24h', label: '24 hours' },
  { value: '7d', label: '7 days' },
  { value: '30d', label: '30 days' },
];

/**
 * Detailed request-metrics screen: range + environment filters, summary stats,
 * and both a vertical bar chart and a classic line/area chart, backed by real
 * backend data. Supports a general (all-environments) view and per-environment
 * scoping.
 */
export function MetricsView() {
  const [range, setRange] = useState<MetricsRange>('7d');
  const [env, setEnv] = useState<string>('');
  const [data, setData] = useState<RequestMetricsResult | null>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading');

  useEffect(() => {
    let cancelled = false;
    setStatus('loading');
    fetchRequestMetrics(range, env || undefined)
      .then((res) => {
        if (cancelled) return;
        setData(res);
        setStatus('ready');
      })
      .catch(() => {
        if (cancelled) return;
        setStatus('error');
      });
    return () => {
      cancelled = true;
    };
  }, [range, env]);

  // The environment picker is populated from the global view so it stays stable
  // regardless of the active scope.
  const [envOptions, setEnvOptions] = useState<{ id: string; name: string }[]>([]);
  useEffect(() => {
    let cancelled = false;
    fetchRequestMetrics(range)
      .then((res) => {
        if (cancelled) return;
        setEnvOptions(res.environments.map((e) => ({ id: e.id, name: e.name })));
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, [range]);

  const successRate = useMemo(() => {
    if (!data || data.total === 0) return 0;
    return Math.round((data.success / data.total) * 100);
  }, [data]);

  const selectedEnvName = env ? envOptions.find((e) => e.id === env)?.name ?? 'Environment' : 'All environments';

  return (
    <div className="metrics-view">
      <div className="metrics-toolbar">
        <div className="metrics-filter-group" role="group" aria-label="Time range">
          {RANGES.map((r) => (
            <button
              key={r.value}
              type="button"
              className={r.value === range ? 'metrics-filter active' : 'metrics-filter'}
              onClick={() => setRange(r.value)}
            >
              {r.label}
            </button>
          ))}
        </div>
        <label className="metrics-env-select">
          <span>Environment</span>
          <select value={env} onChange={(e) => setEnv(e.target.value)}>
            <option value="">All environments</option>
            {envOptions.map((e) => (
              <option key={e.id} value={e.id}>
                {e.name}
              </option>
            ))}
          </select>
        </label>
      </div>

      {status === 'loading' && (
        <EmptyDataPanel icon={Activity} title="Loading metrics…" detail="Fetching request telemetry from the backend." />
      )}

      {status === 'error' && (
        <EmptyDataPanel
          icon={Activity}
          title="Metrics unavailable"
          detail="Sign in and make sure the backend is reachable to see request telemetry."
        />
      )}

      {status === 'ready' && data && (
        <>
          <div className="metrics-summary">
            <div className="metrics-summary-card">
              <span className="metrics-summary-value">{data.total}</span>
              <span className="metrics-summary-label">Total requests · {selectedEnvName}</span>
            </div>
            <div className="metrics-summary-card">
              <span className="metrics-summary-value metric-ok">{data.success}</span>
              <span className="metrics-summary-label">Successful ({successRate}%)</span>
            </div>
            <div className="metrics-summary-card">
              <span className="metrics-summary-value metric-bad">{data.failed}</span>
              <span className="metrics-summary-label">Failed</span>
            </div>
            <div className="metrics-summary-card">
              <span className="metrics-summary-value">{data.avg_latency_ms}ms</span>
              <span className="metrics-summary-label">Average latency</span>
            </div>
          </div>

          <div className="metrics-charts-grid">
            <article className="metrics-chart-card">
              <div className="metrics-chart-heading">
                <BarChart3 size={15} />
                <span>Requests per {data.bucket} (success / failed)</span>
              </div>
              <RequestBarChart series={data.series} title="Requests per bucket" />
            </article>
            <article className="metrics-chart-card">
              <div className="metrics-chart-heading">
                <LineChart size={15} />
                <span>Request volume trend</span>
              </div>
              <RequestAreaChart series={data.series} title="Request volume trend" />
            </article>
          </div>

          <article className="metrics-env-breakdown">
            <div className="metrics-chart-heading">
              <Activity size={15} />
              <span>Requests by environment</span>
            </div>
            {data.environments.length === 0 ? (
              <p className="empty-flows-message">No environments have handled traffic yet.</p>
            ) : (
              <ul className="metrics-env-list">
                {data.environments.map((e) => {
                  const share = data.total > 0 ? Math.round((e.count / data.total) * 100) : 0;
                  return (
                    <li key={e.id} className="metrics-env-row">
                      <button
                        type="button"
                        className={env === e.id ? 'metrics-env-name active' : 'metrics-env-name'}
                        onClick={() => setEnv(env === e.id ? '' : e.id)}
                      >
                        <span className={e.active ? 'metrics-env-dot active' : 'metrics-env-dot'} aria-hidden="true" />
                        {e.name}
                      </button>
                      <div className="metrics-env-bar-track">
                        <div className="metrics-env-bar-fill" style={{ width: `${share}%` }} />
                      </div>
                      <span className="metrics-env-count">{e.count}</span>
                    </li>
                  );
                })}
              </ul>
            )}
          </article>
        </>
      )}
    </div>
  );
}
