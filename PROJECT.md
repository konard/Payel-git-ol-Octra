# CrewAI — Подробная документация проекта

## 📖 Оглавление

1. [Общая концепция](#общая-концепция)
2. [Архитектура системы](#архитектура-системы)
3. [Поток выполнения задачи](#поток-выполнения-задачи)
4. [Подробное описание сервисов](#подробное-описание-сервисов)
   - [Agents Service](#agents-service)
   - [Boss Service](#boss-service)
   - [Manager Service](#manager-service)
   - [Worker Service](#worker-service)
   - [Apigateway](#apigateway)
5. [Межсервисное взаимодействие](#межсервисное-взаимодействие)
6. [Работа с LLM-провайдерами](#работа-с-llm-провайдерами)
7. [Модели данных](#модели-данных)
8. [Протоколы и контракты](#протоколы-и-контракты)
9. [Ключевые архитектурные решения](#ключевые-архитектурные-решения)
10. [Производительность и ограничения](#производительность-и-ограничения)
11. [Развёртывание](#развёртывание)
12. [Разработка и тестирование](#разработка-и-тестирование)

---

## Общая концепция

CrewAI — это **многоагентная система** (multi-agent system), которая моделирует работу реальной IT-компании. Вместо того чтобы один ИИ-агент пытался решить всю задачу, система использует **иерархию специализированных агентов**, каждый из которых выполняет свою роль:

```
Пользователь → Apigateway → Boss (CEO) → Manager(s) → Worker(s) → Готовый проект
```

### Аналогия с реальной компанией

| Роль | Аналог в компании | Что делает |
|------|-------------------|------------|
| **Boss** | CEO / CTO | Анализирует задачу клиента, определяет общую архитектуру, стек технологий, назначает руководителей команд |
| **Manager** | Team Lead | Получает зону ответственности (frontend, backend, testing), набирает команду разработчиков, проверяет их работу |
| **Worker** | Developer / QA | Получает конкретную роль, создаёт файлы проекта, генерирует код |
| **Agents** | Внешние подрядчики (LLM) | Единый интерфейс ко всем ИИ-провайдерам (OpenAI, Gemini, Claude и др.) |

### Ключевое отличие от других систем

В отличие от простых «чат-ботов с кодом», CrewAI:
- **Делегирует** — Boss не пишет код, он планирует
- **Координирует** — Manager проверяет и ревьюит работу
- **Контекст** — воркеры видят результаты друг друга
- **Обратная связь** — Manager может отправить код на доработку (ReviewWorker)
- **Валидация** — Boss проверяет финальный результат перед отправкой

---

## Архитектура системы

### Общая схема

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                     Пользователь                         │
                    │              (Web UI / API / WebSocket)                  │
                    └────────────────────────┬────────────────────────────────┘
                                             │ HTTP / WebSocket
                                             ▼
                    ┌─────────────────────────────────────────────────────────┐
                    │                     Apigateway                           │
                    │  Порт: 3111                                              │
                    │  - Принимает HTTP/WebSocket запросы                      │
                    │  - Проксирует в Boss через gRPC                          │
                    │  - Стримит обновления прогресса обратно клиенту          │
                    └────────────────────────┬────────────────────────────────┘
                                             │ gRPC
                                             ▼
                    ┌─────────────────────────────────────────────────────────┐
                    │                      Boss (CEO)                          │
                    │  Порт: 50051                                             │
                    │  - Анализирует задачу через AI (Agents)                  │
                    │  - Определяет: стек, кол-во менеджеров, роли             │
                    │  - Вызывает КАЖДОГО менеджера ПАРАЛЛЕЛЬНО                │
                    │  - Валидирует итоговый результат через AI                │
                    └────────────┬───────────────┬───────────────┬────────────┘
                                 │ gRPC          │ gRPC          │ gRPC
                                 ▼               ▼               ▼
                    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
                    │  Manager #1     │ │  Manager #2     │ │  Manager #3     │
                    │  (frontend)     │ │  (backend)      │ │  (testing)      │
                    │  Порт: 50052    │ │  Порт: 50052    │ │  Порт: 50052    │
                    │                 │ │                 │ │                 │
                    │  - Нанимает     │ │  - Нанимает     │ │  - Нанимает     │
                    │    воркеров     │ │    воркеров     │ │    воркеров     │
                    │  - Ревьюит      │ │  - Ревьюит      │ │  - Ревьюит      │
                    │    работу       │ │    работу       │ │    работу       │
                    │  - Собирает ZIP │ │  - Собирает ZIP │ │  - Собирает ZIP │
                    └────────┬────────┘ └────────┬────────┘ └────────┬────────┘
                             │ gRPC              │ gRPC              │ gRPC
                             ▼                   ▼                   ▼
                    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
                    │  Worker(s) #1   │ │  Worker(s) #2   │ │  Worker(s) #3   │
                    │  Порт: 50053    │ │  Порт: 50053    │ │  Порт: 50053    │
                    │                 │ │                 │ │                 │
                    │  1. Определяет  │ │  1. Определяет  │ │  1. Определяет  │
                    │     файлы       │ │     файлы       │ │     файлы       │
                    │  2. Генерирует  │ │  2. Генерирует  │ │  2. Генерирует  │
                    │     каждый      │ │     каждый      │ │     каждый      │
                    │     отдельно    │ │     отдельно    │ │     отдельно    │
                    │  3. Возвращает  │ │  3. Возвращает  │ │  3. Возвращает  │
                    │     ZIP         │ │     ZIP         │ │     ZIP         │
                    └────────┬────────┘ └────────┬────────┘ └────────┬────────┘
                             │                   │                   │
                             └───────────────────┼───────────────────┘
                                                 │ gRPC
                                                 ▼
                    ┌─────────────────────────────────────────────────────────┐
                    │                   Agents Service                         │
                    │  Порт: 50053 (gRPC)                                      │
                    │  - Единый интерфейс к 6 LLM-провайдерам                  │
                    │  - OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok    │
                    │  - Per-request токены (ключ не хранится)                 │
                    │  - Retry при transient-ошибках (3 попытки)               │
                    │  - Streaming генерация                                   │
                    └────────────────────────┬────────────────────────────────┘
                                             │ HTTP
                                             ▼
                    ┌─────────────────────────────────────────────────────────┐
                    │                    LLM Providers                         │
                    │  OpenRouter API  │  Gemini API  │  OpenAI API  │  ...   │
                    └─────────────────────────────────────────────────────────┘
```

### Хранение данных

Все сервисы используют **единую PostgreSQL** базу:

```
PostgreSQL (crewai)
├── tasks              — задачи от пользователей
├── boss_decisions     — решения Boss (архитектура, стек, роли)
├── managers           — менеджеры и их команды
├── workers            — воркеры и их результаты
└── worker_solutions   — файлы, сгенерированные воркерами
```

---

## Поток выполнения задачи

### Пошаговый процесс

#### Этап 1: Приём задачи (Apigateway)
```
User отправляет JSON через WebSocket:
{
  "username": "pavel",
  "title": "Прокси сервер",
  "description": "Напиши прокси сервис на go",
  "tokens": { "openrouter": "sk-or-v1-..." },
  "meta": { "provider": "openrouter", "model": "qwen/qwen3-coder" }
}

Apigateway:
  1. Парсит запрос
  2. Отправляет подтверждение: "Connected to task creation service"
  3. Вызывает Boss.CreateTaskStream()
  4. Стримит обновления клиенту по WebSocket
```

#### Этап 2: Планирование (Boss)
```
Boss:
  1. Сохраняет задачу в БД (status: boss_planning)
  2. Отправляет в Agents: "Проанализируй задачу, определи стек и менеджеров"
  3. Парсит JSON-ответ:
     {
       "managers_count": 3,
       "manager_roles": [
         {"role": "frontend", "description": "...", "priority": 2},
         {"role": "backend", "description": "...", "priority": 1},
         {"role": "testing", "description": "...", "priority": 3}
       ],
       "technical_description": "...",
       "tech_stack": ["Go", "React", "PostgreSQL", "Docker"],
       "architecture_notes": "..."
     }
  4. Сохраняет BossDecision в БД
  5. Обновляет статус задачи на managers_assigned
```

#### Этап 3: Параллельное выполнение (Boss → Managers)
```
Boss вызывает AssignManager для КАЖДОГО менеджера ПАРАЛЛЕЛЬНО:

  ┌─ Manager #1 (backend, priority 1) ──┐
  ├─ Manager #2 (frontend, priority 2) ──┤  ← вызываются одновременно
  └─ Manager #3 (testing, priority 3) ───┘
```

#### Этап 4: Работа менеджера
```
Manager (backend):
  1. Создаёт запись в БД (статус: active)
  2. Отправляет в Agents: "Каких воркеров нанять для backend-команды?"
     Ответ:
     {
       "worker_roles": [
         {"role": "Core Go Systems Engineer", "description": "..."},
         {"role": "Routing & Middleware Developer", "description": "..."},
         {"role": "Backend API Developer", "description": "..."}
       ]
     }
  3. Сохраняет WorkerRoles в БД
  4. Вызывает Worker.AssignWorkersAndWait() для ВСЕХ воркеров
```

#### Этап 5: Работа воркера
```
Worker (Core Go Systems Engineer):
  1. Создаёт TASK.md (через Agents): "Разбери задачу для моей роли"
  2. Определяет файлы (через Agents): "Какие файлы создать?"
     → ["main.go", "proxy.go", "config.go"]
  3. Для КАЖДОГО файла — отдельный запрос к Agents:
     → "Напиши main.go" → plain text код
     → "Напиши proxy.go" → plain text код
     → "Напиши config.go" → plain text код
  4. Добавляет TASK.md в файлы
  5. Создаёт ZIP-архив
  6. Возвращает ZIP + метаданные воркера
```

#### Этап 6: Review (Manager)
```
Manager получает результаты от всех воркеров:
  1. Для каждого воркера отправляет в Agents: "Проверь качество кода"
     {
       "approved": false,
       "feedback": "Нет обработки ошибок в HTTP handler"
     }
  2. Если НЕ одобрен → вызывает Worker.ReviewWorker(feedback)
  3. Воркер исправляет → повторная проверка
  4. Собирает все файлы в единый ZIP
  5. Возвращает ZIP + результаты всех воркеров
```

#### Этап 7: Финальная валидация (Boss)
```
Boss получает ZIP от всех менеджеров:
  1. Объединяет все ZIP в один
  2. Отправляет в Agents: "Проверь итоговое решение"
     {
       "approved": true,
       "feedback": "Архитектура соблюдана, все компоненты присутствуют"
     }
  3. Если НЕ одобрен → можно отправить на доработку (TODO)
  4. Сохраняет ZIP в Task.Solution
  5. Обновляет статус на "done"
```

#### Этап 8: Возврат результата
```
Boss → Apigateway → User (WebSocket):
{
  "type": "success",
  "task_id": "uuid-...",
  "message": "🎉 Project ready!",
  "progress": 100,
  "data": {
    "managers": 3,
    "techStack": ["Go", "React", "PostgreSQL", "Docker"]
  }
}
```

---

## Подробное описание сервисов

### Agents Service

**Расположение:** `agents/`
**Порт:** 50053 (gRPC)
**Роль:** Единый интерфейс ко всем LLM-провайдерам

#### Архитектура провайдеров

```
agents/
└── internal/fetcher/providers/
    ├── openrouter/    — OpenRouter API (совместим с OpenAI SDK)
    ├── gemini/        — Google Gemini SDK
    ├── openai/        — OpenAI SDK
    ├── claude/        — Anthropic Claude SDK
    ├── deepseek/      — DeepSeek SDK
    └── grok/          — xAI Grok SDK
```

#### Ключевые особенности

**1. Per-request токены**

Все провайдеры принимают API-ключ **в каждом запросе** через поле `tokens`, а не при инициализации:

```go
// OpenAI провайдер
func (p *OpenAIProvider) Generate(ctx, prompt, tokens) (string, error) {
    client := p.getClient(tokens)  // Создаёт клиент из токена запроса
    // ...
}

func (p *OpenAIProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
    // Ищет ключ по именам: "openai", "api_key", "apiKey", "token"
    apiKey := extractAPIKey(tokens)
    return openai.NewClient(option.WithAPIKey(apiKey)), nil
}
```

**2. Lazy client creation**

Клиент создаётся **лениво** — только при первом вызове `Generate`. Это позволяет:
- Регистрировать все провайдеры при старте без ключей
- Создавать клиент только когда он реально нужен
- Переиспользовать клиент для последующих запросов

**3. Retry-логика (OpenRouter)**

```go
for attempt := 0; attempt < 3; attempt++ {
    response, err := client.Chat.Completions.New(ctx, params)
    if err == nil { return response.Choices[0].Message.Content, nil }
    
    if isTransient(err) {  // EOF, connection reset, timeout
        time.Sleep(time.Duration(attempt) * 2 * time.Second)
        continue
    }
    return "", err  // Auth error, rate limit — не retry
}
```

**4. Streaming**

```protobuf
rpc GenerateStream(GenerateRequest) returns (stream GenerateStreamChunk) {}

message GenerateStreamChunk {
  string content = 1;      // partial content
  bool done = 2;           // true when generation is complete
  string error = 3;        // error message if any
  int32 tokens_used = 4;   // tokens used so far
}
```

---

### Boss Service

**Расположение:** `boss/`
**Порт:** 50051 (gRPC)
**Роль:** CEO/CTO — принимает задачи, планирует архитектуру, координирует менеджеров

#### Основные методы

| Метод | Тип | Описание |
|-------|-----|----------|
| `CreateTask` | Unary | Принимает задачу, выполняет весь цикл, возвращает ZIP |
| `CreateTaskStream` | Server Stream | То же + стримит прогресс по мере выполнения |
| `GetTaskStatus` | Unary | Возвращает статус задачи по ID |

#### Логика планирования

Boss отправляет в Agents промпт:
```
You are CTO. Task:
Title: {title}
Description: {description}

Reply ONLY with JSON:
{
  "managers_count": 3,
  "manager_roles": [
    {"role": "frontend", "description": "...", "priority": 1},
    {"role": "backend", "description": "...", "priority": 2},
    {"role": "testing", "description": "...", "priority": 3}
  ],
  "technical_description": "...",
  "tech_stack": ["Go", "React", "PostgreSQL", "Docker"],
  "architecture_notes": "..."
}
```

#### Параллельный вызов менеджеров

```go
func (s *BossService) assignManagersParallel(ctx, taskID, decision, req) {
    var wg sync.WaitGroup
    for i, role := range decision.ManagerRoles {
        wg.Add(1)
        go func(idx int, role models.ManagerRole) {
            defer wg.Done()
            
            // Wait for higher-priority managers to produce context
            if idx > 0 {
                time.Sleep(time.Duration(idx) * 5 * time.Second)
            }
            
            // Collect context from already-completed managers
            contextResults := collectResults()
            
            result, err := s.managerClient.AssignManager(ctx, &AssignManagerRequest{
                ManagerId:            uuid.New().String(),
                Role:                 role.Role,
                Description:          role.Description,
                TechnicalDescription: decision.TechnicalDescription,
                Priority:             role.Priority,
                OtherWorkersResults:  contextResults,
            })
            // ...
        }(i, role)
    }
    wg.Wait()
}
```

#### Финальная валидация

```go
func (s *BossService) validateSolution(ctx, agentsClient, provider, model, tokens, decision, managerResults, zipData) {
    prompt := `You are the CTO reviewing the final deliverable.
    
    ORIGINAL TASK: {title}
    ARCHITECTURE: {technical_description}
    MANAGERS RESULTS: {summary}
    GENERATED FILES: {fileCount} total
    ZIP size: {zipSize} bytes
    
    Review:
    1. Does the solution meet the requirements?
    2. Is the architecture followed?
    3. Are all managers completed their work?
    
    Reply ONLY with JSON:
    {"approved": true/false, "feedback": "..."}`
    
    // Parses response and returns ValidationResult
}
```

---

### Manager Service

**Расположение:** `manager/`
**Порт:** 50052 (gRPC)
**Роль:** Team Lead — набирает команду, распределяет задачи, ревьюит работу

#### Основные методы

| Метод | Тип | Описание |
|-------|-----|----------|
| `AssignManager` | Unary | Назначить ОДНОГО менеджера (вызывается Boss для каждого) |
| `AssignManagersAndWait` | Unary | Legacy — назначить всех последовательно |
| `AssignManagersAndWaitStream` | Server Stream | Стриминг прогресса |

#### Логика managerThink

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

#### Review-цикл

```go
func (s *ManagerService) processManager(ctx, req) {
    // 1. Determine workers via AI
    workerRoles := s.managerThink(ctx, provider, model, tokens, req.TechnicalDescription, req.Role)
    
    // 2. Call Worker service
    workerResp := s.workerClient.AssignWorkersAndWait(ctx, &AssignWorkersRequest{
        WorkerRoles:         workerRoles,
        OtherWorkersResults: contextResults,  // from other workers
    })
    
    // 3. Review each worker via AI
    for _, wr := range workerResp.WorkerResults {
        review := s.reviewWorkerResult(ctx, provider, model, tokens, req.Role, wr)
        
        if !review.Approved {
            // Send back for fixes
            reviewResp := s.workerClient.ReviewWorker(ctx, &ReviewRequest{
                Feedback:       review.Feedback,
                OriginalFiles:  wr.Files,
            })
            // Update files with reviewed version
        }
    }
    
    // 4. Merge all files into ZIP
    finalZip := createZipArchive(allFiles)
    return finalZip
}
```

#### reviewWorkerResult

```
You are a {managerRole} manager reviewing work from a {workerRole} developer.

TASK: {task_md}
PROPOSED SOLUTION: {solution_md}
FILES: {files_list}

Review the work:
1. Does it fulfill the task requirements?
2. Is the code quality acceptable?
3. Are there obvious bugs or missing pieces?
4. Does it integrate well with the team?

Reply ONLY with JSON:
{"approved": true/false, "feedback": "..."}
```

---

### Worker Service

**Расположение:** `worker/`
**Порт:** 50053 (gRPC)
**Роль:** Developer — генерирует код, создаёт файлы проекта

#### Ключевое отличие от старой версии

**Раньше:** Worker ходил в LLM напрямую через `pkg/llm/client.go` (OpenRouter/Gemini HTTP).

**Теперь:** Worker использует **Agents сервис** через gRPC. Это:
- Единая точка входа ко всем провайдерам
- Централизованная retry-логика
- Возможность переключать провайдеров без изменения Worker

#### N+1 подход к генерации файлов

```
Шаг 1: Планирование (1 запрос к AI)
  → "Какие файлы создать для моей роли?"
  → ["main.go", "config.go", "README.md"]

Шаг 2: Генерация (N запросов к AI — по одному на файл)
  → "Напиши main.go" → plain text
  → "Напиши config.go" → plain text
  → "Напиши README.md" → plain text

Шаг 3: Добавление TASK.md
  → TASK.md (созданный на Шаге 0) добавляется в файлы
```

**Почему plain text, а не JSON?**

JSON с кодом ломается из-за кавычек, экранирования, обрезания по max_tokens. Plain text:
- LLM возвращает raw code
- `stripMarkdownCodeBlock()` убирает ````go ... ````
- Остаётся чистый код

#### Координация между воркерами

Каждый воркер видит результаты предыдущих:

```go
// Build accumulated context from previous workers
accumulatedContext := ""
for _, prevResult := range workerResults {
    accumulatedContext += fmt.Sprintf("\n=== Previous worker %s (%s) ===\n",
        prevResult.WorkerId, prevResult.Role)
    for path, content := range prevResult.Files {
        accumulatedContext += fmt.Sprintf("\n--- File: %s ---\n%s\n", path, content)
    }
}
```

Промпт для генерации файла:
```
You are a {role} developer. Role: {description}

TASK: {task_md}
CONTEXT FROM OTHER WORKERS:
=== Previous worker Go API Engineer ===
--- File: project/backend/Go API Engineer/cmd/api/main.go ---
package main
...

Generate ONLY the most important files (1-5 files max).
Each file should be compact and functional.
Return the file content as PLAIN TEXT. NO JSON. NO markdown. Just the raw code.
```

#### ReviewWorker

```go
func (s *WorkerService) ReviewWorker(ctx, req *ReviewRequest) (*ReviewResponse, error) {
    // Fix each file individually based on manager feedback
    for filePath, oldContent := range req.OriginalFiles {
        prompt := fmt.Sprintf(`You are a %s developer. Your previous work was reviewed.
FILE: %s
MANAGER FEEDBACK: %s
PREVIOUS CODE: %s
FIX the code based on the feedback. Return the FULL corrected file as PLAIN TEXT.`,
            req.WorkerRole, filePath, req.Feedback, oldContent)
        
        fixedContent := s.agentsClient.Generate(ctx, prompt)
        fixedFiles[filePath] = stripMarkdownCodeBlock(fixedContent)
    }
    return &ReviewResponse{Files: fixedFiles, Status: "fixed"}
}
```

---

### Apigateway

**Расположение:** `apigateway/`
**Порт:** 3111 (HTTP)
**Роль:** HTTP/WebSocket шлюз между клиентом и Boss

#### WebSocket стриминг

```
Client ↔ Apigateway ↔ Boss (CreateTaskStream)

Client получает сообщения:
{"type": "processing", "progress": 5, "message": "📝 Task received..."}
{"type": "processing", "progress": 10, "message": "💾 Task saved..."}
{"type": "processing", "progress": 15, "message": "🤖 AI client initialized..."}
{"type": "processing", "progress": 30, "message": "✅ Architecture planned..."}
{"type": "processing", "progress": 40, "message": "🏗️ Creating 3 managers in parallel"}
{"type": "processing", "progress": 70, "message": "👷 Managers completed..."}
{"type": "processing", "progress": 80, "message": "🔍 Boss validating..."}
{"type": "processing", "progress": 90, "message": "📦 Packaging..."}
{"type": "success", "progress": 100, "message": "🎉 Project ready!"}
```

---

## Межсервисное взаимодействие

### gRPC контракты

#### Agents → LLM
```protobuf
service AgentService {
  rpc Generate(GenerateRequest) returns (GenerateResponse) {}
  rpc GenerateStream(GenerateRequest) returns (stream GenerateStreamChunk) {}
}

message GenerateRequest {
  string provider = 1;           // openrouter, gemini, openai, claude, deepseek, grok
  string model = 2;              // qwen/qwen3-coder, gemini-2.5-flash, ...
  string prompt = 3;
  map<string, string> tokens = 4; // API keys per-request
  int32 max_tokens = 5;
  float temperature = 6;
}
```

#### Boss → Manager
```protobuf
service ManagerService {
  rpc AssignManager(AssignManagerRequest) returns (ManagerResult) {}
  rpc AssignManagersAndWait(AssignManagersRequest) returns (AssignManagersResponse) {}
}

message AssignManagerRequest {
  string task_id = 1;
  string manager_id = 2;
  string role = 3;                  // frontend, backend, testing
  string description = 4;
  string technical_description = 5;
  int32 priority = 6;
  map<string, string> metadata = 7; // tokens, model, provider
  repeated WorkerResult other_workers_results = 8; // context
}

message ManagerResult {
  string task_id = 1;
  string manager_id = 2;
  string role = 3;
  string status = 4;
  bytes solution = 5;               // ZIP archive
  repeated WorkerResult worker_results = 6;
  string review_summary = 7;
}
```

#### Manager → Worker
```protobuf
service WorkerService {
  rpc AssignWorkersAndWait(AssignWorkersRequest) returns (AssignWorkersResponse) {}
  rpc ReviewWorker(ReviewRequest) returns (ReviewResponse) {}
}

message AssignWorkersRequest {
  string task_id = 1;
  string manager_id = 2;
  string manager_role = 3;
  repeated WorkerRole worker_roles = 4;
  string task_md = 5;
  map<string, string> metadata = 6;
  repeated WorkerResult other_workers_results = 7; // context
}

message ReviewRequest {
  string task_id = 1;
  string manager_id = 2;
  string worker_id = 3;
  string worker_role = 4;
  string feedback = 5;
  string original_solution_md = 6;
  map<string, string> original_files = 7;
  map<string, string> metadata = 8;
}
```

---

## Работа с LLM-провайдерами

### Формат запроса от клиента

```json
{
  "username": "pavel",
  "title": "Прокси сервер",
  "description": "Напиши прокси сервис на go",
  "tokens": {
    "openrouter": "sk-or-v1-...",
    "gemini": "AIzaSy...",
    "openai": "sk-..."
  },
  "meta": {
    "provider": "openrouter",
    "model": "qwen/qwen3-coder"
  }
}
```

### Поток токенов

```
Client → Apigateway → Boss → Agents
  tokens: {"openrouter": "sk-or-v1-..."}
  meta.provider: "openrouter"
  meta.model: "qwen/qwen3-coder"
                                    │
                                    ▼
                              OpenRouterProvider
                                getClient(tokens)
                                  ↓
                              extractAPIKey(tokens)
                                → ищет: "openrouter", "api_key", "apiKey"
                                  ↓
                              openai.NewClient(option.WithAPIKey(apiKey),
                                option.WithBaseURL("https://openrouter.ai/api/v1"))
```

### Рекомендуемые модели

| Модель | Провайдер | Цена | Скорость | Для чего |
|--------|-----------|------|----------|----------|
| `qwen/qwen3.6-plus:free` | OpenRouter | Бесплатно | Медленно | Тесты |
| `qwen/qwen3-coder` | OpenRouter | ~$0.02/1M | Быстро | Код ⭐ |
| `gemini-2.5-flash` | Gemini | 20 запросов/мин | Быстро | Планирование |
| `gpt-4o-mini` | OpenAI | ~$0.15/1M | Быстро | Код |
| `claude-sonnet-4` | OpenRouter | ~$3/1M | Быстро | Сложные задачи |

---

## Модели данных

### Task
```go
type Task struct {
    ID          uuid.UUID
    UserID      string
    Username    string
    Title       string
    Description string
    Tokens      string  // JSON map[string]string
    Meta        string  // JSON map[string]string
    Status      string  // pending → boss_planning → managers_assigned → processing → reviewing → done/error
    Solution    []byte  // ZIP archive
}
```

### BossDecision
```go
type BossDecision struct {
    ID                   uuid.UUID
    TaskID               uuid.UUID
    Status               string
    ManagersCount        int32
    ManagerRoles         string  // JSON []ManagerRole
    TechnicalDescription string
    TechStack            string  // JSON []string
    ArchitectureNotes    string
}

type ManagerRole struct {
    Role        string  // "frontend", "backend", "testing"
    Description string
    Priority    int32
}
```

### Manager
```go
type Manager struct {
    ID           uuid.UUID
    TaskID       uuid.UUID
    Role         string  // "frontend", "backend", "testing"
    AgentID      string
    Status       string  // active, completed, error
    TaskDesc     string
    WorkerRoles  string  // JSON []WorkerRole
    WorkersCount int32
}

type WorkerRole struct {
    Role        string  // "React Developer", "Go API Engineer", ...
    Description string
}
```

### Worker
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
    Files      string  // JSON map[string]string {path: content}
    Success    bool
    Approved   bool
    Feedback   string
}
```

---

## Ключевые архитектурные решения

### 1. Agents как единая точка входа к LLM

**Проблема:** Каждый сервис ходил в LLM напрямую → дублирование кода, нет централизованной retry-логики, сложно добавить новый провайдер.

**Решение:** Выделенный Agents сервис:
- Все провайдеры в одном месте
- Per-request токены → не нужно хранить ключи
- Retry-логика одна для всех
- Добавление нового провайдера = один новый пакет

### 2. Параллельные менеджеры

**Проблема:** Boss вызывал менеджеров последовательно → 3 менеджера × 10 минут = 30 минут.

**Решение:** Boss вызывает `AssignManager` для каждого менеджера **параллельно** через goroutines. Менеджеры с более высоким приоритетом начинают первыми, остальные ждут немного чтобы получить контекст.

### 3. N+1 генерация файлов

**Проблема:** Один запрос «сгенерируй все файлы в JSON» → JSON обрезается по max_tokens, кавычки в коде ломают парсинг.

**Решение:**
1. Отдельный запрос: «какие файлы создать?» → простой JSON со списком
2. Для каждого файла — отдельный запрос: «напиши содержимое» → plain text
3. `stripMarkdownCodeBlock()` убирает markdown обёртку

### 4. Plain text вместо JSON для кода

**Проблема:** JSON с кодом содержит кавычки, переносы строк, спецсимволы → ломается при парсинге.

**Решение:** LLM возвращает raw code без JSON обёртки. Промпт: «Return the file content as PLAIN TEXT. NO JSON. NO markdown.»

### 5. Контекст между воркерами

**Проблема:** Воркеры работают изолированно → дублируют код, не знают интерфейсов друг друга.

**Решение:** Каждый воркер получает `other_workers_results` — файлы и результаты предыдущих воркеров. Менеджер передаёт это в запросе.

### 6. Review-цикл

**Проблема:** Менеджер просто собирал ZIP без проверки.

**Решение:**
1. Manager через AI проверяет каждый файл воркера
2. Если не одобрен → `ReviewWorker` с feedback
3. Воркер исправляет → повторная проверка
4. Boss через AI валидирует итоговый ZIP

---

## Производительность и ограничения

### Количество запросов к LLM

Для типичной задачи с 3 менеджерами и ~6 воркерами каждый:

| Этап | Запросов |
|------|----------|
| Boss planning | 1 |
| Manager think (×3) | 3 |
| Worker TASK.md (×18) | 18 |
| Worker file planning (×18 × 4 файлов) | 72 |
| Worker file generation (×18 × 4 файлов) | 72 |
| Manager review (×18) | 18 |
| Boss validation | 1 |
| **ИТОГО** | **~185** |

### Время выполнения

| Модель | Время на запрос | Общее время |
|--------|-----------------|-------------|
| `qwen/qwen3.6-plus:free` | 30-60 сек | **1.5-3 часа** |
| `qwen/qwen3-coder` | 10-20 сек | **30-60 минут** |
| `gpt-4o-mini` | 5-10 сек | **15-30 минут** |

### Ограничения

1. **Бесплатные модели** — медленные, с лимитами
2. **Таймаут** — 30 минут (настраивается в apigateway)
3. **Нет реальной параллелизации файлов** — воркеры внутри команды работают последовательно
4. **Нет сборки проекта** — файлы просто собираются в ZIP, но проект может не компилироваться

---

## Развёртывание

### Docker Compose

```yaml
services:
  postgres:
    image: postgres:15-alpine
    ports: ["5432:5432"]
  
  agents:
    build: ./agents
    ports: ["50053:50053"]
  
  boss:
    build: ./boss
    ports: ["50051:50051"]
    depends_on: [postgres, agents]
  
  manager:
    build: ./manager
    ports: ["50052:50052"]
    depends_on: [postgres, agents, boss]
  
  worker:
    build: ./worker
    depends_on: [postgres, agents, manager]
  
  apigateway:
    build: ./apigateway
    ports: ["3111:3111"]
    depends_on: [boss, agents]
```

### Переменные окружения

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `DB_DNS` | boss, manager, worker | PostgreSQL connection string |
| `AGENTS_SERVICE_HOST` | boss, manager, worker | Адрес Agents сервиса |
| `WORKER_SERVICE_HOST` | manager | Адрес Worker сервиса |
| `BOSS_SERVICE_HOST` | apigateway | Адрес Boss сервиса |

---

## Разработка и тестирование

### Генерация proto кода

```powershell
# PowerShell — все сервисы
.\generate-proto.ps1

# Или вручную для agents
cd agents
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --proto_path=proto proto/agents.proto
```

### Локальный запуск

```bash
# Terminal 1 — PostgreSQL
docker run -d --name crewai-postgres -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=crewai -p 5432:5432 postgres:15-alpine

# Terminal 2 — Agents
cd agents && go run cmd/app/main.go

# Terminal 3 — Boss
cd boss && go run cmd/app/main.go

# Terminal 4 — Manager
cd manager && go run cmd/app/main.go

# Terminal 5 — Worker
cd worker && go run cmd/app/main.go

# Terminal 6 — Apigateway
cd apigateway && go run cmd/app/main.go
```

### Тестирование через Node.js клиент

```bash
cd frontend/npm_client_test
node index.js
```

Клиент подключается по WebSocket к `ws://localhost:3111/task/create` и стримит обновления.
