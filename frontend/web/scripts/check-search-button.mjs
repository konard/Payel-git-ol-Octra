import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression test for issue #63: the globe / search button in the task composer
// was a decorative, aria-hidden <span> that did nothing. It must become an
// interactive button that opens Lefine in a new browser tab — the local Vite
// server in development, lefine.pro in production (overridable via VITE_LEFINE_URL).

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const bottomInput = read('src/app/components/chat/BottomInput.tsx');

// The Globe icon must live inside a real <button>, not an aria-hidden span.
const globeIndex = bottomInput.indexOf('<Globe');
assert.ok(globeIndex >= 0, 'the search button must still render the Globe icon');
const buttonOpen = bottomInput.lastIndexOf('<button', globeIndex);
const spanOpen = bottomInput.lastIndexOf('<span', globeIndex);
assert.ok(buttonOpen >= 0 && buttonOpen > spanOpen, 'the Globe icon must be wrapped in a <button>, not a <span>');

assert.match(bottomInput, /onClick=\{openLefineSearch\}/, 'the search button must open Lefine on click');
assert.match(bottomInput, /window\.open\(/, 'opening Lefine must spawn a new browser window/tab');
assert.match(bottomInput, /'_blank'/, 'Lefine must open in a new tab (_blank)');
assert.match(bottomInput, /noopener,noreferrer/, 'the new tab must be opened safely (noopener,noreferrer)');
assert.match(bottomInput, /import\.meta\.env\.DEV \? 'http:\/\/localhost:5173\/' : 'https:\/\/lefine\.pro\/'/, 'the destination must default to localhost in dev and lefine.pro in production');
assert.match(bottomInput, /import\.meta\.env\.VITE_LEFINE_URL/, 'an explicit VITE_LEFINE_URL must override the default Lefine destination');

// The button must no longer be hidden from assistive tech and must carry a label.
assert.ok(!/aria-hidden="true">\s*<Globe/.test(bottomInput), 'the search button must not be aria-hidden anymore');
assert.match(bottomInput, /aria-label=\{t\('bottomInput\.searchOnLefine'\)\}/, 'the search button must expose an accessible label');
assert.match(bottomInput, /title=\{t\('bottomInput\.searchOnLefine'\)\}/, 'the search button must show a tooltip');

// Every shipped translation file must define the new label so the tooltip is
// never the raw i18n key.
for (const dir of ['languages', 'public/languages']) {
  for (const file of readdirSync(resolve(root, dir)).filter((f) => f.endsWith('.json'))) {
    const data = JSON.parse(read(`${dir}/${file}`));
    assert.ok(
      data.bottomInput && typeof data.bottomInput.searchOnLefine === 'string' && data.bottomInput.searchOnLefine.length > 0,
      `${dir}/${file} must define bottomInput.searchOnLefine`,
    );
  }
}

console.log('check-search-button: all assertions passed');
