FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y build-essential && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt && \
    pip install --force-reinstall grpcio==1.62.1 grpcio-tools==1.62.1

COPY src/ ./src/
COPY model.pkl .

ENV PYTHONPATH=/app/src

CMD ["uvicorn", "src.server.api:app", "--host", "0.0.0.0", "--port", "50055"]
