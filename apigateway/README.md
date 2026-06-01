# API Gateway

HTTP/WebSocket gateway — entry point for all client requests. Routes tasks to the Boss service via gRPC, streams progress back over WebSocket, and handles authentication, rate limiting, and reconnection.

## Architecture

```
Client (Browser/App) ──WebSocket──► API Gateway (3111) ──gRPC──► Nodes/Boss (50051)
                                       │
                                       ├── PostgreSQL (task state)
                                       └── Redis (stream state, pub/sub)
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/task/create` | WebSocket — create task, stream progress |
| `GET` | `/task/reconnect` | WebSocket — resume task after disconnect |
| `GET` | `/task/status` | HTTP — query task status |
| `POST` | `/task/:id/stop` | HTTP — stop a running task |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3111` | HTTP listen port |
| `JWT_SECRET` | — | HMAC secret for JWT validation |
| `BOSS_SERVICE_HOST` | `nodes:50051` | gRPC address of Boss service |
| `REDIS_URL` | — | Redis connection string |
| `REDIS_ENABLED` | `false` | Enable Redis for stream persistence |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `RATE_LIMIT_TASK_CREATE` | `10/60` | Rate limit: max/seconds |
| `RATE_LIMIT_TASK_STATUS` | `60/60` | Rate limit: max/seconds |

## Development

```bash
go run cmd/app/main.go
```

## Docker

```bash
docker build -t octra-apigateway .
```
