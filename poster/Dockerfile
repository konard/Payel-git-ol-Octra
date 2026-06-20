FROM python:3.11-slim

RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY pyproject.toml .
RUN pip install uv && uv sync --no-dev

COPY src/ src/
CMD ["uv", "run", "python", "-m", "src.main"]
