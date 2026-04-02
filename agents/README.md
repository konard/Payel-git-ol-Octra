# Agents Service

Единый интерфейс для работы с различными AI провайдерами через gRPC.

## Поддерживаемые провайдеры

- ✅ **Claude** (Anthropic)
- ✅ **Gemini** (Google)
- ✅ **OpenAI**
- ✅ **DeepSeek**
- ✅ **Grok** (xAI)

## Установка

### 1. Генерация Proto файлов

```powershell
cd agents
.\generate-proto.ps1
```

Это создаст:
- `pb/agents.pb.go` - Proto structures
- `pb/agents_grpc.pb.go` - gRPC service code

### 2. Запуск сервиса локально

```powershell
cd agents
go run .\cmd\app\main.go
```

Expected output:
```
✅ Available AI providers: [claude gemini openai deepseek grok]
Starting Agents gRPC server on port 50053...
✅ Agents gRPC server started on port 50053
```

## Использование

### Архитектура потока данных

```
User Request (WebSocket)
    ↓
API Gateway (3111)
    ↓
Boss Service (50051) ──gRPC──┐
                              ├──→ Agents Service (50053)
Manager Service (50052) ──gRPC┤
                              └─→ AI Providers (Claude, OpenAi, etc.)
Worker Service (internal) ──gRPC┘
```

### Ключевой момент: API ключи в запросе

**API ключи НЕ загружаются из environment переменных**. Вместо этого:

1. Пользователь отправляет запрос с API ключями в поле `tokens`:

```json
{
  "title": "Build API",
  "description": "Create REST API with authentication",
  "tokens": {
    "claude": "sk-ant-...",
    "openai": "sk-...",
    "gemini": "AIza...",
    "deepseek": "sk-...",
    "grok": "sk-..."
  },
  "meta": {
    "provider": "claude",
    "model": "claude-opus-4-6"
  }
}
```

2. Boss отправляет GenerateRequest в agents gRPC:

```go
agentClient, _ := grpc.NewAgentClient("agents:50053")
result, _ := agentClient.Generate(
    ctx,
    "claude",              // provider
    "claude-opus-4-6",     // model
    "Your prompt",         // prompt
    tokens,                // tokens from request (contains API key)
    4096,                  // maxTokens
    0.7,                   // temperature
)
```

3. Agents сервис использует API ключ из tokens для вызова провайдера

## Использование из Boss/Manager/Worker

### Boss Service

```go
import "agents/pkg/fetcher/grpc"

// Create client wrapper
agentsClient, _ := NewAgentClientWrapper()
defer agentsClient.Close()

// Generate content
result, _ := agentsClient.GenerateFromTask(
    ctx,
    "claude",           // provider
    "claude-opus-4-6",  // model
    "Your prompt",      // prompt
    req.Tokens,         // tokens from user request
)
```

### Manager Service

```go
agentClient, _ := grpc.NewAgentClient("agents:50053")
content, _ := agentClient.Generate(
    ctx,
    "openai",
    "gpt-4-mini",
    prompt,
    tokens,  // от пользовательского запроса
    2048,
    0.7,
)
```

## Docker Deployment

### Build & Run

```bash
# From project root
docker-compose up -d --build

# Logs
docker-compose logs -f agents
```

### Services Ports

```
API Gateway:    3111   (WebSocket)
Boss gRPC:      50051
Manager gRPC:   50052
Agents gRPC:    50053
Postgres:       5432
```

## Troubleshooting

### Proto compilation fails

```powershell
# Verify protoc installed
protoc --version

# Verify plugins installed
Get-Command protoc-gen-go
Get-Command protoc-gen-go-grpc

# If missing, install:
go install github.com/golang/protobuf/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Connection refused

- Ensure agents service is running: `docker-compose logs agents`
- Ensure other services can reach agents on `agents:50053`
- Check `AGENTS_SERVICE_HOST` environment variable

### Invalid API Key

- Verify API key is correct and has permissions
- Check model name is available for API tier
- Review provider's rate limits and quotas

## Architecture Details

### Provider Interface

All providers implement:

```go
type AIProvider interface {
    Generate(ctx context.Context, req *AgentRequest) (*AgentResponse, error)
    Name() string
    IsConfigured() bool
}
```

### Model Selection Priority

1. User-provided model from `req.Meta["model"]` ← **Highest priority**
2. Environment variable `{PROVIDER}_MODEL`
3. Provider's hardcoded default ← **Lowest priority**

### Request Flow Example

```
WebSocket /task/create
    ↓
APIGateway extracts provider/model from meta
    ↓
Boss creates GenerateRequest with:
  - provider from meta
  - model from meta
  - prompt (task description)
  - tokens (from user)
    ↓
Boss calls agents gRPC Generate()
    ↓
Agents routes to Claude/OpenAI/etc.
    ↓
Provider uses token from request to call API
    ↓
Return response back to Boss
```

## Performance

- **Max message size**: 100 MB (configurable)
- **Concurrent requests**: Limited by provider API rate limits
- **Default timeout**: Standard gRPC timeout

## Monitoring

All calls are logged with:
- Provider name
- Model used
- Request timestamp
- Response time
- Errors

Example logs:
```
🤖 Calling agents service (provider=claude, model=claude-opus-4-6)
✅ Connection to Agents service at agents:50053
```

## License

Part of CrewAI microservices architecture
