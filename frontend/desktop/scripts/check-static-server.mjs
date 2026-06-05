#!/usr/bin/env node
/*
 * Tests electron/staticServer.ts — the loopback http server that serves the
 * built web app so Vite's absolute `/assets/...` URLs resolve (and the renderer
 * gets a real http origin for the backend WebSocket). Covers MIME mapping, the
 * SPA fallback, the traversal guard, and a real end-to-end fetch.
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const {
  contentTypeFor,
  resolveRequestPath,
  startStaticServer,
} = await import('../dist-electron/staticServer.js');

assert.equal(contentTypeFor('a.js'), 'text/javascript; charset=utf-8');
assert.equal(contentTypeFor('a.css'), 'text/css; charset=utf-8');
assert.equal(contentTypeFor('a.unknown'), 'application/octet-stream');

const dist = fs.mkdtempSync(path.join(os.tmpdir(), 'octra-dist-'));
try {
  fs.mkdirSync(path.join(dist, 'assets'));
  fs.writeFileSync(path.join(dist, 'index.html'), '<!doctype html><title>Octra</title>');
  fs.writeFileSync(path.join(dist, 'assets', 'app.js'), 'console.log(1)');

  // Existing asset resolves to the file.
  assert.equal(resolveRequestPath(dist, '/assets/app.js'), path.join(dist, 'assets', 'app.js'));
  // Unknown extensionless route falls back to index.html (SPA).
  assert.equal(resolveRequestPath(dist, '/projects/123'), path.join(dist, 'index.html'));
  // Root resolves to index.html.
  assert.equal(resolveRequestPath(dist, '/'), path.join(dist, 'index.html'));
  // Traversal is refused.
  assert.equal(resolveRequestPath(dist, '/../../etc/passwd'), null);

  // End-to-end: start the server and fetch a real asset + an SPA route.
  const server = await startStaticServer(dist);
  try {
    const js = await fetch(`${server.url}/assets/app.js`);
    assert.equal(js.status, 200);
    assert.match(js.headers.get('content-type'), /javascript/);
    assert.equal(await js.text(), 'console.log(1)');

    const spa = await fetch(`${server.url}/some/client/route`);
    assert.equal(spa.status, 200);
    assert.match(await spa.text(), /Octra/);
  } finally {
    await server.close();
  }

  console.log('check-static-server: OK');
} finally {
  fs.rmSync(dist, { recursive: true, force: true });
}
