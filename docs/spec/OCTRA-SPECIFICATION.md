# OCTRA Custom Client Specification (v3)

## 1. Introduction

This document describes how to build a custom client for Octra, the
**MCP aggregator**. Octra exposes a small HTTP/JSON API: you register to get an
`api_key`, create an environment (a Nix profile with an AI CLI and skills), and
then send prompts to your personal endpoint.

> The previous WebSocket task-streaming protocol (boss/manager/worker pipeline)
> has been removed along with the old microservices. Octra is now a single
> backend with a plain REST API.

---

## 2. Base URL & authentication

```
http://localhost:8080          # local
https://octra.env.pm           # production example
```

All authenticated requests carry the user's API key in a header:

```
octra-api-token: octra_…
```

Send and receive `application/json`.

---

## 3. Endpoints

| Method | Path           | Auth | Purpose |
|--------|----------------|------|---------|
| `GET`  | `/health`      | no   | Liveness probe |
| `POST` | `/register`    | no   | Create a user, return `api_key` |
| `POST` | `/environment` | yes  | Create/update the user's environment (CLI + skills) |
| `POST` | `/api/chat`    | yes  | Send a prompt, get a response |
| `GET`  | `/billing/balance` | yes | Read credits and margin settings |
| `PATCH` | `/billing/settings` | yes | Update margin and auto-pay preferences |
| `GET`  | `/billing/transactions` | yes | List credit ledger entries |
| `POST` | `/billing/topup` | yes | Add credits from a payment flow |
| `POST` | `/billing/lefine-reward` | yes | Add credits from LeFine rewards |
| `POST` | `/billing/usage` | yes | Record usage and debit hosting credits |

### 3.1 `POST /register`

Request:
```json
{ "email": "me@example.com", "password": "secret" }
```
Response `201`:
```json
{ "user_id": "…", "api_key": "octra_…", "balance": 100 }
```

### 3.2 `POST /environment`

Request:
```json
{
  "llm":   { "api_key": "sk-…", "base_url": "https://api.anthropic.com", "model": "claude-sonnet-4-6" },
  "agent": { "cli": "claude-code", "priority": 10 },
  "skills": ["filesystem", "github", "brave-search"]
}
```
Response `200`:
```json
{ "user_id": "…", "agent_id": "…", "api_key": "octra_…" }
```

- `agent.cli` is optional. Supported values include `claude-code`, `opencode`,
  `codex`. If omitted, Octra runs as a plain LLM proxy.
- `agent.priority` is optional and is used by safe margin mode.
- Octra creates the user's Nix environment, installs the CLI (if any), then
  installs each skill and records its install status.

### 3.3 `POST /api/chat`

Request:
```json
{ "prompt": "write a csv parser", "skills": ["filesystem"] }
```
Response `200`:
```json
{ "response": "…" }
```

`skills[]` selects which installed skills are active for this request — omit a
skill to disable it without uninstalling it.

---

### 3.4 Billing

`GET /billing/balance` response:
```json
{
  "user_id": "…",
  "balance": 100,
  "margin_mode": "unlimited",
  "safe_margin_limit": 0,
  "auto_pay_interval": "month",
  "auto_pay_day": 1
}
```

`PATCH /billing/settings` request:
```json
{
  "margin_mode": "safe",
  "safe_margin_limit": 5,
  "auto_pay_interval": "week",
  "auto_pay_day": 2
}
```

`POST /billing/topup` and `POST /billing/lefine-reward` request:
```json
{ "amount": 50, "agent_id": "optional-agent-uuid" }
```

`POST /billing/usage` request:
```json
{
  "cpu_seconds": 120,
  "memory_mb_hours": 256,
  "disk_mb": 512,
  "load_percent": 150,
  "standard_payment": 40,
  "agent_id": "optional-agent-uuid"
}
```

Usage charges are calculated as `standard_payment * load_percent / 100`. A
successful top-up, reward, or usage debit returns a transaction with `type`,
`amount`, `reason`, `balance_after`, and `created_at`.

New users receive 100 credits. Unlimited margin allows negative balances; a
negative balance blocks creation of a new environment with HTTP `402`. Safe
margin preserves `safe_margin_limit` and suspends the current agent if a hosting
charge cannot be paid.

---

## 4. Request flow (server side)

```
request → validate octra-api-token → load user + agent
  → if CLI configured:
      → check Redis: is the process alive?
        → alive → send prompt to its stdin
        → dead  → launch subprocess in the user's Nix env
  → else:
      → direct LLM call (proxy mode)
  → return { "response": … }
```

The CLI subprocess is long-lived (reused across requests) with its state in
Redis and a TTL; idle processes are reaped.

---

## 5. Errors

Errors are returned as JSON with the relevant HTTP status:

```json
{ "error": "invalid token" }
```

| Status | Meaning |
|--------|---------|
| `400`  | Malformed request body |
| `401`  | Missing/invalid `octra-api-token` |
| `409`  | Email already registered |
| `402`  | Balance or safe margin prevents starting/charging an agent |
| `5xx`  | Provisioning / upstream LLM failure |

---

## 6. Language-specific clients

Minimal HTTP examples in several languages:

- [Python](OCTRA-CLIENT-PYTHON.md)
- [TypeScript / Node.js](OCTRA-CLIENT-TYPESCRIPT.md)
- [Go](OCTRA-CLIENT-GO.md)
- [Java](OCTRA-CLIENT-JAVA.md)
- [C#](OCTRA-CLIENT-CS.md)
- [Ruby](OCTRA-CLIENT-RUBY.md)
- [Crystal](OCTRA-CLIENT-CRYSTAL.md)
- [Rust](OCTRA-CLIENT-RUST.md)
