# User Service

Authentication, subscription management, chat history, and custom LLM provider storage. Exposes a REST API consumed by the frontend and API Gateway.

## Endpoints

### Auth
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/register` | Register new user |
| `POST` | `/login` | Login (email + password) |
| `POST` | `/refresh` | Refresh JWT token pair |
| `POST` | `/logout` | Clear session |
| `GET` | `/me` | Get current user info |
| `GET` | `/auth/google` | Google OAuth login |
| `GET` | `/auth/github` | GitHub OAuth login |
| `GET` | `/auth/lefine` | LeFine identity provider login |

### Chat
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/chat/history` | List user chats |
| `POST` | `/chat/create` | Create new chat |
| `GET` | `/chat/:id` | Get chat with messages |
| `POST` | `/chat/:id/messages` | Add message |
| `DELETE` | `/chat/:id` | Delete chat |

### Workflows
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workflows/library` | List public workflows |
| `POST` | `/workflows` | Create workflow |
| `GET` | `/workflows/my` | List user's workflows |

### Subscriptions & Payments
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/plans` | List subscription plans |
| `POST` | `/subscribe` | Subscribe to plan |
| `POST` | `/subscribe/promo` | Activate promo code |
| `POST` | `/payments/create` | Create YooKassa payment |
| `POST` | `/payments/webhook` | YooKassa webhook |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_PORT` | `3112` | HTTP listen port |
| `DB_DNS` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | HMAC secret for access tokens |
| `JWT_REFRESH_SECRET` | — | HMAC secret for refresh tokens |
| `YOOKASSA_SHOP_ID` | `1339826` | Payment gateway shop ID |
| `YOOKASSA_SECRET_KEY` | — | Payment gateway secret |
| `FRONTEND_URL` | — | URL for OAuth redirects |
| `REDIS_URL` | — | Redis connection (optional) |

## Development

```bash
go run cmd/app/main.go
```

## Docker

```bash
docker build -t octra-user -f Dockerfile ..
```
