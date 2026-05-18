package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"nodes/internal/service/rules"
)

// parseWorkerMetadata — извлекает provider/model/tokens/tech_stack из metadata
type workerMeta struct {
	provider  string
	model     string
	tokens    map[string]string
	techStack string
}

func parseWorkerMetadata(metadata map[string]string) workerMeta {
	m := workerMeta{
		provider:  metadata["provider"],
		model:     metadata["model"],
		techStack: metadata["tech_stack"],
		tokens:    map[string]string{},
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
	return m
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

// writeContextFile — пишет .crewai/context.json для текущей задачи
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
	crewaiDir := filepath.Join(basePath, ".crewai")
	os.MkdirAll(crewaiDir, 0755)
	os.WriteFile(filepath.Join(crewaiDir, "context.json"), contextJSON, 0644)
}
