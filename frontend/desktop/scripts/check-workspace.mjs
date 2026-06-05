import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Structural guard for the PR #49 review fixes:
//  1. the real Octra mascot image is used (NOT an emoji / placeholder);
//  2. no hardcoded recent-project list (LeFine / TradeFast / "octra" path);
//  3. the File/Edit/… menu bar AND the workspace are hidden until a project
//     is open ("если не открыт проект то меню файлов быть не должно");
//  4. the desktop reproduces the live web three-pane UI.

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');
const exists = (rel) => existsSync(resolve(here, '..', rel));

const html = read('src/renderer/index.html');
const appJs = read('src/renderer/js/app.js');
const apiJs = read('src/renderer/js/api.js');
const workspaceCss = read('src/renderer/styles/workspace.css');

// 1. Real mascot artwork (shipped asset, referenced from the renderer; no emoji).
assert.ok(exists('src/renderer/assets/octra-mascot.png'), 'the real Octra mascot PNG is bundled');
assert.match(appJs, /assets\/octra-mascot\.png/, 'the renderer points at the real mascot PNG');
assert.ok(!/🐙/.test(html) && !/🐙/.test(appJs), 'no emoji octopus mascot anywhere');
for (const id of ['appbar-mascot', 'welcome-mascot', 'chat-mascot']) {
  assert.match(html, new RegExp(`id="${id}"`), `mascot image #${id} is present`);
}

// 2. No hardcoded project data (the mock recents start empty).
assert.match(apiJs, /let recent = \[\];/, 'the mock recent list starts empty');
for (const bad of ['LeFine', 'TradeFast', 'Вакансии столяр', '/home/user/dev/octra']) {
  assert.ok(!apiJs.includes(bad), `no hardcoded "${bad}" in the mock`);
  assert.ok(!appJs.includes(bad), `no hardcoded "${bad}" in app.js`);
}

// 3. Menu + workspace gated on an open project.
assert.match(html, /data-has-project="false"/, 'the shell boots with no project open');
assert.match(appJs, /setAttribute\('data-has-project'/, 'app toggles data-has-project');
assert.match(
  workspaceCss,
  /body\[data-has-project='false'\] #menubar[\s\S]*?display:\s*none/,
  'the menu bar is hidden when no project is open',
);
assert.match(
  workspaceCss,
  /body\[data-has-project='false'\] #workspace/,
  'the workspace is hidden when no project is open',
);

// 4. The three-pane web UI is reproduced (Sessions / Canvas+Chat / Solution + Console).
assert.match(html, /class="appbar"/, 'the Octra application toolbar is present');
assert.match(html, /id="workspace"/, 'the three-pane workspace is present');
assert.match(html, /pane--sessions/, 'Sessions pane present');
assert.match(html, /pane--center/, 'center Canvas/Chat pane present');
assert.match(html, /pane--solution/, 'Solution pane present');
assert.match(html, /class="canvas"/, 'canvas is present');
assert.match(html, /class="composer"/, 'shared composer is present');
assert.match(html, /class="console"/, 'console dock is present');
assert.match(html, /Ready\./, 'chat shows the Ready state');
assert.match(html, /Solution files/, 'Solution files header present');

console.log('check-workspace: all assertions passed');
