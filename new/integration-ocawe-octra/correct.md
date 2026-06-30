# Octra + Ocawe: correct

```
Пользователь → Octra → Ocawe → LLM / CLI
```

## Octra (инфраструктура)

- Регистрация / аутентификация
- Биллинг / баланс
- Nix-окружение (изолированная папка + софт)
- Запуск Ocawe как HTTP-сервера внутри окружения
- Прокси всех AI-запросов на Ocawe

## Ocawe (AI-движок)

Принимает запросы через `POST /v1/chat/completions`.

Если `model = "openai/gpt-4o"` — шлёт в OpenAI через `OPENAI_API_KEY` из env.

Если `model = "cli/opencode"` — запускает opencode как подпроцесс,
пишет промпт в stdin, читает ответ из stdout. CLI при этом настроен слать
свои LLM-запросы обратно в Ocawe (`ANTHROPIC_BASE_URL = localhost:port`).

## Request flow

```
POST /api/chat {prompt, model}
  → Octra: проверил Auth, биллинг
  → Octra: POST http://localhost:<port>/v1/chat/completions {model, messages}
    → Ocawe: model = "openai/gpt-4o"
      → ChatCompletionProvider: шлёт в OpenAI
      → OpenAI → ответ
    → или: model = "cli/opencode"
      → CliProvider: запускает opencode, пишет в stdin
      → Opencode → ответ
    → Ocawe возвращает ответ
  → Octra возвращает пользователю
```

## Что меняется в Octra

- `service/chat.go` — полная переработка: прокси на Ocawe вместо прямой LLM/CLI
- `llm/llm.go` — удалить
- `cli/` — упрощается: только запуск Ocawe-сервера
- `api/api.go` — адаптировать роутинг

## Что уже есть в Ocawe

- `providers/chat_completions_provider.cr` — вызов LLM через OpenAI API
- `providers/cli_provider.cr` — запуск CLI как подпроцесса
- `cli_provider.cr` уже умеет запускать все CLI из Octra (opencode, claude-code, codex, antigravity, cline, openhands, hermes)
- `cli_provider.cr` уже сам настраивает env-переменные CLI на Ocawe
- Читает `OCAWE_PORT` из окружения (выставляется Octra при запуске)
