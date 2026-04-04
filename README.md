# CrewAI — Компания из ИИ-агентов

Многоагентная система, которая работает как реальная IT-компания с иерархией:
**User → Apigateway → Boss → Manager(s) → Worker(s) → ZIP-архив**

## 🏗 Архитектура

```
┌─────────────┐      HTTP/WS     ┌─────────────┐      gRPC       ┌─────────────┐
│   User UI   │ ───────────────► │  Apigateway │ ──────────────► │    Boss     │
│  (frontend) │     :3111        │   (proxy)   │                 │  (port 50051)│
└─────────────┘                  └─────────────┘                 └──────┬──────┘
                                                                        │
                                                         gRPC (ПАРАЛЛЕЛЬНО)
                                                        ┌───────┴───────┐
                                                        ▼               ▼
                                               ┌─────────────┐  ┌─────────────┐
                                               │  Manager #1 │  │  Manager #2 │  ...
                                               │  (port 50052)│  │  (port 50052)│
                                               └──────┬──────┘  └──────┬──────┘
                                                      │                 │
                                               gRPC (последовательно)   │
                                                      ▼                 ▼
                                               ┌─────────────┐  ┌─────────────┐
                                               │   Worker(s) │  │   Worker(s) │
                                               │  (port 50053)│  │  (port 50053)│
                                               └──────┬──────┘  └──────┬──────┘
                                                      │                 │
                                                      ▼                 ▼
                                               ┌─────────────┐  ┌─────────────┐
                                               │   Agents    │  │   Agents    │
                                               │  (port 50053)│  │  (port 50053)│
                                               └─────────────┘  └─────────────┘
                                                        │
                                              LLM API (OpenRouter, Gemini,
                                              OpenAI, Claude, DeepSeek, Grok)
```

## 📋 Компоненты

### Agents Service (gRPC :50053)
Централизованный сервис для работы с LLM-провайдерами:
- **6 провайдеров**: OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok
- **Per-request токены** — API-ключ передаётся в каждом запросе, не хранится на сервере
- **Streaming** (`GenerateStream`) — потоковая генерация ответа
- **Retry-логика** — 3 попытки при transient-ошибках (EOF, timeout, connection reset)
- Все остальные сервисы обращаются к LLM **только через Agents**

### Apigateway (:3111)
- HTTP/WebSocket шлюз для клиентских запросов
- Проксирует задачи в Boss сервис через gRPC
- **WebSocket стриминг** — real-time обновления прогресса
- Эндпоинты:
  - `GET /task/create` — WebSocket для создания задачи
  - `GET /task/status?task_id=...` — статус задачи
  - `GET /health` — проверка здоровья

### Boss Service (gRPC :50051)
- Принимает задачи от apigateway
- **ИИ-планирование**: через Agents анализирует задачу, определяет стек и архитектуру
- Назначает **N менеджеров** (обычно 1-3) с ролями и приоритетами
- **Параллельный вызов** `AssignManager` для каждого менеджера
- **Финальная валидация** — Boss проверяет итоговый ZIP через AI перед отправкой
- Сохраняет решение в PostgreSQL

### Manager Service (gRPC :50052)
- `AssignManager` — назначить **одного** менеджера (вызывается Boss для каждого отдельно)
- **ИИ-распределение**: через Agents решает, каких воркеров нанять для своей команды
- **Review-цикл**: проверяет работу каждого воркера через AI
  - Если код не прошёл → отправляет `ReviewWorker` с замечаниями
  - Воркер исправляет → повторная проверка
- Собирает файлы всех воркеров в единую структуру
- Возвращает ZIP + результаты воркеров

### Worker Service (gRPC :50053)
- Получает задачу от менеджера с ролью и описанием
- **Использует Agents сервис** для генерации (не ходит в LLM напрямую)
- **N+1 подход**:
  1. Запрос к AI: «Какие файлы создать?» → список файлов
  2. Для каждого файла — отдельный запрос: «Напиши содержимое»
- **Координация**: каждый воркер видит результаты предыдущих (контекст)
- **Plain text** — файлы генерируются как raw code, без JSON-обёрток
- Поддержка `ReviewWorker` — исправление кода по замечаниям менеджера

## 🚀 Быстрый старт

### 1. Клонирование
```bash
git clone <repository>
cd crewai
```

### 2. Настройка переменных окружения
```bash
cp .env.example .env
# Укажите API-ключи провайдеров
```

### 3. Запуск через Docker Compose
```bash
docker-compose up -d --build
```

Сервисы запустятся:
- Apigateway: http://localhost:3111
- Boss: localhost:50051
- Manager: localhost:50052
- Worker: localhost:50053
- Agents: localhost:50053
- PostgreSQL: localhost:5432

### 4. Тестирование

Создать задачу (через WebSocket):
```json
{
  "userId": "user-123",
  "username": "pavel",
  "title": "Мини прокси",
  "description": "Простой HTTP прокси на Go. Один файл main.go. Базовая маршрутизация.",
  "tokens": {
    "openrouter": "sk-or-v1-...",
    "gemini": "AIzaSy..."
  },
  "meta": {
    "provider": "openrouter",
    "model": "qwen/qwen3-coder"
  }
}
```

Проверить статус:
```bash
curl http://localhost:3111/task/status?task_id=<task_id>
```

## 🤖 Поддерживаемые LLM-провайдеры

| Провайдер | Ключ в `tokens` | Модель по умолчанию | Бесплатно? |
|-----------|-----------------|---------------------|------------|
| **OpenRouter** | `"openrouter": "sk-or-v1-..."` | `qwen/qwen3.6-plus:free` | ✅ Да |
| **Gemini** | `"gemini": "AIzaSy..."` | `gemini-2.5-flash` | ✅ 20 запросов/мин |
| **OpenAI** | `"openai": "sk-..."` | `gpt-4o` | ❌ |
| **Claude** | `"claude": "sk-ant-..."` | `claude-opus-4-6` | ❌ |
| **DeepSeek** | `"deepseek": "sk-..."` | `deepseek-chat` | ❌ (дешёвый) |
| **Grok** | `"grok": "xai-..."` | `grok-3` | ❌ |

### Рекомендуемые бесплатные модели:
- `qwen/qwen3.6-plus:free` — OpenRouter, хорошая для кода
- `meta-llama/llama-3-8b-instruct:free` — OpenRouter, быстрая
- `gemini-2.5-flash` — Gemini, 20 запросов/мин

### Рекомендуемые платные модели (дешёвые):
- `qwen/qwen3-coder` — OpenRouter, ~$0.02/1M tokens ⭐ лучшая для кода
- `openai/gpt-4o-mini` — ~$0.15/1M tokens

## 📁 Структура проекта

```
crewai/
├── agents/             # Agents сервис (LLM-маршрутизатор)
│   ├── cmd/app/
│   ├── internal/fetcher/providers/
│   │   ├── openrouter/
│   │   ├── gemini/
│   │   ├── openai/
│   │   ├── claude/
│   │   ├── deepseek/
│   │   └── grok/
│   ├── pkg/fetcher/grpc/
│   ├── pkg/models/
│   └── proto/
├── apigateway/         # HTTP/WebSocket шлюз
│   ├── cmd/app/
│   ├── internal/fetcher/
│   └── pkg/requests/
├── boss/               # Boss сервис (CEO)
│   ├── cmd/app/
│   ├── internal/
│   │   ├── fetcher/grpc/
│   │   │   ├── boss/
│   │   │   └── manager/
│   │   └── service/
│   ├── pkg/
│   │   ├── database/
│   │   └── models/
│   └── proto/
├── manager/            # Manager сервис
│   ├── cmd/app/
│   ├── internal/
│   │   ├── fetcher/grpc/
│   │   │   ├── managerpb/
│   │   │   └── worker/
│   │   └── service/
│   ├── pkg/
│   │   ├── database/
│   │   └── models/
│   └── proto/
├── worker/             # Worker сервис
│   ├── cmd/app/
│   ├── internal/
│   │   ├── fetcher/grpc/
│   │   │   └── workerpb/
│   │   └── service/
│   ├── pkg/
│   │   ├── database/
│   │   └── models/
│   └── proto/
├── frontend/           # Тестовый клиент (Node.js)
│   └── npm_client_test/
├── docker-compose.yml
├── go.work
└── .env.example
```

## 🔄 Поток выполнения задачи

```
1. User → Apigateway (WebSocket)
2. Apigateway → Boss (gRPC: CreateTaskStream)
3. Boss → Agents: «Проанализируй задачу, определи стек и менеджеров»
4. Boss → Manager #1, #2, #3 (gRPC: AssignManager — ПАРАЛЛЕЛЬНО)
   ├── Manager → Agents: «Каких воркеров нанять для моей команды?»
   ├── Manager → Worker(s) (gRPC: AssignWorkersAndWait)
   │   ├── Worker → Agents: «Какие файлы создать?»
   │   ├── Worker → Agents: «Напиши файл 1»
   │   ├── Worker → Agents: «Напиши файл 2»
   │   └── Worker → ZIP-архив
   ├── Manager → Agents: «Проверить работу воркера»
   ├── Manager → Worker (gRPC: ReviewWorker — если не одобрено)
   └── Manager → ZIP + результаты
5. Boss → Agents: «Валидировать итоговое решение»
6. Boss → Apigateway → User: ZIP-архив
```

## 📝 Модели данных

### Task
- ID, UserID, Username, Title, Description
- Tokens (JSON), Meta (JSON)
- Status: `pending → boss_planning → managers_assigned → processing → reviewing → done/error`
- Solution (ZIP архив)

### BossDecision
- TaskID, ManagersCount, ManagerRoles (JSON)
- TechnicalDescription, TechStack (JSON), ArchitectureNotes

### Manager
- TaskID, Role, Status
- WorkerRoles (JSON), WorkersCount

### Worker
- TaskID, ManagerID, Role, Status
- TaskMD, SolutionMD
- Files (JSON), Success, Approved, Feedback

## 🔧 Разработка

### Генерация proto кода
```powershell
# PowerShell — генерация для всех сервисов
.\generate-proto.ps1
```

Или вручную:
```bash
cd agents && protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --proto_path=proto proto/agents.proto

cd boss && protoc ... proto/boss-manager.proto
cd manager && protoc ... proto/boss-manager.proto proto/manager-worker.proto
cd worker && protoc ... proto/manager-worker.proto
```

### Локальный запуск без Docker
```bash
# Terminal 1 — PostgreSQL
docker run -d --name crewai-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=crewai \
  -p 5432:5432 postgres:15-alpine

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

## ⚠️ Важные замечания

1. **API-ключи обязательны** — хотя бы один провайдер должен быть настроен
2. **PostgreSQL общий** для всех сервисов
3. **gRPC порты** не должны конфликтовать
4. **Миграции БД** запускаются автоматически при старте
5. **Бесплатные модели медленные** — `:free` модели на OpenRouter могут отвечать 1-5 минут. Весь пайплайн с 15+ воркерами займёт 30+ минут
6. **Для тестов** используйте простые задачи: «Маленький скрипт на go», «Один файл main.go»
7. **Платные модели дешёвые** — `qwen/qwen3-coder` обойдётся в ~5-10 центов за весь пайплайн

## 📄 Лицензия

MIT
