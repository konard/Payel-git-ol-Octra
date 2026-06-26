# Octra — the MCP aggregator

Octra lets anyone run ready-made AI CLIs (claude-code, opencode, codex, …) with
configured skills **on our infrastructure** — no VPS, no SSH, no JSON config, no
fighting with ports, processes, or systemd.

We don't build models and we don't write our own agents. We give every user a
**personal MCP endpoint** they can call over HTTP, backed by an isolated [Nix](https://nixos.org)
environment with their chosen CLI and skills already installed.

## Why

Setting up an MCP server by hand normally means renting a VPS, SSH-ing in,
installing dependencies, writing JSON configs, and understanding ports,
processes, and systemd. Most "vibe coders" and AI users never cross that
threshold — they give up or lose days. Octra removes the whole setup.

**Who it's for:** vibe coders who want AI with tools without the hassle,
freelancers who need a 24/7 AI helper, and teams who want to share a single
agent.

## How it works

```
1. Register                → get a personal api_key
2. POST /environment       → Octra builds your Nix env, installs the CLI + skills
3. POST /api/chat          → talk to your agent through your MCP endpoint
```

If you configure a CLI, Octra runs it as a long-lived subprocess inside your Nix
environment and pipes prompts to it. If you don't, Octra works as a plain LLM
proxy with your skills as tools.

### 1. Register

```bash
curl http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{ "email": "me@example.com", "password": "secret" }'
# → { "user_id": "…", "api_key": "octra_…", "balance": 100 }
```

### 2. Create your environment

```bash
curl http://localhost:8080/environment \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "llm":   { "api_key": "sk-…", "base_url": "https://api.anthropic.com", "model": "claude-sonnet-4-6" },
    "agent": { "cli": "claude-code", "priority": 10 },
    "skills": ["filesystem", "github", "brave-search"]
  }'
# → { "user_id": "…", "agent_id": "…", "api_key": "octra_…" }
```

Octra creates the user's Nix environment, installs the CLI (if given), then
installs and configures each skill. Leave `cli` empty to run as a pure LLM proxy.

### 3. Chat

```bash
curl http://localhost:8080/api/chat \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "prompt": "write a csv parser", "skills": ["filesystem"] }'
# → { "response": "…" }
```

Pass only the skills you want active in this request — omit a skill to disable it
without uninstalling it.

### 4. Billing

Every new user receives 100 credits. One credit is worth 10 cents. Octra charges
for active hosted environments by recording usage and debiting credits:

```bash
curl http://localhost:8080/billing/usage \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "cpu_seconds": 120, "memory_mb_hours": 256, "disk_mb": 512, "load_percent": 150, "standard_payment": 40 }'
# → { "usage": { … }, "transaction": { "amount": 60, "reason": "hosting", … } }
```

Top-ups and LeFine rewards credit the same balance ledger:

```bash
curl http://localhost:8080/billing/topup \
  -H "octra-api-token: $OCTRA_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ "amount": 50 }'
```

By default users run in unlimited margin mode, so hosting charges may make the
balance negative. A negative balance blocks creation of new agents, but does not
delete an existing environment. Safe margin mode preserves `safe_margin_limit`
and suspends the current agent when a charge cannot be paid.

## Request flow

```
request → validate token → load user + agent
  → if CLI configured:
      → check Redis: is the process alive?
        → alive → send prompt to its stdin
        → dead  → launch subprocess in the user's Nix env (state saved to Redis)
  → else:
      → direct LLM call (proxy mode)
  → return response
```

### CLI process management

- The CLI runs as a subprocess inside the user's Nix environment.
- The process is **not killed after each request** — it is reused.
- State lives in Redis: `user:{id}:cli_state` (PID, port, start time) and
  `user:{id}:cli_ttl` (TTL). On each request Octra checks liveness; a dead or
  expired process is relaunched, and idle processes are reaped once the TTL
  elapses.
- Default transport is an stdin/stdout pipe (fast, native for AI CLIs).

### Skills

Skills are ready-made packages from nixpkgs or custom configurations. Users pick
them on the canvas (frontend); Octra installs them into the user's Nix
environment and records how each one is installed in the database. Per request,
the `skills[]` field selects which installed skills are active.

## Architecture

Octra is a **single Go backend** — the former `user`, `agents`, `orchestrator`,
and `apigateway` microservices (and the boss/manager/worker agent hierarchy) are
gone. The `frontend`, `poster`, and `tgbot` are unchanged.

| Component    | Technology |
|--------------|------------|
| Language     | Go         |
| HTTP         | fasthttp   |
| ORM          | GORM       |
| Database     | PostgreSQL |
| Cache/state  | Redis      |
| Environments | Nix        |

```
octra/
  backend/             — the monolith (see backend/README.md)
    cmd/server/        — entrypoint
    internal/
      api/             — fasthttp handlers, middleware, routes
      config/          — env-based configuration
      model/           — GORM models (User, Agent, Skill, UserSkill)
      service/         — business logic (auth, environment, chat)
      repository/      — persistence interfaces + GORM implementations
      nix/             — per-user Nix environment management
      cli/             — long-lived AI CLI subprocess management (Redis state)
      llm/             — direct LLM proxy (Anthropic Messages API)
      storage/         — DB + Redis wiring
  frontend/            — web canvas + chat UI (unchanged)
  poster/              — unchanged
  tgbot/               — unchanged
  docker-compose.yml
  docker-compose.prod.yml
```

### Data models

| Model         | Fields |
|---------------|--------|
| `User`        | ID, email, password hash, api_key, subscription, balance, margin settings |
| `Agent`       | ID, user ID, LLM config (api_key, base_url, model), CLI (optional), active, priority |
| `Skill`       | ID, name, type (`built-in` / `nixpkgs` / `custom`), install command, description |
| `UserSkill`   | user ID, agent ID, skill ID, install status |
| `Transaction` | user ID, type, amount, reason, optional agent ID, balance after, created_at |
| `UsageMetric` | user ID, date, CPU seconds, memory MB-hours, disk MB, load percent |

See [`CONCEPT_USER_PAYMENT.md`](CONCEPT_USER_PAYMENT.md) and
[`NEW_CONCEPT_USER_BALANCE.md`](NEW_CONCEPT_USER_BALANCE.md) for the payment and
balance concept details.

## Quick start

```bash
docker compose up -d --build
```

This starts PostgreSQL, Redis, and the Octra backend (plus the frontend and
poster). The API listens on `:8080`. See [`backend/README.md`](backend/README.md)
for configuration variables and local development.

```bash
cd backend
go build ./...
go test ./...   # in-memory SQLite + fakes, no external services needed
```

---

*Octra: your personal MCP endpoint, without the setup.*
