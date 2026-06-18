import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { createServer } from 'vite';

// Regression test for issue #67: the Settings → Statistics tab renders a Chart.js
// chart of tokens spent, grouped by task category and filterable. This checks the
// statisticsStore logic (the data layer behind the chart) plus the wiring that
// records usage on task completion.

// Minimal localStorage stub so zustand's persist middleware has somewhere to write
// when the store module is evaluated under Node's SSR.
const memory = new Map();
globalThis.localStorage = {
  getItem: (key) => (memory.has(key) ? memory.get(key) : null),
  setItem: (key, value) => memory.set(key, String(value)),
  removeItem: (key) => memory.delete(key),
  clear: () => memory.clear(),
};

// The recording is wired in the websocket success handler against the classified
// task type, so the chart can later be filtered by category.
const websocketSource = readFileSync(resolve(process.cwd(), 'src/hooks/useWebSocket.ts'), 'utf8');
assert.match(
  websocketSource,
  /recordTaskUsage\(currentTaskType\.current/,
  'task completion should record token usage by the classified task type',
);

// The Statistics component must be mounted behind its own settings tab.
const topBarSource = readFileSync(resolve(process.cwd(), 'src/app/components/shell/TopBar.tsx'), 'utf8');
assert.match(topBarSource, /id: 'statistics'/, 'a statistics settings tab should exist');
assert.match(topBarSource, /<TokenStatistics \/>/, 'the statistics tab should render TokenStatistics');

// The chart component must use Chart.js (issue requirement).
const chartSource = readFileSync(resolve(process.cwd(), 'src/components/user/TokenStatistics.tsx'), 'utf8');
assert.match(chartSource, /react-chartjs-2/, 'the chart must be built with Chart.js (react-chartjs-2)');

const server = await createServer({
  root: process.cwd(),
  logLevel: 'error',
  mode: 'test',
  envFile: false,
  server: { middlewareMode: true },
  optimizeDeps: { noDiscovery: true, include: [] },
  appType: 'custom',
});

try {
  const { useStatisticsStore, mapTaskTypeToCategory, TASK_CATEGORIES } =
    await server.ssrLoadModule('/src/stores/statisticsStore.ts');

  // --- Category mapping: backend task types -> the four UI categories. ---
  assert.equal(mapTaskTypeToCategory('research'), 'search', 'research maps to search');
  assert.equal(mapTaskTypeToCategory('search'), 'search', 'search maps to search');
  assert.equal(mapTaskTypeToCategory('code'), 'development', 'code maps to development');
  assert.equal(mapTaskTypeToCategory('presentation'), 'presentation', 'presentation maps through');
  assert.equal(mapTaskTypeToCategory('document'), 'document', 'document maps through');
  // Unknown / empty values fall back to development (matches orchestrator default).
  assert.equal(mapTaskTypeToCategory(''), 'development', 'empty falls back to development');
  assert.equal(mapTaskTypeToCategory(undefined), 'development', 'undefined falls back to development');
  assert.equal(mapTaskTypeToCategory('PRESENTATION'), 'presentation', 'mapping is case-insensitive');

  assert.deepEqual(
    TASK_CATEGORIES,
    ['search', 'development', 'presentation', 'document'],
    'the four task categories should be exposed in order',
  );

  const store = useStatisticsStore.getState();
  store.clearStatistics();

  // --- Recording: positive token counts are stored, junk is ignored. ---
  store.recordUsage('search', 100);
  store.recordUsage('search', 50);
  store.recordUsage('development', 200);
  store.recordTaskUsage('presentation', 300);
  store.recordTaskUsage('research', 25); // research -> search

  // These must be ignored so the chart never gets empty/garbage bars.
  store.recordUsage('search', 0);
  store.recordUsage('development', -10);
  store.recordUsage('document', Number.NaN);
  store.recordUsage('not-a-category', 999);

  const totals = useStatisticsStore.getState().getTotalsByCategory();
  assert.equal(totals.search, 175, 'search total = 100 + 50 + 25');
  assert.equal(totals.development, 200, 'development total = 200');
  assert.equal(totals.presentation, 300, 'presentation total = 300');
  assert.equal(totals.document, 0, 'document total stays 0 (NaN ignored)');

  assert.equal(useStatisticsStore.getState().getTotalTokens(), 675, 'grand total = 675');

  // getTotalsByCategory always returns every category, even when empty.
  assert.deepEqual(
    Object.keys(totals).sort(),
    ['development', 'document', 'presentation', 'search'],
    'totals should always include all four categories',
  );

  // --- Clearing wipes the data (the UI "Clear" button). ---
  store.clearStatistics();
  assert.equal(useStatisticsStore.getState().getTotalTokens(), 0, 'clearStatistics resets totals');
  assert.deepEqual(
    useStatisticsStore.getState().getTotalsByCategory(),
    { search: 0, development: 0, presentation: 0, document: 0 },
    'after clear every category is zero',
  );
} finally {
  await server.close();
}

console.log('check-token-statistics: all assertions passed');
