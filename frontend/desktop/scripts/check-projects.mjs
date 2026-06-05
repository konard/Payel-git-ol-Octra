#!/usr/bin/env node
/*
 * Unit test for electron/projects.ts — the persisted recent-projects store. Uses
 * a temp userData dir so it never touches the real Electron profile. Covers
 * remember/forget ordering, the on-disk JSON, dropping deleted folders, and the
 * createProject guard.
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const {
  readRecent,
  rememberProject,
  forgetProject,
  createProject,
  describeProject,
  MAX_RECENT,
} = await import('../dist-electron/projects.js');

const userData = fs.mkdtempSync(path.join(os.tmpdir(), 'octra-store-'));
const work = fs.mkdtempSync(path.join(os.tmpdir(), 'octra-proj-'));
try {
  // Empty to start.
  assert.deepEqual(readRecent(userData), []);

  // Two real folders on disk.
  const alpha = path.join(work, 'alpha');
  const beta = path.join(work, 'beta');
  fs.mkdirSync(alpha);
  fs.mkdirSync(beta);

  const r1 = rememberProject(userData, alpha, 1000);
  assert.equal(r1.project.name, 'alpha');
  assert.equal(r1.recent.length, 1);

  const r2 = rememberProject(userData, beta, 2000);
  // Most recent first.
  assert.equal(r2.recent[0].name, 'beta');
  assert.equal(r2.recent[1].name, 'alpha');

  // Re-opening alpha moves it to the front without duplicating.
  const r3 = rememberProject(userData, alpha, 3000);
  assert.equal(r3.recent.length, 2);
  assert.equal(r3.recent[0].name, 'alpha');

  // Persisted to disk as JSON.
  const onDisk = JSON.parse(fs.readFileSync(path.join(userData, 'recent-projects.json'), 'utf8'));
  assert.equal(onDisk.length, 2);

  // Deleting a folder drops it from the list on next read.
  fs.rmSync(beta, { recursive: true, force: true });
  const afterDelete = readRecent(userData);
  assert.equal(afterDelete.length, 1);
  assert.equal(afterDelete[0].name, 'alpha');

  // forgetProject removes the entry.
  assert.deepEqual(forgetProject(userData, alpha), []);

  // describeProject derives the name from the folder.
  assert.equal(describeProject('/tmp/foo/bar').name, 'bar');

  // createProject makes a folder + README and refuses bad names.
  const created = createProject(work, 'fresh');
  assert.ok(fs.existsSync(path.join(work, 'fresh', 'README.md')));
  assert.equal(created.name, 'fresh');
  assert.throws(() => createProject(work, ''));
  assert.throws(() => createProject(work, 'a/b'));

  assert.equal(typeof MAX_RECENT, 'number');

  console.log('check-projects: OK');
} finally {
  fs.rmSync(userData, { recursive: true, force: true });
  fs.rmSync(work, { recursive: true, force: true });
}
