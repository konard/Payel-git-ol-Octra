import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// The application menu (reference 3) must expose the full Zed menu bar restyled
// for Octra, and the native + in-window menus must come from ONE shared model.

const here = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);
const { MENU_MODEL } = require(resolve(here, '..', 'src', 'renderer', 'js', 'menuModel.js'));
const { toElectronTemplate } = require(resolve(here, '..', 'src', 'main', 'menu.js'));

const labels = MENU_MODEL.map((m) => m.label);
// Reference 3 menu bar, with the leading menu restyled from "Zed" to "Octra".
for (const expected of ['Octra', 'File', 'Edit', 'Selection', 'View', 'Go', 'Run', 'Window', 'Help']) {
  assert.ok(labels.includes(expected), `menu bar must include "${expected}"`);
}

// The leading menu is the Octra application menu (matching reference 3's items).
const appMenu = MENU_MODEL[0];
assert.equal(appMenu.label, 'Octra', 'leading menu is the Octra app menu');
assert.ok(appMenu.app === true, 'leading menu is flagged as the app menu');
const appLabels = appMenu.items.filter((i) => i.label).map((i) => i.label);
for (const expected of [
  'About Octra',
  'Check for Updates',
  'Open Settings',
  'Select Theme…',
  'Extensions',
  'Quit Octra',
]) {
  assert.ok(appLabels.includes(expected), `app menu must include "${expected}"`);
}

// Filesystem entry points the issue requires (open / create projects).
const fileLabels = MENU_MODEL.find((m) => m.label === 'File').items.filter((i) => i.label).map((i) => i.label);
assert.ok(fileLabels.some((l) => /New Project/.test(l)), 'File menu offers New Project');
assert.ok(fileLabels.some((l) => /Open Folder/.test(l)), 'File menu offers Open Folder');

// The native template is derived from the same model (no second source of truth).
const template = toElectronTemplate(() => {});
assert.equal(template.length, MENU_MODEL.length, 'native template mirrors the model 1:1');
assert.equal(template[0].label, 'Octra', 'native template leads with Octra');
const quit = template[0].submenu.find((i) => i.label === 'Quit Octra');
assert.equal(quit.accelerator, 'Ctrl+Q', 'single-chord accelerators map to Electron form');
const theme = template[0].submenu.find((i) => i.label === 'Open Keymap');
assert.equal(theme.accelerator, undefined, 'multi-chord accelerators are left visual-only');

console.log('check-menu-model: all assertions passed');
