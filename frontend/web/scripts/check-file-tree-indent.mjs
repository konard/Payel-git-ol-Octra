import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression test for issue #31: the file tree nesting "looked bad, no spacing,
// unclear whether a file is in a folder". The fix draws a vertical guide line per
// nesting level instead of relying on bare left padding.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const viewer = read('src/app/components/SolutionViewer.tsx');

// A dedicated indent component with one guide line per depth level.
assert.match(viewer, /function TreeIndent\(/, 'SolutionViewer must define a TreeIndent component');
assert.match(viewer, /Array\.from\(\{ length: depth \}\)/, 'TreeIndent must render one guide per nesting level');
assert.match(viewer, /border-l border-\[var\(--code-tree-guide\)\]/, 'TreeIndent guides must use the tree-guide colour');
assert.match(viewer, /--code-tree-guide/, 'a --code-tree-guide colour variable must exist');

// Both folder and file rows must render the indent guides.
const treeIndentUses = viewer.match(/<TreeIndent depth=\{depth\} \/>/g) || [];
assert.ok(treeIndentUses.length >= 2, 'both folder and file rows must render <TreeIndent depth={depth} />');

// The old flat "10 + depth * 16" inline padding must be gone.
assert.doesNotMatch(viewer, /10 \+ depth \* 16/, 'the old flat paddingLeft must be replaced by indent guides');

console.log('check-file-tree-indent: all assertions passed');
