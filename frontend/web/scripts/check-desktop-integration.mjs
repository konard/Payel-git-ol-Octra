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
  'desktop/DesktopFileViewer.tsx',
]) {
  const code = read(file);
  assert.match(code, /if \(!isDesktopApp\(\)\) return null/, `${file} must no-op when not in Electron`);
}

// The title bar must be mounted once at the root so it survives App's early
// returns (landing / payment screens).
const main = read('main.tsx');
assert.match(main, /DesktopTitleBar/, 'main.tsx must mount DesktopTitleBar');

// The workspace must only add the Explorer dock inside the desktop app.
const workspace = read('app/components/Workspace.tsx');
assert.match(workspace, /isDesktopApp\(\)/, 'Workspace must gate the explorer on isDesktopApp');
assert.match(workspace, /DesktopFileExplorer/, 'Workspace must render the explorer');
assert.match(workspace, /DesktopFileViewer/, 'Workspace must render the file viewer');

// The store maps file extensions to Monaco languages for syntax highlighting.
const store = read('desktop/desktopStore.ts');
assert.match(store, /typescript/, 'desktop store must map extensions to languages');
assert.match(store, /refreshRecent/, 'desktop store must expose recent-project loading');

console.log('check-desktop-integration: all assertions passed');
