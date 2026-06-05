'use strict';

// Resolves how the desktop shell reaches the Octra web renderer and backend.
//
// The desktop app does NOT reimplement the UI — it loads the real web React app
// (frontend/web) as its renderer, so the IDE looks and behaves exactly like the
// web product (issue #50 points 1, 2, 6, 7, 8). Two modes:
//
//   • dev  — load the running Vite dev server (npm run dev in frontend/web).
//            The Vite proxy forwards /api + /auth to the local backend, so the
//            backend interaction is identical to the web app with zero config.
//   • prod — serve the built frontend/web/dist over a local http server and
//            point the renderer at the configured backend via VITE_* envs baked
//            into the build.
//
// Everything is overridable through environment variables so CI / packagers can
// point at any backend without code changes.

import * as path from 'path';

export interface DesktopConfig {
  // Set when we should load a remote dev server instead of the bundled dist.
  devServerUrl: string | null;
  // Absolute path to the built web app (index.html lives here) for prod.
  webDistDir: string;
  // Backend base URLs surfaced to the renderer via window.octra.getConfig().
  apiBaseUrl: string;
  wsUrl: string;
  // Whether to open devtools on launch.
  openDevtools: boolean;
}

export function resolveConfig(env: NodeJS.ProcessEnv = process.env): DesktopConfig {
  const devServerUrl = env.OCTRA_DEV_SERVER_URL || (env.OCTRA_DEV === '1' ? 'http://localhost:5173' : null);

  // In a checkout the built web app lives at ../web/dist relative to this file's
  // compiled location (dist-electron/). Allow an explicit override for packaging.
  const webDistDir = env.OCTRA_WEB_DIST
    ? path.resolve(env.OCTRA_WEB_DIST)
    : path.resolve(__dirname, '..', '..', 'web', 'dist');

  return {
    devServerUrl,
    webDistDir,
    apiBaseUrl: env.OCTRA_API_URL || '',
    wsUrl: env.OCTRA_WS_URL || '',
    openDevtools: env.OCTRA_DEVTOOLS === '1',
  };
}
