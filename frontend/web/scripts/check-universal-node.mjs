import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { createServer } from 'vite';

const canvasSource = readFileSync(resolve(process.cwd(), 'src/app/components/canvas/Canvas.tsx'), 'utf8');
assert.match(canvasSource, /universal:\s*UniversalNode/, 'Canvas must register the universal React Flow node type');

const sidebarSource = readFileSync(resolve(process.cwd(), 'src/app/components/canvas/NodeSidebar.tsx'), 'utf8');
assert.match(sidebarSource, /type:\s*'universal'/, 'Sidebar must expose the universal node template');
assert.match(sidebarSource, /accent:\s*'#ffffff'/, 'Universal node template must use a white accent');

const websocketSource = readFileSync(resolve(process.cwd(), 'src/hooks/useWebSocket.ts'), 'utf8');
assert.match(websocketSource, /using a single universal node/, 'WebSocket handler must react to the backend fast-path event');
assert.match(websocketSource, /Universal node finished/, 'WebSocket handler must mark the universal node done');
assert.match(websocketSource, /type:\s*'universal'/, 'WebSocket handler must materialize a universal canvas node');

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
  const { useTaskStore, isGeneratedAgentNodeId } = await server.ssrLoadModule('/src/stores/taskStore.ts');
  const { buildWorkflowConfigFromGraph } = await server.ssrLoadModule('/src/app/taskPayload.ts');

  assert.equal(isGeneratedAgentNodeId('universal-1'), true, 'generated universal nodes must be classified as runtime nodes');
  assert.equal(isGeneratedAgentNodeId('canvas-universal'), false, 'user-created universal nodes must remain editable canvas nodes');

  const universalOnly = buildWorkflowConfigFromGraph(
    [{ id: 'canvas-universal', type: 'universal', role: 'Universal', status: 'pending', position: { x: 0, y: 0 } }],
    [],
  );
  assert.ok(universalOnly, 'a manual universal node must produce a predefined workflow with role universal');
  assert.equal(universalOnly.managers.length, 1);
  assert.equal(universalOnly.managers[0].role, 'universal');

  const workflow = buildWorkflowConfigFromGraph(
    [
      { id: 'canvas-universal', type: 'universal', role: 'Universal', status: 'pending', position: { x: 0, y: 0 } },
      { id: 'canvas-manager', type: 'manager', role: 'Coordinator', status: 'pending', position: { x: 0, y: 120 } },
      { id: 'canvas-worker', type: 'worker', role: 'Specialist', status: 'pending', position: { x: 0, y: 240 } },
    ],
    [{ from: 'canvas-manager', to: 'canvas-worker' }],
  );
  assert.ok(workflow, 'manager-worker graphs must still produce a predefined workflow');
  assert.equal(workflow.managers.length, 1, 'universal nodes must not be serialized as manager/worker workflow entries');
  assert.equal(workflow.managers[0].workers.length, 1, 'normal manager-worker edges must still serialize');

  const store = useTaskStore.getState();
  store.setGraph(
    [
      { id: 'universal-1', type: 'universal', role: 'Universal', status: 'done', position: { x: 0, y: 0 } },
      { id: 'canvas-universal', type: 'universal', role: 'Direct AI', status: 'pending', position: { x: 20, y: 20 } },
    ],
    [{ from: 'universal-1', to: 'canvas-universal' }],
  );
  store.resetTask();
  const state = useTaskStore.getState();
  assert.deepEqual(
    state.nodes.map((node) => node.id),
    ['canvas-universal'],
    'resetTask must remove generated universal nodes and keep user-created universal nodes',
  );
  assert.equal(state.edges.length, 0, 'edges touching generated universal nodes must be removed on reset');

  store.setGraph(
    [
      { id: 'boss-1', type: 'boss', role: 'CEO', status: 'done', position: { x: 0, y: 0 } },
      { id: 'universal-1', type: 'universal', role: 'Universal', status: 'done', position: { x: 0, y: 120 } },
      { id: 'canvas-universal', type: 'universal', role: 'Direct AI', status: 'pending', position: { x: 20, y: 20 } },
    ],
    [
      { from: 'boss-1', to: 'universal-1' },
      { from: 'universal-1', to: 'canvas-universal' },
    ],
  );
  store.resetTaskExecution();
  const executionResetState = useTaskStore.getState();
  assert.deepEqual(
    executionResetState.nodes.map((node) => node.id),
    ['canvas-universal'],
    'new task execution must remove generated Boss/Universal runtime nodes',
  );
  assert.equal(executionResetState.edges.length, 0, 'new task execution must remove stale generated-node edges');

  console.log('check-universal-node: all assertions passed');
} finally {
  await server.close();
}
