# OCTRA Custom Client Specification (v4)

## 1. Introduction

This document describes how to build a custom client for Octra, the
**MCP aggregator**. Octra exposes a REST API: you register to get an
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

Most authenticated requests carry the user's API key in a header:

```
octra-api-token: octra_…
```

Some endpoints (login, register, /me) use Bearer JWT tokens. The WebSocket
canvas endpoint accepts auth via `token` query param or the `octra-api-token`
header.

Send and receive `application/json` unless noted otherwise (streaming chat uses
`text/event-stream`).

---

## 3. Endpoints

| Method   | Path                              | Auth            | Purpose |
|----------|-----------------------------------|-----------------|---------|
| `GET`    | `/health`                         | no              | Liveness probe |
| `POST`   | `/register`                       | no              | Create a user, return `api_key` |
| `POST`   | `/login`                          | no              | JWT login, returns access + refresh tokens |
| `POST`   | `/logout`                         | no              | Logout (no-op server-side) |
| `GET`    | `/me`                             | Bearer JWT/key  | Get current user profile |
| `POST`   | `/refresh`                        | no              | Refresh JWT tokens |
| `GET`    | `/auth/google`                    | no              | Google OAuth redirect |
| `GET`    | `/auth/google/callback`           | no              | Google OAuth callback |
| `GET`    | `/auth/github`                    | no              | GitHub OAuth redirect |
| `GET`    | `/auth/github/callback`           | no              | GitHub OAuth callback |
| `GET`    | `/auth/lefine`                    | no              | LeFine OAuth redirect |
| `GET`    | `/auth/lefine/callback`           | no              | LeFine OAuth callback |
| `POST`   | `/environment`                    | `octra-api-token` | Create/update the user's environment (LLM + CLI + skills) |
| `POST`   | `/api/chat`                       | `octra-api-token` | Send a prompt, get a response |
| `POST`   | `/api/chat/environments/{id}`     | API key in body | Chat with a specific dashboard environment |
| `POST`   | `/api/chat/openai/environments/{id}` | API key in body | OpenAI-compatible chat with a specific environment |
| `GET`    | `/billing/balance`                | `octra-api-token` | Read credits and margin settings |
| `PATCH`  | `/billing/settings`               | `octra-api-token` | Update margin and auto-pay preferences |
| `GET`    | `/billing/transactions`           | `octra-api-token` | List credit ledger entries |
| `POST`   | `/billing/topup`                  | `octra-api-token` | Add credits from a payment flow |
| `POST`   | `/billing/lefine-reward`          | `octra-api-token` | Add credits from LeFine rewards |
| `POST`   | `/billing/usage`                  | `octra-api-token` | Record usage and debit hosting credits |
| `GET`    | `/api/metrics/requests`           | `octra-api-token` | Request metrics (chat volume, latency, env breakdown) |
| `POST`   | `/api/keys`                       | `octra-api-token` | Create a new user API key |
| `GET`    | `/api/keys`                       | `octra-api-token` | List user API keys |
| `DELETE` | `/api/keys/{id}`                  | `octra-api-token` | Delete a user API key |
| `POST`   | `/api/environments`               | `octra-api-token` | Create a dashboard environment |
| `GET`    | `/api/environments`               | `octra-api-token` | List dashboard environments |
| `POST`   | `/api/environments/patch`         | `octra-api-token` | Update environment (active/visibility) |
| `DELETE` | `/api/environments/{id}`          | `octra-api-token` | Delete a dashboard environment |
| `GET`    | `/api/environments/{id}/canvas`   | `octra-api-token` | Get workflow canvas nodes for an environment |
| `PUT`    | `/api/environments/{id}/canvas`   | `octra-api-token` | Replace workflow canvas nodes |
| `GET`    | `/ws/canvas/{id}`                 | token/header     | WebSocket for real-time canvas sync |
| `GET`    | `/skills/search`                  | no              | Search skills catalogue |
| `GET`    | `/api/cli`                        | no              | List all available CLIs |
| `GET`    | `/api/cli/search`                 | no              | Search CLI catalogue |
| `GET`    | `/api/providers`                  | no              | List all LLM providers |
| `GET`    | `/api/providers/search`           | no              | Search provider catalogue |
| `GET`    | `/api/catalog/search`             | no              | Unified catalogue search (providers + CLIs + skills + MCP) |
| `ANY`    | `/v1/workflows/{rest:*}`          | `octra-api-token` | Proxy to user's Ocawe workflow API |
| `ANY`    | `/v1/mcp/{rest:*}`                | `octra-api-token` | Proxy to user's Ocawe MCP API |
| `ANY`    | `/v1/hitl/{rest:*}`               | `octra-api-token` | Proxy to user's Ocawe HITL API |
| `ANY`    | `/v1/triggers/{rest:*}`           | `octra-api-token` | Proxy to user's Ocawe triggers API |

---

### 3.1 Auth

#### `POST /register`

Request:
```json
{ "username": "me", "email": "me@example.com", "password": "secret" }
```
Response `201`:
```json
{ "user_id": "…", "api_key": "octra_…", "balance": 100 }
```

Errors: `409` if email or username already taken, `400` if validation fails.

#### `POST /login`

Request:
```json
{ "email": "me@example.com", "password": "secret" }
```
Response `200`:
```json
{
  "status": "ok",
  "data": {
    "access_token": "eyJ…",
    "refresh_token": "eyJ…",
    "user": { "id": "…", "username": "me", "email": "me@example.com", "balance": 100, "created_at": "…" }
  }
}
```

- `access_token` expires in 15 minutes.
- `refresh_token` expires in 7 days.

#### `POST /logout`

Response `200`:
```json
{ "status": "ok", "message": "logged out successfully" }
```

#### `GET /me`

Header: `Authorization: Bearer <access_token>` or `Authorization: Bearer <api_key>`

Response `200`:
```json
{
  "status": "ok",
  "data": {
    "user_id": "…",
    "username": "me",
    "email": "me@example.com",
    "api_key": "octra_…",
    "balance": 100,
    "created_at": "…",
    "has_subscription": false,
    "subscription_end": null
  }
}
```

#### `POST /refresh`

Request:
```json
{ "refresh_token": "eyJ…" }
```
Response `200`:
```json
{
  "status": "ok",
  "data": {
    "access_token": "eyJ…",
    "refresh_token": "eyJ…",
    "user": { "id": "…", "username": "me", "email": "me@example.com", "balance": 100, "created_at": "…" }
  }
}
```

#### OAuth

All OAuth flows share the same pattern:
- `GET /auth/{provider}` redirects to the provider's consent page.
- `GET /auth/{provider}/callback` exchanges the authorization code, creates or logs in the user, and redirects to `<FrontendURL>/app?token=<access_token>&refresh_token=<refresh_token>`.

Supported providers: `google`, `github`, `lefine`.

---

### 3.2 `POST /environment`

Request:
```json
{
  "llm":   { "provider": "claude", "api_key": "sk-…", "base_url": "https://api.anthropic.com", "model": "claude-sonnet-4-6" },
  "agent": { "cli": "claude-code", "priority": 10 },
  "skills": ["filesystem", "github", "brave-search"]
}
```

- `llm.provider` is optional. Known values: `claude`, `openai`, `gemini`, `deepseek`, `qwen`, `kimi`, `grok`, `openrouter`, `zed`.
- `llm.api_key`, `llm.base_url`, `llm.model` are all optional.
- `agent.cli` is optional. Supported values include `claude-code`, `opencode`, `codex`, `cursor`, `antigravity`, `cline`, `openhands`, `hermes`, `ocawecore`. If omitted, Octra runs as a plain LLM proxy.
- `agent.priority` is optional (default `100`) and is used by safe margin mode.
- `skills` is optional — a list of skill names to install.

Response `200`:
```json
{ "user_id": "…", "agent_id": "…", "api_key": "octra_…" }
```

Error `402` if balance is negative (unlimited margin) or safe margin would be exceeded.

---

### 3.3 Chat

#### `POST /api/chat`

Request:
```json
{ "prompt": "write a csv parser", "skills": ["filesystem"], "stream": false }
```

- `skills[]` selects which installed skills are active for this request — omit a skill to disable it without uninstalling it.
- `stream` (optional, default `false`): when `true` returns `text/event-stream` SSE.

Response `200` (sync):
```json
{ "response": "…" }
```

Response (streaming, `text/event-stream`):
```
data: {"response":"chunk 1"}
data: {"response":"chunk 2"}
data: {"error":"…"}
```

#### `POST /api/chat/environments/{id}`

Chat with a specific dashboard environment by UUID.

Request:
```json
{ "prompt": "hello", "api_key": "octra_…", "stream": false }
```

- `api_key` is **required** in the body (auth via API key, not header).
- `stream` is optional (default `false`).

Response same as `/api/chat`.

#### `POST /api/chat/openai/environments/{id}`

OpenAI-compatible chat endpoint. Auth via `api_key` in body.

Request (sync):
```json
{ "api_key": "octra_…", "messages": [{"role": "user", "content": "hello"}], "stream": false }
```

Request (streaming):
```json
{ "api_key": "octra_…", "stream": true, "messages": […] }
```

When `stream: true`, the `messages` array is optional (omit for a blank slate).

---

### 3.4 Billing

#### `GET /billing/balance`

Response:
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

#### `PATCH /billing/settings`

All fields optional:
```json
{
  "margin_mode": "safe",
  "safe_margin_limit": 5,
  "auto_pay_interval": "week",
  "auto_pay_day": 2
}
```

- `margin_mode`: `"unlimited"` or `"safe"`.
- `auto_pay_interval`: `"day"`, `"week"`, `"month"`, `"half_year"`, `"year"`.
- `auto_pay_day`: 1–31.

Response: same shape as `GET /billing/balance`.

#### `GET /billing/transactions`

Query params: `limit` (default 100), `offset` (default 0).

Response: array of transaction objects:
```json
[
  {
    "id": "…",
    "user_id": "…",
    "type": "credit",
    "amount": 100,
    "reason": "registration",
    "agent_id": "…",
    "balance_after": 100,
    "created_at": "…"
  }
]
```

- `type`: `"credit"` or `"debit"`.
- `reason`: `"registration"`, `"hosting"`, `"lefine_reward"`, `"topup"`.
- `agent_id` is optional (omitted if not set).

#### `POST /billing/topup` and `POST /billing/lefine-reward`

Request:
```json
{ "amount": 50, "agent_id": "optional-agent-uuid" }
```

- `amount` must be > 0.
- `agent_id` is optional.

Response: a transaction object (same shape as list items above).

#### `POST /billing/usage`

Request:
```json
{
  "date": "2025-01-01T00:00:00Z",
  "cpu_seconds": 120,
  "memory_mb_hours": 256,
  "disk_mb": 512,
  "load_percent": 150,
  "standard_payment": 40,
  "agent_id": "optional-agent-uuid"
}
```

- All fields optional except standard_payment semantics.
- Usage charge is calculated as `standard_payment * load_percent / 100`.
- `date` defaults to now UTC if omitted.

Response `200`:
```json
{
  "usage": {
    "id": "…",
    "user_id": "…",
    "date": "…",
    "cpu_seconds": 120,
    "memory_mb_hours": 256,
    "disk_mb": 512,
    "load_percent": 150
  },
  "transaction": {
    "id": "…",
    "user_id": "…",
    "type": "debit",
    "amount": 60,
    "reason": "hosting",
    "balance_after": 40,
    "created_at": "…"
  }
}
```

- `transaction` is omitted when the charge is zero.

New users receive 100 credits. Unlimited margin allows negative balances; a
negative balance blocks creation of a new environment with HTTP `402`. Safe
margin preserves `safe_margin_limit` and suspends the current agent if a hosting
charge cannot be paid.

---

### 3.5 Request Metrics

#### `GET /api/metrics/requests`

Query params:
- `env` (UUID, optional) — filter to a specific environment.
- `range` (string, optional) — `"24h"`, `"7d"` (default), or `"30d"`.

Response:
```json
{
  "range": "7d",
  "bucket": "day",
  "total": 142,
  "success": 138,
  "failed": 4,
  "avg_latency_ms": 2340,
  "series": [
    { "start": "…", "label": "Mon", "count": 20, "success": 19, "failed": 1 }
  ],
  "environments": [
    { "id": "…", "name": "My Env", "count": 50, "active": true }
  ]
}
```

---

### 3.6 User API Keys

#### `POST /api/keys`

Request:
```json
{ "name": "my-key", "expires_at": "2026-01-01T00:00:00Z" }
```

- `expires_at` is optional.

Response `201`:
```json
{ "id": "…", "name": "my-key", "key": "octra_…", "expires_at": "…", "created_at": "…" }
```

#### `GET /api/keys`

Response: array of key objects (same shape as create response, without `key`).

#### `DELETE /api/keys/{id}`

Response: `{ "status": "deleted" }`.

---

### 3.7 Dashboard Environments

#### `POST /api/environments`

Request:
```json
{ "name": "My Project", "visibility": "private" }
```

- `visibility` is optional: `"private"` (default) or `"public"`.

Response `201`:
```json
{ "id": "…", "name": "My Project", "visibility": "private", "active": true, "created_at": "…" }
```

#### `GET /api/environments`

Response: array of environment objects.

#### `POST /api/environments/patch`

Request:
```json
{ "id": "…", "active": false, "visibility": "public" }
```

- At least one of `active` or `visibility` required.
- `visibility` must be `"private"` or `"public"`.

Response: updated environment object.

#### `DELETE /api/environments/{id}`

Response: `{ "status": "deleted" }`.

---

### 3.8 Workflow Canvas

#### `GET /api/environments/{id}/canvas`

Response: array of canvas node objects:
```json
[
  {
    "id": "…",
    "item_id": "claude",
    "kind": "provider",
    "name": "Claude",
    "detail": "",
    "description": "",
    "meta": {},
    "position_x": 100,
    "position_y": 200,
    "sort_order": 0,
    "created_at": "…",
    "updated_at": "…"
  }
]
```

- `kind`: `"provider"`, `"custom_provider"`, `"cli"`, or any string.
- `meta` is a JSON-encoded map.

#### `PUT /api/environments/{id}/canvas`

Request:
```json
{
  "nodes": [
    {
      "item_id": "claude",
      "kind": "provider",
      "name": "Claude",
      "detail": "",
      "description": "",
      "meta": {},
      "position_x": 100,
      "position_y": 200,
      "sort_order": 0
    }
  ]
}
```

Response: `{ "status": "saved" }`.

#### `GET /ws/canvas/{id}` (WebSocket)

Auth via `?token=octra_…` query param or `octra-api-token` header.

Server sends init message:
```json
{ "type": "init", "nodes": […] }
```

Client can send:
```json
{ "type": "save", "nodes": […] }
```

Server responds:
```json
{ "type": "saved" }
```
or
```json
{ "type": "error", "error": "…" }
```

---

### 3.9 Search / Catalogue

All search endpoints support `?q=<query>&limit=<1-100>` (limit defaults to 20).

#### `GET /skills/search`

Response:
```json
{
  "query": "filesystem",
  "skills": [
    { "id": "…", "skill_id": "…", "name": "Filesystem", "source": "npm", "install_cmd": "npx …" }
  ],
  "count": 1
}
```

#### `GET /api/cli`

Returns array:
```json
[
  { "id": "…", "name": "opencode", "nix_attr": "", "install_cmd": "curl -fsSL https://opencode.ai/install | bash" }
]
```

#### `GET /api/cli/search`

Response:
```json
{ "query": "code", "clis": […], "count": 1 }
```

#### `GET /api/providers`

Returns array:
```json
[
  { "id": "…", "key": "claude", "name": "Claude", "base_url": "https://api.anthropic.com", "auth_env": "ANTHROPIC_API_KEY", "default_model": "claude-sonnet-4-6", "description": "" }
]
```

#### `GET /api/providers/search`

Response:
```json
{ "query": "claude", "providers": […], "count": 1 }
```

#### `GET /api/catalog/search`

Params: `?q=<query>&category=<category>&limit=<1-100>`

- `category`: `"all"` (default), `"providers"`, `"cli"`, `"skills"`, `"custom"`, `"mcp"`.

Response:
```json
{
  "query": "search",
  "category": "all",
  "items": [
    { "id": "…", "type": "provider", "name": "…", "subtitle": "…", … },
    { "id": "…", "type": "cli", "name": "…", … },
    { "id": "…", "type": "skill", "name": "…", … },
    { "id": "custom-provider", "type": "custom_provider", "name": "Custom provider", … },
    { "id": "mcp-filesystem", "type": "mcp_server", "name": "Filesystem", … }
  ],
  "count": 5
}
```

---

### 3.10 Proxy Endpoints

The following wildcard routes forward requests to the user's Ocawe process:

- `ANY /v1/workflows/{rest:*}` — Ocawe workflow API.
- `ANY /v1/mcp/{rest:*}` — Ocawe MCP API.
- `ANY /v1/hitl/{rest:*}` — Ocawe Human-In-The-Loop API.
- `ANY /v1/triggers/{rest:*}` — Ocawe triggers API.

`404` if the proxy is not configured (no Ocawe running).
`502` if Ocawe is unavailable.

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
| `400`  | Malformed request body or validation failure |
| `401`  | Missing/invalid `octra-api-token` or JWT |
| `402`  | Balance or safe margin prevents starting/charging an agent |
| `403`  | Not your resource (environment ownership) |
| `404`  | Resource not found |
| `409`  | Email or username already registered |
| `502`  | Upstream LLM/Ocawe failure |
| `503`  | Metrics service unavailable |
| `5xx`  | Provisioning / internal failure |

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
