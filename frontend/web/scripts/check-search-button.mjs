import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression coverage for the task composer search button:
// - issue #63 made the globe a real button that can still open Lefine safely;
// - issue #94 changes the click behavior to a provider picker with Apodex as
//   the configurable default search backend.

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

assert.match(bottomInput, /window\.open\(/, 'opening Lefine must spawn a new browser window/tab');
assert.match(bottomInput, /'_blank'/, 'Lefine must open in a new tab (_blank)');
assert.match(bottomInput, /noopener,noreferrer/, 'the new tab must be opened safely (noopener,noreferrer)');
assert.match(bottomInput, /import\.meta\.env\.DEV \? 'http:\/\/localhost:5173\/' : 'https:\/\/lefine\.pro\/'/, 'the destination must default to localhost in dev and lefine.pro in production');
assert.match(bottomInput, /import\.meta\.env\.VITE_LEFINE_URL/, 'an explicit VITE_LEFINE_URL must override the default Lefine destination');

// The button must no longer be hidden from assistive tech and must carry a label.
assert.ok(!/aria-hidden="true">\s*<Globe/.test(bottomInput), 'the search button must not be aria-hidden anymore');
assert.match(bottomInput, /aria-label=\{t\('bottomInput\.searchProviders'\)\}/, 'the search button must expose an accessible provider-picker label');
assert.match(bottomInput, /title=\{t\('bottomInput\.searchProviders'\)\}/, 'the search button must show a provider-picker tooltip');
assert.match(bottomInput, /setShowSearchPicker/, 'clicking the globe must open the provider picker instead of directly opening Lefine');
assert.match(bottomInput, /handleApodexSearchSelect/, 'the picker must include an Apodex branch that opens Search settings');
assert.match(bottomInput, /handleLefineSearchSelect/, 'the picker must keep the Lefine branch');
assert.match(bottomInput, /octra:open-settings/, 'choosing Apodex must request the Settings -> Search tab');
assert.match(bottomInput, /text-orange-500|text-orange-600|bg-orange-500/, 'configured Apodex/custom search must make the globe visibly orange');
assert.match(bottomInput, /data-search-picker-layout="octra-popover"/, 'the provider picker must use the Octra-native popover layout');
assert.match(bottomInput, /data-search-provider-preview="apodex"/, 'the Apodex branch must include a restrained provider preview area');
assert.match(bottomInput, /data-search-provider-preview="lefine"/, 'the Lefine branch must include a restrained provider preview area');
assert.match(bottomInput, /bg-\[var\(--background\)\]/, 'the provider picker must reuse Octra surface tokens');
assert.ok(!/data-search-picker-layout="split-visual"/.test(bottomInput), 'the old oversized split visual layout must not be used');
assert.ok(!/ring-8 ring-black/.test(bottomInput), 'the Lefine image must not use the old heavy black frame');
assert.ok(!/bg-black\/35/.test(bottomInput), 'the provider previews must not use the old heavy black image block');
assert.match(bottomInput, /apodex\.png/, 'the provider picker must reserve the apodex.png image slot');
assert.match(bottomInput, /lefine\.pro\.jpg/, 'the provider picker must use the Lefine image asset');
assert.match(bottomInput, /data-search-provider-visual="apodex"/, 'the picker must expose the Apodex visual panel');
assert.match(bottomInput, /data-search-provider-visual="lefine"/, 'the picker must expose the Lefine visual panel');

// Every shipped translation file must define the new label so the tooltip is
// never the raw i18n key.
for (const dir of ['languages', 'public/languages']) {
  for (const file of readdirSync(resolve(root, dir)).filter((f) => f.endsWith('.json'))) {
    const data = JSON.parse(read(`${dir}/${file}`));
    assert.ok(
      data.bottomInput && typeof data.bottomInput.searchProviders === 'string' && data.bottomInput.searchProviders.length > 0,
      `${dir}/${file} must define bottomInput.searchProviders`,
    );
    assert.ok(
      data.bottomInput && typeof data.bottomInput.searchWithApodex === 'string' && data.bottomInput.searchWithApodex.length > 0,
      `${dir}/${file} must define bottomInput.searchWithApodex`,
    );
    assert.ok(
      data.bottomInput && typeof data.bottomInput.searchWithLefine === 'string' && data.bottomInput.searchWithLefine.length > 0,
      `${dir}/${file} must define bottomInput.searchWithLefine`,
    );
    assert.ok(
      data.settings && typeof data.settings.search === 'string' && data.settings.search.length > 0,
      `${dir}/${file} must define settings.search`,
    );
  }
}

console.log('check-search-button: all assertions passed');
