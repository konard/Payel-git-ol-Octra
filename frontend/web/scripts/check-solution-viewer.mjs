import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { createServer } from 'vite';

// Regression test for issue #17: the "Code" tab was renamed to "Solution" and
// the viewer now renders Markdown results (research reports, documents, tables)
// as a formatted preview by default. This guards both the rename and the
// Markdown detection used to pick the preview vs. the Monaco source view.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

// --- 1. Source-level rename assertions. -----------------------------------
const app = read('src/app/App.tsx');
assert.match(app, /from '\.\/components\/SolutionViewer'/, 'App must import SolutionViewer');
assert.match(app, /<SolutionViewer \/>/, 'App must render <SolutionViewer />');
assert.match(app, /'canvas' \| 'chat' \| 'solution'/, "App mode union must use 'solution'");
assert.doesNotMatch(app, /CodeViewer/, 'App must not reference the old CodeViewer');
assert.doesNotMatch(app, /mode === 'code'/, "App must not use the old 'code' mode");

const topBar = read('src/app/components/TopBar.tsx');
assert.match(topBar, /onModeChange\('solution'\)/, "TopBar tab must switch to 'solution'");
assert.match(topBar, />\s*Solution\s*</, 'TopBar tab label must read "Solution"');
assert.doesNotMatch(topBar, /onModeChange\('code'\)/, "TopBar must not switch to the old 'code' mode");

const viewer = read('src/app/components/SolutionViewer.tsx');
assert.match(viewer, /export function SolutionViewer\(/, 'SolutionViewer must be exported');
assert.match(viewer, /react-markdown/, 'SolutionViewer must use react-markdown');
assert.match(viewer, /remark-gfm/, 'SolutionViewer must use remark-gfm (tables, task lists)');

const pkg = JSON.parse(read('package.json'));
assert.ok(pkg.dependencies['react-markdown'], 'react-markdown must be a dependency');
assert.ok(pkg.dependencies['remark-gfm'], 'remark-gfm must be a dependency');

// --- 2. Behavioural test of the Markdown detection. ------------------------
const server = await createServer({
  root,
  logLevel: 'error',
  mode: 'test',
  envFile: false,
  server: { middlewareMode: true },
  optimizeDeps: { noDiscovery: true, include: [] },
  appType: 'custom',
});

try {
  const { isMarkdownPath } = await server.ssrLoadModule('/src/lib/markdown.ts');

  for (const path of ['report.md', 'docs/Research.MARKDOWN', 'a/b/notes.mdx', 'README.md']) {
    assert.equal(isMarkdownPath(path), true, `${path} should be detected as Markdown`);
  }
  for (const path of ['main.go', 'src/index.ts', 'style.css', 'noextension', 'mdfile']) {
    assert.equal(isMarkdownPath(path), false, `${path} should NOT be detected as Markdown`);
  }

  console.log('check-solution-viewer: all assertions passed');
} finally {
  await server.close();
}
