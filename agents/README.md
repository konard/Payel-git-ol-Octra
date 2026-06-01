# Agents

Unified gRPC service for LLM provider access. Routes prompts to the appropriate AI provider (Claude, Gemini, OpenAI, DeepSeek, Grok, OpenRouter, Qwen, Z.AI, or custom) using per-request API keys.

## Protocol Buffers

```protobuf
service AgentService {
  rpc Generate(GenerateRequest) returns (GenerateResponse);
  rpc GenerateStream(GenerateRequest) returns (stream GenerateStreamChunk);
}
```

## Supported Providers

| Provider | SDK | Default Model |
|----------|-----|---------------|
| Claude | anthropic-sdk-go | claude-3-5-sonnet |
| Gemini | google/generative-ai-go | gemini-2.0-flash |
| OpenAI | openai-go/v3 | gpt-4o-mini |
| DeepSeek | cohesion-org/deepseek-go | deepseek-chat |
| Grok | SimonMorphy/grok-go | grok-2 |
| OpenRouter | openai-go (compatible) | openai/gpt-4o-mini |
| Qwen | openai-go (compatible) | qwen-turbo |
| Z.AI | openai-go (compatible) | glm-4.5-air |

## Architecture

```
Nodes (Boss/Manager/Worker) ──gRPC──► Agents (50053)
                                          │
                    ┌─────┬──────┬───────┼──────┬──────┬─────┐
                    ▼     ▼      ▼       ▼      ▼      ▼     ▼
                 Claude Gemini OpenAI DeepSeek Grok OpenRouter ...
```

API keys are provided **per-request** via the `tokens` field — not loaded from environment at startup.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTS_PORT` | `50053` | gRPC server port |

## Development

```bash
go run cmd/app/main.go
```

## Docker

```bash
docker build -t octra-agents -f Dockerfile ..
```
