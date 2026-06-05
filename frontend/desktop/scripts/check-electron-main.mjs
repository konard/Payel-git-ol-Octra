#!/usr/bin/env node
/*
 * Static contract test for the Electron main + preload. Electron itself can't
 * boot headlessly in CI, so instead we assert the compiled main process and the
 * preload agree on every IPC channel: each channel the preload invokes must have
 * a matching ipcMain.handle in main, and the preload must expose window.octra.
 * Also asserts the window is frameless (issue #50.5 — fix the window chrome).
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dir = path.dirname(fileURLToPath(import.meta.url));
const dist = path.join(dir, '..', 'dist-electron');

const main = fs.readFileSync(path.join(dist, 'main.js'), 'utf8');
const preload = fs.readFileSync(path.join(dist, 'preload.js'), 'utf8');

// Preload must expose the bridge.
assert.match(preload, /exposeInMainWorld\(["']octra["']/, 'preload must expose window.octra');
assert.match(preload, /isElectron: true/, 'bridge must flag isElectron');

// Every channel invoked by the preload must be handled by main.
const invoked = [...preload.matchAll(/ipcRenderer\.invoke\(["']([^"']+)["']/g)].map((m) => m[1]);
const handled = new Set(
  [...main.matchAll(/ipcMain\.handle\(["']([^"']+)["']/g)].map((m) => m[1]),
);
assert.ok(invoked.length > 0, 'preload should invoke at least one channel');
for (const channel of invoked) {
  assert.ok(handled.has(channel), `main is missing a handler for "${channel}"`);
}

// The headline filesystem channels must be present (issue #50.4).
for (const channel of ['fs:read-tree', 'fs:read-file', 'projects:open', 'app:get-config']) {
  assert.ok(handled.has(channel), `main must handle ${channel}`);
}

// Frameless window with a custom title bar (issue #50.5).
assert.match(main, /frame:\s*false/, 'window must be frameless');
assert.match(main, /preload\.js/, 'window must load the preload bridge');

console.log('check-electron-main: OK');
