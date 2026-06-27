# Octra Backend (monolith)

The Octra backend is a single Go service that implements the
[new concept](../README.md): an **MCP aggregator** that lets users run ready-made
AI CLIs (claude-code, opencode, codex, …) with configured skills on our
infrastructure — no SSH, no VPS, no JSON config.

Each user gets a personal MCP endpoint, authenticated with a per-user API key.

## Stack

| Component   | Technology  |
|-------------|-------------|
| Language    | Go          |
| HTTP        | fasthttp    |
| ORM         | GORM        |
| Database    | PostgreSQL  |
| Cache/state | Redis       |
| Environments| Nix         |

## Layout

```
backend/
  cmd/server/        — entrypoint
  internal/
    api/             — fasthttp handlers, middleware, routes
    config/          — env-based configuration
    model/           — GORM models (User, Agent, Skill, UserSkill)
    repository/      — persistence interfaces + GORM implementations
    service/         — business logic (auth, environment, chat)
    nix/             — per-user Nix environment management
    cli/             — long-lived AI CLI subprocess management (Redis state)
    llm/             — direct LLM proxy (Anthropic Messages API)
    storage/         — DB + Redis wiring
  Dockerfile
  go.mod
```

## Configuration

All settings come from environment variables (with sensible defaults):

| Variable           | Default                              | Description                              |
|--------------------|--------------------------------------|------------------------------------------|
| `HTTP_ADDR`        | `:8080`                              | HTTP listen address                      |
| `DB_DNS`           | `host=localhost user=octra …`        | PostgreSQL DSN                           |
| `REDIS_URL`        | `redis://localhost:6379/0`           | Redis connection URL                     |
| `ENVIRONMENTS_DIR` | `/var/lib/octra/environments`        | Root dir for per-user Nix environments   |
| `CLI_TTL`          | `30m`                                | How long an idle CLI process is kept     |

## API

### `POST /register`
```json
{ "email": "me@example.com", "password": "secret" }
```
→
```json
{ "user_id": "…", "api_key": "octra_…", "balance": 100 }
```

### `POST /environment` (auth: `octra-api-token`)
```json
{
  "llm":   { "api_key": "sk-…", "base_url": "https://api.anthropic.com", "model": "claude-sonnet-4-6" },
  "agent": { "cli": "claude-code", "priority": 10 },
  "skills": ["filesystem", "github", "brave-search"]
}
```
→
```json
{ "user_id": "…", "agent_id": "…", "api_key": "octra_…" }
```

If `cli` is empty Octra runs as a plain LLM proxy.

### `POST /api/chat` (auth: `octra-api-token`)
```json
{ "prompt": "write a csv parser", "skills": ["filesystem"] }
```
→
```json
{ "response": "…" }
```

### Billing endpoints (auth: `octra-api-token`)

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/billing/balance` | Current credits and margin settings |
| `PATCH` | `/billing/settings` | Update `margin_mode`, `safe_margin_limit`, `auto_pay_interval`, `auto_pay_day` |
| `GET` | `/billing/transactions` | Newest-first balance ledger |
| `POST` | `/billing/topup` | Add credits from a payment flow |
| `POST` | `/billing/lefine-reward` | Add credits from LeFine winnings |
| `POST` | `/billing/usage` | Record resource usage and debit hosting credits |

New users receive 100 credits. `margin_mode` defaults to `unlimited`, which lets
hosting charges create debt. Negative balances block starting a new agent. Safe
mode preserves `safe_margin_limit` and suspends the current agent if a charge
cannot be paid.

## Request flow

```
request → validate token → load user + agent
  → if CLI configured:
      → check Redis: is the process alive?
        → alive  → send prompt to stdin
        → dead   → launch subprocess in the user's Nix env
  → else:
      → direct LLM call (proxy mode)
  → return response
```

## Development

```bash
cd backend
go build ./...
go test ./...
```

Tests use an in-memory SQLite database and fakes for Nix/CLI/LLM, so no external
services are required.
