FROM python:3.11-slim

WORKDIR /app

# Устанавливаем зависимости
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем код
COPY src/ ./src/

# Запускаем бота (токен передаётся через env_file в docker-compose)
CMD ["python", "-m", "src.main"]
