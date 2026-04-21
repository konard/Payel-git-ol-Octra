# CrewAI Frontend

## Dynamic Model Loading

The application now dynamically loads models from OpenRouter API instead of using hardcoded model lists.

### Setup

1. Create a `.env` file in the `frontend/web` directory:
```bash
cp frontend/web/.env.example frontend/web/.env
```

2. Add your OpenRouter API key to `.env`:
```env
VITE_OPENROUTER_API_KEY=your_openrouter_api_key_here
```

### Features

- **Automatic Model Updates**: Models are fetched from OpenRouter API and cached for 24 hours
- **Filtered Models**: Only models from approved providers are shown (OpenAI, Google, Qwen, Z.AI, Anthropic, DeepSeek)
- **Loading States**: UI shows loading indicators when fetching models
- **Free Model Indicators**: Free models are marked with a "Free" badge
- **Fallback Support**: If API is unavailable, fallback models are used

### How It Works

1. When opening the model selector for OpenRouter provider, models are loaded from API
2. Results are cached in localStorage for 24 hours
3. If API fails, fallback models are displayed
4. Models are filtered to only include supported providers

### API Requirements

- OpenRouter API key with models endpoint access
- Internet connection for initial model loading
- Cached models work offline for up to 24 hours

### Development

To test with different API keys, modify the `VITE_OPENROUTER_API_KEY` in `.env` and restart the dev server.