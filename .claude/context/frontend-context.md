# Контекст фронтенда CrewAI

## Структура приложения
- **main.tsx** - точка входа
- **App.tsx** - основной компонент приложения
- **components/** - переиспользуемые компоненты
- **app/components/** - компоненты приложения (Canvas, TopBar, BottomInput, etc.)
- **stores/** - состояние приложения (taskStore, authStore, etc.)
- **hooks/** - кастомные хуки (useWebSocket, useI18n, etc.)
- **services/** - сервисы для API (authService, workflowService, etc.)

## Ключевые компоненты
- **TopBar**: верхняя панель с логотипом, настройками, темой
- **Canvas**: холст для создания рабочих процессов с узлами
- **BottomInput**: нижняя панель для ввода задач
- **StatusBar**: панель статуса задач
- **ConsolePanel**: консоль логов

## Состояние приложения
- **taskStore**: управление задачами, статусами, логами
- **authStore**: аутентификация пользователей
- **settingsStore**: настройки приложения
- **themeStore**: тема (светлая/темная)
- **i18nStore**: локализация

## WebSocket
- Подключение к серверу через useWebSocket хук
- Отправка задач и получение обновлений статуса

## Стилизация
- CSS переменные для тем
- Tailwind CSS классы
- Кастомные стили в styles/

## Задача: добавить режим чата
Нужно добавить переключатель режимов Canvas/Chat в TopBar, где Chat - для общения с боссом без создания задач, Canvas - для создания рабочих процессов.</content>
<parameter name="filePath">.claude/context/frontend-context.md