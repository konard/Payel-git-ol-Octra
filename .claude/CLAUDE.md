# Octra — VPS Aggregator for AI Agents

**Суть проекта:** Платформа, предоставляющая персональные MCP-эндпоинты с AI-агентами в изолированных Nix-окружениях. Пользователь получает API-ключ, создаёт окружение и общается с агентом через HTTP без аренды VPS.

## Стек технологий

| Компонент | Технология |
|-----------|-----------|
| Backend | Go 1.25, fasthttp, GORM |
| DB | PostgreSQL 15 |
| Cache | Redis 7 |
| Search | Typesense |
| Изоляция | Nix (per-user окружения) |
| Auth | JWT, OAuth2, bcrypt |
| Frontend | Next.js 16, React 18, TypeScript, @xyflow/react |
| TGBot | Python, aiogram, aiohttp, websockets |
| Poster | Python 3.11+, aiogram, asyncpg, OpenAI SDK |
| Workflow runtime | Crystal (ocawe) |

## Архитектура: Octra ↔ Ocawe

```
Пользователь → Octra (Go) → Ocawe (Crystal) → LLM / CLI / Workflow
                                                  ├── openai/gpt-4o
                                                  ├── anthropic/claude-sonnet-4
                                                  ├── cli/opencode
                                                  ├── workflow/orator
                                                  └── ...
```

**Octra** — VPS-хостинг для AI-агентов. Управляет пользователями, биллингом, Nix-окружениями, запуском Ocawe-процессов.

**Ocawe** — AI-движок и workflow-рантайм на Crystal. Принимает все AI-запросы от Octra и решает, куда их направить (LLM, CLI, workflow). Запускается как HTTP-сервер на случайном порту внутри каждого Nix-окружения пользователя.

**Deployment modes:**
- **OCAWE_ADDR set** — Octra подключается к удалённому Ocawe (например, `http://octra-ocawe:4111` в Docker)
- **OCAWE_ADDR not set** — Octra запускает `ocawecore` как подпроцесс на случайном порту через `OcaweLauncher`

## Структура проекта

```
octra/
├── backend/                    # Go-монолит
│   ├── cmd/server/main.go      # entrypoint — инициализация всего
│   └── internal/
│       ├── api/                # fasthttp handlers, middleware, routes
│       │   ├── api.go          # все эндпоинты, роутер
│       │   └── api_test.go     # интеграционные тесты (in-memory SQLite + фейковый ocawe)
│       ├── cli/                # управление процессом Ocawe (не CLIs!)
│       │   ├── manager.go      # Manager — реестр процессов, EnsureOcawe, OcawePort
│       │   ├── exec.go         # profileBinPaths, prependPath (для Nix-окружения)
│       │   ├── packages.go     # BuiltinCLIs() — каталог для фронтенда (не lifecycle!)
│       │   ├── state.go        # RedisStateStore — хранение состояния процесса
│       │   ├── ocawe_launcher.go  # OcaweLauncher — запуск ocawecore как subprocess
│       │   ├── ocawe_remote.go    # RemoteOcaweLauncher — подключение к удалённому Ocawe
│       │   └── manager_test.go    # тесты Manager
│       ├── config/             # env-based конфигурация
│       ├── model/              # GORM модели
│       │   └── model.go        # User, Agent, Skill, CLI, Provider, CanvasNode и др.
│       ├── service/            # бизнес-логика
│       │   ├── auth.go         # регистрация, логин, JWT, API-ключи
│       │   ├── environment.go  # создание/обновление окружений
│       │   ├── chat.go         # прокси на Ocawe /v1/chat/completions
│       │   ├── billing.go      # баланс, транзакции, usage-charge
│       │   ├── metrics.go      # метрики запросов (PR #117)
│       │   └── skill_sync.go   # синхронизация скиллов из skills.sh
│       ├── repository/         # GORM implementations
│       │   └── repository.go   # все репозитории (User, Agent, Skill, CLI, Provider, Metrics...)
│       ├── nix/                # управление Nix-окружениями
│       │   ├── nix.go          # CreateEnvironment, InstallSkill — без ProvisionSystem!
│       │   └── nix_test.go
│       ├── oauth/              # OAuth2 (Google, GitHub, LeFine)
│       ├── provider/           # каталог провайдеров
│       ├── llm/                # [УДАЛЁН] — прямой LLM-клиент, всё через Ocawe
│       ├── storage/            # PostgreSQL + Redis wiring
│       ├── skillapi/           # HTTP-клиент skills.sh API
│       └── typesense/          # Typesense search client
├── frontend/web/               # Next.js 16
│   └── app/
│       ├── page.tsx            # главная /app (canvas + чат + боковая панель)
│       ├── components/
│       │   ├── WorkflowCanvas.tsx    # Canvas на @xyflow/react
│       │   ├── EditNodeModal.tsx     # модалка редактирования ноды
│       │   ├── RequestMetricsCharts.tsx    # SVG-чарты метрик (PR #117)
│       │   └── RequestMetricsOverview.tsx  # компактный виджет метрик
│       ├── dashboard/          # дашборд
│       │   ├── page.tsx        # /dashboard — редирект
│       │   └── metrics/        # /dashboard/metrics — страница метрик
│       └── server/metrics.ts   # типизированный клиент GET /api/metrics/requests
├── ocawe/                      # Crystal workflow runtime
│   └── src/
│       ├── ocawe.cr            # entrypoint HTTP-сервера (--port, --workflows-root)
│       ├── cli/                # CLI-утилита ocawe (build, up)
│       └── framework/
│           ├── providers/      # AI-провайдеры
│           │   ├── client.cr         # маршрутизация provider/model → провайдер
│           │   ├── cli_provider.cr   # запуск CLI как подпроцесс (opencode и др.)
│           │   ├── chat_completions_provider.cr
│           │   ├── openai_provider.cr
│           │   └── gonka_provider.cr
│           ├── mcp/            # MCP-сервер менеджмент (CRUD + transport)
│           ├── workflows/      # декларативный workflow engine
│           │   └── declarative/
│           │       ├── service.cr   # start_run, resume_run, cancel_run, list_runs...
│           │       ├── engine.cr    # ядро движка
│           │       ├── types.cr     # RunStatus, NodeKind, EventType
│           │       ├── guardrails.cr # валидация ввода/вывода
│           │       └── acp_executor.cr  # ACP-протокол
│           ├── http/endpoints/ # ~40 HTTP-эндпоинтов
│           │   ├── compat.cr         # /v1/chat/completions (со streaming)
│           │   ├── mcp.cr            # /v1/mcp/* (серверы, каталог)
│           │   ├── workflows.cr      # /v1/workflows/*
│           │   ├── runs.cr           # /v1/workflows/:id/runs/*
│           │   ├── agents.cr         # /v1/agents/*
│           │   ├── tools.cr          # /v1/tools
│           │   ├── skills.cr         # /v1/skills/*
│           │   ├── datasets/         # /v1/datasets/*
│           │   ├── models.cr         # /v1/models
│           │   ├── triggers.cr       # /v1/triggers/*
│           │   ├── keys.cr           # /v1/keys/*
│           │   ├── hitl.cr           # /v1/hitl/* (Human-in-the-Loop)
│           │   └── federation*.cr    # ActivityPub/ForgeFed
│           ├── api/             # Cawfile struct API types
│           ├── registry/        # Registry API (NodeKind, функции, ресурсы)
│           ├── agents/          # агент-лоадер
│           ├── skills/          # скилл-лоадер
│           ├── discovery/       # поиск .caw файлов
│           ├── dataset/         # CRUD датасетов
│           └── acp/             # ACP-клиент
├── docker-compose.yml
├── docker-compose.prod.yml
├── PLAN.md                     # план интеграции Ocawe → Octra
└── .claude/CLAUDE.md           # этот файл
```

## Что изменилось: CLI lifecycle moved to Ocawe

**Было:** Octra управляла CLI-процессами напрямую — `cli/manager.go` имел `Send()`, `Process` interface с stdin/stdout, `llmEnv()` настраивал переменные окружения, `nix.ProvisionSystem()` устанавливал CLI через Nix.

**Стало:** Octra только запускает Ocawe-процесс. Ocawe через `CliProvider` сам управляет CLI-подпроцессами (opencode, claude-code, codex...). Octra шлёт HTTP-запросы в Ocawe с моделью `cli/opencode`, Ocawe запускает CLI и перехватывает его LLM-запросы.

**Удалено из Octra:**
- `cli/manager.go` — `Process.Send()`, `LLMConfig`, `LaunchSpec.CLI`, `Manager.Send()`
- `cli/exec.go` — `llmEnv()`
- `nix/nix.go` — `ProvisionSystem()`, `cliPackage()`, `CreateEnvironment()` больше не принимает `CLIType`
- `cmd/server/main.go` — фоновая Nix-провижининг CLI
- `internal/llm/` — весь пакет удалён

**Оставлено в Octra (каталог для фронтенда):**
- `cli/packages.go` — `BuiltinCLIs()`, `CLIPackage`
- `model.CLI` — модель в БД
- `repository.CLIRepository` — CRUD для каталога
- Эндпоинты `GET /api/cli`, `GET /api/cli/search`, `GET /api/catalog/search`
- `main.go` — `seedCLIs()` для заполнения каталога

## Основные модели данных

- **User** — ID, email, password_hash, api_key, subscription, balance, margin settings
- **Agent** — ID, user_id, LLM config (provider, api_key, base_url, model), CLI, active, priority
- **Skill** — ID, name, type (built-in/nixpkgs/custom), install_cmd, description, skill_id, source
- **CLI** — ID, name, nix_attr, install_cmd (каталог для фронтенда)
- **Provider** — ID, key, name, base_url, auth_env, default_model, description
- **UserSkill** — user_id, agent_id, skill_id, status (pending/installed/failed)
- **Transaction** — user_id, type (credit/debit), amount, reason (registration/hosting/topup/...), balance_after
- **UsageMetric** — user_id, date, cpu_seconds, memory_mb_hours, disk_mb, load_percent
- **DashboardEnvironment** — ID, user_id, name, visibility (private/public), active, building
- **CanvasNode** — environment_id, user_id, item_id, kind (provider/cli/...), name, meta (JSONB), position
- **UserAPIKey** — ID, user_id, name, key, expires_at
- **RequestMetric** — ID, user_id, environment_id (optional), success, latency_ms, timestamp (PR #117)

## Ключевые эндпоинты Octra API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | healthcheck |
| POST | `/register` | регистрация (+100 кредитов) |
| POST | `/login` | JWT-логин |
| POST | `/environment` | создание окружения (agent + nix + skills) |
| POST | `/api/chat` | чат с агентом (прокси на Ocawe) |
| POST | `/api/chat/environments/{id}` | чат для dashboard-окружения |
| POST | `/api/keys` | создать API-ключ (для MCP) |
| GET | `/api/keys` | список ключей |
| GET | `/api/cli` | список CLI (каталог) |
| GET | `/api/cli/search` | поиск CLI (Typesense или in-memory) |
| GET | `/api/providers` | список провайдеров |
| GET | `/api/catalog/search` | агрегированный поиск (providers+CLI+skills) |
| GET | `/skills/search` | поиск скиллов |
| GET | `/api/metrics/requests` | метрики запросов (PR #117) |
| GET/POST | `/v1/workflows/*` | прокси на Ocawe |
| GET | `/billing/balance` | баланс |
| POST | `/billing/topup` | пополнение |
| POST | `/billing/usage` | запись usage + debit |
| POST | `/api/environments` | создать dashboard-окружение |
| GET/PUT | `/api/environments/{id}/canvas` | Canvas-ноды |
| GET | `/auth/google|github|lefine` | OAuth2 |

## Ocawe: что НЕ используется Octra (план интеграции)

### 🔥 MUST HAVE — прямо сейчас
1. **Streaming** — Ocawe умеет `stream: true` SSE, Octra не передаёт флаг
2. **MCP Manager** — Ocawe: полный CRUD MCP-серверов + каталог инструментов. Octra: 0
3. **Workflow Runs API** — Ocawe: `start_run`, `resume_run`, `cancel_run`, `list_runs`. Octra: только pass-through прокси

### 👍 ПОТОМ
4. **HITL** — приостановка воркфлоу для ввода пользователя
5. **Triggers/Webhooks** — HTTP-триггеры для автоматизации

### 👎 НЕ НАДО
- Federation (ActivityPub) — оверкиллинг
- ACP Protocol — CliProvider покрывает тот же юзкейс
- Datasets API — непонятный юзкейс
- Keys API — у Octra своя система
- Docs/Swagger — Octra — публичный API, ocawe — нет
- RAG / Guardrails — только вместе с воркфлоу

## Как запустить тесты

```bash
cd backend && go test ./...          # SQLite in-memory, без внешних сервисов
cd backend && go test ./... -count=1 # без кеша
cd frontend/web && npm test          # frontend tests
```

## Важные замечания

- Тесты используют `github.com/glebarez/sqlite` (чистый Go SQLite, без CGO)
- Все тесты работают in-memory, без PostgreSQL/Redis
- Ocawe запускается как `ocawecore --port=0 --workflows-root=<path>`
- Nix-окружения создаются в `cfg.EnvironmentsDir/{userID}/`
- Состояние в Redis: `user:{id}:cli_state` (PID, port, start time)
- Ветка разработки: `create/integration-ocawe`
- После мержа в `master`: fast-forward

## Ссылки

- `PLAN.md` — план интеграции Ocawe → Octra
- `new/integration-ocawe-octra/ocawe+octra.md` — полный дизайн интеграции
- `docker-compose.yml` — полный стек (typesense, postgres, redis, ocawe, backend, frontend)
