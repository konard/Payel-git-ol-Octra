import assert from 'node:assert/strict';
import { createServer } from 'vite';

// Regression coverage for issue #19 follow-up: when the user builds a Canvas
// workflow and then starts work from Chat, the task payload must include the
// backend's expected WorkflowConfig shape. Sending raw nodes/edges makes the
// gateway decode an empty predefined workflow, so managers/workers disappear.

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
  const { buildWorkflowConfigFromGraph } = await server.ssrLoadModule('/src/app/taskPayload.ts');

  const workflow = buildWorkflowConfigFromGraph(
    [
      { id: 'canvas-boss-search', type: 'boss', role: 'Search Boss', status: 'pending', techStack: ['web search'], position: { x: 0, y: 0 } },
      { id: 'canvas-manager-links', type: 'manager', role: 'Links Search Manager', status: 'pending', position: { x: 0, y: 120 } },
      { id: 'canvas-worker-docs', type: 'worker', role: 'Documentation Search Worker', status: 'pending', position: { x: 0, y: 240 } },
      { id: 'canvas-worker-news', type: 'worker', role: 'News Search Worker', status: 'pending', position: { x: 160, y: 240 } },
      { id: 'manager-0', type: 'manager', role: 'Auto Manager', status: 'done', position: { x: 320, y: 120 } },
    ],
    [
      { from: 'canvas-manager-links', to: 'canvas-worker-docs' },
      { from: 'canvas-manager-links', to: 'canvas-worker-news' },
      { from: 'manager-0', to: 'canvas-worker-news' },
    ],
  );

  assert.ok(workflow, 'custom manager/worker graph should produce a workflow');
  assert.equal(workflow.useAiPlanning, false);
  assert.equal(workflow.architecture, 'Custom workflow with Search Boss and 1 managers');
  assert.deepEqual(workflow.techStack, ['web search']);
  assert.equal(workflow.managers.length, 1, 'auto-generated manager ids must be ignored');
  assert.equal(workflow.managers[0].role, 'Links Search Manager');
  assert.equal(workflow.managers[0].workers.length, 2);
  assert.deepEqual(
    workflow.managers[0].workers.map((worker) => worker.role),
    ['Documentation Search Worker', 'News Search Worker'],
  );

  const noWorkflow = buildWorkflowConfigFromGraph(
    [{ id: 'boss-only', type: 'boss', role: 'Boss', status: 'pending', position: { x: 0, y: 0 } }],
    [],
  );
  assert.equal(noWorkflow, null, 'a graph without custom managers should not force an empty predefined workflow');

  console.log('check-task-payload-workflow: all assertions passed');
} finally {
  await server.close();
}
