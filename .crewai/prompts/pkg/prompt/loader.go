package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var promptsPath string

func init() {
	if path := os.Getenv("CREWAI_PROMPTS_PATH"); path != "" {
		promptsPath = path
	} else {
		execPath, _ := os.Getwd()
		promptsPath = filepath.Join(execPath, ".crewai", "prompts")
	}
}

func SetPromptsPath(path string) {
	promptsPath = path
}

func GetPromptsPath() string {
	return promptsPath
}

func Load(service, name string, data map[string]string) (string, error) {
	path := filepath.Join(promptsPath, service, name+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt %s/%s: %w", service, name, err)
	}

	tmpl, err := template.New("prompt").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt: %w", err)
	}

	return buf.String(), nil
}