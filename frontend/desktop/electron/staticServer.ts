'use strict';

// Minimal static file server for the built web app.
//
// The web app is built by Vite with `base: '/'`, so its index.html references
// assets with absolute paths like /assets/index.js. Loading that over file://
// breaks those paths, so in production we serve frontend/web/dist over a local
// loopback http server and point the BrowserWindow at http://127.0.0.1:<port>.
// This also gives the renderer a normal http origin (needed for the WebSocket
// backend and SPA history routing) — exactly matching how the web app runs.
//
// Uses only Node's built-in http/fs so the desktop app needs no extra runtime
// dependency. Kept dependency-free and pure enough to unit-test the routing.

import * as http from 'http';
import * as fs from 'fs';
import * as path from 'path';

const MIME: Record<string, string> = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.map': 'application/json; charset=utf-8',
};

export function contentTypeFor(filePath: string): string {
  return MIME[path.extname(filePath).toLowerCase()] || 'application/octet-stream';
}

// Map a request URL to a file inside the root, falling back to index.html for
// client-side routes (SPA). Returns null if the resolved path escapes the root.
export function resolveRequestPath(rootDir: string, requestUrl: string): string | null {
  const absRoot = path.resolve(rootDir);
  const pathname = decodeURIComponent((requestUrl || '/').split('?')[0].split('#')[0]);
  let candidate = path.resolve(absRoot, '.' + pathname);

  const rel = path.relative(absRoot, candidate);
  if (rel.startsWith('..') || path.isAbsolute(rel)) return null;

  try {
    const stat = fs.statSync(candidate);
    if (stat.isDirectory()) candidate = path.join(candidate, 'index.html');
  } catch {
    // Unknown path with no file extension → SPA route, serve index.html.
    if (!path.extname(candidate)) candidate = path.join(absRoot, 'index.html');
  }
  return candidate;
}

export interface RunningServer {
  url: string;
  port: number;
  close: () => Promise<void>;
}

// Start serving `rootDir` on a loopback port. Resolves once listening.
export function startStaticServer(rootDir: string): Promise<RunningServer> {
  const server = http.createServer((req, res) => {
    const filePath = resolveRequestPath(rootDir, req.url || '/');
    if (!filePath) {
      res.writeHead(403).end('Forbidden');
      return;
    }
    fs.readFile(filePath, (err, data) => {
      if (err) {
        res.writeHead(404).end('Not found');
        return;
      }
      res.writeHead(200, { 'Content-Type': contentTypeFor(filePath) });
      res.end(data);
    });
  });

  return new Promise((resolve, reject) => {
    server.on('error', reject);
    // Port 0 → let the OS pick a free port, avoiding clashes.
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      const port = typeof address === 'object' && address ? address.port : 0;
      resolve({
        url: `http://127.0.0.1:${port}`,
        port,
        close: () =>
          new Promise<void>((res) => {
            server.close(() => res());
          }),
      });
    });
  });
}
