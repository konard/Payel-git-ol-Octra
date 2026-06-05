#!/usr/bin/env node
/*
 * Guards against a Rules-of-Hooks regression in App.tsx.
 *
 * App renders the landing page (unauthenticated) or the payment-success page via
 * early `return` statements. React requires every hook to run in the same order
 * on every render, so those early returns MUST come AFTER the last hook call.
 *
 * Previously they sat in the middle of the component, before `useWebSocket`, the
 * `useTaskStore` selector and several `useEffect`s. The moment `isAuthenticated`
 * flipped (e.g. on login) the rendered hook count changed and React threw
 * "Rendered fewer/more hooks than expected", white-screening the whole app — and
 * the Electron desktop shell that loads it. This test keeps the early returns
 * below every hook so that can't happen again.
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dir = path.dirname(fileURLToPath(import.meta.url));
const appPath = path.join(dir, '..', 'src', 'app', 'App.tsx');
const src = fs.readFileSync(appPath, 'utf8');

// The hooks that used to sit AFTER the early returns. Each must appear before
// the landing-page guard for the render to be hook-safe.
const hookMarkers = [
  'useWebSocket(',
  'useTaskStore((state)',
];

const landingGuard = src.indexOf('if (isLandingPage)');
const paymentGuard = src.indexOf('if (isPaymentSuccess)');

assert.ok(landingGuard !== -1, 'App.tsx must contain the isLandingPage guard');
assert.ok(paymentGuard !== -1, 'App.tsx must contain the isPaymentSuccess guard');

for (const marker of hookMarkers) {
  const at = src.indexOf(marker);
  assert.ok(at !== -1, `App.tsx must still call ${marker}`);
  assert.ok(
    at < landingGuard,
    `Hook "${marker}" must run BEFORE the isLandingPage early return ` +
      '(early returns belong after every hook — see Rules of Hooks).',
  );
  assert.ok(
    at < paymentGuard,
    `Hook "${marker}" must run BEFORE the isPaymentSuccess early return.`,
  );
}

// No hook may appear after the early returns. Catch any new hook accidentally
// added below the guards.
const afterGuards = src.slice(Math.min(landingGuard, paymentGuard));
const strayHook = afterGuards.match(/\buse[A-Z]\w*\s*\(/);
assert.ok(
  !strayHook,
  `No hook may be called after the early returns, found: ${strayHook?.[0]}`,
);

console.log('check-app-render-guards: all assertions passed');
