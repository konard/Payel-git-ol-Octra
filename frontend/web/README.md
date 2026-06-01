# Frontend

React-based web interface for Octra. Provides a visual node-based workflow editor, real-time task monitoring, chat interface, and solution viewer.

## Tech Stack

- **Framework:** React 18 + TypeScript
- **Build:** Vite 6
- **Canvas:** @xyflow/react (ReactFlow)
- **State:** Zustand
- **UI:** MUI 7, Radix UI, Tailwind CSS 4
- **Editor:** Monaco Editor
- **Icons:** Lucide React
- **i18n:** 33 languages

## Pages

| Route | Description |
|-------|-------------|
| `/` | Landing page (unauthenticated) or main app |
| `/payment-success` | Payment confirmation |

## App Modes

- **Canvas** — Node-based workflow builder with drag-and-drop
- **Chat** — Conversational interface with real-time progress
- **Solution** — Code viewer and document reader

## Development

```bash
npm install
npm run dev       # http://localhost:5173
```

### Proxy (vite.config.ts)

- `/auth/*`, `/me`, `/workflows/*`, `/payments/*`, `/chat/*` → `localhost:3112` (User Service)
- `/api/*` → `localhost:3111` (API Gateway) with WebSocket support

## Build

```bash
npm run build     # Output: dist/
```

## Docker

```bash
docker build -t octra-frontend .
```

Served via nginx with reverse proxy to User Service and API Gateway.

## Environment

| Variable | Description |
|----------|-------------|
| `VITE_OPENROUTER_API_KEY` | Default OpenRouter API key |
