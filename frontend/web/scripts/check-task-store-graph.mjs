import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { createServer } from 'vite';

// Regression test for issue #13: switching between chats used to leak the
// previous chat's nodes onto the canvas. `resetTask()` preserved "user" nodes,
// so the next chat's saved workflow was appended on top of the old one. The
// fix introduces `setGraph()`, a full graph swap, used when switching chats.
const websocketSource = readFileSync(resolve(process.cwd(), 'src/hooks/useWebSocket.ts'), 'utf8');
assert.match(websocketSource, /addGitHubNode\(\)/, 'code packaging should add the GitHub sink before the final URL arrives');
assert.match(websocketSource, /currentTaskType\.current === 'code'/, 'GitHub sink should be limited to code tasks');

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

  // --- Simulate chat A: a user builds a small workflow on the canvas. ---
  store.setGraph(
    [
      { id: 'user-1', type: 'manager', role: 'Designer', status: 'done', position: { x: 0, y: 0 } },
      { id: 'user-2', type: 'worker', role: 'Coder', status: 'done', position: { x: 10, y: 10 } },
    ],
    [{ from: 'user-1', to: 'user-2' }],
  );

  let state = useTaskStore.getState();
  assert.equal(state.nodes.length, 2, 'chat A should have 2 nodes loaded');
  assert.equal(state.edges.length, 1, 'chat A should have 1 edge loaded');
  // setGraph must normalize progress so nodes render correctly.
  assert.ok(state.nodes.every((n) => n.progress === 0), 'nodes must default progress to 0');

  // --- Switch to chat B: a full graph swap must replace, not merge. ---
  store.setGraph(
    [{ id: 'b-1', type: 'boss', role: 'Boss', status: 'done', position: { x: 0, y: 0 } }],
    [],
  );

  state = useTaskStore.getState();
  assert.equal(state.nodes.length, 1, 'chat B must NOT inherit chat A nodes (no contamination)');
  assert.equal(state.nodes[0].id, 'b-1', 'only chat B node should remain');
  assert.equal(state.edges.length, 0, 'stale chat A edges must be cleared');

  // --- Switching to a chat with no workflow yields a clean canvas. ---
  store.setGraph([], []);
  state = useTaskStore.getState();
  assert.equal(state.nodes.length, 0, 'empty graph must clear the canvas');
  assert.equal(state.edges.length, 0, 'empty graph must clear edges');

  // --- setGraph also resets execution state tied to the old graph. ---
  store.setTaskId('task-123');
  store.setTaskStatus('executing');
  store.setGraph([{ id: 'c-1', type: 'worker', role: 'W', status: 'pending', position: { x: 0, y: 0 } }], []);
  state = useTaskStore.getState();
  assert.equal(state.taskId, null, 'setGraph must clear the previous taskId');
  assert.equal(state.status, 'idle', 'setGraph must reset task status to idle');

  console.log('check-task-store-graph: all assertions passed');
} finally {
  await server.close();
}
