import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Structural guard for the Electron main process & preload bridge: a frameless,
// context-isolated window with an explicit IPC surface for window controls and
// project filesystem operations (issue #48 technical requirements).

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');

const main = read('src/main/main.js');
const preload = read('src/main/preload.js');
const pkg = JSON.parse(read('package.json'));

// Frameless window so we can draw the Zed-style title bar ourselves.
assert.match(main, /frame:\s*false/, 'main window is frameless');
assert.match(main, /backgroundColor:\s*'#111111'/, 'window background is the Octra dark base');

// Secure renderer: context isolation on, node integration off, preload bridge.
assert.match(main, /contextIsolation:\s*true/, 'context isolation is enabled');
assert.match(main, /nodeIntegration:\s*false/, 'node integration is disabled');
assert.match(main, /preload:/, 'a preload script is configured');

// Window-control IPC.
for (const channel of ['window:minimize', 'window:toggle-maximize', 'window:close']) {
  assert.ok(main.includes(channel), `main handles ${channel}`);
}

// Project IPC (open / create / recent / forget).
for (const channel of ['projects:list-recent', 'projects:open', 'projects:create', 'projects:open-path']) {
  assert.ok(main.includes(channel), `main handles ${channel}`);
}
assert.match(main, /dialog\.showOpenDialog/, 'open/create use a native folder dialog');

// The preload bridge exposes a minimal, explicit API — never raw Node.
assert.match(preload, /contextBridge\.exposeInMainWorld\('octra'/, 'preload exposes window.octra');
assert.match(preload, /isElectron:\s*true/, 'bridge marks the Electron environment');
for (const api of ['minimize', 'toggleMaximize', 'close', 'listRecent', 'open', 'create']) {
  assert.ok(preload.includes(api), `bridge exposes ${api}`);
}

// Cross-platform packaging metadata (Windows/Linux/macOS per the issue).
assert.equal(pkg.main, 'src/main/main.js', 'package main points at the Electron entry');
assert.ok(pkg.devDependencies && pkg.devDependencies.electron, 'electron is a dev dependency');

console.log('check-electron-main: all assertions passed');
