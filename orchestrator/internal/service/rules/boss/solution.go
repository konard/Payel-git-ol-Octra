package boss

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"orchestrator/internal/prompts"
	"orchestrator/internal/service/rules"
)

type streamedSolutionFile struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Language    string `json:"language"`
	Encoding    string `json:"encoding,omitempty"`
	WorkerRole  string `json:"worker_role"`
	ManagerRole string `json:"manager_role"`
	Status      string `json:"status"`
	UpdatedAt   int64  `json:"updated_at"`
}

// collectSolutionMarkdown — собирает Markdown-документы из папки solution/,
// произведённые воркерами не-кодовых задач, в один текст. Возвращает также
// количество исходных документов (нужно, чтобы решить, требуется ли синтез).
func collectSolutionMarkdown(results []*rules.ManagerResult) (string, int) {
	type doc struct {
		path    string
		content string
	}
	var docs []doc
	seen := map[string]bool{}
	for _, mr := range results {
		for _, wr := range mr.WorkerResults {
			for path, content := range wr.Files {
				lower := strings.ToLower(path)
				if !strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, ".markdown") {
					continue
				}
				if seen[path] {
					continue
				}
				seen[path] = true
				docs = append(docs, doc{path: path, content: content})
			}
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].path < docs[j].path })

	var b strings.Builder
	for _, d := range docs {
		if b.Len() > 0 {
			b.WriteString("\n\n---\n\n")
		}
		b.WriteString(d.content)
	}
	return b.String(), len(docs)
}

func collectCodeFilesPayload(results []*rules.ManagerResult) (string, int) {
	type fileEntry struct {
		path        string
		content     string
		workerRole  string
		managerRole string
		language    string
		encoding    string
	}

	var files []fileEntry
	seen := map[string]bool{}
	totalFiles := 0
	for _, mr := range results {
		for _, wr := range mr.WorkerResults {
			for path, content := range wr.Files {
				if strings.TrimSpace(path) == "" || seen[path] {
					continue
				}
				seen[path] = true
				totalFiles++
				language := languageForSolutionPath(path)
				encoding := ""
				if isBinarySolutionPath(path) {
					content = base64.StdEncoding.EncodeToString([]byte(content))
					encoding = "base64"
				}
				files = append(files, fileEntry{
					path:        path,
					content:     content,
					workerRole:  wr.Role,
					managerRole: mr.Role,
					language:    language,
					encoding:    encoding,
				})
			}
		}
	}
	if len(files) == 0 {
		return "", totalFiles
	}

	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })
	now := time.Now().Unix()
	payload := make([]streamedSolutionFile, 0, len(files))
	for _, file := range files {
		payload = append(payload, streamedSolutionFile{
			Path:        file.path,
			Content:     file.content,
			Language:    file.language,
			Encoding:    file.encoding,
			WorkerRole:  file.workerRole,
			ManagerRole: file.managerRole,
			Status:      "ready",
			UpdatedAt:   now,
		})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", totalFiles
	}
	return string(data), totalFiles
}

func isBinarySolutionPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pptx", ".docx", ".xlsx", ".pdf", ".png", ".jpg", ".jpeg", ".gif", ".zip":
		return true
	default:
		return false
	}
}

func languageForSolutionPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cc", ".cpp", ".cxx", ".hpp", ".hxx":
		return "cpp"
	case ".css":
		return "css"
	case ".html", ".htm":
		return "html"
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".sh", ".bash":
		return "shell"
	case ".sql":
		return "sql"
	case ".pptx", ".docx", ".xlsx", ".pdf", ".png", ".jpg", ".jpeg", ".gif", ".zip":
		return "binary"
	default:
		return "plaintext"
	}
}

// synthesizeSolution — менеджерский синтез: сливает Markdown-результаты
// нескольких воркеров в один связный документ, проверяя факты и убирая
// повторы. Если AI недоступен — возвращает "" (вызывающий код оставит исходную
// механическую склейку).
func (s *Service) synthesizeSolution(
	ctx context.Context, provider, model string, tokens map[string]string,
	decision *DecisionResult, topic, workerOutputs string,
) string {
	managerRole := "lead"
	if len(decision.ManagerRoles) > 0 && decision.ManagerRoles[0].Role != "" {
		managerRole = decision.ManagerRoles[0].Role
	}
	outputs := workerOutputs
	if len(outputs) > 12000 {
		outputs = outputs[:12000]
	}
	prompt := prompts.ManagerSynthesis(managerRole, decision.TaskType, topic, outputs)
	resp, err := s.agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil || strings.TrimSpace(resp) == "" {
		return ""
	}
	return strings.TrimSpace(resp)
}

// buildChatSummary — босс готовит короткий текстовый ответ для чата по итогам
// research/document/presentation задачи (полная версия уже лежит во вкладке Solution).
func (s *Service) buildChatSummary(
	ctx context.Context, provider, model string, tokens map[string]string,
	taskType, topic, fullDocument string,
) string {
	full := fullDocument
	if len(full) > 6000 {
		full = full[:6000]
	}
	prompt := prompts.BossChatSummary(taskType, topic, full)
	resp, err := s.agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil || strings.TrimSpace(resp) == "" {
		// Фолбэк: краткое сообщение без AI.
		return "Your " + taskType + " is ready — open the Solution tab to see the full result."
	}
	return strings.TrimSpace(resp)
}

