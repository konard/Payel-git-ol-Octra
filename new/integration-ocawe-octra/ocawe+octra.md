# Octra + Ocawe: Интеграция

## Концепция

**Octra** — VPS-хостинг для AI-агентов. Каждый пользователь получает изолированное
Nix-окружение с предустановленным Ocawe (AI-движок) и набором скиллов, без
необходимости арендовать целый VPS, заходить по SSH и настраивать всё вручную.

**Ocawe** — AI-движок и workflow-рантайм. Умеет вызывать LLM-провайдеров
(OpenAI, Anthropic, Gemini и т.д.), запускать CLI как подпроцессы, исполнять
multi-agent воркфлоу (Cawfile), управлять MCP-серверами. Ocawe — это
**дефолтный AI-провайдер** для всех окружений Octra.

**Логика:**
- Octra управляет инфраструктурой: пользователи, биллинг, Nix-окружения, запуск
  и остановка Ocawe-процессов
- Ocawe делает всю AI-работу: сам решает, куда слать запрос — в LLM API, в CLI
  подпроцесс, или в workflow
- CLI (opencode, claude-code, codex) — это всего лишь **один из типов провайдеров**
  внутри Ocawe, а не основная функция
- Octra не вызывает LLM напрямую и не управляет CLI-процессами — всё через Ocawe

```
Пользователь → Octra (Go) → Ocawe (Crystal) → LLM / CLI / Workflow
                                                 ├── openai/gpt-4o
                                                 ├── anthropic/claude-sonnet-4
                                                 ├── cli/opencode
                                                 ├── workflow/orator
                                                 └── ...
```

---

## Конечная цель

Пользователь регистрируется в Octra, создаёт окружение через простой HTTP API
(или веб-интерфейс), указывает preferred LLM-провайдера и/или CLI, и сразу
получает рабочий AI-агент с MCP-энтрипоинтом.

Никакого SSH, никаких консолей, никаких ручных Nix-настроек. Octra даёт
"комнату в общежитии для AI-агентов" вместо аренды целого здания (VPS).

Ocawe — это "движок" этой комнаты. Он работает внутри окружения каждого
пользователя, принимает запросы от Octra и сам решает, как их обработать:
вызвать OpenAI, запустить opencode, исполнить Cawfile или скомбинировать всё
сразу.

**Модель провайдеров в Ocawe (формат):**
- `openai/gpt-4o` → OpenAI Chat Completions
- `anthropic/claude-sonnet-4-20250514` → Anthropic Messages API
- `gemini/gemini-2.5-pro` → Google Gemini
- `cli/opencode` → запуск CLI opencode как подпроцесс
- `cli/claude-code` → запуск CLI claude-code как подпроцесс
- `workflow/orator` → исполнение Cawfile-воркфлоу

В перспективе:
- Пользователь может собирать multi-agent пайплайны визуально через Canvas
  (фронтенд Octra) и запускать их через Ocawe как workflow
- Ocawe выступает единым AI-бэкендом для всех окружений
- Octra масштабируется горизонтально, запуская Ocawe рядом с каждым
  пользовательским Nix-окружением

---

## Задачи

### 0. Исследовать конфигурацию каждого CLI

Перед интеграцией нужно понять, через какие переменные окружения каждый CLI
настраивается на кастомного провайдера (base URL, API key, модель). Это нужно,
чтобы Ocawe правильно проксировал запросы.

#### opencode

| Переменная | Описание |
|------------|----------|
| `ANTHROPIC_API_KEY` | API-ключ Anthropic |
| `ANTHROPIC_BASE_URL` | Базовый URL для Anthropic (если не стандартный) |
| `OPENAI_API_KEY` | API-ключ OpenAI |
| `OPENAI_BASE_URL` | Базовый URL для OpenAI-compatible эндпоинта |
| `GEMINI_API_KEY` | API-ключ Google Gemini |
| `OPENCODE_CONFIG` | Путь к кастомному конфиг-файлу |
| `OPENCODE_CONFIG_CONTENT` | Инлайн-конфиг (JSON строка) |

Opencode также поддерживает `/connect` (интерактивный выбор провайдера) и
`opencode auth login`. Приоритет: переменные окружения > конфиг > интерактив.

**Для Ocawe:** установить `ANTHROPIC_BASE_URL=http://localhost:<port>/v1`
и `ANTHROPIC_API_KEY=ocawe_internal` — opencode будет слать все запросы
в Ocawe, а Ocawe проксирует в реального провайдера.

#### claude-code

| Переменная | Описание |
|------------|----------|
| `ANTHROPIC_API_KEY` | Стандартный API-ключ Anthropic |
| `ANTHROPIC_AUTH_TOKEN` | Raw Bearer token для кастомного base_url |
| `ANTHROPIC_BASE_URL` | Эндпоинт для всех запросов CLI |
| `ANTHROPIC_MODEL` | Модель (если нужен оверрайд) |

Также поддерживает `claude auth login` (OAuth). Настройки можно прописать в
`~/.claude/settings.json` через `env`-блок.

Ключевой момент: `ANTHROPIC_AUTH_TOKEN` имеет приоритет над `ANTHROPIC_API_KEY`
при использовании кастомного `ANTHROPIC_BASE_URL`.

**Для Ocawe:** установить `ANTHROPIC_BASE_URL=http://localhost:<port>`
(без `/v1` — claude-code сам добавляет) и `ANTHROPIC_AUTH_TOKEN=ocawe_internal`.

#### codex

| Переменная | Описание |
|------------|----------|
| `OPENAI_API_KEY` | API-ключ |
| `OPENAI_BASE_URL` | Базовый URL для OpenAI-compatible эндпоинта |
| `CODEX_HOME` | Корневая директория для состояния (по умолч. `~/.codex`) |

Настройки также можно задать в `~/.codex/config.toml` через
`model_provider_config.openai_base_url`.

**Для Ocawe:** установить `OPENAI_BASE_URL=http://localhost:<port>/v1`
и `OPENAI_API_KEY=ocawe_internal`.

#### antigravity (agy)

| Переменная | Описание |
|------------|----------|
| `GOOGLE_GEMINI_BASE_URL` | Gateway root URL (заменяет дефолтный Google) |
| `GEMINI_API_KEY` | API-ключ |
| `GEMINI_MODEL` | Модель по умолчанию |

Использует сервис root (не полный API path) — antigravity сам добавляет пути.

**Для Ocawe:** установить `GOOGLE_GEMINI_BASE_URL=http://localhost:<port>`
и `GEMINI_API_KEY=ocawe_internal`.

#### cline

| Переменная | Описание |
|------------|----------|
| `ANTHROPIC_API_KEY` | API-ключ Anthropic |
| `OPENAI_API_KEY` | API-ключ OpenAI |
| `CLINE_API_KEY` | API-ключ Cline API |

Cline поддерживает OpenAI-compatible эндпоинты. Настройка через
`cline auth` или переменные окружения. Можно задать кастомного провайдера
через `--provider custom` и `--model` флаги или через конфиг.

**Для Ocawe:** установить `ANTHROPIC_API_KEY=ocawe_internal` и базовый URL
через OpenAI-compatible адрес Ocawe, либо через конфиг Cline указать
`base_url=http://localhost:<port>/v1`.

#### openhands

| Переменная | Описание |
|------------|----------|
| `LLM_API_KEY` | API-ключ для LLM |
| `LLM_MODEL` | Модель (с префиксом, напр. `openhands/claude-sonnet-4`) |
| `LLM_BASE_URL` | Базовый URL (по умолчанию игнорируется, нужен флаг `--override-with-envs`) |

Конфигурация хранится в `~/.openhands/`. По умолчанию env-переменные
игнорируются — нужно передавать `--override-with-envs`.

**Для Ocawe:** установить `LLM_API_KEY=ocawe_internal`,
`LLM_MODEL=openai/gpt-4o`, `LLM_BASE_URL=http://localhost:<port>/v1`
и запускать openhands с флагом `--override-with-envs`.

#### hermes

| Переменная | Описание |
|------------|----------|
| `OPENAI_API_KEY` | API-ключ для кастомного OpenAI-compatible эндпоинта |
| `OPENAI_BASE_URL` | Базовый URL для кастомного эндпоинта |
| `OPENROUTER_API_KEY` | API-ключ OpenRouter |
| `HERMES_PORTAL_BASE_URL` | Override URL для Nous Portal |

Конфигурация через `~/.hermes/config.yaml` и `~/.hermes/.env`.
Поддерживает кастомные эндпоинты через `base_url` в конфиге.

**Для Ocawe:** установить `OPENAI_BASE_URL=http://localhost:<port>/v1`
и `OPENAI_API_KEY=ocawe_internal` в окружении процесса hermes.

---

### 1. Ocawe — дефолтный AI-провайдер при создании окружения

При создании окружения (`POST /environment`) Octra:
- Создаёт Nix-окружение пользователя
- Устанавливает Ocawe (бинарник `ocawecore`)
- Запускает `ocawecore` как HTTP-сервер на случайном порту
- Передаёт Ocawe настройки LLM (API-ключи, base_url, модель) через переменные
  окружения либо через его REST API (`POST /v1/mcp/servers`)
- Сохраняет порт Ocawe в Redis: `user:{id}:ocawe_port`

Ocawe работает **по умолчанию** — пользователь не выбирает "использовать Ocawe
или нет", Ocawe есть всегда. Пользователь выбирает только **backned внутри Ocawe**:
какой LLM провайдер или какой CLI запустить.

### 2. Octra запускает Ocawe как HTTP-сервер

Сейчас Octra запускает CLI через stdin/stdout пайп (`cli/exec.go`).
Нужен новый Launcher, который запускает полноценный HTTP-сервер Ocawe:

- Новый `OcaweLauncher` в `cli/` (или модификация `ExecLauncher`)
- Запускает `ocawecore --port=0 --workflows-root=<envPath>/workflows`
- Определяет порт старта (вычитывает из stdout)
- Сохраняет в Redis: `user:{id}:ocawe_port`
- При остановке: SIGTERM процессу Ocawe
- health-check: `GET /health`

### 3. Переписать ChatService на прокси через Ocawe

**Файл:** `backend/internal/service/chat.go`

Вместо:
```go
if agent.CLI != "" {
    cliMgr.Send(spec, prompt)
} else {
    llm.Complete(req)
}
```

Сделать:
```go
// Все запросы всегда идут через Ocawe
ocaweURL := fmt.Sprintf("http://localhost:%d/v1/chat/completions", ocawePort)
body := map[string]any{
    "model":    resolvedModel,  // "openai/gpt-4o" | "cli/opencode" | etc.
    "messages": []map[string]string{{"role": "user", "content": prompt}},
}
response := http.Post(ocaweURL, jsonBody)
```

Где `resolvedModel` формируется так:
- Если пользователь указал LLM-провайдера: `<provider>/<model>` (напр. `openai/gpt-4o`)
- Если пользователь указал CLI: `cli/<name>` (напр. `cli/opencode`)
- Если не указал ничего: из дефолтных настроек Ocawe

### 4. Удалить прямой LLM-клиент из Octra

**Файл:** `backend/internal/llm/llm.go`

Ocawe сам вызывает LLM. Octra не нужен прямой LLM-клиент. Пакет удалить.

### 5. Упростить CLI-менеджер Octra

**Файлы:** `backend/internal/cli/manager.go`, `exec.go`, `state.go`

Оставить только запуск Ocawe-сервера. Управление CLI-подпроцессами переходит
в Ocawe (через CLI-провайдер внутри Ocawe).

### 6. Добавить в Ocawe CLI-провайдер

**Новый файл:** `src/framework/providers/cli_provider.cr`

Провайдер, который:
- Принимает модель вида `cli/opencode`, `cli/claude-code`, `cli/codex`
- Запускает указанный CLI как долгоживущий подпроцесс
- Держит процесс в памяти, перезапускает при падении
- Проксирует промпты в stdin, читает ответ из stdout
- Поддерживает TTL и таймауты
- Использует переменные окружения пользователя (API-ключи) проксированные
  Octra в окружение процесса

**Файл:** `src/framework/providers/client.cr`

Добавить в `provider_for`:
```crystal
when "cli"
  CliProvider.new
```

### 7. Настройка CLI на использование Ocawe как LLM-провайдера

Когда пользователь подключает CLI (opencode, claude-code и т.д.), CLI нужно
настроить так, чтобы он слал LLM-запросы в Ocawe, а не напрямую провайдеру.
Это делается через переменные окружения CLI.

**Принцип:**
```
Запрос пользователя → Octra → Ocawe (cli/opencode)
                                    │
                                    ↓
                              Ocawe запускает opencode
                              как подпроцесс с настройками:
                                ANTHROPIC_BASE_URL = http://localhost:<ocawe_port>
                                ANTHROPIC_API_KEY  = ocawe_internal
                                OPENAI_BASE_URL    = http://localhost:<ocawe_port>
                                OPENAI_API_KEY     = ocawe_internal
                                    │
                                    ↓
                              Opencode думает, что шлёт запрос в Anthropic/OpenAI
                              Но на самом деле шлёт в Ocawe
                                    │
                                    ↓
                              Ocawe получает запрос, проверяет внутренний ключ,
                              проксирует в реального провайдера (OpenAI/Anthropic/...)
                                    │
                                    ↓
                              Ответ идёт обратно: провайдер → Ocawe → CLI → Ocawe → Octra → пользователь
```

**Что это даёт:**
- API-ключ пользователя не попадает в CLI — он остаётся в Ocawe
- Octra/Ocawe контролирует биллинг (считает токены, таймауты, лимиты)
- Можно добавлять guardrails, логирование, модерацию между CLI и LLM
- Единая точка контроля всех AI-запросов

**Реализация в CLI-провайдере Ocawe (`cli_provider.cr`):**
```crystal
# При запуске CLI как подпроцесса устанавливаем переменные окружения,
# которые перенаправляют LLM-запросы CLI обратно в Ocawe
ENV["ANTHROPIC_BASE_URL"] = "http://localhost:#{ocawe_port}"
ENV["ANTHROPIC_API_KEY"]  = ocawe_internal_key
ENV["OPENAI_BASE_URL"]    = "http://localhost:#{ocawe_port}"
ENV["OPENAI_API_KEY"]     = ocawe_internal_key
ENV["ANTHROPIC_MODEL"]    = ""       # Ocawe сам решает какую модель использовать
ENV["OPENAI_MODEL"]       = ""
```

Ocawe должен уметь принимать эти запросы от CLI (на своём же `/v1/chat/completions`
или `/v1/messages`) и проксировать их в реального провайдера, используя
настоящие API-ключи пользователя.

### 8. Выпилить Nix-установку и управление CLI из Octra

Сейчас `backend/internal/nix/nix.go` устанавливает CLI в Nix-профиль, а
`backend/internal/cli/` управляет процессами. После интеграции:

- Установка CLI в Nix остаётся (чтобы бинарник был доступен в окружении)
- Но управление процессом CLI полностью переходит в Ocawe
- Octra только говорит Ocawe: "для этого окружения используй cli/opencode"
- Ocawe сам запускает `opencode` из PATH (который настроен Octra через Nix)

### 9. Octra передаёт конфигурацию LLM в Ocawe

При старте Ocawe в Nix-окружении Octra устанавливает переменные окружения,
которые Ocawe уже читает:

```
OPENAI_API_KEY=sk-...
OPENAI_BASE_URL=https://api.openai.com/v1
ANTHROPIC_API_KEY=sk-...
ANTHROPIC_BASE_URL=https://api.anthropic.com
GEMINI_API_KEY=...
DEEPSEEK_API_KEY=...
```

Ocawe использует их в своих провайдерах.

### 10. Адаптировать API Octra под новую схему

**Файл:** `backend/internal/api/api.go`

- `/api/chat` — прокси на Ocawe `/v1/chat/completions`, контракт для
  пользователя не меняется
- `/environment` — при создании окружения запускает Ocawe и конфигурирует
  его (передаёт LLM-настройки, регистрирует CLI-провайдер если нужно)
- Добавить эндпоинты `/v1/workflows/*` как прокси на Ocawe (для будущих
  воркфлоу)

### 11. Фронтенд: Canvas → Cawfile

Когда фронтенд Octra (WorkflowCanvas) будет собирать multi-agent пайплайны,
конвертировать ноды Canvas в Cawfile и отправлять в Ocawe через
`POST /v1/workflows`.

### 12. Билд Ocawe внутри Nix-окружения

Убедиться, что Crystal-компилятор доступен через Nix. Добавить в
nixpkgs-оверлей для Octra. Либо добавить бинарник Ocawe в nixpkgs.

---

## Что остаётся в Octra (не меняется)

| Компонент | Статус |
|-----------|--------|
| Auth (регистрация, логин, API-ключи) | остаётся |
| Billing (баланс, транзакции, usage) | остаётся |
| Nix-менеджер (создание окружений) | остаётся (добавляется сборка Ocawe) |
| API-роутинг, мидлвари, OAuth | остаётся |
| Фронтенд (регистрация, дашборд, canvas) | остаётся |
| Poster / TGBot | остаётся |

## Что меняется в Octra

| Компонент | Изменение |
|-----------|-----------|
| `service/chat.go` | полная переработка: прокси на Ocawe вместо прямой LLM/CLI |
| `llm/llm.go` | удалить — Ocawe сам вызывает LLM |
| `cli/manager.go` | упрощается: только запуск Ocawe-сервера |
| `cli/exec.go` | новый OcaweLauncher вместо stdin/stdout пайпа |
| `api/api.go` | адаптировать роутинг под прокси на Ocawe |

## Что добавляется в Ocawe

| Компонент | Описание |
|-----------|----------|
| `providers/cli_provider.cr` | Провайдер для запуска CLI как подпроцесса (opencode, claude-code и т.д.) |
| `providers/client.cr` | + маршрутизация на `cli/` |
