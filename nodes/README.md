# Nodes

The core orchestration engine. Contains **Boss**, **Manager**, and **Worker** agents in a single process. Receives tasks via gRPC, plans architecture using AI, spawns parallel manager/worker teams, and publishes results to GitHub.

## Architecture

```
API Gateway ──gRPC──► Nodes (50051)
                        │
              ┌─────────┼─────────┐
              ▼         ▼         ▼
            Boss     Manager   Worker
              │         │         │
              └─────────┼─────────┘
                        ▼
                   Agents (gRPC, 50053)
                        │
                        ▼
                   AI Providers
```

### Flow

1. **Boss** receives task, calls Grademodel for complexity score, plans architecture via AI
2. **Boss** spawns **Managers** in parallel goroutines
3. **Managers** hire **Workers**, review their output, request fixes
4. **Workers** generate code, research, documents, or presentations
5. **Boss** validates solution and pushes to GitHub (new repo or PR)

## Task Types

| Type | Output | Worker Produces |
|------|--------|----------------|
| `code` | Source files | Code files, git commits |
| `research` | Markdown report | Web-sourced document |
| `document` | Markdown document | Structured document |
| `presentation` | PPTX + Markdown | Slide deck |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NODES_GRPC_PORT` | `50051` | gRPC server port |
| `AGENTS_SERVICE_HOST` | `localhost:50053` | Agents gRPC address |
| `DB_DNS` | — | PostgreSQL connection string |
| `REDIS_URL` | — | Redis connection string |
| `REDIS_ENABLED` | `false` | Enable stream persistence |
| `GITHUB_TOKEN` | — | GitHub API token for publishing |
| `GIT_USER_NAME` | `CrewAI Bot` | Git commit author name |
| `GIT_USER_EMAIL` | `bot@crewai.local` | Git commit author email |
| `PROJECTS_DIR` | `/workspace/projects` | Workspace directory |
| `WORKER_MODE` | `multypass` | Code generation mode |
| `WEB_SEARCH_DISABLED` | — | Disable web search |

## Development

```bash
go run cmd/app/main.go
```

## Docker

```bash
docker build -t octra-nodes .
```
