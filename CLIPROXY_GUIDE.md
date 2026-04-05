# CLIProxyAPI + Qwen Code — Полное руководство

## 📋 Оглавление

1. [Что это такое?](#что-это-такое)
2. [Зачем это нужно?](#зачем-это-нужно)
3. [Установка](#установка)
4. [Настройка OAuth](#настройка-oauth)
5. [Запуск с CrewAI](#запуск-с-crewai)
6. [Использование](#использование)
7. [Troubleshooting](#troubleshooting)
8. [Архитектура интеграции](#архитектура-интеграции)

---

## Что это такое?

**CLIProxyAPI** — это прокси-сервер, который оборачивает CLI-инструменты (включая Qwen Code CLI) в OpenAI-совместимый API.

**Qwen Code** — это ИИ-ассистент для программирования от Alibaba Cloud, который можно использовать **бесплатно** через OAuth авторизацию.

### Как это работает вместе?

```
CrewAI Workers → Agents Service → CLIProxyAPI → Qwen Code CLI (OAuth) → Бесплатный ИИ!
```

CLIProxyAPI принимает запросы в формате OpenAI API и передаёт их Qwen Code CLI, который работает с твоей учётной записью через OAuth.

---

## Зачем это нужно?

### Проблема

Твой проект CrewAI использует несколько ИИ-провайдеров:
- OpenAI (платный)
- Claude (платный)
- OpenRouter (есть бесплатные модели, но ограниченные)
- Gemini (бесплатный, но с лимитами)

**Для тестирования и разработки** нужно тратить деньги на API-ключи.

### Решение

CLIProxyAPI + Qwen Code = **полностью бесплатный ИИ** для тестов!

### Преимущества

✅ **Бесплатно** — не нужны API-ключи, работает через OAuth  
✅ **Локально** — работает на твоей машине, нет rate limits от облачных провайдеров  
✅ **OpenAI-совместимый** — легко интегрируется  
✅ **Qwen Code мощный** — одна из лучших моделей для программирования  
✅ **Для тестов идеально** — можно запускать сколько угодно задач без затрат  

### Ограничения

⚠️ **Медленнее облачных API** — локальный CLI работает медленнее  
⚠️ **Требует авторизации** — нужно один раз войти через браузер  
⚠️ **Только для тестов** — для production лучше использовать облачные API  

---

## Установка

### Шаг 1: Установка Qwen Code CLI

Qwen Code CLI требуется для работы CLIProxyAPI.

```bash
# Требуется Node.js 18+
# Проверить версию Node.js
node --version

# Установить Qwen Code CLI глобально
npm install -g @qwen/code

# Проверить установку
qwen --version
```

### Шаг 2: Проверка Docker Compose

CLIProxyAPI уже настроен в твоём `docker-compose.yml`. Проверь что файл существует:

```bash
# Проверить конфигурацию
cat cliproxy/config.yaml
```

Файл должен содержать:

```yaml
port: 8317
api-keys:
  - "cliproxy-dev-key-change-me"

oauth-model-alias:
  qwen:
    - name: "qwen3-coder-plus"
      alias: "qwen-code"
```

---

## Настройка OAuth

### Шаг 1: Первичная авторизация

При первом запуске нужно авторизоваться через браузер:

#### Вариант A: Через Docker (рекомендуется)

```bash
# Запустить только CLIProxyAPI
docker-compose up -d cliproxy

# Выполнить авторизацию внутри контейнера
docker exec -it crewai-cliproxy qwen login
```

Откроется браузер с просьбой войти через Alibaba Cloud аккаунт.

#### Вариант B: Локально

Если CLIProxyAPI запущен без Docker:

```bash
qwen login
```

### Шаг 2: Проверка авторизации

```bash
# Проверить что OAuth токен сохранён
qwen whoami

# Должен вернуть информацию о твоём аккаунте
```

### Шаг 3: Проверка работы CLIProxyAPI

```bash
# Проверить что CLIProxyAPI видит модели
curl http://localhost:8317/v1/models

# Должен вернуть JSON со списком моделей включая "qwen-code"
```

---

## Запуск с CrewAI

### Полный запуск всех сервисов

```bash
# Запустить все сервисы включая CLIProxyAPI
docker-compose up -d --build

# Проверить что все работают
docker-compose ps
```

Должны быть запущены:
- `crewai-cliproxy` — CLIProxyAPI (порт 8317)
- `crewai-agents` — Agents Service (порт 50053)
- `crewai-boss` — Boss Service (порт 50051)
- `crewai-manager` — Manager Service (порт 50052)
- `crewai-worker` — Worker Service (порт 50053)
- `crewai-apigateway` — API Gateway (порт 3111)
- `crewai-postgres` — PostgreSQL (порт 5432)

### Проверка интеграции

```bash
# Проверить что Agents Service видит CLIProxyAPI провайдер
docker logs crewai-agents | grep -i "cliproxy"

# Должно быть что-то вроде:
# ✅ Available AI providers: [claude gemini openai deepseek grok openrouter cliproxy]
```

---

## Использование

### Через WebSocket API

Отправь задачу через WebSocket на `ws://localhost:3111/task/create`:

```json
{
  "userId": "user-123",
  "username": "pavel",
  "title": "Тестовая задача",
  "description": "Напиши простой HTTP сервер на Go с одним эндпоинтом /health",
  "tokens": {},
  "meta": {
    "provider": "cliproxy",
    "model": "qwen-code"
  }
}
```

**Важно:** Поле `tokens` может быть пустым — CLIProxyAPI не требует API-ключей!

### Через Node.js тестовый клиент

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:3111/task/create');

ws.on('open', () => {
  ws.send(JSON.stringify({
    userId: 'user-123',
    username: 'test',
    title: 'Simple Go Server',
    description: 'Write a basic HTTP server with /health endpoint',
    tokens: {},  // Не нужны ключи!
    meta: {
      provider: 'cliproxy',
      model: 'qwen-code'
    }
  }));
});

ws.on('message', (data) => {
  console.log('Progress:', JSON.parse(data));
});
```

### Сравнение с другими провайдерами

#### С API-ключами (OpenRouter):
```json
{
  "tokens": {
    "openrouter": "sk-or-v1-..."
  },
  "meta": {
    "provider": "openrouter",
    "model": "qwen/qwen3.6-plus:free"
  }
}
```

#### Без ключей (CLIProxyAPI):
```json
{
  "tokens": {},
  "meta": {
    "provider": "cliproxy",
    "model": "qwen-code"
  }
}
```

---

## Troubleshooting

### CLIProxyAPI не запускается

**Проблема:** Контейнер не стартует или сразу падает

```bash
# Проверить логи
docker logs crewai-cliproxy

# Проверить что порт 8317 свободен
netstat -an | findstr 8317
# или
lsof -i :8317

# Если порт занят — убить процесс
# Windows:
netstat -ano | findstr 8317
taskkill /PID <PID> /F

# Linux/Mac:
lsof -ti:8317 | xargs kill -9
```

### OAuth ошибка

**Проблема:** CLIProxyAPI не может авторизоваться

```bash
# Вариант 1: Переавторизоваться
docker exec -it crewai-cliproxy qwen login

# Вариант 2: Полная перезагрузка OAuth
docker-compose down -v  # Удалит volume с токенами
docker-compose up -d cliproxy
docker exec -it crewai-cliproxy qwen login
```

### Agents Service не видит CLIProxyAPI

**Проблема:** В логах Agents нет провайдера `cliproxy`

```bash
# Проверить переменную окружения
docker exec crewai-agents env | grep CLIPROXY

# Должно быть:
# CLIPROXY_API_URL=http://cliproxy:8317/v1

# Если нет — проверить docker-compose.yml
# Agents должен иметь:
# environment:
#   - CLIPROXY_API_URL=http://cliproxy:8317/v1
# depends_on:
#   - cliproxy
```

### CLIProxyAPI работает, но запросы не проходят

```bash
# Проверить что CLIProxyAPI отвечает
curl -v http://localhost:8317/v1/models

# Проверить что Agents может достучаться до CLIProxyAPI
docker exec crewai-agents wget -O- http://cliproxy:8317/v1/models

# Проверить конфигурацию
docker exec crewai-cliproxy cat /CLIProxyAPI/config.yaml
```

### Qwen Code CLI не установлен

```bash
# Проверить
qwen --version

# Если ошибка — установить
npm install -g @qwen/code

# Проверить что npm доступен
npm --version
node --version
```

### Медленные ответы

Это нормально для локального CLI. Qwen Code через OAuth работает медленнее облачных API.

**Рекомендации:**
- Для тестов используй простые задачи (1-2 файла)
- Не запускай сложные много-воркерные пайплайны
- Для production используй облачные API (OpenRouter, OpenAI, etc.)

---

## Архитектура интеграции

### Общая схема

```
┌─────────────────────────────────────────────────────────────────┐
│                         CrewAI System                           │
│                                                                 │
│  User → Apigateway → Boss → Manager → Worker                    │
│                                                  │               │
│                                                  ▼               │
│                                        ┌──────────────┐         │
│                                        │   Agents     │         │
│                                        │  Service     │         │
│                                        │  (gRPC :50053)│        │
│                                        └──────┬───────┘         │
│                                               │                 │
│                                    provider: cliproxy           │
│                                               │                 │
└───────────────────────────────────────────────┼─────────────────┘
                                                │
                                                ▼
┌───────────────────────────────────────────────────────────────┐
│                    CLIProxyAPI (:8317)                         │
│                                                                 │
│  OpenAI-compatible API endpoint                                 │
│  ↓                                                              │
│  Translates to Qwen Code CLI calls                              │
│  ↓                                                              │
│  Uses OAuth tokens for authentication                           │
└───────────────────────────┬───────────────────────────────────┘
                            │
                            ▼
┌───────────────────────────────────────────────────────────────┐
│                      Qwen Code CLI                             │
│                                                                 │
│  Local CLI tool authenticated via OAuth                         │
│  ↓                                                              │
│  Makes requests to Qwen API using your account                  │
│  ↓                                                              │
│  Returns AI-generated code                                      │
└─────────────────────────────────────────────────────────────────┘
```

### Поток запроса

1. **Worker** получает задачу сгенерировать код
2. **Worker** вызывает `Agents.Generate()` с `provider=cliproxy`
3. **Agents Service** маршрутизирует запрос к `CLIProxyProvider`
4. **CLIProxyProvider** отправляет HTTP POST на `http://cliproxy:8317/v1/chat/completions`
5. **CLIProxyAPI** получает запрос, использует OAuth токен для Qwen Code CLI
6. **Qwen Code CLI** генерирует ответ через твой аккаунт
7. **CLIProxyAPI** возвращает ответ в формате OpenAI API
8. **Agents Service** возвращает код обратно в **Worker**

### Код интеграции

#### Провайдер (`agents/internal/fetcher/providers/qwen/qwen.go`)

```go
type CLIProxyProvider struct {
    client *openai.Client  // OpenAI SDK с кастомным BaseURL
    config *models.ProviderConfig
}

func (p *CLIProxyProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
    // BaseURL из переменной окружения CLIPROXY_API_URL
    baseURL := os.Getenv("CLIPROXY_API_URL")
    if baseURL == "" {
        baseURL = "http://localhost:8317/v1"
    }

    // API-key не важен для CLIProxyAPI (использует OAuth)
    apiKey := "cliproxy-oauth"

    client := openai.NewClient(
        option.WithAPIKey(apiKey),
        option.WithBaseURL(baseURL),
    )
    return &client, nil
}
```

#### Docker Compose (`docker-compose.yml`)

```yaml
cliproxy:
  image: eceasy/cli-proxy-api:latest
  ports:
    - "8317:8317"
  volumes:
    - ./cliproxy/config.yaml:/CLIProxyAPI/config.yaml
    - cliproxy_auth:/root/.cli-proxy-api  # OAuth токены

agents:
  environment:
    - CLIPROXY_API_URL=http://cliproxy:8317/v1
  depends_on:
    - cliproxy
```

---

## Дополнительные ресурсы

- [CLIProxyAPI GitHub](https://github.com/router-for-me/CLIProxyAPI)
- [CLIProxyAPI Documentation](https://help.router-for.me/)
- [Qwen Code CLI](https://github.com/QwenLM/qwen-code)
- [CrewAI PROJECT.md](./PROJECT.md)
- [CrewAI SERVICES.md](./SERVICES.md)

---

## FAQ

**Q: CLIProxyAPI работает только с Qwen Code?**  
A: Нет, CLIProxyAPI поддерживает много CLI-инструментов: Gemini CLI, Claude Code, ChatGPT Codex, iFlow и другие. Но в CrewAI мы используем Qwen Code.

**Q: Можно ли использовать CLIProxyAPI для production?**  
A: Технически да, но для production лучше использовать облачные API с SLA и гарантированной производительностью. CLIProxyAPI идеален для тестов и разработки.

**Q: Что если OAuth токен истечёт?**  
A: CLIProxyAPI автоматически обновляет токены. Если что — просто выполни `qwen login` снова.

**Q: Сколько стоит Qwen Code через OAuth?**  
A: Полностью бесплатно! Qwen Code предоставляет бесплатный доступ через OAuth авторизацию.

**Q: Можно ли использовать несколько провайдеров одновременно?**  
A: Да! CrewAI поддерживает 7 провайдеров. Ты можешь использовать CLIProxyAPI для тестов и облачные API для production задач.

---

**Готово! 🎉**

Теперь у тебя есть бесплатный ИИ для тестирования CrewAI!
