# CrewAI — Многоагентная система генерации кода

Современная веб-платформа для автоматической генерации программного кода с использованием ИИ-агентов. Работает как полноценная IT-компания с иерархией ролей.

## 🏗 Архитектура

```
┌─────────────┐      HTTP/WS     ┌─────────────┐      gRPC       ┌─────────────┐
│   Frontend  │ ───────────────► │  Apigateway │ ──────────────► │    Boss     │
│   (React)   │     :80          │   (proxy)   │   :3111         │  (port 50051)│
└──────┬──────┘                  └──────┬──────┘                 └──────┬──────┘
       │                                 │                               │
       │                                 │                gRPC (ПАРАЛЛЕЛЬНО)
       │                                 │               ┌───────┴───────┐
       │                                 ▼               ▼               ▼
       │                        ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
       │                        │ User Service│  │  Manager #1 │  │  Manager #2 │
       │                        │  (auth)     │  │  (port 50052)│  │  (port 50052)│
       │                        │   :3112     │  └──────┬──────┘  └──────┬──────┘
       │                        └─────────────┘         │                 │
       │                                                 │                 │
       │                                       gRPC (последовательно)     │
       │                                                 ▼                 ▼
       │                                  ┌─────────────┐  ┌─────────────┐
       │                                  │   Worker(s) │  │   Worker(s) │
       │                                  │  (port 50053)│  │  (port 50053)│
       │                                  └──────┬──────┘  └──────┬──────┘
       │                                         │                 │
       │                                         ▼                 ▼
       │                                  ┌─────────────┐  ┌─────────────┐
       │                                  │   Agents    │  │   Agents    │
       │                                  │  (port 50053)│  │  (port 50053)│
       └─────────────────────────────────────────────────┼─────────────────┘
                                                         │
                                               LLM API (OpenRouter, Gemini,
                                               OpenAI, Claude, DeepSeek, Grok)
```

## 💡 Что такое CrewAI?

CrewAI — это веб-платформа, которая позволяет пользователям описывать программные проекты на естественном языке, а система автоматически генерирует полный рабочий код.

### Как это работает:
1. **Пользователь** описывает задачу: "Создай REST API на Go с аутентификацией"
2. **Boss агент** анализирует задачу и планирует архитектуру
3. **Manager агенты** распределяют работу между разработчиками
4. **Worker агенты** пишут код для каждого компонента
5. **Система** собирает всё в готовый проект и возвращает ZIP-архив

## ✨ Возможности

- 🎯 **Интеллектуальное планирование** — ИИ анализирует требования и выбирает оптимальную архитектуру
- 👥 **Командная работа** — несколько агентов работают параллельно над разными частями проекта
- 🔄 **Итеративная разработка** — автоматические review и исправления кода
- 💳 **Подписка Pro** — расширенные возможности для активных пользователей
- 🌐 **Веб-интерфейс** — удобный интерфейс с real-time обновлениями
- 📱 **Чат и канвас** — два режима работы: чат и визуальный редактор

## 🚀 Быстрый старт

### 1. Клонирование и настройка
```bash
git clone <repository>
cd crewai
cp .env.example .env
# Настройте переменные окружения (API ключи провайдеров)
```

### 2. Запуск через Docker Compose
```bash
docker compose up -d --build
```

### 3. Доступ к приложению
- **Веб-интерфейс**: http://localhost
- **API Gateway**: http://localhost:3111
- **User Service**: http://localhost:3112

### 4. Первое использование
1. Откройте http://localhost
2. Зарегистрируйтесь или войдите в аккаунт
3. Оформите подписку Pro (тестовая оплата доступна для разработки)
4. Создайте задачу в режиме Canvas или Chat
5. Скачайте готовый проект в ZIP архиве

## 💳 Подписка Pro

CrewAI работает по модели подписки для обеспечения качества и доступности сервиса.

### Тарифы:
- **1 месяц** — $10
- **3 месяца** — $25 (экономия 17%)
- **6 месяцев** — $50 (экономия 17%)
- **1 год** — $100 (экономия 17%)

### Что дает подписка Pro:
- ✅ Доступ ко всем LLM провайдерам (OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok)
- ✅ Неограниченное количество задач
- ✅ Приоритетная обработка
- ✅ Расширенные возможности генерации кода
- ✅ Кастомные провайдеры и модели
- ✅ Техническая поддержка

### Платежи:
- 💳 Банковские карты через YooKassa
- 🔄 Автоматическое продление подписки
- 🛡️ Безопасная обработка платежей
- 📧 Квитанции и история платежей

### Для разработки:
- 🧪 Тестовая оплата доступна для проверки функционала
- 🔧 Настройка платежей через переменные окружения

## 📁 Структура проекта

```
crewai/
├── frontend/web/        # React веб-приложение
│   ├── src/
│   │   ├── app/         # Основное приложение
│   │   ├── components/  # UI компоненты
│   │   ├── services/    # API сервисы
│   │   ├── stores/      # Zustand стейт менеджмент
│   │   └── hooks/       # React хуки
│   ├── nginx.conf       # Nginx конфигурация
│   └── Dockerfile
├── user/                # User сервис (аутентификация, подписки)
│   ├── cmd/app/         # Основное приложение
│   ├── internal/core/   # Бизнес логика
│   ├── pkg/             # Вспомогательные пакеты
│   └── Dockerfile
├── apigateway/          # API Gateway
│   ├── cmd/app/
│   ├── internal/
│   └── pkg/
├── boss/                # Boss агент (координатор)
├── manager/             # Manager агенты
├── worker/              # Worker агенты
├── agents/              # LLM провайдеры
├── docker-compose.yml   # Docker конфигурация
├── go.work             # Go workspace
└── .env.example        # Пример переменных окружения
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

### Требования
- Docker & Docker Compose
- Go 1.25+
- Node.js 18+ (для frontend)
- PostgreSQL 15+
- Redis 7+

### Запуск в режиме разработки
```bash
# Клонирование
git clone <repository>
cd crewai

# Настройка переменных окружения
cp .env.example .env
# Отредактируйте .env файл

# Запуск всех сервисов
docker compose up -d

# Или только необходимые для разработки
docker compose up -d postgres redis user frontend
```

### Frontend разработка
```bash
cd frontend/web
npm install
npm run dev  # Запуск на http://localhost:5173
```

### Backend разработка
```bash
# Локальный запуск сервисов (после запуска БД)
cd user && go run cmd/app/main.go
cd apigateway && go run cmd/app/main.go
# ... другие сервисы
```

### Работа с protobuf
```bash
# Генерация Go кода из proto файлов
# Используйте скрипт генерации или protoc вручную
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/*.proto
```

## ⚠️ Важные замечания

1. **Подписка обязательна** — для использования сервиса требуется активная подписка Pro
2. **API ключи провайдеров** — настройте хотя бы один LLM провайдер в переменных окружения
3. **PostgreSQL и Redis** — используются для хранения данных и кэширования
4. **Docker Compose** — рекомендуемый способ запуска для разработки и продакшена
5. **Тестовая оплата** — доступна для проверки платежного функционала без реальных денег
6. **Производительность** — время генерации зависит от сложности задачи и выбранной модели
7. **Безопасность** — API ключи хранятся securely и не передаются в логи

---

## ⚙️ Переменные окружения

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

### Обязательные переменные:

```bash
# JWT токены (сгенерируйте случайные строки)
JWT_SECRET="your-jwt-secret-here"
JWT_REFRESH_SECRET="your-refresh-secret-here"

# YooKassa платежи (тестовые данные)
YOOKASSA_SHOP_ID="1339826"
YOOKASSA_SECRET_KEY="test_StL4_VJfVFbOJ7_BbolU2VhoR1zjIQ7Qf2gcwN3Gngw"

# API ключи провайдеров (минимум один)
OPENROUTER_API_KEY="sk-or-v1-..."
GEMINI_API_KEY="AIzaSy..."
# Другие провайдеры опциональны
```

### Опциональные переменные:

```bash
# Порты сервисов (по умолчанию)
AUTH_PORT=3112
API_GATEWAY_PORT=3111

# Rate limiting
RATE_LIMIT_TASK_CREATE=10/60
RATE_LIMIT_TASK_STATUS=60/60

# Redis (для кэширования)
REDIS_URL=redis://redis:6379/0
```

## 📞 Поддержка

- 📧 Email: support@crewai.com
- 💬 Discord: [Присоединяйтесь к сообществу](https://discord.gg/crewai)
- 📚 Документация: [docs.crewai.com](https://docs.crewai.com)
- 🐛 Issues: [GitHub Issues](https://github.com/crewai/crewai/issues)

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE) файл для деталей.
