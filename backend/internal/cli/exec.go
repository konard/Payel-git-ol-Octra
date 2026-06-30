package cli

import (
	"os"
	"path/filepath"
	"strings"
)

// llmEnv translates LLM settings into environment variables understood by the
// common AI CLIs.
func llmEnv(c LLMConfig) []string {
	var env []string
	provider := strings.ToLower(c.Provider)
	if c.APIKey != "" {
		env = append(env,
			"ANTHROPIC_API_KEY="+c.APIKey,
			"OPENAI_API_KEY="+c.APIKey,
		)
		switch provider {
		case "gemini":
			env = append(env, "GEMINI_API_KEY="+c.APIKey)
		case "deepseek":
			env = append(env, "DEEPSEEK_API_KEY="+c.APIKey)
		case "qwen":
			env = append(env, "DASHSCOPE_API_KEY="+c.APIKey)
		case "kimi":
			env = append(env, "MOONSHOT_API_KEY="+c.APIKey)
		case "grok":
			env = append(env, "XAI_API_KEY="+c.APIKey)
		case "openrouter":
			env = append(env, "OPENROUTER_API_KEY="+c.APIKey)
		}
	}
	if c.BaseURL != "" {
		env = append(env,
			"ANTHROPIC_BASE_URL="+c.BaseURL,
			"OPENAI_BASE_URL="+c.BaseURL,
			"OPENAI_API_BASE="+c.BaseURL,
		)
	}
	if c.Model != "" {
		env = append(env,
			"ANTHROPIC_MODEL="+c.Model,
			"OPENAI_MODEL="+c.Model,
		)
	}
	return env
}

func profileBinPaths(envPath string) []string {
	baseDir := filepath.Dir(envPath)
	return []string{
		filepath.Join(envPath, ".octra", "nix-profile", "bin"),
		filepath.Join(envPath, "home", ".nix-profile", "bin"),
		filepath.Join(baseDir, ".system", "nix-profile", "bin"),
		filepath.Join(baseDir, ".system", "home", ".nix-profile", "bin"),
	}
}

func prependPath(env []string, dirs []string) []string {
	currentPath := os.Getenv("PATH")
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			currentPath = strings.TrimPrefix(entry, "PATH=")
			break
		}
	}
	pathValue := strings.Join(append(dirs, currentPath), string(os.PathListSeparator))
	next := make([]string, 0, len(env)+1)
	added := false
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			next = append(next, "PATH="+pathValue)
			added = true
			continue
		}
		next = append(next, entry)
	}
	if !added {
		next = append(next, "PATH="+pathValue)
	}
	return next
}


