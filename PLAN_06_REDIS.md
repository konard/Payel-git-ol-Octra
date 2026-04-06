# PLAN_06_REDIS — Redis для WebSocket стримов + Оптимизация токенов

## 📖 Оглавление

1. [Проблемы](#проблемы)
2. [Redis архитектура](#redis-архитектура)
3. [Оптимизация токенов](#оптимизация-токенов)
4. [План реализации](#план-реализации)
5. [Изменения в коде](#изменения-в-коде)

---

## Проблемы

### 1. WebSocket стримы хранятся в памяти

**Сейчас:**
- Apigateway хранит соединения в `map[*azurewebsockets.Conn]bool` (мёртвый код)
- Hub использует `gorilla/websocket` (несовместим с основным хендлером)
- При 100+ пользователях: память заканчивается, нет масштабирования

**Почему это проблема:**
- Невозможно запустить несколько подов Apigateway
- Потерянные соединения не восстанавливаются
- Нет истории обновлений для reconnect

### 2. Чрезмерный расход токенов

**Сейчас (N+1 подход в Worker):**
```
Воркер #1:
  1. "Какие файлы создать?" → ["main.go", "config.go", "README.md"]     ~800 tokens
  2. "Напиши main.go" → content1                                        ~1500 tokens
  3. "Напиши config.go" → content2                                      ~1500 tokens
  4. "Напиши README.md" → content3                                      ~1000 tokens
  ИТОГО: ~4800 tokens

Воркер #2:
  1. "Какие файлы создать?" → ["api.go", "handler.go"]                  ~800 tokens
  2. "Напиши api.go" → content1                                         ~1500 tokens
  3. "Напиши handler.go" → content2                                     ~1500 tokens
  ИТОГО: ~3800 tokens

ВСЕГО для 6 воркеров: ~25,000+ tokens
```

**Почему это проблема:**
- Каждый запрос к LLM = контекст + промпт + ответ
- Бесплатные модели медленные (1-5 мин на запрос)
- Платные модели дорогие (при 100 задачах/день = $50-100/день)

---

## Redis архитектура

### Структура данных

```
┌─────────────────────────────────────────────────────────────┐
│                    Redis Keys Structure                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│ 1. STREAM:{task_id}  (Hash)                                  │
│    └─ Состояние стрима задачи                                │
│    {                                                          
│      "status": "processing",      // processing|done|error   
│      "progress": 45,                                          
│      "message": "Manager working: backend",                   
│      "task_id": "uuid-123",                                   
│      "user_id": "user-456",                                   
│      "created_at": "2026-04-06T10:00:00Z",                    
│      "updated_at": "2026-04-06T10:05:00Z"                     
│    }                                                          
│                                                               │
│ 2. STREAM:{task_id}:UPDATES  (List)                          │
│    └─ Буфер последних 50 обновлений (для reconnect)          │
│    [                                                          
│      {"progress": 40, "message": "Starting backend..."},      
│      {"progress": 45, "message": "Manager working..."},       
│      ... (LTRIM хранит только 50 последних)                   
│    ]                                                          
│                                                               │
│ 3. WS:{task_id}  (Set)                                       │
│    └─ Активные WebSocket соединения (для scale-out)          │
│    [                                                          
│      "apigateway-1:conn-abc123",                             
│      "apigateway-2:conn-def456"                              
│    ]                                                          
│                                                               │
│ 4. CHANNEL:task:{task_id}  (Pub/Sub)                         │
│    └─ Канал для real-time обновлений                         │
│    Publish → Все Apigateway поды получают обновления          │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Поток с Redis

```
User → Apigateway #1
         │
         ├─ 1. Создаёт задачу → Boss.CreateTaskStream()
         ├─ 2. Boss стримит TaskUpdate
         │
         ├─ 3. Apigateway публикует в Redis:
         │     HSET STREAM:{task_id} {state}
         │     RPUSH STREAM:{task_id}:UPDATES {update}
         │     LTRIM STREAM:{task_id}:UPDATES -50 -1  # храним 50 последних
         │     PUBLISH CHANNEL:task:{task_id} {update}
         │
         └─ 4. Отправляет клиенту через WebSocket

User reconnect → Apigateway #2
         │
         ├─ 1. HGETALL STREAM:{task_id} → текущий статус
         ├─ 2. LRANGE STREAM:{task_id}:UPDATES 0 -1 → история
         ├─ 3. Подписывается на CHANNEL:task:{task_id} → новые
         └─ 4. SADD WS:{task_id} "apigateway-2:conn-xyz"
```

### Масштабирование

```
┌──────────┐    ┌───────────────┐    ┌──────────┐
│  User #1 │───►│ Apigateway #1 │    │  User #50│───►│ Apigateway #2 │
└──────────┘    └───────┬───────┘    └──────────┘    └───────┬───────┘
                        │                                     │
                        └─────────────┬───────────────────────┘
                                      │
                              ┌───────────────┐
                              │    Redis      │
                              │  Pub/Sub +    │
                              │  Streams      │
                              └───────┬───────┘
                                      │
                              ┌───────┴───────┐
                              │    Boss       │
                              │  (gRPC)       │
                              └───────────────┘
```

---

## Оптимизация токенов

### Стратегия: Объединение запросов (Single-Pass Generation)

**Было (N+1 подход):**
```
Запрос 1: "Какие файлы создать?" → JSON ["main.go", "config.go"]
Запрос 2: "Напиши main.go" → plain text
Запрос 3: "Напиши config.go" → plain text
```

**Станет (Single-Pass):**
```
Запрос 1: "Создай main.go и config.go.
Формат ответа:
=== FILE: main.go ===
<полный код main.go>
=== FILE: config.go ===
<полный код config.go>"

Результат: парсим по маркеру "=== FILE: {path} ==="
```

### Экономия

| Метрика | N+1 (было) | Single-Pass (стало) | Экономия |
|---------|------------|---------------------|----------|
| Запросов на воркера | 4-6 | 1-2 | **70-80%** |
| Токенов на воркера | ~4000 | ~2200 | **~45%** |
| Время на воркера | 4-20 мин | 1-5 мин | **~75%** |
| Всего на задачу (6 воркеров) | ~24,000 | ~13,200 | **~45%** |

### Промпт для Single-Pass генерации

```
You are a {role} developer. Role: {description}

TASK: {task_md}
CONTEXT FROM OTHER WORKERS:
{accumulated_context}

Create the following files: {file_list}

RETURN FORMAT (STRICT - follow exactly):
=== FILE: path/to/file1.go ===
<complete code for file1.go - no placeholders, no TODOs>
=== FILE: path/to/file2.go ===
<complete code for file2.go - no placeholders, no TODOs>
=== FILE: path/to/file3.go ===
<complete code for file3.go - no placeholders, no TODOs>

RULES:
1. Each file MUST start with "=== FILE: <path> ===" on its own line
2. File content MUST be complete code - no placeholders, no TODOs
3. Use proper imports and exports
4. Keep code compact but functional
5. Do NOT include markdown code fences around the entire response
6. If you can't create a file, skip it and move to the next one
```

### Парсинг ответа

```go
func parseMultiFileResponse(content string) map[string]string {
    files := make(map[string]string)
    
    // Регулярка: === FILE: path ===
    re := regexp.MustCompile(`=== FILE:\s+(.+?)\s+===\n`)
    matches := re.FindAllStringSubmatchIndex(content, -1)
    
    for i, match := range matches {
        path := content[match[2]:match[3]]
        
        // Контент от конца маркера до начала следующего (или конца строки)
        start := match[1]
        end := len(content)
        if i+1 < len(matches) {
            end = matches[i+1][0]
        }
        
        content := strings.TrimSpace(content[start:end])
        files[path] = stripMarkdownCodeBlock(content)
    }
    
    return files
}
```

### Fallback механизм

Если парсинг провалился (LLM вернул не в том формате):
```go
files := parseMultiFileResponse(response)
if len(files) == 0 {
    // Fallback: старый N+1 подход
    log.Warn("Multi-file parsing failed, falling back to N+1")
    return generateFilesOneByOne(ctx, fileList, context)
}
```

---

## План реализации

### Фаза 1: Redis инфраструктура

#### 1.1 Обновить docker-compose.yml
- Добавить сервис Redis (Alpine image)
- Добавить volume для persistence
- Настроить healthcheck

#### 1.2 Создать пакет Redis клиента для Apigateway
- `apigateway/internal/core/redis/client.go` — подключение, команды
- `apigateway/internal/core/redis/stream.go` — управление стримами
- `apigateway/internal/core/redis/pubsub.go` — pub/sub подписки

#### 1.3 Интегрировать Redis в WebSocket хендлер
- При получении TaskUpdate → публикация в Redis
- При reconnect → восстановление из Redis
- Добавить endpoint `/task/status?task_id=...` для проверки статуса

### Фаза 2: Оптимизация Worker

#### 2.1 Обновить генерацию файлов
- `worker/internal/service/worker/code.go` — новый `generateCodeMultiPass()`
- Добавить `parseMultiFileResponse()` 
- Добавить fallback на старый N+1

#### 2.2 Обновить промпты
- Создать единый промпт для генерации всех файлов
- Убрать отдельный запрос "Какие файлы создать?" (список файлов определяется в том же запросе)

#### 2.3 Тестирование
- Проверить что парсинг работает с разными форматами ответа
- Проверить fallback механизм
- Сравнить качество кода (N+1 vs Single-Pass)

### Фаза 3: Интеграция и тесты

#### 3.1 Обновить环境变量
- Добавить `REDIS_URL` в `.env.example`
- Обновить конфигурацию Apigateway

#### 3.2 Тесты масштабирования
- Запустить 2 пода Apigateway
- Проверить что reconnect работает на другом поде
- Проверить Pub/Sub рассылку

---

## Изменения в коде

### Файлы для создания

```
apigateway/
├── internal/core/redis/
│   ├── client.go           # Redis клиент + инициализация
│   ├── stream.go           # StreamStateManager
│   └── pubsub.go           # PubSub менеджер

worker/
├── internal/service/worker/
│   ├── multifile.go        # parseMultiFileResponse()
│   └── code_multpass.go    # generateCodeMultiPass()
```

### Файлы для изменения

```
docker-compose.yml          # + Redis сервис
.env.example               # + REDIS_URL
apigateway/
├── cmd/app/main.go        # + Redis инициализация
├── internal/fetcher/http.go  # + Redis stream updates
└── go.mod                 # + redis/go-redis/v9

worker/
├── cmd/app/main.go        # + конфигурация multi-pass
├── internal/service/worker/code.go   # ~30-50 строк на файл
└── internal/service/worker/assign.go # ~10-20 строк на файл
```

---

## Миграция и обратная совместимость

### WebSocket стримы
- Старая логика (memory map) остаётся как fallback
- Если Redis недоступен → работает без него
- Graceful degradation

### Генерация файлов
- Multi-Pass — основной режим
- N+1 — fallback при ошибке парсинга
- Можно отключить через环境变量 `WORKER_MODE=multypass|nplus1`

---

## Метрики для мониторинга

### Redis
```go
// Счётчики
redis_stream_updates_total    // всего обновлений в Redis
redis_reconnects_total        // reconnect с восстановлением
redis_errors_total            // ошибки Redis
redis_memory_bytes            // используемая память
```

### Токены
```go
llm_requests_total{method="nplus1"}      // N+1 запросов
llm_requests_total{method="multypass"}   // Multi-Pass запросов
llm_tokens_used_total                    // всего токенов
llm_avg_tokens_per_task                  // среднее на задачу
llm_savings_percentage                   // % экономии
```

---

## Таймлайн выполнения

| Фаза | Задачи | Результат |
|------|--------|-----------|
| **1. Redis** | docker-compose, клиент, стримы | Рабочее масштабирование |
| **2. Токены** | Multi-Pass генерация, fallback | -45% токенов |
| **3. Тесты** | reconnect, fallback, метрики | Production-ready |

---

## Риски и митигация

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| LLM не соблюдает формат | Высокая | Fallback на N+1 |
| Redis падает | Средняя | Graceful degradation |
| Большой размер ответа | Средняя | LTRIM 50 последних |
| Конфликт имён файлов | Низкая | sanitizeProjectName() |

---

## Итоговый ожидаемый результат

✅ **Масштабируемость:** 1000+ пользователей, multiple Apigateway поды  
✅ **Надёжность:** Reconnect с восстановлением истории  
✅ **Экономия:** -45% токенов на задачу  
✅ **Скорость:** -75% времени генерации  
✅ **Стоимость:** С $50-100/день → $25-50/день при 100 задачах  
