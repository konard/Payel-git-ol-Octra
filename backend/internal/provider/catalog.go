// Package provider contains the built-in model provider catalogue.
package provider

// Package describes a provider template that can be selected from the search
// catalogue. Most non-Anthropic providers expose an OpenAI-compatible chat API,
// which Octra can use in proxy mode and pass through to CLIs via env vars.
type Package struct {
	Key          string
	Name         string
	BaseURL      string
	AuthEnv      string
	DefaultModel string
	Description  string
}

// BuiltinProviders returns the providers Octra exposes in the search catalogue.
func BuiltinProviders() []Package {
	return []Package{
		{
			Key:          "claude",
			Name:         "Claude",
			BaseURL:      "https://api.anthropic.com",
			AuthEnv:      "ANTHROPIC_API_KEY",
			DefaultModel: "claude-sonnet-4-6",
			Description:  "Anthropic Messages API provider for Claude models.",
		},
		{
			Key:          "openai",
			Name:         "OpenAI",
			BaseURL:      "https://api.openai.com/v1",
			AuthEnv:      "OPENAI_API_KEY",
			DefaultModel: "gpt-4.1",
			Description:  "OpenAI-compatible chat completions provider.",
		},
		{
			Key:          "gemini",
			Name:         "Gemini",
			BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai",
			AuthEnv:      "GEMINI_API_KEY",
			DefaultModel: "gemini-2.5-pro",
			Description:  "Google Gemini through its OpenAI-compatible endpoint.",
		},
		{
			Key:          "deepseek",
			Name:         "DeepSeek",
			BaseURL:      "https://api.deepseek.com",
			AuthEnv:      "DEEPSEEK_API_KEY",
			DefaultModel: "deepseek-chat",
			Description:  "DeepSeek OpenAI-compatible API provider.",
		},
		{
			Key:          "qwen",
			Name:         "Qwen",
			BaseURL:      "https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
			AuthEnv:      "DASHSCOPE_API_KEY",
			DefaultModel: "qwen-plus",
			Description:  "Alibaba Cloud DashScope OpenAI-compatible Qwen endpoint.",
		},
		{
			Key:          "kimi",
			Name:         "Kimi",
			BaseURL:      "https://api.moonshot.ai/v1",
			AuthEnv:      "MOONSHOT_API_KEY",
			DefaultModel: "kimi-k2-0711-preview",
			Description:  "Moonshot AI Kimi OpenAI-compatible endpoint.",
		},
		{
			Key:          "grok",
			Name:         "Grok",
			BaseURL:      "https://api.x.ai/v1",
			AuthEnv:      "XAI_API_KEY",
			DefaultModel: "grok-4",
			Description:  "xAI Grok OpenAI-compatible endpoint.",
		},
		{
			Key:          "openrouter",
			Name:         "OpenRouter",
			BaseURL:      "https://openrouter.ai/api/v1",
			AuthEnv:      "OPENROUTER_API_KEY",
			DefaultModel: "openai/gpt-4.1",
			Description:  "OpenRouter gateway for hosted and third-party models.",
		},
		{
			Key:          "zed",
			Name:         "Zed AI",
			BaseURL:      "",
			AuthEnv:      "ZED_AI_API_KEY",
			DefaultModel: "",
			Description:  "Zed-compatible provider template for custom CLI configuration.",
		},
	}
}
