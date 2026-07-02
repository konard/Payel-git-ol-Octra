'use client';

import { useEffect, useState } from 'react';
import { Activity } from 'lucide-react';
import { EmptyDataPanel } from './EmptyDataPanel';
import { RequestBarChart } from './RequestMetricsCharts';
import { fetchRequestMetrics, type MetricsRange, type RequestMetricsResult } from '../server/metrics';
import { ROUTES } from '../config/routes';

type RequestMetricsOverviewProps = {
  range?: MetricsRange;
  env?: string;
  compact?: boolean;
};

/**
 * Compact request-metrics widget shown in the workspace side panel and on the
 * dashboard overview. It pulls real data from the backend and links through to
 * the detailed metrics screen. Falls back to the shared empty/loading panels so
 * it blends into octra's design system.
 */
export function RequestMetricsOverview({ range = '7d', env, compact = false }: RequestMetricsOverviewProps) {
  const [data, setData] = useState<RequestMetricsResult | null>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading');

  useEffect(() => {
    let cancelled = false;
    setStatus('loading');
    fetchRequestMetrics(range, env)
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

  if (status === 'loading') {
    return (
      <EmptyDataPanel
        compact={compact}
        icon={Activity}
        title="Loading metrics…"
        detail="Fetching request telemetry from the backend."
      />
    );
  }

  if (status === 'error' || !data) {
    return (
      <EmptyDataPanel
        compact={compact}
        icon={Activity}
        title="Metrics unavailable"
        detail="Sign in and make sure the backend is reachable to see request telemetry."
        actionHref={ROUTES.DASHBOARD_METRICS}
        actionLabel="Open metrics"
      />
    );
  }

  if (data.total === 0) {
    return (
      <EmptyDataPanel
        compact={compact}
        icon={Activity}
        title="No requests yet"
        detail="Request counts will appear here once your environments start handling traffic."
        actionHref={ROUTES.DASHBOARD_METRICS}
        actionLabel="Open metrics"
      />
    );
  }

  const successRate = data.total > 0 ? Math.round((data.success / data.total) * 100) : 0;

  return (
    <div className="metrics-overview">
      <div className="metrics-overview-stats">
        <div className="metrics-stat">
          <span className="metrics-stat-value">{data.total}</span>
          <span className="metrics-stat-label">Requests</span>
        </div>
        <div className="metrics-stat">
          <span className="metrics-stat-value metric-ok">{successRate}%</span>
          <span className="metrics-stat-label">Success</span>
        </div>
        <div className="metrics-stat">
          <span className="metrics-stat-value">{data.avg_latency_ms}ms</span>
          <span className="metrics-stat-label">Avg latency</span>
        </div>
      </div>
      <RequestBarChart series={data.series} title="Requests over time" />
      <a className="metrics-overview-link" href={ROUTES.DASHBOARD_METRICS}>
        Open detailed metrics
      </a>
    </div>
  );
}
