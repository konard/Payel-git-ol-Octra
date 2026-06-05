import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Structural guard for the Zed-style top panel (references 2 & 3). Rendering the
// Electron chrome in CI isn't practical, so we assert the wiring invariants on
// the source instead.

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');

const html = read('src/renderer/index.html');
const chromeCss = read('src/renderer/styles/chrome.css');
const tokens = read('src/renderer/styles/tokens.css');
const appJs = read('src/renderer/js/app.js');

// Frameless title bar with the three required regions: app menu, breadcrumb,
// window controls (issue #48 "top panel must contain project name, menu, and
// window control buttons").
assert.match(html, /class="titlebar"/, 'a titlebar element is rendered');
assert.match(html, /id="app-menu-button"[\s\S]*?Open Application Menu/, 'hamburger labelled "Open Application Menu"');
assert.match(html, /id="project-name"/, 'breadcrumb shows the project name');
assert.match(html, /id="menubar"/, 'the application menu bar is present');
assert.match(html, /id="win-min"/, 'window controls include minimize');
assert.match(html, /id="win-max"/, 'window controls include maximize');
assert.match(html, /id="win-close"/, 'window controls include close');

// Zed reference 2 detail: project name followed by branch chips.
assert.match(html, /id="branch-remote"/, 'top bar shows the remote/branch chip');
assert.match(html, /id="branch-local"/, 'top bar shows the local branch chip');

// Frameless dragging: the bar drags the window, interactive items opt out.
assert.match(chromeCss, /-webkit-app-region:\s*drag/, 'titlebar is draggable');
assert.match(chromeCss, /-webkit-app-region:\s*no-drag/, 'interactive items opt out of dragging');

// Octra design system reused (same token names as frontend/web/src/styles/theme.css).
for (const token of ['--background', '--surface', '--accent', '--text', '--text-muted', '--border']) {
  assert.ok(tokens.includes(token), `tokens.css defines ${token}`);
}
assert.match(tokens, /--accent:\s*#f97316/, 'accent matches the Octra orange');
assert.match(chromeCss, /var\(--accent\)/, 'chrome uses the Octra accent token');
assert.match(chromeCss, /var\(--background\)/, 'chrome uses the Octra background token');

// Window controls are wired to the IPC bridge.
assert.match(appJs, /Api\.window\.minimize\(\)/, 'minimize wired to the bridge');
assert.match(appJs, /Api\.window\.toggleMaximize\(\)/, 'maximize wired to the bridge');
assert.match(appJs, /Api\.window\.close\(\)/, 'close wired to the bridge');

console.log('check-topbar: all assertions passed');
