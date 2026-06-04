import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Structural guard for the desktop three-pane workspace (issue #40). SSR-rendering
// the workspace would pull in the ReactFlow canvas, so instead we assert the
// wiring invariants against the source: the layout must keep Sessions, the
// centre, and the Solution panes as independently toggleable resizable docks and
// must reuse the existing components rather than forking them.

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');

const workspace = read('src/app/components/Workspace.tsx');

// Resizable three-pane layout.
assert.match(workspace, /from 'react-resizable-panels'/, 'workspace must use react-resizable-panels');
assert.match(workspace, /PanelGroup/, 'workspace must render a PanelGroup');
assert.match(workspace, /id="sessions"/, 'workspace must declare a Sessions pane');
assert.match(workspace, /id="center"/, 'workspace must declare a centre pane');
assert.match(workspace, /id="solution"/, 'workspace must declare a Solution pane');

// Docks are independently toggleable.
assert.match(workspace, /sessionsOpen &&/, 'Sessions pane must be conditional on sessionsOpen');
assert.match(workspace, /solutionOpen &&/, 'Solution pane must be conditional on solutionOpen');

// Reuse the existing components instead of duplicating them.
assert.match(workspace, /variant="dock"/, 'workspace must render the Sidebar as a dock');
assert.match(workspace, /<SolutionViewer/, 'workspace must reuse the SolutionViewer');
assert.match(workspace, /<Canvas/, 'workspace must reuse the Canvas');
assert.match(workspace, /<Chat\b/, 'workspace must reuse the Chat view');
assert.match(workspace, /<BottomInput/, 'workspace must reuse the task input');
assert.match(workspace, /<ChatInput/, 'workspace must reuse the chat input');

// The colour palette must be preserved (issue #40 explicitly keeps the colours).
assert.match(workspace, /var\(--surface\)/, 'workspace panes must keep the existing surface colour');
assert.match(workspace, /var\(--border\)/, 'workspace panes must keep the existing border colour');

const app = read('src/app/App.tsx');

// The desktop layout is gated behind the media query, with the single-pane
// fallback preserved for narrow screens.
assert.match(app, /useIsDesktop/, 'App must detect desktop width');
assert.match(app, /<Workspace/, 'App must mount the desktop workspace');
assert.match(app, /isDesktop \?/, 'App must branch the layout on isDesktop');

const topbar = read('src/app/components/TopBar.tsx');

// The header exposes the dock toggles only on desktop.
assert.match(topbar, /onToggleSessions/, 'TopBar must expose a Sessions toggle');
assert.match(topbar, /onToggleSolution/, 'TopBar must expose a Solution toggle');

console.log('check-workspace-layout: all assertions passed');
