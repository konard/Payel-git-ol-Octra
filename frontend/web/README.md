# CrewAI Frontend

Визуальный редактор для управления AI агентами в стиле n8n.

## 🚀 Описание

CrewAI — это платформа для визуального управления командой AI агентов. Приложение позволяет создавать, настраивать и соединять агентов различных типов в единую рабочую систему.

## 📋 Стек технологий

- **TypeScript** — типизированный JavaScript
- **Parcel** — быстрый сборщик проектов
- **CSS Variables** — система дизайна на переменных
- **Vanilla JS** — без фреймворков, чистая производительность

## 🎨 Дизайн-система

Приложение использует тему в стиле n8n:
- Тёмная цветовая схема
- Яркие акценты для категорий агентов
- Плавные анимации и переходы
- Сетка на фоне канваса

## 📁 Структура проекта

```
frontend/
├── src/
│   ├── components/       # UI компоненты
│   │   ├── BaseComponent.ts
│   │   ├── ButtonComponent.ts
│   │   ├── InputComponent.ts
│   │   ├── NodeComponent.ts
│   │   ├── AgentSidebarComponent.ts
│   │   ├── PropertiesPanelComponent.ts
│   │   ├── ConnectionLayerComponent.ts
│   │   ├── ToolbarComponent.ts
│   │   ├── MinimapComponent.ts
│   │   └── HeaderComponent.ts
│   ├── core/            # Ядро приложения
│   │   ├── NodeEngine.ts
│   │   └── CanvasEngine.ts
│   ├── models/          # Типы и интерфейсы
│   │   ├── Types.ts
│   │   └── AgentDefinitions.ts
│   ├── utils/           # Утилиты
│   │   ├── IdGenerator.ts
│   │   ├── MathUtils.ts
│   │   ├── EventDispatcher.ts
│   │   └── StorageUtils.ts
│   ├── styles/          # Стили
│   │   ├── variables.css
│   │   ├── main.css
│   │   ├── node.css
│   │   └── components.css
│   ├── assets/          # Ресурсы
│   ├── App.ts           # Главное приложение
│   ├── index.ts         # Точка входа
│   └── index.html       # HTML шаблон
├── package.json
├── tsconfig.json
├── .eslintrc.json
├── .prettierrc.json
└── README.md
```

## 🛠️ Установка

```bash
# Установка зависимостей
npm install

# Запуск dev-сервера
npm run dev

# Сборка продакшн версии
npm run build

# Линтинг
npm run lint

# Форматирование
npm run format
```

## 🎯 Возможности

### Типы агентов

#### Руководители (Chiefs)
- 👔 Chief Agent — Главный агент
- 📋 Manager — Менеджер
- 🎯 Team Lead — Тимлид
- 🏢 Director — Директор

#### Дизайнеры (Designers)
- 🎨 Designer — Дизайнер
- 🧭 UX Designer — UX дизайнер
- 🖼️ UI Designer — UI дизайнер
- ✏️ Graphic Designer — Графический дизайнер

#### Frontend разработчики
- 🌐 Frontend Developer — Frontend разработчик
- ⚛️ React Developer — React разработчик
- 💚 Vue Developer — Vue разработчик
- 🅰️ Angular Developer — Angular разработчик

#### Backend разработчики
- 🔧 Backend Developer — Backend разработчик
- 🐹 Go Developer — Go разработчик
- 🐍 Python Developer — Python разработчик
- ☕ Java Developer — Java разработчик
- 📦 Node.js Developer — Node.js разработчик

#### Тестировщики (Testers)
- 🧪 QA Tester — Тестировщик
- 🤖 QA Automation — Автоматизатор
- 📝 QA Manual — Ручной тестировщик
- ⚡ Performance Tester — Тестировщик производительности

#### DevOps
- 🔄 DevOps Engineer — DevOps инженер
- ☁️ Cloud Engineer — Облачный инженер
- 🛡️ SRE — Site Reliability Engineer

#### Аналитики (Analysts)
- 📊 Analyst — Аналитик
- 💼 Business Analyst — Бизнес-аналитик
- 📈 Data Analyst — Data аналитик
- 🖥️ System Analyst — Системный аналитик

### Функционал

- ✅ Drag-and-drop создание агентов
- ✅ Соединение агентов между собой
- ✅ Редактирование параметров агентов
- ✅ Зум и панорамирование канваса
- ✅ Мини-карта для навигации
- ✅ Сохранение проекта в localStorage
- ✅ Экспорт/импорт проекта в JSON
- ✅ Горячие клавиши

## ⌨️ Горячие клавиши

| Клавиша | Действие |
|---------|----------|
| `Delete` / `Backspace` | Удалить выделенную ноду |
| `Ctrl/Cmd + S` | Сохранить проект |
| `Ctrl/Cmd + Z` | Отменить (TODO) |
| `Escape` | Снять выделение |
| `Колесо мыши` | Зум |
| `Alt + ЛКМ` | Панорамирование |
| `СКМ` | Панорамирование |

## 🎨 Категории агентов

Каждая категория имеет свой цвет:
- **Chiefs** — 🔴 Красный (#ff6d5a)
- **Managers** — 🟠 Оранжевый (#ff9f4d)
- **Designers** — 🟣 Фиолетовый (#d44dff)
- **Frontend** — 🔵 Голубой (#00e5ff)
- **Backend** — 🟢 Зелёный (#4dff91)
- **Testers** — 🩷 Розовый (#ff4d94)
- **DevOps** — 🟡 Жёлтый (#ffd04d)
- **Analysts** — 🔵 Синий (#4d94ff)

## 📦 Архитектура

### NodeEngine
Ядро системы управления нодами:
- Создание и удаление нод
- Управление соединениями
- Выделение и перемещение
- Экспорт/импорт проекта

### CanvasEngine
Движок канваса:
- Зум и панорамирование
- Преобразование координат
- Привязка к сетке
- Отслеживание viewport

### Компоненты
Все компоненты наследуются от `BaseComponent`:
- Жизненный цикл (create, render, destroy)
- Работа с DOM
- Обработка событий
- Управление стилями

## 🔧 Конфигурация

### TypeScript (tsconfig.json)
- Target: ES2022
- Module: ESNext
- Strict mode: включён
- Path aliases: @/, @components/, @core/, @models/, @utils/

### ESLint
- TypeScript recommended rules
- Custom rules для форматирования
- Интеграция с Prettier

### Prettier
- Single quotes
- Tab width: 4
- Trailing comma: ES5
- Print width: 120

## 🚧 Roadmap

- [ ] Undo/Redo система
- [ ] Групповое выделение нод
- [ ] Копирование/вставка нод
- [ ] Поиск по проекту
- [ ] История изменений
- [ ] Совместная работа (WebSocket)
- [ ] Интеграция с backend API
- [ ] Запуск воркфлоу
- [ ] Логирование выполнения
- [ ] Плагины и расширения

## 📄 Лицензия

MIT

## 👥 Команда

CrewAI Team

---

**CrewAI** — Ваша команда AI агентов 🤖
