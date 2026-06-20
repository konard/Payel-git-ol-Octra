package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/search"
)

// parseWorkerMetadata — извлекает provider/model/tokens/tech_stack из metadata
type workerMeta struct {
	provider     string
	model        string
	tokens       map[string]string
	techStack    string
	taskType     string
	title        string
	description  string
	searchConfig *search.ModelConfig
}

func parseWorkerMetadata(metadata map[string]string) workerMeta {
	m := workerMeta{
		provider:    metadata["provider"],
		model:       metadata["model"],
		techStack:   metadata["tech_stack"],
		taskType:    metadata["task_type"],
		title:       metadata["title"],
		description: metadata["description"],
		tokens:      map[string]string{},
	}
	if m.taskType == "" {
		m.taskType = "code"
	}
	if m.provider == "" {
		m.provider = "openai"
	}
	if m.model == "" {
		m.model = "gpt-4o-mini"
	}
	if m.techStack == "" {
		m.techStack = "go"
	}
	if tokensJSON, ok := metadata["tokens"]; ok {
		json.Unmarshal([]byte(tokensJSON), &m.tokens)
	}
	if apiKey, ok := metadata[m.provider]; ok {
		m.tokens[m.provider] = apiKey
	}
	m.searchConfig = parseSearchConfig(metadata["search"])
	return m
}

type searchMetadata struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	BaseURL   string `json:"base-url"`
	APIKey    string `json:"api-key"`
	Striming  bool   `json:"striming"`
	Streaming *bool  `json:"streaming,omitempty"`
}

func parseSearchConfig(raw string) *search.ModelConfig {
	if raw == "" {
		return nil
	}

	var meta searchMetadata
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil
	}
	streaming := meta.Striming
	if meta.Streaming != nil {
		streaming = *meta.Streaming
	}
	cfg := search.ModelConfig{
		Provider:  meta.Provider,
		Model:     meta.Model,
		BaseURL:   meta.BaseURL,
		APIKey:    meta.APIKey,
		Streaming: streaming,
	}
	if cfg.Provider == "" || cfg.Model == "" || cfg.BaseURL == "" || cfg.APIKey == "" {
		return nil
	}
	return &cfg
}

// resolveBasePath — определяет директорию проекта для группы воркеров
func resolveBasePath(req *rules.AssignWorkersRequest) string {
	if req.ProjectPath != "" {
		return req.ProjectPath
	}
	if repoPath, ok := req.Metadata["repo_path"]; ok && repoPath != "" {
		return repoPath
	}
	projectsDir := os.Getenv("PROJECTS_DIR")
	if projectsDir == "" {
		projectsDir = "/workspace/projects"
	}
	if filepath.Base(projectsDir) == "projects" || projectsDir == "/workspace/projects" {
		return filepath.Join(projectsDir, req.TaskId)
	}
	return projectsDir
}

// buildContextSummary — собирает текстовое описание результатов других команд воркеров
func buildContextSummary(other []*rules.WorkerResult) string {
	contextSummary := ""
	for _, wr := range other {
		contextSummary += fmt.Sprintf("\n--- Worker %s (%s) result ---\n", wr.WorkerId, wr.Role)
		contextSummary += fmt.Sprintf("Task: %s\n", wr.TaskMd)
		if wr.Success {
			contextSummary += fmt.Sprintf("Solution: %s\n", wr.SolutionMd)
		}
		if wr.Feedback != "" {
			contextSummary += fmt.Sprintf("Feedback: %s\n", wr.Feedback)
		}
	}
	return contextSummary
}

// writeContextFile — пишет .octra/context.json для текущей задачи
func writeContextFile(basePath, taskID, managerID, managerRole string, workerRoles []*rules.WorkerRole, contextSummary string) {
	contextData := map[string]interface{}{
		"task_id": taskID,
		"manager": map[string]string{
			"id":   managerID,
			"role": managerRole,
		},
		"workers": workerRoles,
		"context": contextSummary,
		"history": []map[string]interface{}{},
	}
	contextJSON, _ := json.MarshalIndent(contextData, "", "  ")
	octraDir := filepath.Join(basePath, ".octra")
	os.MkdirAll(octraDir, 0755)
	os.WriteFile(filepath.Join(octraDir, "context.json"), contextJSON, 0644)
}
