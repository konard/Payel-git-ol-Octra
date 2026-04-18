# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CrewAI is a multi-agent system that models a real IT company hierarchy:
```
User → Apigateway → Boss → Manager(s) → Worker(s) → ZIP archive
```

Services communicate via gRPC. All LLM calls route through the Agents service (not direct).

## Service Architecture

| Service | Port | Language | Role |
|---------|------|----------|------|
| Apigateway | 3111 | Go | HTTP/WebSocket entry point, proxies to Boss |
| Boss | 50051 | Go | CEO — analyzes task, plans architecture, coordinates managers |
| Manager | 50052 | Go | Team Lead — hires workers, reviews code |
| Worker | 50053 | Go | Developer — generates code files |
| Agents | 50053 | Go | LLM router — single interface to 6+ LLM providers |
| Auth | 3112 | Go | JWT authentication |
| Frontend | 80 | React/Vite | Web UI |
| tgbot | - | Python | Telegram bot |

External dependencies: PostgreSQL (5432), Redis (6379)

## Common Commands

### Full stack
```bash
docker-compose up -d --build
```

### Proto code generation
```powershell
.\generate-proto.ps1
```

### Individual Go service (local dev without Docker)
```bash
# Requires PostgreSQL and Redis running locally
cd agents && go run cmd/app/main.go   # Terminal 2
cd boss && go run cmd/app/main.go      # Terminal 3
cd manager && go run cmd/app/main.go   # Terminal 4
cd worker && go run cmd/app/main.go    # Terminal 5
cd apigateway && go run cmd/app/main.go # Terminal 6
```

### Frontend
```bash
cd frontend/web && npm run dev
```

### Test client
```bash
cd frontend/npm_client_test && node index.js
```

## Go Workspace

This project uses Go workspace (`go.work`). All modules are under `./agents`, `./apigateway`, `./boss`, `./manager`, `./worker`, `./auth`.

## Key Architecture Decisions

### Agents Service is the only LLM caller
Boss, Manager, and Worker do NOT call LLM providers directly — they call `Agents.Generate()` via gRPC. This provides:
- Single retry/logging point
- Per-request API keys (tokens passed in request, not stored server-side)
- Easy provider additions

### N+1 File Generation (Worker)
Workers use a two-step approach to avoid JSON truncation:
1. Ask AI "what files to create?" → simple JSON array
2. For each file, ask AI individually → plain text code (no JSON wrapper)
3. Strip markdown code blocks with `stripMarkdownCodeBlock()`

### Managers Run in Parallel (Boss)
Boss calls each manager's `AssignManager` in parallel goroutines. Higher priority managers execute first; lower priority ones wait briefly to gather context from completed managers.

### Worker Modes
Worker service supports two modes via `WORKER_MODE` env var:
- `multypass` (default, optimized) — uses Redis context caching, ~45% fewer tokens
- `nplus1` (legacy) — sequential file generation without context optimization

## Environment Variables

Key variables per service (see `.env.example` for full list):
- `DB_DNS` — PostgreSQL connection string
- `AGENTS_SERVICE_HOST` — Agents gRPC address (e.g., `agents:50053`)
- `REDIS_URL` / `REDIS_ENABLED` — Redis for WebSocket streaming and context caching
- `WORKER_MODE` — `multypass` or `nplus1`

## Database

PostgreSQL with auto-migrations on service startup. GORM models in each service's `pkg/models/`.

## Frontend

React app with Vite. Frontend guidelines are in `frontend/web/guidelines/Guidelines.md` (currently a template — add project-specific rules there).