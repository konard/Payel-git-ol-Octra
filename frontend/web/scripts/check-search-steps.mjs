import assert from 'node:assert/strict';
import { createServer } from 'vite';

// Regression test for issue #19: research workers stream their web-search
// progress to the chat as a collapsible "Searching the web" panel. The store
// accumulates de-duplicated steps via recordSearchStep(), tracks the search
// phase ('searching' → 'done'), and exposes the final step count for the
// "Completed N steps" label. clearCodeFiles()/resetTaskExecution() must wipe it
// so a new task starts with an empty panel.

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
  const { useTaskStore } = await server.ssrLoadModule('/src/stores/taskStore.ts');
  const store = useTaskStore.getState();

  // --- A worker emits two distinct search steps while searching. ---
  store.recordSearchStep('Searching the web for «httpx install python»', 'searching', 0);
  store.recordSearchStep('Searching the web for «httpx documentation»', 'searching', 0);

  let state = useTaskStore.getState();
  assert.equal(state.searchSteps.length, 2, 'two distinct steps must be recorded');
  assert.equal(state.searchPhase, 'searching', 'phase must be "searching" while in progress');

  // --- Duplicate step text must be ignored (no double rendering). ---
  store.recordSearchStep('Searching the web for «httpx install python»', 'searching', 0);
  state = useTaskStore.getState();
  assert.equal(state.searchSteps.length, 2, 'duplicate step text must be de-duplicated');

  // --- The "done" event flips the phase and records the final count. ---
  store.recordSearchStep('', 'done', 2);
  state = useTaskStore.getState();
  assert.equal(state.searchPhase, 'done', 'phase must become "done" on completion');
  assert.equal(state.searchStepsCount, 2, 'final step count must be stored for the label');

  // --- A new searching step (next worker) reactivates the panel. ---
  store.recordSearchStep('Searching the web for «bm25 ranking»', 'searching', 0);
  state = useTaskStore.getState();
  assert.equal(state.searchPhase, 'searching', 'a new step must reopen the active state');
  assert.equal(state.searchSteps.length, 3, 'new worker step must append');
  assert.equal(state.searchStepsCount, 2, 'count must not shrink below a previous done count');

  // --- clearSearchSteps() resets everything. ---
  store.clearSearchSteps();
  state = useTaskStore.getState();
  assert.equal(state.searchSteps.length, 0, 'clearSearchSteps must empty the steps');
  assert.equal(state.searchPhase, 'idle', 'clearSearchSteps must reset the phase');
  assert.equal(state.searchStepsCount, 0, 'clearSearchSteps must reset the count');

  // --- clearCodeFiles() (called when a new task starts) also wipes search. ---
  store.recordSearchStep('Searching the web for «something»', 'searching', 0);
  store.clearCodeFiles();
  state = useTaskStore.getState();
  assert.equal(state.searchSteps.length, 0, 'clearCodeFiles must clear search steps for a new task');
  assert.equal(state.searchPhase, 'idle', 'clearCodeFiles must reset the search phase');

  console.log('check-search-steps: all assertions passed');
} finally {
  await server.close();
}
