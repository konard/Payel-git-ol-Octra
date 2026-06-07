# OCTRA Custom Client Specification (v2)

## 1. Introduction

This document describes how to build a custom client that can replace or supplement the official React frontend.

The official frontend connects to the API Gateway using WebSocket and supports automatic reconnection with task state restoration.

---

## 2. Connection Endpoints

The frontend uses the following relative paths (proxied through `/api`):

- **Create new task**: `/task/create`
- **Resume existing task**: `/task/reconnect`

**Full production URL examples**:
```
wss://your-domain.com/api/task/create
wss://your-domain.com/api/task/reconnect
```

**Local development** (via Vite proxy):
```
ws://localhost:5173/api/task/create
```

---

## 3. Reconnection & Resume Protocol

### Important discovery from real frontend

When reconnecting, the client does **not** send `{"type": "resume"}`.

Instead, it sends:

```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

This tells the server to restore the task state from Redis and continue streaming updates.

### Reconnection flow

1. Client stores `activeTaskId` locally
2. On disconnect → wait with exponential backoff
3. On reconnect → connect to `/task/reconnect`
4. Immediately send `{ "task_id": "..." }`
5. Server resumes sending `TaskUpdate` messages

---

## 4. Message Format

### Outgoing messages

**Create task** (правильный формат):
```json
{
  "username": "",
  "user_id": "",
  "title": "Need a mini proxy in Go without frontend and tests",
  "description": "Need a mini proxy in Go without frontend and tests",
  "meta": {
    "model": "your-model",
    "provider": "provider"
  },
  "tokens": {
    "provider": "your-api-key",
    "base_url": "https://api.example.com/v1"
  }
}
```

**Resume task**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Incoming messages

The server streams `TaskUpdate` objects with progress, logs, and final results.

---

## 5. Heartbeat

The real frontend starts a heartbeat after connection to keep the connection alive and detect dead connections early.

---

## 6. Recommended Client Features

- Store `taskId` immediately after first update
- Implement exponential backoff reconnection
- Support both new task and resume flows
- Handle authentication token refresh
- Show clear connection status to the user

---

## 7. Language-Specific Implementations

See detailed examples with correct reconnection logic:

- [Python + FastAPI](OCTRA-CLIENT-PYTHON.md)
- [TypeScript](OCTRA-CLIENT-TYPESCRIPT.md)
- [Go](OCTRA-CLIENT-GO.md)
- [Java + Spring](OCTRA-CLIENT-JAVA.md)
- [C# + ASP.NET](OCTRA-CLIENT-CS.md)
- [Ruby](OCTRA-CLIENT-RUBY.md)
- [Crystal](OCTRA-CLIENT-CRYSTAL.md)
- [Rust](OCTRA-CLIENT-RUST.md)

---

This specification is aligned with how the official frontend actually works.