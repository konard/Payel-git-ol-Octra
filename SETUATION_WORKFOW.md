Пользователь отправляет запрос по веб сокетам такой:
```json
{
  "userId": "user-123", 
  "username": "pavel",
  "title": "Тест",
  "description": "Тестовая задача с Gemini",
  "tokens": ["AIzaSyB-BUpO-Kf1eXx4oMtWelMZrC_YQvmvTWY"],
  "meta": {
    "model": "gemini-2.5-flash-lite",
    "modelUrl": "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent"
  }
}
```

Потом ему постеменно приходять от сервисов (внимание там не будет что за сервисы):

- Контекст увеличен ! 
- Создано $n менеджеров
- Менеджеры создали $n воркеров
- Фронтенд создаётся 
...
- Результат: $name_project.zip

И это приходит постепенно 