# Алгоритм использования промптов

## Проблема
Промпты сейчас отправляются кучей — все сразу, без учёта задачи. Boss получает тонну контекста, большая часть которого нерелевантна. Это засоряет контекстное окно, увеличивает токены, снижает качество.

## Решение: гибридная фильтрация промптов (детерминизм + AI)

Два слоя фильтрации:

**Слой 1 — Детерминированный (код):** триггер-слова в заголовке/описании задачи отсекают заведомо ненужные промпты и подключают релевантные гайды.

**Слой 2 — AI выбирает из суженного списка.** Модель получает не 50 вариантов, а 3–5, и выбирает нужные под конкретную задачу.

AI не получает полной свободы — он выбирает из контролируемого набора, который уже отфильтрован кодом.

---

## Слой 1: Детерминированные триггер-слова

```javascript
// Пример таблицы триггеров
const triggers = {
  // Языки и фреймворки
  nodejs:      { add_prompts: ["node_npm", "node_express"],   add_guides: ["backend-api"] },
  express:     { add_prompts: ["node_express"],                add_guides: ["backend-api"] },
  go:          { add_prompts: ["go_std"],                      add_guides: ["backend-api", "backend-concurrency"] },
  golang:      { add_prompts: ["go_std"],                      add_guides: ["backend-api", "backend-concurrency"] },
  "c#":        { add_prompts: ["dotnet"],                      add_guides: ["backend-api"] },
  "csharp":    { add_prompts: ["dotnet"],                      add_guides: ["backend-api"] },
  "asp.net":   { add_prompts: ["dotnet_asp"],                  add_guides: ["backend-api", "backend-security"] },
  python:      { add_prompts: ["python_std"],                  add_guides: ["backend-api"] },
  django:      { add_prompts: ["python_django"],               add_guides: ["backend-api", "backend-data"] },
  fastapi:     { add_prompts: ["python_fastapi"],              add_guides: ["backend-api"] },
  pandas:      { add_prompts: ["python_pandas"],               add_guides: ["backend-data"] },
  react:       { add_prompts: ["frontend_react"],              add_guides: ["frontend-component", "frontend-state"] },
  vue:         { add_prompts: ["frontend_vue"],                add_guides: ["frontend-component", "frontend-state"] },
  rust:        { add_prompts: ["rust_std"],                    add_guides: ["backend-concurrency", "backend-security"] },
  php:         { add_prompts: ["php_std"],                     add_guides: ["backend-api"] },
  laravel:     { add_prompts: ["php_laravel"],                 add_guides: ["backend-api", "backend-data"] },

  // Типы задач
  "issue":     { add_prompts: ["github_issue"],                set_type: "github" },
  "feature":   { add_prompts: ["github_issue"],                set_type: "github" },
  "bug":       { add_prompts: ["github_issue"],                set_type: "github" },
  "pull request":    { add_prompts: ["github_pr"],             set_type: "github" },
  "github.com/issues": { add_prompts: ["github_issue"],        set_type: "github" },

  // Типы контента
  presentation: { set_type: "presentation",                    add_guides: ["presentation-slide", "presentation-visual"] },
  slides:       { set_type: "presentation",                    add_guides: ["presentation-slide", "presentation-story"] },
  pptx:         { set_type: "presentation",                    add_guides: ["presentation-slide"] },
  research:     { set_type: "research",                        add_guides: ["research-sources", "research-facts"] },

  // Инфраструктура
  docker:       { add_guides: ["devops-containers"] },
  kubernetes:   { add_guides: ["devops-containers", "devops-observability"] },
  cicd:         { add_guides: ["devops-cicd"] },
  vpn:          { add_guides: ["vpn-security", "vpn-network"] },
  proxy:        { add_guides: ["proxy-headers", "proxy-security"] },
}
```

Триггер-слова сканируются в: `title`, `description`, `issue body` (если задача типа github).

---

## Слой 2: AI выбирает из суженного списка

После детерминированной фильтрации формируется структура:

```json
{
    "task_title": "Fix login button not working in Express.js app",
    "task_description": "When user clicks login button on /login page, nothing happens. Need to fix the handler.",
    "system_prompt": "<базовый системный промпт Octra>",
    "task_type": "github",
    "or_prompts": [
        "github_issue",
        "node_express"
    ],
    "available_guides": [
        "backend-api",
        "backend-security"
    ],
    "locked_prompts": [],
    "locked_guides": []
}
```

Поля:
- `or_prompts` — AI **может** выбрать (не обязан). Выбор означает: «добавь этот промпт в контекст моей работы».
- `available_guides` — AI **может** запросить эти гайды для справки.
- `locked_prompts` — промпты, которые **обязательно** попадают в контекст (например, базовый промпт босса/менеджера/воркера). AI не выбирает — они всегда включены.
- `locked_guides` — гайды, которые **обязательно** применяются (например, tech stack language mapping).

---

## Полный поток

```
[Input: Title + Description]
    ↓
AnalyzeTaskInput():
  Слой 1:
  1. Сканирует триггер-слова в title + description + issue body
  2. Определяет task_type (code / github / research / document / presentation)
  3. Собирает or_prompts из совпавших триггеров
  4. Собирает available_guides из совпавших триггеров
  5. Определяет tech_stack
    ↓
  Формирует структуру с or_prompts + available_guides + locked_prompts
    ↓
[LLM получает системный промпт + структуру]
    ↓
  Слой 2:
  6. LLM анализирует задачу и выбирает из or_prompts нужные
  7. LLM может запросить дополнительные гайды из available_guides
    ↓
[LLM возвращает выбранные промпты + гайды]
```

Пример ответа AI:

```json
{
    "reflection": "This is a Node.js Express bug fix from a GitHub issue. The login handler is missing form body parsing. I need the Express.js prompt for syntax reference and the GitHub issue prompt for PR workflow.",
    "selected_prompts": ["node_express"],
    "requested_guides": ["backend-api"]
}
```

Пример ответа системы:
```json
{
    "context": [
      "user": "Fix login button not working in Express.js app. When user clicks login button on /login page, nothing happens. Need to fix the handler.",
      "agent": "This is a Node.js Express bug fix from a GitHub issue. The login handler is missing form body parsing. I need the Express.js prompt for syntax reference and the GitHub issue prompt for PR workflow."
    ],
    "selected_promts": [
         ""
    ]
}
```


---

## Преимущества подхода

| Аспект | Сейчас (всё сразу) | Гибрид (триггеры + AI) |
|---|---|---|
| Токены в промпте | Все гайды независимо от задачи | Только выбранные |
| Релевантность | Средняя — AI фильтрует сам | Высокая — гайды под задачу |
| Контроль | AI может игнорировать гайды | AI выбирает из отфильтрованного |
| Гибкость | Никакой — всё жёстко | AI адаптируется под нетипичные задачи |
| Масштабирование | Каждый новый гайд = больше токенов | Можно добавить 100 гайдов — выберутся 2-3 |

---

## Как это ложится на текущую архитектуру Octra

**AnalyzeTaskInput()** — новый слой перед Boss:

- `internal/service/github/analyze.go` — сканирование триггер-слов
- `internal/service/prompts/selector.go` — формирование `or_prompts` + `available_guides`
- `internal/prompts/types.go` — структура `TaskPromptConfig`

Boss получает `TaskPromptConfig` в Meta и передаёт его Manager'у, Manager — Worker'у. Каждый этап видит только свои `or_prompts` (Boss — одни, Worker — другие).

---

## Границы ответственности

| Что определяет код | Что определяет AI |
|---|---|
| tech stack (по триггер-словам) | Какой промпт из or_prompts нужен |
| task_type (code/github/...) | Какие гайды запросить |
| or_prompts (суженный список) | |
| language mapping (lang.go) | |
| fork / clone решение | |
