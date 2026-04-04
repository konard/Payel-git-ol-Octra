# CrewAI — Подробное описание сервисов

## 📖 Оглавление

1. [Agents Service](#1-agents-service)
2. [Boss Service](#2-boss-service)
3. [Manager Service](#3-manager-service)
4. [Worker Service](#4-worker-service)
5. [Apigateway](#5-apigateway)

---

## 1. Agents Service

**Каталог:** `agents/`
**Порт:** 50053 (gRPC)
**Роль:** Централизованный маршрутизатор запросов к LLM-провайдерам
**Язык:** Go 1.25

### 1.1. Назначение

Agents Service — это **единая точка входа** ко всем ИИ-провайдерам в системе. Вместо того чтобы каждый сервис (Boss, Manager, Worker) самостоятельно подключался к OpenAI, Gemini и другим, они все обращаются к Agents через gRPC.

**Преимущества такого подхода:**
- Все провайдеры в одном месте — легко добавить новый
- Централизованная retry-логика
- Per-request токены — API-ключи не хранятся на сервере
- Единый интерфейс независимо от провайдера
- Возможность вести логирование и мониторинг всех запросов к LLM

### 1.2. Структура файлов

```
agents/
├── cmd/app/main.go                          # Точка входа
├── internal/
│   ├── service/agent_service.go             # Роутер запросов к провайдерам
│   └── fetcher/
│       ├── grpc/
│       │   ├── server.go                    # gRPC сервер (Generate, GenerateStream)
│       │   └── client.go                    # gRPC клиент (для других сервисов)
│       └── providers/
│           ├── openrouter/openrouter.go     # OpenRouter (через OpenAI SDK)
│           ├── gemini/gemini.go             # Google Gemini SDK
│           ├── openai/openai.go             # OpenAI SDK
│           ├── claude/claude.go             # Anthropic Claude SDK
│           ├── deepseek/deepseek.go         # DeepSeek SDK
│           └── grok/grok.go                 # xAI Grok SDK
├── pkg/
│   ├── fetcher/grpc/
│   │   ├── server.go                        # Дубликат (используется в main.go)
│   │   └── client.go                        # Клиентская библиотека
│   └── models/agent.go                      # Интерфейсы и модели
├── proto/agents.proto                       # Proto-контракт
└── pb/
    ├── agents.pb.go                         # Сгенерированный код
    └── agents_grpc.pb.go                    # Сгенерированный gRPC код
```

### 1.3. Точка входа (`cmd/app/main.go`)

```go
func main() {
    agentService := service.NewAgentService()

    // Все провайдеры инициализируются с пустыми API-ключами
    // Ключи будут переданы per-request через tokens поле
    providers := map[string]*models.ProviderConfig{
        "claude":     {Model: "claude-opus-4-6"},
        "gemini":     {Model: "gemini-2.5-flash"},
        "openai":     {Model: "gpt-4o"},
        "deepseek":   {Model: "deepseek-chat"},
        "grok":       {Model: "grok-3"},
        "openrouter": {Model: "qwen/qwen3.6-plus:free"},
    }

    // Регистрация всех провайдеров
    agentService.RegisterProvider("claude", claude.New(providers["claude"]))
    agentService.RegisterProvider("gemini", gemini.New(providers["gemini"]))
    // ... и так далее

    // Запуск gRPC сервера
    grpc.Start(os.Getenv("AGENTS_PORT"), agentService)
}
```

### 1.4. Интерфейс провайдера (`pkg/models/agent.go`)

```go
type AIProvider interface {
    // Generate генерирует ответ по промпту
    Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error)

    // Name возвращает имя провайдера ("openai", "gemini", ...)
    Name() string

    // IsConfigured — всегда true (ключи per-request)
    IsConfigured() bool
}

type AgentRequest struct {
    Provider    string  // "openrouter", "gemini", "openai", ...
    Model       string  // "qwen/qwen3-coder", "gpt-4o", ...
    Prompt      string
    Tokens      map[string]interface{}  // {"openrouter": "sk-or-v1-..."}
    MaxTokens   int
    Temperature float32
}
```

### 1.5. Паттерн lazy client creation

Все провайдеры создают HTTP-клиент **лениво** — только при первом вызове `Generate`. Это позволяет:

1. Зарегистрировать все провайдеры при старте без ключей
2. Не создавать HTTP-клиенты для неиспользуемых провайдеров
3. Переиспользовать один клиент для всех запросов

```go
// OpenAI провайдер
type OpenAIProvider struct {
    mu     sync.Mutex
    client *openai.Client      // nil до первого запроса
    config *models.ProviderConfig
}

func (p *OpenAIProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if p.client != nil {
        return p.client, nil  // Уже создан — возвращаем
    }

    // Ищем API-ключ в токенах запроса
    apiKey := extractAPIKey(tokens)  // ищет: "openai", "api_key", "apiKey"
    if apiKey == "" {
        return nil, fmt.Errorf("OpenAI API key not found in tokens")
    }

    p.client = openai.NewClient(option.WithAPIKey(apiKey))
    return p.client, nil
}
```

### 1.6. Retry-логика (OpenRouter)

OpenRouter провайдер — единственный с явной retry-логикой, т.к. бесплатные модели нестабильны:

```go
func (p *OpenRouterProvider) Generate(ctx, prompt, tokens) (string, error) {
    params := openai.ChatCompletionNewParams{...}

    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        if attempt > 0 {
            delay := time.Duration(attempt) * 2 * time.Second  // 0s, 2s, 4s
            time.Sleep(delay)
        }

        response, err := client.Chat.Completions.New(ctx, params)
        if err == nil {
            return response.Choices[0].Message.Content, nil
        }

        // Retry только на transient-ошибки
        if isTransient(err) {  // EOF, connection reset, timeout
            continue
        }
        return "", err  // Auth error, rate limit — сразу фейл
    }
    return "", lastErr
}
```

### 1.7. gRPC API (`proto/agents.proto`)

```protobuf
service AgentService {
  // Унарный запрос — полный ответ после завершения
  rpc Generate(GenerateRequest) returns (GenerateResponse) {}

  // Потоковый запрос — ответ по частям
  rpc GenerateStream(GenerateRequest) returns (stream GenerateStreamChunk) {}
}

message GenerateRequest {
  string provider = 1;           // "openrouter", "gemini", "openai", "claude", "deepseek", "grok"
  string model = 2;              // "qwen/qwen3-coder", "gpt-4o-mini", ...
  string prompt = 3;             // Текст промпта
  map<string, string> tokens = 4; // {"openrouter": "sk-or-v1-..."}
  int32 max_tokens = 5;          // Максимум токенов в ответе
  float temperature = 6;         // Креативность (0.0 - 1.0)
}

message GenerateResponse {
  string provider = 1;
  string model = 2;
  string content = 3;
  int32 tokens_used = 4;
  string error = 5;
  string error_code = 6;
}

message GenerateStreamChunk {
  string content = 1;      // Часть контента
  bool done = 2;           // true = генерация завершена
  string error = 3;        // Ошибка если возникла
  int32 tokens_used = 4;   // Токенов использовано
}
```

### 1.8. Поддерживаемые провайдеры

| Провайдер | Пакет | SDK | Модель по умолчанию | Извлечение ключа |
|-----------|-------|-----|---------------------|------------------|
| **OpenRouter** | `openrouter/` | `openai-go` + `BaseURL` | `qwen/qwen3.6-plus:free` | `"openrouter"`, `"api_key"`, `"apiKey"` |
| **Gemini** | `gemini/` | `google/generative-ai-go` | `gemini-2.5-flash` | `"gemini"` |
| **OpenAI** | `openai/` | `openai-go/v3` | `gpt-4o` | `"openai"`, `"api_key"`, `"apiKey"` |
| **Claude** | `claude/` | `anthropic-sdk-go` | `claude-opus-4-6` | `"claude"`, `"anthropic"`, `"api_key"` |
| **DeepSeek** | `deepseek/` | `deepseek-go` | `deepseek-chat` | `"deepseek"`, `"api_key"` |
| **Grok** | `grok/` | `grok-go` | `grok-3` | `"grok"`, `"xai"`, `"api_key"` |

### 1.9. Как добавить новый провайдер

1. Создать пакет: `internal/fetcher/providers/newprovider/newprovider.go`
2. Реализовать интерфейс `AIProvider`:

```go
package newprovider

type NewProvider struct {
    mu     sync.Mutex
    client *SomeClient
    config *models.ProviderConfig
}

func New(config *models.ProviderConfig) (*NewProvider, error) {
    return &NewProvider{config: config}, nil
}

func (p *NewProvider) getClient(tokens map[string]interface{}) (*SomeClient, error) {
    // Lazy client creation
}

func (p *NewProvider) Generate(ctx, prompt, tokens) (string, error) {
    client := p.getClient(tokens)
    // ... вызвать API ...
    return content, nil
}

func (p *NewProvider) Name() string { return "newprovider" }
func (p *NewProvider) IsConfigured() bool { return true }
```

3. Зарегистрировать в `cmd/app/main.go`:
```go
agentService.RegisterProvider("newprovider", newprovider.New(providers["newprovider"]))
```

---

## 2. Boss Service

**Каталог:** `boss/`
**Порт:** 50051 (gRPC)
**Роль:** CEO/CTO — принимает задачи, планирует архитектуру, координирует менеджеров
**Язык:** Go 1.25

### 2.1. Назначение

Boss — это **первый сервис** в цепке после Apigateway. Его задачи:

1. **Принять задачу** от пользователя
2. **Проанализировать** через AI (Agents) что нужно сделать
3. **Определить стек технологий** и архитектуру
4. **Назначить менеджеров** с ролями и приоритетами
5. **Дождаться** завершения работы всех менеджеров
6. **Валидировать** итоговый результат через AI
7. **Вернуть** ZIP-архив с готовым проектом

### 2.2. Структура файлов

```
boss/
├── cmd/app/main.go                              # Точка входа
├── internal/
│   ├── fetcher/grpc/
│   │   ├── server.go                            # gRPC сервер
│   │   ├── bosspb/                              # Proto от gateway-boss.proto
│   │   └── manager/
│   │       ├── client.go                        # gRPC клиент к Manager
│   │       └── managerpb/                       # Proto от boss-manager.proto
│   └── service/
│       ├── boss.go                              # Основная бизнес-логика
│       └── agents_client_wrapper.go             # Обёртка к Agents сервису
├── pkg/
│   ├── database/database.go                     # PostgreSQL инициализация
│   ├── llm/client.go                            # LLM клиенты (не используются напрямую)
│   └── models/task.go                           # GORM модели
└── proto/
    ├── gateway-boss.proto                       # Контракт Apigateway → Boss
    └── boss-manager.proto                       # Контракт Boss → Manager
```

### 2.3. Основные методы (`internal/service/boss.go`)

#### CreateTask (Unary RPC)

```go
func (s *BossService) CreateTask(ctx, req) (*BossDecision, error) {
    // 1. Сохранить задачу в БД
    task := &models.Task{
        ID: uuid.New(), Status: "boss_planning", ...
    }

    // 2. Boss думает через AI
    decision := s.thinkAboutTask(ctx, agentsClient, provider, model, req)

    // 3. Сохранить решение
    bossDecision := &models.BossDecision{...}

    // 4. Вызвать менеджеров ПАРАЛЛЕЛЬНО
    managerResults, zipData := s.assignManagersParallel(ctx, task.ID, decision, req)

    // 5. Boss валидирует результат
    validationResult := s.validateSolution(ctx, agentsClient, ..., zipData)

    // 6. Сохранить ZIP и вернуть
    task.Solution = zipData
    task.Status = "done"
    return &BossDecision{Solution: zipData}, nil
}
```

#### CreateTaskStream (Server Stream RPC)

```go
func (s *BossService) CreateTaskStream(req, stream) error {
    // То же что CreateTask, но отправляет обновления по мере выполнения:
    stream.Send(&TaskUpdate{Progress: 5,  Message: "📝 Task received..."})
    stream.Send(&TaskUpdate{Progress: 10, Message: "💾 Task saved..."})
    stream.Send(&TaskUpdate{Progress: 15, Message: "🤖 AI client initialized..."})
    stream.Send(&TaskUpdate{Progress: 30, Message: "✅ Architecture planned..."})
    stream.Send(&TaskUpdate{Progress: 40, Message: "🏗️ Creating N managers..."})
    stream.Send(&TaskUpdate{Progress: 70, Message: "👷 Managers completed..."})
    stream.Send(&TaskUpdate{Progress: 80, Message: "🔍 Boss validating..."})
    stream.Send(&TaskUpdate{Progress: 90, Message: "📦 Packaging..."})
    stream.Send(&TaskUpdate{Progress: 100, Message: "🎉 Project ready!"})
}
```

### 2.4. Логика планирования (`thinkAboutTask`)

Boss отправляет в Agents промпт:

```
You are CTO. Task:

Title: Прокси сервер
Description: Напиши прокси сервис на go

Reply ONLY with JSON:
{
  "managers_count": 3,
  "manager_roles": [
    {"role": "backend", "description": "Core proxy engine...", "priority": 1},
    {"role": "frontend", "description": "Admin dashboard...", "priority": 2},
    {"role": "testing", "description": "Load and security testing...", "priority": 3}
  ],
  "technical_description": "High-performance reverse proxy...",
  "tech_stack": ["Go", "React", "PostgreSQL", "Docker"],
  "architecture_notes": "Modular monolith architecture..."
}
```

Парсинг ответа:
```go
// Извлекает JSON из markdown блоков ```json ... ```
jsonStr := extractJSONFromMarkdown(resp)
var decision BossDecisionResult
json.Unmarshal([]byte(jsonStr), &decision)
```

### 2.5. Параллельный вызов менеджеров

```go
func (s *BossService) assignManagersParallel(ctx, taskID, decision, req) {
    var (
        mu         sync.Mutex
        allResults []*managerpb.ManagerResult
        allZipData [][]byte
        firstErr   error
    )

    var wg sync.WaitGroup
    for i, role := range decision.ManagerRoles {
        wg.Add(1)
        go func(idx int, role models.ManagerRole) {
            defer wg.Done()

            // Менеджеры с высоким приоритетом начинают первыми
            if idx > 0 {
                select {
                case <-time.After(time.Duration(idx) * 5 * time.Second):
                case <-ctx.Done():
                    return
                }
            }

            // Собираем контекст из уже завершённых менеджеров
            mu.Lock()
            var contextResults []*managerpb.WorkerResult
            for _, mr := range allResults {
                contextResults = append(contextResults, mr.WorkerResults...)
            }
            mu.Unlock()

            // Вызываем AssignManager для ОДНОГО менеджера
            result, err := s.managerClient.AssignManager(ctx, &AssignManagerRequest{
                ManagerId:            uuid.New().String(),
                Role:                 role.Role,
                Description:          role.Description,
                TechnicalDescription: decision.TechnicalDescription,
                Priority:             role.Priority,
                Metadata:             metadata,
                OtherWorkersResults:  contextResults,
            })

            mu.Lock()
            allResults = append(allResults, result)
            mu.Unlock()
        }(i, role)
    }
    wg.Wait()

    // Объединяем все ZIP
    finalZip := mergeZipArchives(allZipData)
    return allResults, finalZip, nil
}
```

**Зачем задержка для менеджеров с низким приоритетом?**

Менеджер #1 (backend) начинает сразу. Менеджер #2 (frontend) ждёт 5 секунд, чтобы получить контекст от backend — какие API endpoints будут, какие модели данных. Это помогает frontend-менеджеру создать более согласованный проект.

### 2.6. Финальная валидация

```go
func (s *BossService) validateSolution(ctx, agentsClient, ..., zipData) (*ValidationResult, error) {
    prompt := `You are the CTO reviewing the final deliverable.

ORIGINAL TASK: {title}
ARCHITECTURE: {technical_description}
Stack: {tech_stack}

MANAGERS RESULTS:
{summary of each manager's work}

GENERATED FILES: {count} total
ZIP size: {size} bytes

Review:
1. Does the solution meet the requirements?
2. Is the architecture followed?
3. Are all managers completed their work?

Reply ONLY with JSON:
{"approved": true/false, "feedback": "..."}`

    resp := agentsClient.Generate(ctx, prompt)
    // Parse and return ValidationResult
}
```

### 2.7. Клиент к Manager (`internal/fetcher/grpc/manager/client.go`)

```go
type Client struct {
    conn   *grpc.ClientConn
    client managerpb.ManagerServiceClient
}

// AssignManager — вызывает ОДНОГО менеджера
func (c *Client) AssignManager(ctx, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
    return c.client.AssignManager(ctx, req)
}

// AssignManagersAndWait — legacy, вызывает всех последовательно
func (c *Client) AssignManagersAndWait(ctx, taskID, techDesc string, roles []string, ...) ([]byte, error) {
    return c.client.AssignManagersAndWait(ctx, &AssignManagersRequest{...}).Solution, nil
}
```

### 2.8. Модели данных (`pkg/models/task.go`)

```go
type Task struct {
    ID          uuid.UUID
    UserID      string
    Username    string
    Title       string
    Description string
    Tokens      string  // JSON: {"openrouter": "sk-or-..."}
    Meta        string  // JSON: {"provider": "openrouter", "model": "qwen3-coder"}
    Status      string  // pending → boss_planning → managers_assigned → processing → reviewing → done/error
    Solution    []byte  // ZIP archive
}

type BossDecision struct {
    ID                   uuid.UUID
    TaskID               uuid.UUID
    ManagersCount        int32
    ManagerRoles         string  // JSON: [{role, description, priority}]
    TechnicalDescription string
    TechStack            string  // JSON: ["Go", "React", ...]
    ArchitectureNotes    string
}
```

---

## 3. Manager Service

**Каталог:** `manager/`
**Порт:** 50052 (gRPC)
**Роль:** Team Lead — набирает команду, распределяет задачи, ревьюит работу
**Язык:** Go 1.25

### 3.1. Назначение

Manager — это **посредник** между Boss и Worker. Его задачи:

1. Получить зону ответственности от Boss (frontend, backend, testing)
2. Через AI решить, **каких воркеров нанять** для своей команды
3. Вызвать Worker сервис для генерации кода
4. **Проверить** работу каждого воркера через AI (review-цикл)
5. Если код не прошёл проверку — **отправить на доработку** (ReviewWorker)
6. Собрать все файлы в ZIP и вернуть

### 3.2. Структура файлов

```
manager/
├── cmd/app/main.go                              # Точка входа
├── internal/
│   ├── fetcher/grpc/
│   │   ├── server.go                            # gRPC сервер
│   │   ├── managerpb/                           # Proto от boss-manager.proto
│   │   └── worker/
│   │       └── client.go                        # gRPC клиент к Worker
│   │       └── workerpb/                        # Proto от manager-worker.proto
│   ├── service/manager.go                       # Основная бизнес-логика
│   └── core/services/get_task.go                # Вспомогательные функции
├── pkg/
│   ├── database/database.go                     # PostgreSQL
│   ├── llm/client.go                            # НЕ используется (через Agents)
│   └── models/manager.go                        # GORM модели
└── proto/
    ├── boss-manager.proto                       # Контракт Boss → Manager
    └── manager-worker.proto                     # Контракт Manager → Worker
```

### 3.3. Основные методы (`internal/service/manager.go`)

#### AssignManager (Unary RPC) — новый метод

```go
func (s *ManagerService) AssignManager(ctx, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
    return s.processManager(ctx, req)
}
```

#### AssignManagersAndWait (Unary RPC) — legacy

```go
func (s *ManagerService) AssignManagersAndWait(ctx, req *managerpb.AssignManagersRequest) (*managerpb.AssignManagersResponse, error) {
    // Последовательно вызывает всех менеджеров (обратная совместимость)
    for _, role := range req.Roles {
        result := s.processManager(ctx, mgrReq)
    }
    return &AssignManagersResponse{Solution: finalZip}, nil
}
```

### 3.4. processManager — основная логика

```go
func (s *ManagerService) processManager(ctx, req) (*managerpb.ManagerResult, error) {
    // 1. Создать менеджера в БД
    manager := &models.Manager{ID: uuid.Parse(req.ManagerId), Role: req.Role, Status: "active"}
    database.Db.Create(manager)

    // 2. Manager думает: каких воркеров нанять?
    workerRolesList := s.managerThink(ctx, provider, model, tokens, req.TechnicalDescription, req.Role, req.Description)

    // 3. Вызвать Worker сервис
    workerResp := s.workerClient.AssignWorkersAndWait(ctx, &AssignWorkersRequest{
        TaskId:              req.TaskId,
        ManagerId:           req.ManagerId,
        ManagerRole:         req.Role,
        WorkerRoles:         workerRoles,
        TaskMd:              req.TechnicalDescription,
        OtherWorkersResults: otherResults,  // Контекст от других воркеров
    })

    // 4. REVIEW: Проверить каждого воркера через AI
    for _, wr := range workerResp.WorkerResults {
        review := s.reviewWorkerResult(ctx, provider, model, tokens, req.Role, wr)

        if !review.Approved {
            // Отправить на доработку
            reviewResp := s.workerClient.ReviewWorker(ctx, &ReviewRequest{
                WorkerId:   wr.WorkerId,
                Feedback:   review.Feedback,
                OriginalFiles: wr.Files,
            })
            // Обновить файлы из reviewed версии
        }
    }

    // 5. Собрать все файлы в ZIP
    allFiles := make(map[string]string)
    for _, wr := range workerResp.WorkerResults {
        for path, content := range wr.Files {
            allFiles[path] = content
        }
    }
    finalZip := createZipArchive(allFiles)

    manager.Status = "done"
    database.Db.Save(manager)

    return &managerpb.ManagerResult{Solution: finalZip, WorkerResults: workerResp.WorkerResults}, nil
}
```

### 3.5. managerThink — AI решает каких воркеров нанять

```
You are a project manager for a {role} team ONLY. Your team's focus: {description}

Task: {technical_description}

You are ONLY responsible for the {role} aspect of this project.
What workers do you need on YOUR team?

IMPORTANT: Only list workers that belong to YOUR team ({role}).
- If you are "frontend" manager → only list frontend workers
- If you are "backend" manager → only list backend workers
- If you are "devops/qa" manager → only list devops/qa workers

Do NOT list workers from other teams. Each team has its own manager.

Reply ONLY with JSON:
{"worker_roles": [{"role": "...", "description": "..."}]}
```

Пример ответа для backend-менеджера:
```json
{
  "worker_roles": [
    {"role": "Core Go Systems Engineer", "description": "Architects the non-blocking I/O layer..."},
    {"role": "Routing & Middleware Developer", "description": "Implements dynamic upstream routing..."},
    {"role": "PostgreSQL Database Engineer", "description": "Designs optimized data models..."}
  ]
}
```

### 3.6. reviewWorkerResult — AI проверяет качество кода

```
You are a {managerRole} manager reviewing work from a {workerRole} developer.

TASK: {task_md}
PROPOSED SOLUTION: {solution_md}
FILES:
--- main.go ---
package main
...

Review the work:
1. Does it fulfill the task requirements?
2. Is the code quality acceptable?
3. Are there obvious bugs or missing pieces?
4. Does it integrate well with the team (if context provided)?

Reply ONLY with JSON:
{"approved": true/false, "feedback": "..."}
```

### 3.7. Клиент к Worker (`internal/fetcher/grpc/worker/client.go`)

```go
type Client struct {
    conn   *grpc.ClientConn
    client workerpb.WorkerServiceClient
}

// AssignWorkersAndWait — назначить воркеров и ждать результат
func (c *Client) AssignWorkersAndWait(ctx, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
    return c.client.AssignWorkersAndWait(ctx, req)
}

// ReviewWorker — отправить замечания воркеру
func (c *Client) ReviewWorker(ctx, req *workerpb.ReviewRequest) (*workerpb.ReviewResponse, error) {
    return c.client.ReviewWorker(ctx, req)
}
```

### 3.8. Конвертация типов WorkerResult

Поскольку `WorkerResult` определён и в `managerpb` и в `workerpb`, Manager использует функции конвертации:

```go
func workerResultToManagerpb(wr *workerpb.WorkerResult) *managerpb.WorkerResult {
    return &managerpb.WorkerResult{
        WorkerId: wr.WorkerId, Role: wr.Role, Files: wr.Files, ...
    }
}

func workerResultsFromManagerpb(wrs []*managerpb.WorkerResult) []*workerpb.WorkerResult {
    result := make([]*workerpb.WorkerResult, len(wrs))
    for i, wr := range wrs {
        result[i] = workerResultFromManagerpb(wr)
    }
    return result
}
```

### 3.9. Модели данных (`pkg/models/manager.go`)

```go
type Manager struct {
    ID           uuid.UUID
    TaskID       uuid.UUID
    Role         string  // "frontend", "backend", "testing"
    AgentID      string
    Status       string  // active, completed, error
    TaskDesc     string  // Техническое описание от Boss
    WorkerRoles  string  // JSON: [{role, description}]
    WorkersCount int32
}

type WorkerRole struct {
    Role        string
    Description string
}
```

---

## 4. Worker Service

**Каталог:** `worker/`
**Порт:** 50053 (gRPC)
**Роль:** Developer — генерирует код, создаёт файлы проекта
**Язык:** Go 1.25

### 4.1. Назначение

Worker — это **исполнитель**. Он получает роль от менеджера и генерирует файлы проекта.

**Ключевые отличия от старой версии:**
- ~~Ходил в LLM напрямую через HTTP~~ → **Использует Agents сервис через gRPC**
- ~~Генерировал все файлы в одном JSON~~ → **N+1 подход: список файлов + по одному на файл**
- ~~JSON с кодом (ломался от кавычек)~~ → **Plain text: raw code без JSON**
- ~~Нет обратной связи~~ → **Поддержка ReviewWorker**

### 4.2. Структура файлов

```
worker/
├── cmd/app/main.go                              # Точка входа
├── internal/
│   ├── fetcher/grpc/
│   │   ├── server.go                            # gRPC сервер
│   │   └── workerpb/                            # Proto от manager-worker.proto
│   └── service/worker.go                        # Основная бизнес-логика
├── pkg/
│   ├── database/database.go                     # PostgreSQL
│   └── models/worker.go                         # GORM модели
└── proto/
    └── manager-worker.proto                     # Контракт Manager → Worker
```

### 4.3. Основные методы (`internal/service/worker.go`)

#### AssignWorkersAndWait (Unary RPC)

```go
func (s *WorkerService) AssignWorkersAndWait(ctx, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
    // Парсим метаданные
    provider := metadata["provider"]  // "openrouter", "gemini", ...
    model := metadata["model"]        // "qwen/qwen3-coder", ...
    tokens := parseTokens(metadata["tokens"])

    allFiles := make(map[string]string)
    workerResults := make([]*workerpb.WorkerResult, 0)

    // Для каждой роли воркера — ПОСЛЕДОВАТЕЛЬНО (чтобы видеть результаты предыдущих)
    for _, wr := range req.WorkerRoles {
        // 1. Создать воркера в БД
        worker := &models.Worker{Role: wr.Role, Status: "thinking"}
        database.Db.Create(worker)

        // 2. Собрать контекст от предыдущих воркеров
        accumulatedContext := buildContext(workerResults)

        // 3. Создать TASK.md (через Agents)
        taskMD := s.createTaskMD(ctx, provider, model, tokens, wr.Role, ...)

        // 4. Сгенерировать файлы (через Agents, N+1 подход)
        files := s.generateCode(ctx, provider, model, tokens, taskMD, wr.Role, ...)

        // 5. Добавить TASK.md в файлы
        files[fmt.Sprintf("%s/%s/TASK.md", basePath, wr.Role)] = taskMD

        // 6. Сохранить результат
        workerResults = append(workerResults, &WorkerResult{Files: files})
    }

    // 7. Создать ZIP
    zipData := createZipArchive(allFiles)

    return &workerpb.AssignWorkersResponse{Solution: zipData, WorkerResults: workerResults}, nil
}
```

#### ReviewWorker (Unary RPC)

```go
func (s *WorkerService) ReviewWorker(ctx, req *workerpb.ReviewRequest) (*workerpb.ReviewResponse, error) {
    fixedFiles := make(map[string]string)

    // Исправить каждый файл индивидуально
    for filePath, oldContent := range req.OriginalFiles {
        prompt := fmt.Sprintf(`You are a %s developer. Your previous work was reviewed.

FILE: %s
MANAGER FEEDBACK: %s
PREVIOUS CODE: %s

FIX the code based on the feedback. Return the FULL corrected file as PLAIN TEXT.`,
            req.WorkerRole, filePath, req.Feedback, oldContent)

        fixedContent := s.agentsClient.Generate(ctx, prompt)
        fixedContent = stripMarkdownCodeBlock(fixedContent)
        fixedFiles[filePath] = fixedContent
    }

    return &workerpb.ReviewResponse{Files: fixedFiles, Status: "fixed"}, nil
}
```

### 4.4. N+1 подход к генерации файлов

#### Шаг 1: Планирование (1 запрос к AI)

```go
func (s *WorkerService) generateCode(ctx, provider, model, tokens, taskMD, role, ...) (map[string]string, error) {
    // Определяем какие файлы создать
    planPrompt := fmt.Sprintf(`You are a %s developer. Role: %s

TASK: %s

List ONLY the 3-5 most important files to create. Be specific.
Return JSON ONLY:
{"files": ["path1.ext", "path2.ext", "path3.ext"]}`, role, description, taskMD)

    planResp := s.agentsClient.Generate(ctx, planPrompt)
    // Parse: {"files": ["main.go", "config.go", "README.md"]}
}
```

#### Шаг 2: Генерация каждого файла (N запросов к AI)

```go
// Для КАЖДОГО файла — отдельный запрос
for _, file := range plan.Files {
    contentPrompt := fmt.Sprintf(`Write the FULL content of file: %s

TASK: %s
Role: %s

Write COMPLETE code. No placeholders. No TODOs. No "implement later".
Return the file content as PLAIN TEXT. NO JSON. NO markdown. Just the raw code.`,
        file, taskMD, role)

    content := s.agentsClient.Generate(ctx, contentPrompt)
    content = stripMarkdownCodeBlock(content)
    files[fmt.Sprintf("%s/%s/%s", basePath, role, file)] = content
}
```

**Почему plain text, а не JSON?**

| Подход | Проблема |
|--------|----------|
| JSON `{"file": "code with \"quotes\""}` | Кавычки в коде ломают JSON |
| JSON с `\n` `\t` | LLM путает экранирование |
| JSON обрезан по max_tokens | `json.Unmarshal` падает |
| **Plain text** | ✅ Работает, `stripMarkdownCodeBlock` убирает обёртку |

### 4.5. stripMarkdownCodeBlock

```go
func stripMarkdownCodeBlock(s string) string {
    s = strings.TrimSpace(s)
    // Убираем открывающий блок
    for _, marker := range []string{"```go\n", "```ts\n", "```tsx\n", "```js\n", "```json\n", "```\n", "```"} {
        if strings.HasPrefix(s, marker) {
            s = s[len(marker):]
            break
        }
    }
    // Убираем закрывающий блок
    if end := strings.LastIndex(s, "```"); end > 0 {
        s = strings.TrimSpace(s[:end])
    }
    return s
}
```

### 4.6. createTaskMD

```go
func (s *WorkerService) createTaskMD(ctx, provider, model, tokens, role, description, taskMD, context, basePath) (string, error) {
    prompt := fmt.Sprintf(`You are a %s developer on a project team.

YOUR ROLE: %s
TASK: %s
CONTEXT FROM OTHER WORKERS: %s

Create TASK.md file with detailed task breakdown for your role. Include:
1. Files to create
2. Functionality to implement
3. Dependencies
4. How your work integrates with other workers

Return ONLY the content of TASK.md file.`, role, description, taskMD, context)

    return s.agentsClient.Generate(ctx, prompt)
}
```

### 4.7. Координация между воркерами

Каждый воркер видит результаты предыдущих:

```go
// Build accumulated context
accumulatedContext := ""
for _, prevResult := range workerResults {
    accumulatedContext += fmt.Sprintf("\n=== Previous worker %s (%s) ===\n", prevResult.Role)
    for path, content := range prevResult.Files {
        accumulatedContext += fmt.Sprintf("\n--- File: %s ---\n%s\n", path, content)
    }
}
```

Промпт с контекстом:
```
You are a React Frontend Developer. Role: Develops core UI components...

TASK: Create admin dashboard for proxy management
CONTEXT FROM OTHER WORKERS:
=== Previous worker Go API Engineer ===
--- File: project/backend/Go API Engineer/cmd/api/main.go ---
package main
import "github.com/gin-gonic/gin"
...

Generate ONLY the most important files (1-5 files max).
```

### 4.8.repairJSON

На случай если JSON всё же обрезается:

```go
func repairJSON(s string) string {
    // Считаем скобки
    braceCount := 0
    bracketCount := 0
    // ... добавляем недостающие } и ]
    // Убираем висящие запятые
    result = strings.ReplaceAll(result, ",}", "}")
    result = strings.ReplaceAll(result, ",]", "]")
    return result
}
```

### 4.9. Клиент к Agents

```go
type WorkerService struct {
    agentsClient *grpc.AgentClient  // Из agents/pkg/fetcher/grpc
}

func NewWorkerService() *WorkerService {
    agentsHost := os.Getenv("AGENTS_SERVICE_HOST")  // "agents:50053"
    agentsClient := grpc.NewAgentClient(agentsHost)
    return &WorkerService{agentsClient: agentsClient}
}

// Использование:
content, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 8192, 0.3)
```

### 4.10. Модели данных (`pkg/models/worker.go`)

```go
type Worker struct {
    ID         uuid.UUID
    TaskID     uuid.UUID
    ManagerID  uuid.UUID
    Role       string  // "React Developer", "Go API Engineer", ...
    AgentID    string
    Status     string  // thinking, coding, done, error
    TaskMD     string  // TASK.md content
    SolutionMD string
    Files      string  // JSON: {path: content}
    Success    bool
    Approved   bool
    Feedback   string
}

type WorkerSolution struct {
    ID       uuid.UUID
    WorkerID uuid.UUID
    TaskID   uuid.UUID
    Files    string  // JSON: {path: content}
    Success  bool
    Feedback string
    Approved bool
}
```

---

## 5. Apigateway

**Каталог:** `apigateway/`
**Порт:** 3111 (HTTP)
**Роль:** HTTP/WebSocket шлюз между клиентом и Boss
**Язык:** Go 1.25
**Фреймворк:** Azure Framework (ultrahttp)

### 5.1. Назначение

Apigateway — это **точка входа** для пользователей. Он:

1. Принимает HTTP/WebSocket запросы от клиентов
2. Проксирует задачи в Boss через gRPC
3. **Стримит** обновления прогресса обратно клиенту через WebSocket
4. Не содержит бизнес-логики — только маршрутизация

### 5.2. Структура файлов

```
apigateway/
├── cmd/app/main.go                              # Точка входа
├── internal/
│   ├── core/services/hub.go                     # WebSocket Hub (broadcasting)
│   └── fetcher/
│       ├── http.go                              # HTTP маршруты
│       ├── grpc/boss/
│       │   ├── client.go                        # gRPC клиент к Boss
│       │   └── bosspb/                          # Proto (локальная копия)
│       └── grpc/boss/bosspb/
│           ├── boss.pb.go                       # Сгенерированный код
│           └── boss_grpc.pb.go                  # gRPC код
├── pkg/requests/                                # Типы запросов/ответов
└── proto/boss.proto                             # Proto от gateway-boss.proto
```

### 5.3. Точка входа (`cmd/app/main.go`)

```go
func main() {
    a := azure.New()
    fetcher.HttpManager(a)
    a.Listen(3111)
}
```

### 5.4. Маршруты (`internal/fetcher/http.go`)

#### GET /health

```go
a.Get("/health", func(c *azure.Context) {
    c.JsonStatus(200, azure.M{"status": "OK"})
})
```

#### GET /task/status

```go
a.Get("/task/status", func(c *azure.Context) {
    taskID := c.GetQueryParam("task_id")
    resp, err := bossClient.GetTaskStatus(context.Background(), taskID)
    c.JsonStatus(200, azure.M{"status": "success", "task_id": resp.TaskId, "progress": resp.Progress})
})
```

#### GET /task/create (WebSocket)

```go
a.Get("/task/create", azurewebsockets.HandlerFunc(handleTaskCreateWebSocket))

func handleTaskCreateWebSocket(ws *azurewebsockets.Conn, opcode int, data []byte) {
    // 1. Парсим запрос
    var taskReq requests.CreateTaskRequest
    json.Unmarshal(data, &taskReq)

    // 2. Отправляем подтверждение
    ws.WriteJSON(azure.M{"type": "connected", "task_id": taskReq.UserID})

    // 3. Запускаем стриминг в фоне
    go processTaskStreamAzureWS(ws, taskReq)
}
```

### 5.5. Стриминг прогресса

```go
func processTaskStreamAzureWS(ws *azurewebsockets.Conn, taskReq requests.CreateTaskRequest) {
    // Создаём gRPC запрос к Boss
    grpcReq := &bosspb.CreateTaskRequest{
        UserId:      taskReq.UserID,
        Username:    taskReq.Username,
        Title:       taskReq.Title,
        Description: taskReq.Description,
        Tokens:      taskReq.Tokens,
        Meta:        taskReq.Meta,
    }

    // 30-минутный таймаут
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    // Вызываем Boss.CreateTaskStream
    stream, err := bossClient.CreateTaskStream(ctx, grpcReq)

    // Читаем сообщения из стрима и отправляем клиенту
    for {
        update, err := stream.Recv()
        if err != nil { break }

        wsUpdate := azure.M{
            "type":      update.Status,     // "processing", "success", "error"
            "task_id":   update.TaskId,
            "message":   update.Message,    // "🏗️ Creating 3 managers..."
            "progress":  update.Progress,   // 5, 10, 15, ... 100
            "timestamp": update.Timestamp,
        }
        if update.Data != nil {
            wsUpdate["data"] = update.Data
        }

        ws.WriteJSON(wsUpdate)

        // Завершение
        if update.Status == "success" || update.Status == "error" {
            return
        }
    }
}
```

### 5.6. Клиент к Boss (`internal/fetcher/grpc/boss/client.go`)

```go
type Client struct {
    conn   *grpc.ClientConn
    client bosspb.BossServiceClient
}

func NewClient(address string) (*Client, error) {
    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    return &Client{conn: conn, client: bosspb.NewBossServiceClient(conn)}, nil
}

// CreateTaskStream
func (c *Client) CreateTaskStream(ctx, req *bosspb.CreateTaskRequest) (bosspb.BossService_CreateTaskStreamClient, error) {
    return c.client.CreateTaskStream(ctx, req)
}

// GetTaskStatus
func (c *Client) GetTaskStatus(ctx, taskID string) (*bosspb.TaskStatusResponse, error) {
    return c.client.GetTaskStatus(ctx, &bosspb.TaskStatusRequest{TaskId: taskID})
}
```

### 5.7. Типы запросов (`pkg/requests/task.go`)

```go
type CreateTaskRequest struct {
    UserID      string            `json:"userId"`
    Username    string            `json:"username"`
    Title       string            `json:"title"`
    Description string            `json:"description"`
    Tokens      map[string]string `json:"tokens"`       // {"openrouter": "sk-or-..."}
    Meta        map[string]string `json:"meta"`         // {"provider": "openrouter", "model": "qwen3-coder"}
}
```

### 5.8. WebSocket сообщения

**От сервера к клиенту:**

```json
{"type": "connected", "task_id": "uuid", "message": "Connected to task creation service", "timestamp": 1234567890}
{"type": "processing", "progress": 5, "message": "📝 Task received and processing started", "timestamp": 1234567890}
{"type": "processing", "progress": 10, "message": "💾 Task saved to database", "timestamp": 1234567890}
{"type": "processing", "progress": 15, "message": "🤖 AI client initialized (openrouter/qwen3-coder)", "timestamp": 1234567890}
{"type": "processing", "progress": 30, "message": "✅ Architecture planned by AI", "data": {"managers": "3"}, "timestamp": 1234567890}
{"type": "processing", "progress": 40, "message": "🏗️ Creating 3 managers in parallel", "timestamp": 1234567890}
{"type": "processing", "progress": 70, "message": "👷 Managers completed code generation", "timestamp": 1234567890}
{"type": "processing", "progress": 80, "message": "🔍 Boss validating solution...", "timestamp": 1234567890}
{"type": "processing", "progress": 90, "message": "📦 Packaging project", "timestamp": 1234567890}
{"type": "success", "progress": 100, "message": "🎉 Project ready!", "data": {"managers": "3", "techStack": "[\"Go\",\"React\"]"}, "timestamp": 1234567890}
```

**В случае ошибки:**

```json
{"type": "error", "message": "❌ Manager failed: rpc error...", "timestamp": 1234567890}
```

### 5.9. WebSocket Hub (`internal/core/services/hub.go`)

```go
type Hub struct {
    Clients map[string]map[*azurewebsockets.Conn]bool  // taskID → connections
    Register   chan *azurewebsockets.Conn
    Unregister chan *azurewebsockets.Conn
    Broadcast  chan BroadcastMessage
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            // Добавить подключение
        case client := <-h.Unregister:
            // Удалить подключение
        case msg := <-h.Broadcast:
            // Отправить сообщение всем клиентам задачи
        }
    }
}
```

---

## Сводная таблица сервисов

| Сервис | Порт | Протокол | Роль | Зависит от |
|--------|------|----------|------|------------|
| **Apigateway** | 3111 | HTTP/WebSocket | Шлюз | Boss, Agents |
| **Boss** | 50051 | gRPC | CEO/CTO | Agents, Manager |
| **Manager** | 50052 | gRPC | Team Lead | Agents, Worker |
| **Worker** | 50053 | gRPC | Developer | Agents |
| **Agents** | 50053 | gRPC | LLM маршрутизатор | Внешние LLM API |
| **PostgreSQL** | 5432 | TCP | Хранение данных | — |

## Порядок запуска

```
1. PostgreSQL    ← база данных
2. Agents        ← LLM маршрутизатор (не зависит от других)
3. Boss          ← зависит от Agents
4. Manager       ← зависит от Agents, Boss
5. Worker        ← зависит от Agents, Manager
6. Apigateway    ← зависит от Boss, Agents
```
