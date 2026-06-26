# Captcha — статус и исправление

## Состояние

Captcha **не реализована**. Присутствует только частичный скелет:

| Компонент | Статус |
|-----------|--------|
| Фронтенд-виджет | Закомментирован (`AuthModal.tsx:12`) |
| npm-пакет | Отсутствует |
| Env-ключи (RECAPTCHA_SITE_KEY и т.д.) | Отсутствуют |
| Бэкенд-проверка токена у провайдера | Не реализована |
| go-модуль для recaptcha | Отсутствует |

## Проблема (исправлена)

Бэкенд требовал `captcha_token != ""` при регистрации и логине:

- `user/internal/fetcher/http/router/accaunt/register.go:17` — `if req.CaptchaToken == ""` → 400
- `user/internal/fetcher/http/router/accaunt/login.go:17` — `if req.CaptchaToken == ""` → 400

Фронтенд передавал пустую строку `''` (`AuthModal.tsx:107`), поэтому регистрация/логин всегда падали с 400.

## Исправление

Убраны проверки на `req.CaptchaToken == ""` из обоих хендлеров.

Сервисные функции `RegisterUser`/`LoginUser` captcha-токен и так игнорируют.

## Оставшиеся артефакты (можно удалить)

- `user/pkg/requests/user.go` — поле `CaptchaToken` в `UserRegisterRequest` и `UserLoginRequest`
- `frontend/web/src/services/authService.ts` — поле `captcha_token` в интерфейсах
- `frontend/web/src/stores/authStore.ts` — параметр `captchaToken` в `register()`
- `frontend/web/src/components/auth/AuthModal.tsx` — закомментированный импорт ReCAPTCHA, передача `''`
