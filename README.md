# CrewAI - Компания из ИИ агентов

Система автономных ИИ агентов которые работают как реальная компания с иерархией:
**User → Apigateway → Boss → Manager → Worker → Решение**

## 🏗 Архитектура

```
┌─────────────┐      HTTP       ┌─────────────┐      gRPC       ┌─────────────┐
│   User UI   │ ──────────────► │  Apigateway │ ──────────────► │    Boss     │
│  (frontend) │     :3111       │   (proxy)   │                 │  (port 50051)│
└─────────────┘                 └─────────────┘                 └──────┬──────┘
                                                                       │
                                                                       │ gRPC
                                                                       ▼
┌─────────────┐                 ┌─────────────┐      gRPC       ┌─────────────┐
│   Solution  │ ◄────────────── │   Worker    │ ◄────────────── │   Manager   │
│  (ZIP code) │     return      │ (port 50053)│                 │ (port 50052)│
└─────────────┘                 └─────────────┘                 └─────────────┘
```

## 📋 Компоненты

### Apigateway (:3111)
- HTTP шлюз для клиентских запросов
- Проксирует задачи в Boss сервис через gRPC
- Эндпоинты:
  - `POST /task/create` - создать задачу
  - `GET /task/status?task_id=...` - статус задачи
  - `GET /health` - проверка здоровья

### Boss Service (gRPC :50051)
- Принимает задачи от apigateway
- **ИИ планирование**: анализирует задачу, определяет нужный стек технологий
- Расписывает сколько менеджеров нужно и их роли
- Сохраняет решение в PostgreSQL

### Manager Service (gRPC :50052)
- Получает указания от Boss
- **ИИ распределение**: создаёт менеджеров, определяет сколько воркеров нужно
- Распределяет роли воркеров (frontend-dev, backend-dev, tester, etc.)

### Worker Service (gRPC :50053)
- Получает задачи от менеджеров
- **ИИ генерация**: генерирует код через LLM агентов
- Создаёт TASK.md и SOLUTION.md
- Возвращает готовый проект (ZIP архив)

## 🚀 Быстрый старт

### 1. Клонирование
```bash
git clone <repository>
cd crewai
```

### 2. Настройка переменных окружения
```bash
cp .env.example .env
# Отредактируйте .env и укажите AZURE_API_KEY
```

### 3. Запуск через Docker Compose
```bash
docker-compose up -d
```

Сервисы запустятся:
- Apigateway: http://localhost:3111
- Boss: localhost:50051
- Manager: localhost:50052
- Worker: localhost:50053
- PostgreSQL: localhost:5432

### 4. Тестирование

Создать задачу:
```bash
curl -X POST http://localhost:3111/task/create \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "username": "pavel",
    "title": "Личный прокси",
    "description": "Создать личный прокси сервер",
    "tokens": ["token1", "token2"],
    "meta": {"priority": "high"}
  }'
```

Проверить статус:
```bash
curl http://localhost:3111/task/status?task_id=<task_id>
```

##  Стек технологий

- **Go 1.25** - основной язык
- **gRPC** - межсервисное общение
- **Azure OpenAI API** - ИИ агенты через `llm-unified-client`
- **GORM** - ORM для работы с БД
- **PostgreSQL** - хранение задач и решений
- **Docker & Docker Compose** - контейнеризация
- **Azure Framework** - HTTP сервер для apigateway

## 📁 Структура проекта

```
crewai/
├── apigateway/          # HTTP шлюз
│   ├── cmd/app/
│   ├── internal/fetcher/
│   ├── pkg/requests/
│   └── proto/
├── boss/               # Boss сервис
│   ├── cmd/app/
│   ├── internal/
│   │   ├── fetcher/grpc/
│   │   └── service/
│   ├── pkg/
│   │   ├── database/
│   │   └── models/
│   └── proto/
├── manager/            # Manager сервис
│   ├── cmd/app/
│   ├── internal/fetcher/grpc/
│   ├── pkg/database/
│   └── proto/
├── worker/             # Worker сервис
│   ├── cmd/app/
│   ├── internal/fetcher/grpc/
│   └── proto/
├── docker-compose.yml
└── .env.example
```

## 🔧 Разработка

### Генерация proto кода
```bash
# Для каждого сервиса
cd boss && .\generate-proto.ps1
cd manager && .\generate-proto.ps1
cd worker && .\generate-proto.ps1
cd apigateway && .\generate-proto.ps1
```

### Локальный запуск без Docker
```bash
# Terminal 1 - PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=crewai \
  -p 5432:5432 postgres:15-alpine

# Terminal 2 - Boss
cd boss && go run cmd/app/main.go

# Terminal 3 - Manager
cd manager && go run cmd/app/main.go

# Terminal 4 - Worker
cd worker && go run cmd/app/main.go

# Terminal 5 - Apigateway
cd apigateway && go run cmd/app/main.go
```

## 📝 Модели данных

### Task
- ID, UserID, Username, Title, Description
- Tokens (JSON), Meta (JSON)
- Status: pending → boss_planning → managers_assigned → workers_assigned → processing → reviewing → done/error
- Solution (ZIP архив)

### BossDecision
- TaskID, ManagersCount, ManagerRoles (JSON)
- TechnicalDescription, TechStack (JSON), ArchitectureNotes

### Manager
- TaskID, Role, AgentID, Status
- WorkerRoles (JSON), WorkersCount

### Worker
- TaskID, ManagerID, Role, AgentID
- TaskMD, SolutionMD, Status
- Files (JSON), Success, Approved

## 🎯 Примеры использования

### Создание простого проекта
```json
POST /task/create
{
  "userId": "user-1",
  "username": "john",
  "title": "REST API для блога",
  "description": "Создать REST API для блога с авторизацией",
  "tokens": ["azure-token-123"],
  "meta": {
    "stack": "Go,PostgreSQL,Docker",
    "priority": "medium"
  }
}
```

### Ответ
```json
{
  "status": "success",
  "task_id": "uuid-...",
  "task_status": "managers_assigned",
  "managers": 3,
  "tech_stack": ["Go", "PostgreSQL", "Docker", "Redis"],
  "description": "Микросервисная архитектура...",
  "architecture": "API Gateway + Services..."
}
```

## ⚠️ Важные замечания

1. **AZURE_API_KEY** обязателен для работы ИИ агентов
2. **PostgreSQL** общая для всех сервисов
3. **gRPC порты** не должны конфликтовать
4. **Миграции БД** запускаются автоматически при старте

## 📄 Лицензия

MIT
