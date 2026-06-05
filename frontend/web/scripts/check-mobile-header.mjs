import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression guard for issue #46: "Mobile and desktop versions don't correspond
// to each other". On narrow screens the app bar used to be a single wrapping
// flex row that also held the Canvas/Chat/Solution switch, so on phones it
// collapsed into a cramped three-row stack that looked nothing like the clean
// desktop header. The header is now split into a non-wrapping top row plus a
// dedicated full-width mode-tab row shown only on mobile.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const topbar = read('src/app/components/TopBar.tsx');

// The <header> itself must NOT rely on flex-wrap any more — that is exactly what
// let the action group spill onto extra rows on phones.
const headerTag = topbar.match(/<header[^>]*>/);
assert.ok(headerTag, 'TopBar must render a <header>');
assert.doesNotMatch(
  headerTag[0],
  /flex-wrap/,
  'the header must not use flex-wrap (it caused the mobile multi-row stack, issue #46)',
);

// The top row keeps the logo and the action buttons on a single line.
assert.match(
  topbar,
  /Top row: kept on a single line/,
  'TopBar must keep a single-line top row',
);

// The Canvas/Chat/Solution switch now lives in its own row, gated on mobile, and
// laid out as an equal three-column grid so it reads as a clean segmented
// control instead of inline buttons that wrap.
assert.match(
  topbar,
  /\{!isDesktop && \(\s*<div className="mt-3 grid grid-cols-3/,
  'the mode switch must render in a dedicated mobile-only full-width row',
);

// All three modes must still be wired to onModeChange.
for (const tab of ['canvas', 'chat', 'solution']) {
  assert.match(
    topbar,
    new RegExp(`onModeChange\\('${tab}'\\)`),
    `the mode switch must still wire up the ${tab} tab`,
  );
}

// The desktop three-pane workspace shows everything at once, so the tab row must
// stay hidden there (it remains behind the !isDesktop guard above).
assert.doesNotMatch(
  topbar,
  /\{isDesktop && \(\s*<div className="mt-3 grid grid-cols-3/,
  'the mobile mode-tab row must not render on desktop (must stay behind !isDesktop)',
);

console.log('check-mobile-header: all assertions passed');
