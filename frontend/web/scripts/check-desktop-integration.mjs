#!/usr/bin/env node
/*
 * Guards the Electron desktop integration that lives inside the web app.
 *
 * The desktop shell (frontend/desktop) runs this web app as its renderer and
 * injects window.octra. Every desktop-only feature MUST be gated so the plain
 * web build is byte-for-byte unaffected. This test asserts those gates exist and
 * that the bridge keys the renderer relies on are present, without needing a
 * browser or Electron.
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dir = path.dirname(fileURLToPath(import.meta.url));
const src = path.join(dir, '..', 'src');
const read = (rel) => fs.readFileSync(path.join(src, rel), 'utf8');

// The bridge gates everything on the injected Electron flag.
const bridge = read('desktop/bridge.ts');
assert.match(bridge, /window\.octra\?\.isElectron/, 'isDesktopApp must check window.octra.isElectron');
assert.match(bridge, /export function isDesktopApp/, 'bridge must export isDesktopApp');
assert.match(bridge, /export function getBridge/, 'bridge must export getBridge');
for (const key of ['readTree', 'readFile', 'minimize', 'toggleMaximize', 'close', 'listRecent']) {
  assert.match(bridge, new RegExp(key), `bridge type must expose ${key}`);
}

// Components render nothing outside Electron.
for (const file of [
  'desktop/DesktopTitleBar.tsx',
  'desktop/DesktopFileExplorer.tsx',
]) {
  const code = read(file);
  assert.match(code, /if \(!isDesktopApp\(\)\) return null/, `${file} must no-op when not in Electron`);
}

// The title bar must be mounted once at the root so it survives App's early
// returns (landing / payment screens).
const main = read('main.tsx');
assert.match(main, /DesktopTitleBar/, 'main.tsx must mount DesktopTitleBar');

// The workspace must only add the Explorer dock inside the desktop app, and must
// NOT mount a separate desktop file viewer window — opened files render in the
// "Solution files" panel instead (issue #50 owner feedback).
const workspace = read('app/components/Workspace.tsx');
assert.match(workspace, /isDesktopApp\(\)/, 'Workspace must gate the explorer on isDesktopApp');
assert.match(workspace, /DesktopFileExplorer/, 'Workspace must render the explorer');
assert.doesNotMatch(workspace, /DesktopFileViewer/, 'Workspace must not render a separate file viewer window');

// The standalone desktop file viewer component must be gone.
assert.ok(
  !fs.existsSync(path.join(src, 'desktop/DesktopFileViewer.tsx')),
  'DesktopFileViewer.tsx should be removed; files open in the Solution files panel',
);

// The store maps file extensions to Monaco languages for syntax highlighting and
// routes opened files into the task store so they appear in "Solution files".
const store = read('desktop/desktopStore.ts');
assert.match(store, /typescript/, 'desktop store must map extensions to languages');
assert.match(store, /refreshRecent/, 'desktop store must expose recent-project loading');
assert.match(store, /useTaskStore/, 'desktop store must push opened files into the task store');
assert.match(store, /upsertCodeFiles/, 'desktop store must upsert opened files as solution files');

// The Solution files panel focuses the file opened from the desktop Explorer.
const solutionViewer = read('app/components/SolutionViewer.tsx');
assert.match(solutionViewer, /useDesktopStore/, 'SolutionViewer must observe the desktop open-file signal');
assert.match(solutionViewer, /openNonce/, 'SolutionViewer must focus the freshly opened desktop file');

console.log('check-desktop-integration: all assertions passed');
