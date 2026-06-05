# Octra Desktop

A desktop IDE for **Octra** built with **Electron** and written entirely in
**TypeScript**. Instead of reimplementing the UI, the desktop app **runs the real
Octra web app (`frontend/web`) as its renderer**, so the IDE looks, behaves and
talks to the backend exactly like the web product. On top of that renderer it adds
a frameless **title bar with window controls**, a native **application menu**, a
**recent-projects** store and a **filesystem-backed file explorer**.

> Implements [issue #50](https://github.com/Payel-git-ol/Octra/issues/50).

## Why load the web app instead of rebuilding it?

The previous shell (issue #48) was a separate vanilla HTML/CSS/JS UI that drifted
from the web app — it didn't look like Octra, most buttons did nothing, and it had
its own backend assumptions. This rewrite deletes that duplication:

| Issue #50 requirement | How it's met |
| --- | --- |
| 1. Look like the web app | The renderer **is** the web app (`frontend/web`). |
| 2. Buttons actually work | Real React app → all existing web functionality works. |
| 3. Rewrite JS → TypeScript | The whole main process is TypeScript (`electron/*.ts`). |
| 4. Read & display project files | `fs:read-tree` / `fs:read-file` IPC + a file explorer and Monaco viewer in the renderer. |
| 5. Fix the window-control buttons | Frameless window + a styled React title bar that reuses Octra design tokens. |
| 6. Same backend as the web app | The web renderer uses its own API/WS client unchanged (dev uses the Vite proxy). |
| 7. Fix the Canvas | The web app's real `@xyflow/react` canvas renders as-is. |
| 8. Use React components | The renderer is the web app's React tree; desktop chrome is React too. |

The desktop-only React pieces live in `frontend/web/src/desktop/` and every one of
them is gated on `window.octra?.isElectron`, so the plain web build is unaffected.

## Architecture

```
frontend/desktop/
  electron/                 Electron main process — TypeScript, compiled by tsc
    main.ts                 frameless window, IPC handlers, renderer lifecycle
    preload.ts              contextBridge → window.octra (the only renderer ↔ Node surface)
    config.ts               resolves dev-server vs bundled-dist + backend URLs from env
    staticServer.ts         loopback http server that serves frontend/web/dist in prod
    fileSystem.ts           readProjectTree / readProjectFile (pure Node, sandboxed to root)
    projects.ts             recent-projects store (open / remember / list / forget)
    menu.ts, menuModel.ts   native application menu built from a shared model
  scripts/                  Node unit tests (no Electron runtime needed)
  tsconfig.json             rootDir electron/ → outDir dist-electron/
  dist-electron/            tsc output; main field points at dist-electron/main.js

frontend/web/src/desktop/   the desktop-only renderer pieces (gated on Electron)
  bridge.ts                 typed wrapper over window.octra + isDesktopApp()
  desktopStore.ts           zustand store: project, file tree, open file, recents
  DesktopTitleBar.tsx       frameless title bar + minimize/maximize/close controls
  DesktopFileExplorer.tsx   project file tree
  DesktopFileViewer.tsx     read-only Monaco viewer for the selected file
```

### How the renderer is loaded

`config.ts` resolves one of two modes (everything overridable via env vars so
packagers/CI can point at any backend without code changes):

- **dev** (`OCTRA_DEV=1`) — load the running Vite dev server
  (`http://localhost:5173`). The Vite proxy forwards `/api` + `/auth` to the local
  backend, so backend interaction is identical to the web app with zero config.
- **prod** — serve the built `frontend/web/dist` over a loopback http server
  (`staticServer.ts`) and load that URL.

Security: the window runs with `contextIsolation: true` and `nodeIntegration:
false`. The renderer reaches the filesystem only through the explicit, whitelisted
channels in `preload.ts`, and `fileSystem.ts` refuses any path that escapes the
opened project root.

## Develop

```bash
# 1. start the web app (the renderer) in one terminal
cd frontend/web && npm install && npm run dev

# 2. launch the desktop shell against it in another terminal
cd frontend/desktop && npm install && npm run dev   # OCTRA_DEV=1 → loads localhost:5173
```

For a production-like run, build the web app first, then start Electron:

```bash
cd frontend/web && npm run build      # produces frontend/web/dist
cd ../desktop && npm start            # serves dist over loopback and loads it
```

## Test

Structural and functional checks (no Electron runtime required). `npm test` first
compiles `electron/*.ts` with `tsc`, then runs the unit tests:

```bash
npm test
```

- `check-filesystem` — `readProjectTree` / `readProjectFile` on a temp dir, including the path-escape guard
- `check-projects` — round-trips open / remember / list / forget recent projects
- `check-menu-model` — the application menu model is well-formed and shared
- `check-static-server` — the loopback dist server serves files with correct MIME types
- `check-electron-main` — frameless, context-isolated window with the expected IPC surface

The web side adds `check-desktop-integration` and `check-app-render-guards`
(run via `npm test` in `frontend/web`) to guarantee the desktop gates stay in
place and the renderer never regresses the Rules-of-Hooks ordering.
