#!/usr/bin/env node
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Guard the first-run /app training flow from issue #52. The implementation is
// intentionally source-checked: rendering App in Node would pull in ReactFlow and
// browser storage, while the risk here is integration drift between the tour and
// the real controls it highlights.

const here = dirname(fileURLToPath(import.meta.url));
const read = (rel) => readFileSync(resolve(here, '..', rel), 'utf8');

const app = read('src/app/App.tsx');
const tour = read('src/app/components/OnboardingTour.tsx');
const tourState = read('src/app/components/onboardingTourState.ts');
const workspace = read('src/app/components/Workspace.tsx');
const topbar = read('src/app/components/TopBar.tsx');
const canvas = read('src/app/components/Canvas.tsx');
const bottomInput = read('src/app/components/BottomInput.tsx');
const sidebar = read('src/components/Sidebar.tsx');
const theme = read('src/styles/theme.css');
const en = JSON.parse(read('public/languages/en.json'));
const ru = JSON.parse(read('public/languages/ru.json'));

assert.match(app, /shouldShowOnboardingTour\(\)/, 'App must check first-run tour state');
assert.match(app, /<OnboardingTour\b/, 'App must mount the onboarding tour in the app shell');

assert.match(tourState, /ONBOARDING_TOUR_STORAGE_KEY/, 'tour must expose a stable storage key');
assert.match(tourState, /localStorage\.getItem\(ONBOARDING_TOUR_STORAGE_KEY\)/, 'tour must read persisted completion');
assert.match(tourState, /localStorage\.setItem\(ONBOARDING_TOUR_STORAGE_KEY,\s*'true'\)/, 'tour must persist completion');
assert.match(tour, /querySelectorAll/, 'tour must locate real UI targets by selector');
assert.match(tour, /getBoundingClientRect/, 'tour must position the spotlight from target geometry');
assert.match(tour, /scrollIntoView/, 'tour must bring highlighted controls into view');

const requiredSelectors = [
  'workflow-workspace',
  'add-agent',
  'task-input',
  'settings-button',
  'new-chat',
  'chat-sessions',
  'solution-pane',
  'solution-tab',
];

for (const selector of requiredSelectors) {
  assert.match(tour, new RegExp(`data-tour="${selector}"`), `tour must include the ${selector} step target`);
}

assert.match(workspace, /dataTour="workflow-workspace"/, 'workspace target must be wired');
assert.match(workspace, /dataTour="chat-sessions"/, 'chat sessions target must be wired');
assert.match(workspace, /dataTour="solution-pane"/, 'solution pane target must be wired');
assert.match(canvas, /data-tour="add-agent"/, 'add-agent button target must be wired');
assert.match(bottomInput, /data-tour="task-input"/, 'task input target must be wired');
assert.match(topbar, /data-tour="settings-button"/, 'settings button target must be wired');
assert.match(topbar, /data-tour="solution-tab"/, 'mobile solution tab target must be wired');
assert.match(sidebar, /data-tour="new-chat"/, 'new chat action target must be wired');

assert.match(tour, /onboarding\.ok/, 'tour must expose an OK button label');
assert.match(tour, /onboarding\.skip/, 'tour must expose a skip-training action');
assert.match(theme, /onboarding-tour__spotlight/, 'tour spotlight CSS must exist');
assert.match(theme, /0 0 0 9999px/, 'spotlight must dim the page around the target');

assert.equal(en.onboarding.skip, 'Skip training', 'English skip copy must be present');
assert.equal(ru.onboarding.skip, 'Пропустить обучение', 'Russian skip copy must be present');

console.log('check-onboarding-tour: all assertions passed');
