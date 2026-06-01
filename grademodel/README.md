# Grademodel

Task complexity grader. Uses a trained Ridge regression model to predict task difficulty (1–100) from a natural language description. Called once per task by the Boss service to optimize token usage.

## Endpoints

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Health check |
| `POST` | `/grade` | Grade a task |

**Request:**
```json
{ "task": "Create a REST API in Go with JWT auth" }
```

**Response:**
```json
{ "task": "...", "translated": "...", "grade": 45 }
```

## Architecture

```
Request ──► Translate (Google Translate) ──► Feature Extraction ──► Ridge Regression ──► Grade (1-100)
```

The model extracts 21 boolean/keyword features (word count, has_code, has_api, has_db, has_auth, etc.) and predicts complexity using `scikit-learn` Ridge regression (trained on 266 task-grade pairs).

## Development

```bash
pip install -r requirements.txt
uvicorn src.server.api:app --host 0.0.0.0 --port 50055
```

### Retraining

```bash
python -m src.model.train --epochs 100
```

## Docker

```bash
docker build -t octra-grademodel .
```
