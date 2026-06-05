import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Structural guard for the project switcher (reference 4): search box, This
// Window / Recent Projects sections, and the open/new actions, all wired to the
// filesystem bridge.

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');

const switcher = read('src/renderer/js/projectSwitcher.js');
const css = read('src/renderer/styles/chrome.css');

// Reference 4 anatomy.
assert.match(switcher, /placeholder="Search projects/, 'switcher has a "Search projects…" box');
assert.match(switcher, /This Window/, 'switcher shows the This Window section');
assert.match(switcher, /Recent Projects/, 'switcher shows the Recent Projects section');
assert.match(switcher, /Open Local Folder/, 'switcher offers Open Local Folder');
assert.match(switcher, /New Project/, 'switcher offers New Project');

// The active project is marked with a check (reference 4) and uses the active style.
assert.match(switcher, /switcher__item--active/, 'active project gets the active style');
assert.match(switcher, /switcher__check/, 'active project shows a check mark');
assert.match(css, /\.switcher__item--active/, 'active style is defined in CSS');

// Filtering is live.
assert.match(switcher, /state\.filter/, 'switcher filters by the search query');

// Wired to the filesystem bridge (issue #48: open existing / create new projects).
assert.match(switcher, /Api\.projects\.open\(\)/, 'Open Folder calls the bridge');
assert.match(switcher, /Api\.projects\.create\(/, 'New Project calls the bridge');
assert.match(switcher, /Api\.projects\.openPath\(/, 'clicking a recent project re-opens it');
assert.match(switcher, /Api\.projects\.listRecent\(\)/, 'recents come from the bridge');

console.log('check-project-switcher: all assertions passed');
