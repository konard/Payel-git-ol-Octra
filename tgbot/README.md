# Telegram Bot

Telegram bot interface for Octra. Allows users to authenticate, select AI providers and models, and create tasks via Telegram chat.

## Architecture

```
User ──Telegram──► tgbot ──REST──► User Service (3112)
                         ──WS────► API Gateway (3111)
```

## Features

- Login via email/password
- Select workflow, provider, and model
- Real-time task progress streaming via WebSocket
- Multi-language model lists (OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok, Qwen)
- Connection retry with exponential backoff

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `TG_BOT_TOKEN` | — | Telegram bot token (from BotFather) |
| `AUTH_API_URL` | `http://localhost:3112` | User service base URL |
| `API_GATEWAY_URL` | `ws://localhost:3111` | API Gateway WebSocket URL |

## Development

```bash
pip install -r requirements.txt
python -m src.main
```

## Docker

```bash
docker build -t octra-tgbot .
```
