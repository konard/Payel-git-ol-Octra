# Octra — VPS Aggregator

**Суть проекта:** Платформа, предоставляющая персональные MCP-эндпоинты с AI-агентами (Claude Code, OpenCode, Codex) в изолированных Nix-окружениях. Пользователь получает API-ключ, создаёт окружение с CLI и навыками, и общается с агентом через HTTP без аренды VPS.

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

## Структура проекта

```
octra/
  backend/             — Go-монолит (cmd/server/ — entrypoint)
    internal/
      api/             — fasthttp handlers, middleware, routes
      config/          — env-based конфигурация
      model/           — GORM модели (User, Agent, Skill, UserSkill, Transaction, UsageMetric)
      service/         — бизнес-логика (auth, environment, chat, billing)
      repository/      — persistence interfaces + GORM implementations
      nix/             — управление Nix-окружениями
      cli/             — управление долгоживущими процессами AI CLI (Redis state)
      llm/             — прямой LLM прокси (Anthropic Messages API)
      storage/         — DB + Redis wiring
  frontend/web/        — Next.js фронтенд (canvas + чат)
  tgbot/               — Telegram-бот
  poster/              — бот для постов о коммитах
  ocawe/               — Crystal-воркфлоу-рантайм
  docs/                — документация
  docker-compose.yml
  docker-compose.prod.yml
```

## Основные модели данных

- **User** — ID, email, password hash, api_key, subscription, balance, margin settings
- **Agent** — ID, user ID, LLM config, CLI, active, priority
- **Skill** — ID, name, type (built-in/nixpkgs/custom), install command, description
- **UserSkill** — user ID, agent ID, skill ID, install status
- **Transaction** — user ID, type, amount, reason, balance after
- **UsageMetric** — user ID, date, CPU seconds, memory MB-hours, disk MB, load percent

## Как работает

1. Регистрация → api_key + 100 бесплатных кредитов
2. POST /environment → создание Nix-окружения с CLI + skills
3. POST /api/chat → общение с агентом через MCP

Управление процессами: CLI запускается как subprocess внутри Nix-окружения, переиспользуется между запросами. Состояние в Redis: `user:{id}:cli_state` (PID, port, start time) и `user:{id}:cli_ttl`. При таймауте — перезапуск.

Биллинг: запись CPU/memory/disk usage, дебетование кредитов. Unlimited margin mode (баланс может уходить в минус). Safe margin mode с лимитом.

## Быстрый старт

```bash
docker compose up -d --build
cd backend && go test ./...   # SQLite in-memory, без внешних сервисов
```

## Ссылки

- `CONCEPT_USER_PAYMENT.md` — дизайн платежей
- `NEW_CONCEPT_USER_BALANCE.md` — дизайн баланса
