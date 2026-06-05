import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { dirname, resolve, join } from 'node:path';
import { mkdtempSync, rmSync, existsSync, readFileSync, mkdirSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';

// Functional test of the filesystem project integration (issue #48): open,
// remember, list-recent, create, and forget must round-trip on a real temp dir.

const here = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);
const projects = require(resolve(here, '..', 'src', 'main', 'projects.js'));

const userData = mkdtempSync(join(tmpdir(), 'octra-desktop-'));
const workspace = mkdtempSync(join(tmpdir(), 'octra-workspace-'));

try {
  // Empty store reads as an empty list.
  assert.deepEqual(projects.readRecent(userData), [], 'fresh store must be empty');

  // Remembering a project persists it and puts it at the front.
  const a = join(workspace, 'alpha');
  mkdirSync(a);
  const r1 = projects.rememberProject(userData, a, 1000);
  assert.equal(r1.project.name, 'alpha', 'project name is the folder base name');
  assert.equal(r1.recent.length, 1, 'recent has one entry');
  assert.ok(existsSync(projects.storePath(userData)), 'store file is written');

  // A second project moves to the front; re-opening the first de-dupes it.
  const b = join(workspace, 'beta');
  mkdirSync(b);
  projects.rememberProject(userData, b, 2000);
  const r3 = projects.rememberProject(userData, a, 3000);
  assert.equal(r3.recent[0].name, 'alpha', 'most recent is first');
  assert.equal(r3.recent.length, 2, 're-opening de-dupes rather than duplicating');

  // Persistence survives a fresh read.
  const reread = projects.readRecent(userData);
  assert.equal(reread.length, 2, 'recents persist across reads');

  // Creating a new project makes the folder + a seed README.
  const created = projects.createProject(workspace, 'gamma');
  assert.ok(existsSync(created.path), 'created project folder exists');
  assert.ok(existsSync(join(created.path, 'README.md')), 'created project has a README');
  assert.match(readFileSync(join(created.path, 'README.md'), 'utf8'), /gamma/, 'README mentions the name');

  // Creating into a non-empty existing folder must fail loudly (no clobber).
  const occupied = join(workspace, 'occupied');
  mkdirSync(occupied);
  writeFileSync(join(occupied, 'keep.txt'), 'data');
  assert.throws(() => projects.createProject(workspace, 'occupied'), /not empty/, 'must not clobber existing work');

  // Names with path separators are rejected.
  assert.throws(() => projects.createProject(workspace, 'a/b'), /separators/, 'path separators rejected');

  // Forgetting removes the entry.
  const afterForget = projects.forgetProject(userData, a);
  assert.ok(!afterForget.some((p) => p.path === a || p.path === resolve(a)), 'forgotten project is gone');

  // A corrupt store degrades to an empty list instead of throwing.
  writeFileSync(projects.storePath(userData), '{ not json');
  assert.deepEqual(projects.readRecent(userData), [], 'corrupt store reads as empty');

  console.log('check-projects: all assertions passed');
} finally {
  rmSync(userData, { recursive: true, force: true });
  rmSync(workspace, { recursive: true, force: true });
}
