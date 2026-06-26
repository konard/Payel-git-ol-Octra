# Octra Desktop

Electron shell for the Octra web frontend. The desktop process loads the real `frontend/web` app instead of maintaining a separate renderer.

## Runtime Modes

- **Development:** set `OCTRA_DEV=1` and run the Next dev server at `http://localhost:3000`.
- **Production:** build `frontend/web`, then the shell serves `frontend/web/dist` over a loopback HTTP server.

The preload bridge exposes `window.octra` for native window controls, recent projects, filesystem reads, shell links, and backend configuration. The Next renderer mounts a desktop-only title bar when that bridge is present and renders nothing extra in the browser.

## Develop

```bash
cd frontend/web
npm install
npm run dev
```

```bash
cd frontend/desktop
npm install
npm run dev
```

## Production Run

```bash
cd frontend/web
npm run build
cd ../desktop
npm start
```

## Test

```bash
npm test
```

The tests compile `electron/*.ts` and check the filesystem reader, recent-projects store, menu model, static file server, and IPC contract.
