FROM python:3.11-slim

RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY pyproject.toml .
RUN pip install --no-cache-dir "setuptools>=64" \
 && pip install --no-cache-dir "aiogram>=3.0.0" "openai>=1.0.0" "python-dotenv>=1.0.0" "asyncpg>=0.29.0"

COPY src/ src/
ENV PYTHONPATH=/app
CMD ["python", "-m", "src.main"]
