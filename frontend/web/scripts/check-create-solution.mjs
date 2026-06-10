import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Regression test for issue #68 ("FIX CREATE SOLUTION").
//
// Symptoms reported:
//   1. The canvas only ever showed Boss + Manager — it never reached the
//      Workers. Cause: workers now run through a "Group Chat" and the backend
//      stopped emitting the per-worker "Starting worker N/M" lines the frontend
//      listened for, so worker nodes were never materialised.
//   2. The canvas auto-zoomed/recentred onto every new node as it appeared.
//      The canvas should instead start zoomed out and stay put.
//   3. Raw Nix build details leaked into the user-facing progress log.

const repoRoot = resolve(process.cwd(), '../..');

// --- 1. Frontend materialises workers from the Group Chat messages. ---
const websocketSource = readFileSync(resolve(process.cwd(), 'src/hooks/useWebSocket.ts'), 'utf8');
assert.match(
  websocketSource,
  /Starting Group Chat/,
  'useWebSocket must react to the "Starting Group Chat" message to create worker nodes',
);
assert.match(
  websocketSource,
  /worker_roles/,
  'useWebSocket must read the worker_roles supplied by the backend',
);
assert.match(
  websocketSource,
  /Group Chat completed:\\s\*\(\\d\+\)\\s\*agents/,
  'useWebSocket must mark workers done when the Group Chat completes',
);

// --- 2. Canvas starts zoomed out and never recentres on new nodes. ---
const canvasSource = readFileSync(resolve(process.cwd(), 'src/app/components/Canvas.tsx'), 'utf8');
assert.doesNotMatch(
  canvasSource,
  /^\s*fitView\s*$/m,
  'Canvas must NOT use the bare fitView prop — it recentres on every new node',
);
assert.match(
  canvasSource,
  /defaultViewport=\{\{[^}]*zoom:/,
  'Canvas must set an explicit zoomed-out defaultViewport',
);

// --- 3. Backend no longer leaks Nix details into the user-facing log. ---
const nixBuildSource = readFileSync(
  resolve(repoRoot, 'orchestrator/internal/service/rules/boss/nix_build.go'),
  'utf8',
);
// Only the user-facing emit() messages must be Nix-free. Internal log.Printf
// lines (server logs) may still mention nix.
const emitLines = nixBuildSource
  .split('\n')
  .filter((line) => line.includes('emit(progress'));
assert.ok(emitLines.length > 0, 'nix_build.go should still emit build progress');
for (const line of emitLines) {
  assert.doesNotMatch(
    line,
    /Nix/,
    `user-facing progress message must not mention Nix: ${line.trim()}`,
  );
}
assert.doesNotMatch(
  nixBuildSource,
  /"nix_output":/,
  'raw nix_output must not be forwarded to the frontend',
);

// --- 4. Backend hands the worker roles to the frontend. ---
const assignSource = readFileSync(
  resolve(repoRoot, 'orchestrator/internal/service/rules/worker/assign.go'),
  'utf8',
);
assert.match(
  assignSource,
  /"worker_roles":/,
  'AssignWorkersAndWait must include worker_roles in the Group Chat start event',
);

console.log('check-create-solution: all assertions passed');
