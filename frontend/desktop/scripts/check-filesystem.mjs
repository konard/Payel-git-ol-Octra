#!/usr/bin/env node
/*
 * Unit test for electron/fileSystem.ts — the project file reading that the old
 * desktop shell was missing entirely (issue #50.4). Builds a temp project on
 * disk, then asserts the tree walk and single-file read behave correctly,
 * including the ignore list, traversal guard, and binary/size handling.
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const {
  readProjectTree,
  readProjectFile,
  isBinaryPath,
  resolveWithinRoot,
} = await import('../dist-electron/fileSystem.js');

const root = fs.mkdtempSync(path.join(os.tmpdir(), 'octra-fs-'));
try {
  // Build a small project.
  fs.mkdirSync(path.join(root, 'src'));
  fs.mkdirSync(path.join(root, 'node_modules'));
  fs.writeFileSync(path.join(root, 'node_modules', 'junk.js'), 'x');
  fs.writeFileSync(path.join(root, 'README.md'), '# Hello\n');
  fs.writeFileSync(path.join(root, 'src', 'index.ts'), 'export const a = 1;\n');

  const tree = readProjectTree(root);
  assert.equal(tree.type, 'directory');
  const names = tree.children.map((c) => c.name);
  assert.ok(names.includes('src'), 'tree should include src/');
  assert.ok(names.includes('README.md'), 'tree should include README.md');
  assert.ok(!names.includes('node_modules'), 'node_modules must be ignored');
  // Directories sort before files.
  assert.equal(tree.children[0].type, 'directory');

  const src = tree.children.find((c) => c.name === 'src');
  assert.equal(src.children[0].name, 'index.ts');
  assert.equal(src.children[0].path, 'src/index.ts');

  // Read a file.
  const file = readProjectFile(root, 'src/index.ts');
  assert.equal(file.content, 'export const a = 1;\n');
  assert.equal(file.encoding, 'utf8');
  assert.equal(file.binary, false);
  assert.equal(file.truncated, false);

  // Path traversal is refused.
  assert.equal(resolveWithinRoot(root, '../escape'), null);
  assert.throws(() => readProjectFile(root, '../../etc/passwd'));

  // Binary detection by extension.
  assert.equal(isBinaryPath('logo.png'), true);
  assert.equal(isBinaryPath('main.ts'), false);

  // Missing root throws.
  assert.throws(() => readProjectTree(path.join(root, 'does-not-exist')));

  console.log('check-filesystem: OK');
} finally {
  fs.rmSync(root, { recursive: true, force: true });
}
