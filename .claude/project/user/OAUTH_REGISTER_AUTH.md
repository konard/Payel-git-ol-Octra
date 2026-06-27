# User Service — OAuth, Register, Auth

## Структура

| Файл | Роль |
|------|------|
| `user/cmd/app/main.go` | Точка входа |
| `user/pkg/models/user.go` | Модель `UserRegister` |
| `user/pkg/requests/user.go` | DTO: `UserRegisterRequest`, `UserLoginRequest` |
| `user/pkg/database/database.go` | GORM AutoMigrate |
| `user/internal/core/services/auth.go` | Бизнес-логика: регистрация, логин, OAuth создание, токены |
| `user/internal/fetcher/http/oauth/oauth.go` | Хендлеры Google + GitHub OAuth |
| `user/internal/fetcher/http/oauth/lefine.go` | Хендлер LeFine OAuth |
| `user/internal/fetcher/http/oauth/middleware/auth.go` | JWT middleware |
| `user/internal/fetcher/http/router/accaunt/register.go` | POST /register |
| `user/internal/fetcher/http/router/accaunt/login.go` | POST /login |
| `user/internal/fetcher/http/router/accaunt/oauth.go` | Регистрация OAuth-роутов |
| `user/internal/fetcher/http/router/accaunt/accaunt.go` | Регистрация всех аккаунт-роутов |

## Модель UserRegister

```go
type UserRegister struct {
    ID              uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time `gorm:"index"`
    Username        string     `gorm:"unique"`        // → uni_user_registers_username
    Email           string     `gorm:"unique"`         // → uni_user_registers_email
    Password        string     `gorm:"default:null"`   // null для OAuth-пользователей
    SubscriptionEnd *int64     `json:"subscription_end"`
}
```

## OAuth Flow

1. `GET /auth/github` (или `/auth/google`) — редирект на провайдера.
2. `GET /auth/github/callback?code=...` — обмен кода на токен, запрос `https://api.github.com/user` (или Google userinfo), парсинг ответа.
3. Вызов `services.GetOrCreateUserFromGithub(email, name)` (или Google/LeFine).
4. Поиск пользователя по email. Если найден — возвращается.
5. Если не найден — создаётся новый `UserRegister` с `Username = name`, `Email = email`, `Password = null`.
6. Генерация JWT access + refresh токенов, установка `refresh_token` cookie, редирект на фронтенд с токенами в URL.

## Исправленные проблемы

### 1. Коллизия username при OAuth

**`user/internal/core/services/auth.go` — все 3 `GetOrCreateUserFrom*` функции**

**Было:** OAuth display name использовался как `Username` напрямую. Если у другого пользователя уже был такой username — `duplicate key value violates unique constraint "uni_user_registers_username"`.

**Стало:**
- Пре-чека: `SELECT count WHERE username = ?` перед insert. Если занят — добавляется случайный суффикс (6 hex-символов).
- Ретрай на race condition: если `Create` упал с duplicate, генерируется новый суффикс и повтор.
- Raw SQL-ошибка заменена на человеческое сообщение.

**`user/internal/fetcher/http/oauth/oauth.go` — GitHub callback**

**Было:** `name = githubUser.Name; if empty → githubUser.Login` (приоритет у display name).

**Стало:** `name = githubUser.Login; if empty → githubUser.Name` (приоритет у GitHub login — он уникален на GitHub).

### 2. Стандартная регистрация (POST /register)

Валидация `Username`, `Email`, `Password` непустые → хэш пароля → `db.Create` → если duplicate → "уже существует" → генерация JWT.

### 3. Токены

| Токен | Срок | Алгоритм | Содержит |
|-------|------|----------|----------|
| Access | 15 мин | HS256 | user_id, username, email, exp, iat |
| Refresh | 7 дней | HS256 (отдельный секрет) | user_id, username, email, exp, iat |

Refresh token устанавливается в httpOnly cookie при логине/OAuth.

### 4. Cross-user data leak (localStorage) — исправлено

**Проблема:** Zustand persist-ключи были глобальными (`crewai-custom-providers`, `crewai-settings`, `crewai-integrations`, `octra-token-statistics`). При смене пользователя на одном браузере данные предыдущего пользователя оставались в localStorage и подмешивались к новым.

**Факторы:**
1. Глобальные persist-ключи без привязки к userId
2. `logout()` не чистил стора
3. `TopBar.tsx` мержил API-ответ с существующим стором (только add, без replace)

**Исправление** (`frontend/web/src/stores/`):
- Создан `storageScope.ts` — обёртка над localStorage, добавляющая к ключу суффикс `_${userId}`
- Все 4 стора используют `scopedStorage` через `createJSONStorage(() => scopedStorage)`
- `authStore.logout()` вызывает `clearUserScopedData()` — удаляет из localStorage ключи уходящего пользователя
- `customProvidersStore` — добавлены `setProviders`/`setModels` для replace вместо merge
- `settingsStore` — добавлен `resetSettings`
- `integrationStore` — добавлен `resetIntegrations`
- `TopBar.tsx` — вместо merge-цикла теперь `setProviders(providers)` / `setModels(models)` (полная замена)

### 5. Subscription

`SubscriptionEnd *int64` — Unix timestamp окончания подписки. Проверяется в `GetMe`: если `time.Now() > time.Unix(*end, 0)` — подписка истекла.
