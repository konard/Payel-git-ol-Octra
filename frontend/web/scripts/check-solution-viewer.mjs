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
assert.match(app, /from '\.\/components\/solution\/SolutionViewer'/, 'App must import SolutionViewer');
assert.match(app, /<SolutionViewer \/>/, 'App must render <SolutionViewer />');
assert.match(app, /'canvas' \| 'chat' \| 'solution'/, "App mode union must use 'solution'");
assert.doesNotMatch(app, /CodeViewer/, 'App must not reference the old CodeViewer');
assert.doesNotMatch(app, /mode === 'code'/, "App must not use the old 'code' mode");

const topBar = read('src/app/components/shell/TopBar.tsx');
assert.match(topBar, /onModeChange\('solution'\)/, "TopBar tab must switch to 'solution'");
assert.match(topBar, />\s*Solution\s*</, 'TopBar tab label must read "Solution"');
assert.doesNotMatch(topBar, /onModeChange\('code'\)/, "TopBar must not switch to the old 'code' mode");

const viewer = read('src/app/components/solution/SolutionViewer.tsx');
assert.match(viewer, /export function SolutionViewer\(/, 'SolutionViewer must be exported');
assert.match(viewer, /react-markdown/, 'SolutionViewer must use react-markdown');
assert.match(viewer, /remark-gfm/, 'SolutionViewer must use remark-gfm (tables, task lists)');
// Binary documents (e.g. generated .pptx presentations) must not be dumped into
// the Monaco editor — the viewer shows a placeholder instead.
assert.match(viewer, /isBinaryPath/, 'SolutionViewer must guard binary files with isBinaryPath');
assert.match(viewer, /binaryFileLabel/, 'SolutionViewer must label binary files');
assert.match(viewer, /Download/, 'SolutionViewer must offer generated binary files as downloads');
assert.match(viewer, /displayedContentByPathRef/, 'SolutionViewer must cache streaming text per file path');
// Document solutions get a clean reader (no explorer/tabs) with a page switcher.
assert.match(viewer, /isDocumentSolution/, 'SolutionViewer must detect document solutions');
assert.match(viewer, /paginateMarkdown/, 'SolutionViewer must paginate document reader pages');
assert.match(viewer, /documentMode/, 'SolutionViewer must branch on documentMode for the reader');
assert.match(viewer, /choosePreferredCodeFilePath/, 'SolutionViewer must preserve the user-selected document');
assert.match(viewer, /PresentationDeckPreview/, 'SolutionViewer must preview generated presentations from slide Markdown');
assert.match(viewer, /findPresentationPreviewFile/, 'SolutionViewer must find Markdown paired with generated pptx files');
// Real images embedded into slides must render as <img>, not just visual text.
assert.match(viewer, /currentSlide\.image/, 'SolutionViewer must branch on the slide image');
assert.match(viewer, /<img/, 'SolutionViewer must render real slide images as <img>');

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
  const { isMarkdownPath, isBinaryPath, binaryFileLabel, isDocumentSolution, paginateMarkdown } =
    await server.ssrLoadModule('/src/lib/markdown.ts');
  const { choosePreferredCodeFilePath } = await server.ssrLoadModule('/src/lib/solutionFiles.ts');
  const { parsePresentationMarkdown, findPresentationPreviewFile } =
    await server.ssrLoadModule('/src/lib/presentation.ts');

  for (const path of ['report.md', 'docs/Research.MARKDOWN', 'a/b/notes.mdx', 'README.md']) {
    assert.equal(isMarkdownPath(path), true, `${path} should be detected as Markdown`);
  }
  for (const path of ['main.go', 'src/index.ts', 'style.css', 'noextension', 'mdfile']) {
    assert.equal(isMarkdownPath(path), false, `${path} should NOT be detected as Markdown`);
  }

  // --- 3. Binary-document detection (presentations, documents, assets). -----
  for (const path of ['solution/deck.pptx', 'a/b/Report.DOCX', 'sheet.xlsx', 'paper.pdf', 'img.png', 'photo.JPEG', 'bundle.zip']) {
    assert.equal(isBinaryPath(path), true, `${path} should be detected as binary`);
  }
  for (const path of ['report.md', 'main.go', 'notes.txt', 'data.json']) {
    assert.equal(isBinaryPath(path), false, `${path} should NOT be detected as binary`);
  }
  assert.equal(binaryFileLabel('solution/deck.pptx'), 'PowerPoint presentation', 'pptx label');
  assert.equal(binaryFileLabel('report.docx'), 'Word document', 'docx label');

  // --- 4. Document-solution detection (reader mode, no file explorer). -------
  // Pure documents / presentations → reader mode.
  assert.equal(isDocumentSolution(['solution/report.md']), true, 'single Markdown is a document solution');
  assert.equal(isDocumentSolution(['solution/report.md', 'solution/deck.pptx']), true, 'md + pptx is a document solution');
  assert.equal(isDocumentSolution(['a.md', 'b.markdown', 'c.mdx']), true, 'all-Markdown is a document solution');
  // Anything with real source code (or plain text/JSON) keeps the IDE view.
  assert.equal(isDocumentSolution(['src/main.go', 'README.md']), false, 'code projects are NOT document solutions');
  assert.equal(isDocumentSolution(['notes.txt']), false, 'plain text is NOT a document solution');
  assert.equal(isDocumentSolution([]), false, 'empty solution is NOT a document solution');

  // --- 5. Markdown pagination for the bottom page switcher. ------------------
  assert.deepEqual(paginateMarkdown('Just one short paragraph.'), ['Just one short paragraph.'], 'short doc is a single page');
  // Explicit thematic breaks (`---`) act as page boundaries.
  const paged = paginateMarkdown('# Page one\n\nIntro.\n\n---\n\n# Page two\n\nMore.');
  assert.equal(paged.length, 2, 'thematic breaks split into pages');
  assert.match(paged[0], /Page one/, 'first page keeps its content');
  assert.match(paged[1], /Page two/, 'second page keeps its content');
  assert.ok(paged.every((page) => !/^\s*---\s*$/m.test(page)), 'page-break markers are stripped');
  // Long break-free documents auto-paginate by length.
  const longDoc = Array.from({ length: 90 }, (_, i) => `Paragraph number ${i} with enough text to add up over many blocks.`).join('\n\n');
  assert.ok(paginateMarkdown(longDoc).length > 1, 'long documents paginate by length');
  // Empty content still yields one (empty) page.
  assert.deepEqual(paginateMarkdown(''), [''], 'empty content yields a single page');

  // --- 6. Document selection priority. --------------------------------------
  // When a workflow writes multiple Markdown files, selecting a document in the
  // Solution tab must not snap back to the latest generated file.
  assert.equal(
    choosePreferredCodeFilePath(['solution/httpx-install.md', 'solution/sources.md'], 'solution/httpx-install.md', 'solution/sources.md'),
    'solution/httpx-install.md',
    'active file should remain selected over latest',
  );
  assert.equal(
    choosePreferredCodeFilePath(['solution/httpx-install.md', 'solution/sources.md'], 'missing.md', 'solution/sources.md'),
    'solution/sources.md',
    'latest file should be the fallback when active file is unavailable',
  );
  assert.equal(
    choosePreferredCodeFilePath(['solution/deck.md', 'solution/deck.pptx', 'solution/final-presentation.md'], null, 'solution/final-presentation.md'),
    'solution/deck.pptx',
    'presentation workflows should open the generated pptx instead of the final markdown summary',
  );
  assert.equal(
    choosePreferredCodeFilePath(['solution/httpx-install.md', 'solution/sources.md'], null, null),
    'solution/httpx-install.md',
    'first file should be the fallback when active and latest are unavailable',
  );
  assert.equal(choosePreferredCodeFilePath([], null, null), null, 'empty file list should return null');

  // --- 7. Presentation preview model. --------------------------------------
  const deck = parsePresentationMarkdown(`# Launch Plan

## Opening
- Define the opportunity
- Align stakeholders
Visual: Product dashboard screenshot
Source: Internal analytics
> Keep this slide under one minute.

## Roadmap
* Phase one
* Phase two`);
  assert.equal(deck.title, 'Launch Plan', 'presentation parser keeps deck title');
  assert.equal(deck.slides.length, 2, 'presentation parser creates slides from h2 headings');
  assert.equal(deck.slides[0].title, 'Opening', 'first slide title');
  assert.deepEqual(deck.slides[0].bullets, ['Define the opportunity', 'Align stakeholders'], 'first slide bullets');
  assert.equal(deck.slides[0].visual, 'Product dashboard screenshot', 'visual direction is parsed');
  assert.deepEqual(deck.slides[0].sources, ['Internal analytics'], 'sources are parsed');
  assert.equal(deck.slides[0].notes, 'Keep this slide under one minute.', 'speaker notes are parsed');
  assert.deepEqual(deck.slides[1].bullets, ['Phase one', 'Phase two'], 'asterisk bullets are parsed');

  // Real images: direct URLs (Markdown image, Image: line) become an embeddable
  // image field; plain text visual ideas stay as visual descriptions only.
  const withImages = parsePresentationMarkdown(`# Deck

## Photo slide
- Point one
![A harbor at sunrise](https://example.com/harbor.jpg)

## Idea slide
- Another point
Image: https://example.com/chart.png

## Text-only slide
- Final point
Visual: A simple diagram with no URL`);
  assert.equal(withImages.slides[0].image?.url, 'https://example.com/harbor.jpg', 'markdown image URL is captured');
  assert.equal(withImages.slides[0].image?.alt, 'A harbor at sunrise', 'markdown image alt is captured');
  assert.equal(withImages.slides[0].visual, 'A harbor at sunrise', 'image alt falls back into visual text');
  assert.equal(withImages.slides[1].image?.url, 'https://example.com/chart.png', 'Image: line URL is captured');
  assert.equal(withImages.slides[2].image, null, 'text-only visual stays without an image');
  assert.equal(withImages.slides[2].visual, 'A simple diagram with no URL', 'plain visual text is preserved');

  const preview = findPresentationPreviewFile([
    { path: 'solution/designer-a1b2.md', content: '# Deck\n\n## Slide\n- point' },
    { path: 'solution/designer-a1b2.pptx', content: 'UEs=', encoding: 'base64' },
    { path: 'solution/final-presentation.md', content: '# Summary' },
  ], 'solution/designer-a1b2.pptx');
  assert.equal(preview?.path, 'solution/designer-a1b2.md', 'paired Markdown should preview the generated pptx');

  console.log('check-solution-viewer: all assertions passed');
} finally {
  await server.close();
}
