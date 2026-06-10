// Reproducible capture of the four landing-page screenshots.
//
// It drives the screenshot harness (screenshot.html + src/screenshot/) with a
// headless Chromium and writes each surface to src/images/main/landing/. The
// harness pins the UI to the dark theme and English and seeds the real
// components, so re-running this produces the same screenshots every time.
//
// Usage:
//   1. npm run dev            # in one terminal (serves http://localhost:5173)
//   2. npm i -D playwright && npx playwright install chromium
//   3. node scripts/capture-landing.mjs [baseUrl]
//
// Playwright is intentionally NOT a project dependency (it is only needed to
// regenerate the assets), so this script imports it lazily and prints a helpful
// hint if it is missing.

import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, '..');
const baseUrl = (process.argv[2] || 'http://localhost:5173').replace(/\/$/, '');

// 1280x800 is exactly the 16:10 aspect ratio of the landing preview pane
// (WorkflowDemo.tsx renders the screenshots with `aspect-[16/10] object-cover`).
const VIEWPORT = { width: 1280, height: 800 };

const SURFACES = [
  { surface: 'code-view', file: 'code-view-dark.png', ready: 'text=Solution files' },
  { surface: 'research-progress', file: 'research-progress-dark.png', ready: 'text=Completed' },
  { surface: 'document-reader', file: 'document-reader-dark.png', ready: 'text=Q2 2026 Business Review' },
  { surface: 'presentation-deck', file: 'presentation-deck-dark.png', ready: 'text=Octra: Agents That Ship Real Work' },
];

async function main() {
  let chromium;
  try {
    ({ chromium } = await import('playwright'));
  } catch {
    console.error(
      'Playwright is not installed. Install it just for capturing assets:\n' +
        '  npm i -D playwright && npx playwright install chromium',
    );
    process.exit(1);
  }

  const browser = await chromium.launch();
  const context = await browser.newContext({ viewport: VIEWPORT, deviceScaleFactor: 2 });
  const page = await context.newPage();

  for (const { surface, file, ready } of SURFACES) {
    const url = `${baseUrl}/screenshot.html?surface=${surface}`;
    console.log(`Capturing ${surface} -> ${file}`);
    await page.goto(url, { waitUntil: 'networkidle' });
    await page.waitForSelector(ready, { timeout: 15000 });
    // Give fonts, Monaco tokenisation and transitions a beat to settle.
    await page.waitForTimeout(1200);
    const outPath = path.join(root, 'src/images/main/landing', file);
    await page.screenshot({ path: outPath });
  }

  await browser.close();
  console.log('Done. Updated screenshots in src/images/main/landing/.');
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
