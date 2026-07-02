'use client';

import { useId } from 'react';
import type { RequestMetricsBucket } from '../server/metrics';

const CHART_WIDTH = 720;
const CHART_HEIGHT = 240;
const PADDING_X = 40;
const PADDING_TOP = 16;
const PADDING_BOTTOM = 32;

function niceMax(value: number): number {
  if (value <= 0) return 4;
  const pow = Math.pow(10, Math.floor(Math.log10(value)));
  const scaled = value / pow;
  let nice = 10;
  if (scaled <= 1) nice = 1;
  else if (scaled <= 2) nice = 2;
  else if (scaled <= 5) nice = 5;
  return nice * pow;
}

// Show at most ~8 x-axis labels so dense ranges (24 hours, 30 days) stay legible.
function labelStride(count: number): number {
  return Math.max(1, Math.ceil(count / 8));
}

type ChartProps = {
  series: RequestMetricsBucket[];
  title: string;
};

/**
 * Vertical bar chart of request counts per bucket, with successes and failures
 * stacked so the split is visible at a glance. Hand-built SVG (no chart lib).
 */
export function RequestBarChart({ series, title }: ChartProps) {
  const max = niceMax(Math.max(1, ...series.map((b) => b.count)));
  const innerW = CHART_WIDTH - PADDING_X * 2;
  const innerH = CHART_HEIGHT - PADDING_TOP - PADDING_BOTTOM;
  const slot = series.length > 0 ? innerW / series.length : innerW;
  const barW = Math.max(2, Math.min(28, slot * 0.6));
  const stride = labelStride(series.length);
  const gridLines = [0, 0.25, 0.5, 0.75, 1];

  return (
    <figure className="metrics-chart" aria-label={title}>
      <svg viewBox={`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`} role="img" preserveAspectRatio="none" className="metrics-chart-svg">
        {gridLines.map((g) => {
          const y = PADDING_TOP + innerH * (1 - g);
          return (
            <g key={g}>
              <line x1={PADDING_X} y1={y} x2={CHART_WIDTH - PADDING_X} y2={y} className="metrics-grid-line" />
              <text x={PADDING_X - 8} y={y + 3} textAnchor="end" className="metrics-axis-label">
                {Math.round(max * g)}
              </text>
            </g>
          );
        })}
        {series.map((bucket, i) => {
          const x = PADDING_X + slot * i + (slot - barW) / 2;
          const successH = (bucket.success / max) * innerH;
          const failedH = (bucket.failed / max) * innerH;
          const failedY = PADDING_TOP + innerH - failedH;
          const successY = failedY - successH;
          return (
            <g key={bucket.start}>
              {bucket.success > 0 && (
                <rect x={x} y={successY} width={barW} height={successH} rx={2} className="metrics-bar-success">
                  <title>{`${bucket.label}: ${bucket.success} ok`}</title>
                </rect>
              )}
              {bucket.failed > 0 && (
                <rect x={x} y={failedY} width={barW} height={failedH} rx={2} className="metrics-bar-failed">
                  <title>{`${bucket.label}: ${bucket.failed} failed`}</title>
                </rect>
              )}
              {i % stride === 0 && (
                <text x={x + barW / 2} y={CHART_HEIGHT - 12} textAnchor="middle" className="metrics-axis-label">
                  {bucket.label}
                </text>
              )}
            </g>
          );
        })}
      </svg>
    </figure>
  );
}

/**
 * Classic line + area chart of total request counts over time. Hand-built SVG.
 */
export function RequestAreaChart({ series, title }: ChartProps) {
  const gradientId = useId();
  const max = niceMax(Math.max(1, ...series.map((b) => b.count)));
  const innerW = CHART_WIDTH - PADDING_X * 2;
  const innerH = CHART_HEIGHT - PADDING_TOP - PADDING_BOTTOM;
  const stride = labelStride(series.length);
  const gridLines = [0, 0.25, 0.5, 0.75, 1];

  const points = series.map((bucket, i) => {
    const x = series.length > 1 ? PADDING_X + (innerW * i) / (series.length - 1) : PADDING_X + innerW / 2;
    const y = PADDING_TOP + innerH * (1 - bucket.count / max);
    return { x, y, bucket };
  });

  const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' ');
  const areaPath =
    points.length > 0
      ? `${linePath} L${points[points.length - 1].x.toFixed(1)},${(PADDING_TOP + innerH).toFixed(1)} L${points[0].x.toFixed(1)},${(PADDING_TOP + innerH).toFixed(1)} Z`
      : '';

  return (
    <figure className="metrics-chart" aria-label={title}>
      <svg viewBox={`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`} role="img" preserveAspectRatio="none" className="metrics-chart-svg">
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="var(--metric-success)" stopOpacity="0.32" />
            <stop offset="100%" stopColor="var(--metric-success)" stopOpacity="0" />
          </linearGradient>
        </defs>
        {gridLines.map((g) => {
          const y = PADDING_TOP + innerH * (1 - g);
          return (
            <g key={g}>
              <line x1={PADDING_X} y1={y} x2={CHART_WIDTH - PADDING_X} y2={y} className="metrics-grid-line" />
              <text x={PADDING_X - 8} y={y + 3} textAnchor="end" className="metrics-axis-label">
                {Math.round(max * g)}
              </text>
            </g>
          );
        })}
        {areaPath && <path d={areaPath} fill={`url(#${gradientId})`} />}
        {linePath && <path d={linePath} className="metrics-line" fill="none" />}
        {points.map((p, i) => (
          <g key={p.bucket.start}>
            <circle cx={p.x} cy={p.y} r={2.5} className="metrics-line-dot">
              <title>{`${p.bucket.label}: ${p.bucket.count} requests`}</title>
            </circle>
            {i % stride === 0 && (
              <text x={p.x} y={CHART_HEIGHT - 12} textAnchor="middle" className="metrics-axis-label">
                {p.bucket.label}
              </text>
            )}
          </g>
        ))}
      </svg>
    </figure>
  );
}
