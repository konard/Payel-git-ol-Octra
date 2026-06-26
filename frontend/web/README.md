# Frontend

Next-based web interface for Octra. The package now ships a new static frontend with a landing page, authentication page, and operational dashboard.

## Tech Stack

- **Framework:** Next 16 + TypeScript
- **UI:** React 18 + Lucide icons
- **Output:** Static export copied from `out/` to `dist/`

## Routes

| Route | Description |
|-------|-------------|
| `/` | New Octra landing page |
| `/auth` | Sign in and create account page with Google, GitHub, and Lefine entry points |
| `/dashboard` | New main dashboard surface |

## Development

```bash
npm install
npm run dev
```

The dev server runs on the standard Next port, `http://localhost:3000`.

## Build

```bash
npm run build
```

The build runs `next build`, exports static files to `out/`, and copies them to `dist/` for nginx and the desktop shell.

## Docker

```bash
docker build -t octra-frontend .
```
