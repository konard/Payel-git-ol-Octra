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
- **7 провайдеров**: OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok, **CLIProxyAPI (Qwen Code)**
- **Per-request токены** — API-ключ передаётся в каждом запросе, не хранится на сервере
- **Streaming** (`GenerateStream`) — потоковая генерация ответа
- **Retry-логика** — 3 попытки при transient-ошибках (EOF, timeout, connection reset)
- Все остальные сервисы обращаются к LLM **только через Agents**

### CLIProxyAPI (:8317) — Опционально
Локальный прокси-сервер для использования Qwen Code CLI через OAuth (бесплатно):
- **Бесплатный доступ** к Qwen Code без API-ключей через OAuth авторизацию
- **OpenAI-совместимый API** — интегрируется как провайдер `cliproxy` в Agents Service
- **Локальный запуск** — работает на твоей машине, нет rate limits
- Запуск: `.\start-cliproxy.ps1` (PowerShell)
- Настройка: `cliproxy/config.yaml` + OAuth авторизация через Management API

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
| **CLIProxyAPI** | не требуется | `qwen-code` | ✅ **Да, через OAuth** |
| **OpenRouter** | `"openrouter": "sk-or-v1-..."` | `qwen/qwen3.6-plus:free` | ✅ Да |
| **Gemini** | `"gemini": "AIzaSy..."` | `gemini-2.5-flash` | ✅ 20 запросов/мин |
| **OpenAI** | `"openai": "sk-..."` | `gpt-4o` | ❌ |
| **Claude** | `"claude": "sk-ant-..."` | `claude-opus-4-6` | ❌ |
| **DeepSeek** | `"deepseek": "sk-..."` | `deepseek-chat` | ❌ (дешёвый) |
| **Grok** | `"grok": "xai-..."` | `grok-3` | ❌ |

### 🎯 Рекомендуемые бесплатные модели:
- `qwen-code` (через CLIProxyAPI) — **лучший вариант для тестов**, полностью бесплатный ⭐
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

1. **API-ключи или CLIProxyAPI** — нужен хотя бы один провайдер (рекомендуется CLIProxyAPI для тестов)
2. **PostgreSQL общий** для всех сервисов
3. **gRPC порты** не должны конфликтовать
4. **Миграции БД** запускаются автоматически при старте
5. **Бесплатные модели медленные** — `:free` модели на OpenRouter могут отвечать 1-5 минут. Весь пайплайн с 15+ воркерами займёт 30+ минут
6. **Для тестов** используйте простые задачи: «Маленький скрипт на go», «Один файл main.go»
7. **Платные модели дешёвые** — `qwen/qwen3-coder` обойдётся в ~5-10 центов за весь пайплайн

---

## 🆓 Бесплатный ИИ через CLIProxyAPI + Qwen Code

### Что такое CLIProxyAPI?

[CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) — это прокси-сервер, который оборачивает CLI-инструменты (включая **Qwen Code CLI**) в OpenAI-совместимый API. Это позволяет использовать Qwen Code **бесплатно** через OAuth авторизацию, без API-ключей.

### Преимущества

✅ **Полностью бесплатно** — Qwen Code работает через OAuth, без API-ключей  
✅ **Локальный запуск** — работает на твоей машине, нет rate limits  
✅ **OpenAI-совместимый API** — интегрируется как провайдер `cliproxy`  
✅ **Идеально для тестов** — не нужно тратить деньги на API-ключи  

### Быстрый старт с CLIProxyAPI

#### 1. Установка Qwen Code CLI

```bash
# Установка Qwen Code CLI (требуется Node.js 18+)
npm install -g @qwen/code
```

#### 2. Авторизация Qwen Code

```bash
# Вход через OAuth (откроется браузер)
qwen login
```

#### 3. Запуск CLIProxyAPI через Docker Compose

CLIProxyAPI уже настроен в твоём `docker-compose.yml`:

```bash
# Запустить все сервисы включая CLIProxyAPI
docker-compose up -d --build
```

CLIProxyAPI запустится на порту **8317** и будет доступен для Agents Service.

#### 4. Проверка работоспособности

```bash
# Проверить что CLIProxyAPI работает
curl http://localhost:8317/v1/models

# Должен вернуть список моделей включая qwen-code
```

#### 5. Использование в запросах

Теперь можно использовать провайдер `cliproxy` без API-ключей:

```json
{
  "userId": "user-123",
  "username": "pavel",
  "title": "Тестовая задача",
  "description": "Напиши простой HTTP сервер на Go",
  "tokens": {},  // ключи НЕ нужны!
  "meta": {
    "provider": "cliproxy",
    "model": "qwen-code"
  }
}
```

### Ручной запуск CLIProxyAPI (без Docker)

Если хочешь запустить CLIProxyAPI отдельно:

```bash
# Клонировать репозиторий (уже сделано в cliproxy-temp/)
cd cliproxy-temp

# Собрать из исходников
go build -o cliproxy

# Запустить с конфигом
./cliproxy --config ../cliproxy/config.yaml
```

### Конфигурация CLIProxyAPI

Файл `cliproxy/config.yaml` уже настроен:

```yaml
port: 8317
api-keys:
  - "cliproxy-dev-key-change-me"

oauth-model-alias:
  qwen:
    - name: "qwen3-coder-plus"
      alias: "qwen-code"
```

### Первичная настройка OAuth

При первом запуске CLIProxyAPI потребуется авторизация:

```bash
# Вариант 1: Через Docker
docker exec -it crewai-cliproxy /bin/sh
# Внутри контейнера:
qwen login

# Вариант 2: Локально (если CLIProxyAPI запущен без Docker)
qwen login
```

После авторизации OAuth токены сохранятся в volume `cliproxy_auth` и будут переиспользоваться.

### Troubleshooting

**CLIProxyAPI не запускается:**
```bash
# Проверить логи
docker logs crewai-cliproxy

# Проверить что порт свободен
netstat -an | findstr 8317
```

**OAuth ошибка:**
```bash
# Переавторизоваться
docker exec -it crewai-cliproxy qwen login

# Или удалить volume и начать заново
docker-compose down -v
docker-compose up -d cliproxy
docker exec -it crewai-cliproxy qwen login
```

**Agents Service не видит CLIProxyAPI:**
```bash
# Проверить переменную окружения
docker exec crewai-agents env | grep CLIPROXY

# Должно быть: CLIPROXY_API_URL=http://cliproxy:8317/v1
```

---

## 📄 Лицензия

MIT
