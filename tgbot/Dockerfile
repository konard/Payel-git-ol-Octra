FROM python:3.12-slim

WORKDIR /app

# Update pip and install dependencies
RUN pip install --upgrade pip
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем код
COPY src/ ./src/

# Запускаем бота (токен передаётся через env_file в docker-compose)
CMD ["python", "-m", "src.main"]
