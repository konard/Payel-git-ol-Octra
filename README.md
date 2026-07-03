# Octra — AI Environment Hosting Platform

Octra provides **personal MCP endpoints** with AI agents running in isolated Nix
environments. Users get an API key, create an environment, and talk to their
agent over HTTP — no VPS, no SSH, no systemd.

```
User → Octra (Go) → Ocawe (Crystal) → LLM / CLI / Workflow
                                          ├── openai/gpt-4o
                                          ├── anthropic/claude-sonnet-4
                                          ├── cli/opencode
                                          └── workflow/orator
```

**Octra** — VPS-hosting platform. Manages users, billing, Nix environments, and
proxies AI/CLI/workflow requests to Ocawe.

**Ocawe** — AI engine and workflow runtime (Crystal). Handles all LLM calls, CLI
subprocess management, MCP server CRUD, workflow execution, and HITL.

## Quick start

```bash
docker compose up -d --build
```

API listens on `:8080`. Frontend on `:3000`.

```bash
cd backend && go build ./... && go test ./...   # SQLite in-memory, no external deps
cd frontend/web && npm run build                # static export
```

## Registration

```bash
curl http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{ "email": "me@example.com", "password": "secret" }'
# → { "user_id": "…", "api_key": "octra_…", "balance": 100 }
```

Save the returned `api_key` — it is used as `octra-api-token` on all subsequent requests.

## Chat & Streaming

Send a prompt to your AI agent. The agent is configured via your environment settings (LLM provider, model, skills).

### Non-streaming

```bash
curl http://localhost:8080/api/chat \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "prompt": "write a csv parser", "skills": ["filesystem"] }'
# → { "response": "…" }
```

Request body:

| Field | Type | Description |
|-------|------|-------------|
| `prompt` | `string` | The prompt to send to the agent. **Required.** |
| `skills` | `string[]` | Skills to activate for this request. Omit to use all installed. |
| `stream` | `boolean` | If `true`, response is sent as SSE. Default `false`. |

### Streaming (SSE)

```bash
curl -N http://localhost:8080/api/chat \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "prompt": "write a csv parser", "stream": true }'
```

Each SSE event is a `data:` line with the chat completion chunk:

```
data: {"id":"chatcmpl_...","object":"chat.completion.chunk","choices":[{"delta":{"content":"Sure"},"index":0}]}

data: {"id":"chatcmpl_...","object":"chat.completion.chunk","choices":[{"delta":{"content":","},"index":0}]}

data: [DONE]
```

### Chat with Dashboard Environment

```bash
curl http://localhost:8080/api/chat/environments/YOUR_ENV_ID \
  -H 'Content-Type: application/json' \
  -d '{ "prompt": "hello", "api_key": "octra_…", "stream": false }'
```

## MCP Manager

Full CRUD for MCP servers. All endpoints are proxied to Ocawe at `/v1/mcp/*`.

### List Servers

```bash
curl http://localhost:8080/v1/mcp/servers \
  -H "octra-api-token: $OCTRA_API_KEY"
# → { "servers": [ { "id": "…", "transport": "stdio", "command": "npx", "args": ["-y","@modelcontextprotocol/server-filesystem"], "enabled": true, "status": "running", … } ] }
```

### Create Server (stdio)

```bash
curl -X POST http://localhost:8080/v1/mcp/servers \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "transport": "stdio",
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem"],
    "env": { "ALLOWED_DIRS": "/tmp" }
  }'
```

### Create Server (HTTP)

```bash
curl -X POST http://localhost:8080/v1/mcp/servers \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "transport": "http",
    "url": "https://my-mcp-server.example.com",
    "bearer_token": "mcp_secret_token"
  }'
```

Request body for create:

| Field | Type | Description |
|-------|------|-------------|
| `transport` | `"stdio"` \| `"http"` | Server transport type. **Required.** |
| `command` | `string` | (stdio) Binary to execute (e.g. `npx`, `uvx`). |
| `args` | `string[]` | (stdio) CLI arguments. |
| `env` | `object` | (stdio) Environment variables for the process. |
| `url` | `string` | (http) Server URL. |
| `bearer_token` | `string` | (http) Bearer token for auth. |

### Update Server

```bash
curl -X PATCH http://localhost:8080/v1/mcp/servers/SERVER_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "enabled": false }'
```

### Delete Server

```bash
curl -X DELETE http://localhost:8080/v1/mcp/servers/SERVER_ID \
  -H "octra-api-token: $OCTRA_API_KEY"
```

### Reconnect

```bash
curl -X POST http://localhost:8080/v1/mcp/servers/SERVER_ID/reconnect \
  -H "octra-api-token: $OCTRA_API_KEY"
```

### Browse Server Catalog (Tools/Resources/Prompts)

```bash
# List all tools from connected MCP servers
curl http://localhost:8080/v1/mcp/catalog \
  -H "octra-api-token: $OCTRA_API_KEY"
# → { "servers": [...], "tools": [...], "resources": [...], "prompts": [...] }
```

### Search Catalog (Octra-level)

```bash
# Search across providers, CLIs, skills, and MCP templates
curl "http://localhost:8080/api/catalog/search?q=filesystem&category=mcp" \
  -H "octra-api-token: $OCTRA_API_KEY"
# → { "items": [ { "id": "mcp-filesystem", "type": "mcp_server", "name": "Filesystem", "subtitle": "@modelcontextprotocol/server-filesystem", ... } ] }
```

Query parameters:

| Param | Type | Description |
|-------|------|-------------|
| `q` | `string` | Search query. |
| `category` | `string` | Filter: `providers`, `cli`, `skills`, `mcp`, or comma-separated. |
| `limit` | `int` | Max results (1–100, default 20). |

## Skills

Skills are tool packages installed into your Nix environment. They are configured on the canvas and activated per request via the `skills` field in chat.

### Search Skills

```bash
curl "http://localhost:8080/api/catalog/search?q=github&category=skills" \
  -H "octra-api-token: $OCTRA_API_KEY"
```

## CLI

CLIs (opencode, claude-code, codex, etc.) are managed by Ocawe — it runs them as subprocesses and intercepts their LLM requests. Octra provides the catalog.

### List Available CLIs

```bash
curl http://localhost:8080/api/cli \
  -H "octra-api-token: $OCTRA_API_KEY"
```

## Workflow Runs

Declarative workflow execution. All endpoints are proxied to Ocawe at `/v1/workflows/*`.

### List Workflow Runs

```bash
curl "http://localhost:8080/v1/workflows/YOUR_WORKFLOW_ID/runs" \
  -H "octra-api-token: $OCTRA_API_KEY"
# → { "runs": [ { "run_id": "…", "workflow_id": "…", "status": "running", "updated_at": "…" } ] }
```

### Start a Run

```bash
curl -X POST http://localhost:8080/v1/workflows/YOUR_WORKFLOW_ID/runs \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "input_data": { "message": "hello" } }'
# → { "workflow_id": "…", "run_id": "…", "status": "running" }
```

### Get Run Details

```bash
curl http://localhost:8080/v1/workflows/YOUR_WORKFLOW_ID/runs/RUN_ID \
  -H "octra-api-token: $OCTRA_API_KEY"
# → { "workflow_id": "…", "run_id": "…", "status": "completed", "output": { … }, "state": { … }, "node_results": { … }, "updated_at": "…" }
```

### Resume a Suspended Run

```bash
curl -X POST http://localhost:8080/v1/workflows/YOUR_WORKFLOW_ID/runs/RUN_ID/resume \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "user_input": "continue with option A" }'
```

### Cancel a Run

```bash
curl -X POST http://localhost:8080/v1/workflows/YOUR_WORKFLOW_ID/runs/RUN_ID/cancel \
  -H "octra-api-token: $OCTRA_API_KEY"
```

### Restart a Run

```bash
curl -X POST http://localhost:8080/v1/workflows/YOUR_WORKFLOW_ID/runs/RUN_ID/restart \
  -H "octra-api-token: $OCTRA_API_KEY"
```

## HITL (Human-in-the-Loop)

When a workflow reaches a `suspend` node, it pauses and waits for human input. HITL endpoints let you view and respond to suspended runs.

### List Suspended Runs

```bash
curl http://localhost:8080/v1/hitl/runs \
  -H "octra-api-token: $OCTRA_API_KEY"
# → { "runs": [ { "workflow_id": "…", "run_id": "…", "resume_labels": ["approve","reject"], "suspend_payload": { … }, "updated_at": "…" } ] }
```

### Get HITL Run Detail

```bash
curl http://localhost:8080/v1/hitl/runs/WORKFLOW_ID/RUN_ID \
  -H "octra-api-token: $OCTRA_API_KEY"
```

### Resume with User Input

```bash
curl -X POST http://localhost:8080/v1/hitl/runs/WORKFLOW_ID/RUN_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "action": "resume", "resume_data": { "choice": "approve", "reason": "looks good" } }'
```

### Cancel HITL Run

```bash
curl -X POST http://localhost:8080/v1/hitl/runs/WORKFLOW_ID/RUN_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "action": "cancel" }'
```

## Triggers / Webhooks

HTTP endpoints for automating workflows, agents, skills, and functions. All proxied to Ocawe at `/v1/triggers/*`.

### Trigger a Workflow

```bash
curl -X POST http://localhost:8080/v1/triggers/workflows/WORKFLOW_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "input_data": { "key": "value" } }'
# → { "workflow_id": "…", "run_id": "…", "status": "running" }
```

### Trigger an Agent

```bash
curl -X POST http://localhost:8080/v1/triggers/agents/AGENT_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "messages": [{ "role": "user", "content": "Hello!" }] }'
# → { "id": "trg_agent_…", "object": "trigger.agent.response", "agent_id": "…", "output_text": "…" }
```

### Trigger a Skill

```bash
curl -X POST http://localhost:8080/v1/triggers/skills/SKILL_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "input": { "param1": "value" } }'
# → { "id": "trg_skill_…", "object": "trigger.skill.response", "status": "ok" }
```

### Trigger a Function

```bash
curl -X POST http://localhost:8080/v1/triggers/functions/FN_ID \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "input": { "arg1": "value" } }'
# → { "id": "trg_fn_…", "object": "trigger.function.response", "output": { "result": "…" } }
```

## Features

| Feature | Description |
|---------|-------------|
| Streaming SSE | `"stream": true` in chat requests → SSE events |
| MCP Manager | CRUD MCP servers (stdio/HTTP), browse tools/resources/prompts |
| Workflow Engine | Declarative workflow runs with resume/cancel/restart |
| HITL | Human-in-the-Loop — suspend workflows for user input |
| Triggers / Webhooks | HTTP triggers for workflow automation |
| Canvas | Visual workflow builder via @xyflow/react |
| Dashboard | MCP, Runs, HITL, Triggers, Metrics, Environments |
| Catalog | 67 MCP server templates + providers + CLI + skills |

## Architecture

```
octra/
  backend/             — Go monolith (fasthttp + GORM)
    cmd/server/        — entrypoint
    internal/
      api/             — handlers, middlewares, routes, proxies
      config/          — env config
      model/           — GORM models
      service/         — business logic (auth, env, chat, billing)
      repository/      — GORM repositories
      nix/             — Nix environment management
      cli/             — Ocawe launcher + process manager
      storage/         — DB + Redis wiring
  frontend/web/        — Next.js 16 (static export)
    app/
      components/      — WorkflowCanvas, CatalogSearch, EditNodeModal, Metrics
      dashboard/       — MCP, Runs, HITL, Triggers, Metrics, Environments
      server/          — typed API clients (auth, mcp, runs, hitl, catalog...)
      app/page.tsx     — main workspace (canvas + chat + sidebar)
  ocawe/               — Crystal workflow runtime
  docker-compose.yml
  docker-compose.prod.yml
```

## API Reference

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/register` | No | Registration (+100 credits) |
| POST | `/login` | No | JWT login |
| POST | `/api/chat` | `octra-api-token` | Chat with agent, `stream: true` for SSE |
| POST | `/api/chat/environments/{id}` | `api_key` in body | Chat for dashboard environment |
| GET | `/api/catalog/search` | `octra-api-token` | Search providers, CLIs, skills, MCP |
| GET | `/api/cli` | `octra-api-token` | List available CLIs |
| GET | `/api/providers` | `octra-api-token` | List LLM providers |
| GET/POST | `/api/environments` | `octra-api-token` | Dashboard environments CRUD |
| POST | `/api/keys` | `octra-api-token` | Create API key |
| GET | `/api/keys` | `octra-api-token` | List API keys |
| GET | `/api/metrics/requests` | `octra-api-token` | Request metrics |
| ANY | `/v1/mcp/{rest:*}` | `octra-api-token` | MCP servers CRUD → Ocawe |
| ANY | `/v1/workflows/{rest:*}` | `octra-api-token` | Workflow runs → Ocawe |
| ANY | `/v1/hitl/{rest:*}` | `octra-api-token` | HITL → Ocawe |
| ANY | `/v1/triggers/{rest:*}` | `octra-api-token` | Triggers → Ocawe |
| GET | `/billing/balance` | `octra-api-token` | User balance |
| POST | `/billing/topup` | `octra-api-token` | Top up balance |
| POST | `/billing/usage` | `octra-api-token` | Record usage + debit |

All authenticated endpoints use the `octra-api-token` header (except chat/environments which uses `api_key` in the body).

See [backend/README.md](backend/README.md) for all configuration variables.

---

*Octra: your personal MCP endpoint, without the setup.*
