export const FAKE_METRICS = [
  {
    label: 'Total requests',
    value: '12.4K',
    delta: '+8.2%',
    tone: 'success' as const,
    bars: [32, 48, 40, 62, 78],
  },
  {
    label: 'Active sessions',
    value: '247',
    delta: '+3',
    tone: 'success' as const,
    bars: [20, 35, 28, 42, 55],
  },
  {
    label: 'Avg latency',
    value: '342ms',
    delta: '-12ms',
    tone: 'success' as const,
    bars: [70, 55, 62, 45, 38],
  },
  {
    label: 'Error rate',
    value: '0.8%',
    delta: '-0.3%',
    tone: 'success' as const,
    bars: [12, 18, 10, 8, 6],
  },
  {
    label: 'Environments',
    value: '18',
    delta: '+2',
    tone: 'success' as const,
    bars: [8, 12, 14, 16, 18],
  },
  {
    label: 'Tokens today',
    value: '2.1M',
    delta: '+14%',
    tone: 'warning' as const,
    bars: [40, 55, 70, 65, 85],
  },
];

export const FAKE_FLOW_NODES = [
  { label: 'POST /api/chat', className: 'flow-node gateway' },
  { label: 'Middleware → token check', className: 'flow-node' },
];

export const FAKE_FLOW_SPLITS = [
  { label: 'environment', value: 'Nix profile', className: 'model-node neutral' },
  { label: 'CLI', value: 'claude code', className: 'model-node ok' },
  { label: 'Redis', value: 'cli_state', className: 'model-node caution' },
];
