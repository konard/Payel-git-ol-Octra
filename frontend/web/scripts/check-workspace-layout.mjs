import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Structural guard for the desktop three-pane workspace (issue #40). SSR-rendering
// the workspace would pull in the ReactFlow canvas, so instead we assert the
// wiring invariants against the source: the layout must keep Sessions (a
// collapsible left dock), the centre (Canvas stacked over the Octra Boss chat),
// and the Solution panes as resizable docks and must reuse the existing
// components rather than forking them.

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');

const workspace = read('src/app/components/Workspace.tsx');

// Resizable three-pane layout.
assert.match(workspace, /from 'react-resizable-panels'/, 'workspace must use react-resizable-panels');
assert.match(workspace, /PanelGroup/, 'workspace must render a PanelGroup');
assert.match(workspace, /id="sessions"/, 'workspace must declare a Sessions pane');
assert.match(workspace, /id="center"/, 'workspace must declare a centre pane');
assert.match(workspace, /id="solution"/, 'workspace must declare a Solution pane');

// The left Sessions dock is collapsible from the header so the centre and
// Solution panes can take the full width.
assert.match(workspace, /sessionsOpen &&/, 'Sessions pane must be conditional on sessionsOpen');

// Matching the reference design (issue #40), the centre column stacks the
// Canvas over the Octra Boss chat as two resizable panes, and the Solution
// dock stays permanently docked on the right (no toggle).
assert.match(workspace, /id="center-canvas"/, 'centre must stack a Canvas pane');
assert.match(workspace, /id="center-chat"/, 'centre must stack a chat pane below the Canvas');

// Reuse the existing components instead of duplicating them.
assert.match(workspace, /variant="dock"/, 'workspace must render the Sidebar as a dock');
assert.match(workspace, /<SolutionViewer/, 'workspace must reuse the SolutionViewer');
assert.match(workspace, /<Canvas/, 'workspace must reuse the Canvas');
assert.match(workspace, /<Chat\b/, 'workspace must reuse the Chat view');

// The Canvas and the Octra Boss chat share a SINGLE unified input docked at the
// bottom of the centre column — the owner asked for one field instead of a
// separate input for the canvas and another for the chat (issue #40 feedback).
const bottomInputs = workspace.match(/<BottomInput/g) ?? [];
assert.equal(bottomInputs.length, 1, 'centre must dock exactly one unified input (no per-pane inputs)');
assert.doesNotMatch(workspace, /<ChatInput/, 'the separate chat input must be removed in favour of the unified input');

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

// The header exposes the Sessions dock toggle only on desktop. The Solution
// dock is always visible in the three-pane workspace, so it has no toggle.
assert.match(topbar, /onToggleSessions/, 'TopBar must expose a Sessions toggle');

console.log('check-workspace-layout: all assertions passed');
