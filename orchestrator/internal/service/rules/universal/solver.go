package universal

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"orchestrator/internal/config"
	"orchestrator/internal/prompts"
	"orchestrator/internal/service/agents"
	"orchestrator/internal/service/git"
	"orchestrator/internal/service/util"
)

type streamedFile struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Language    string `json:"language"`
	WorkerRole  string `json:"worker_role"`
	ManagerRole string `json:"manager_role"`
	Status      string `json:"status"`
	UpdatedAt   int64  `json:"updated_at"`
}

type universalResponse struct {
	Files        map[string]string `json:"files"`
	Dependencies bool              `json:"dependencies"`
}

type ProgressFunc func(progress int32, message string, data map[string]string)

type SolveRequest struct {
	Title       string
	Description string
	TaskType    string
	TechStack   string
	Provider    string
	Model       string
	Tokens      map[string]string
	ProjectPath string
	Progress    ProgressFunc

	// SearchBlock — отформатированные результаты веб-поиска от ноды поиска (issue #97).
	// Когда не пуст, передаётся в промпт, чтобы нода отвечала по реальным источникам.
	SearchBlock string
	// SearchSources — Markdown-список источников, который пишется в solution/sources.md.
	SearchSources string
}

type SolveResult struct {
	Files     map[string]string
	NeedsDeps bool
}

func NewSolver() *Solver {
	return &Solver{}
}

type Solver struct{}

func (s *Solver) Solve(ctx context.Context, agentsClient *agents.Client, req *SolveRequest) *SolveResult {
	emit := req.Progress

	emit(45, "Universal node solving the task directly...", map[string]string{
		"task_type":    req.TaskType,
		"current_role": "universal",
	})

	prompt := prompts.UniversalNode(req.Title, req.Description, req.TaskType, req.TechStack, req.SearchBlock)
	if req.SearchBlock != "" {
		log.Printf("[Universal] solving with web search context (%d chars)", len(req.SearchBlock))
	}

	tokens := req.Tokens
	if tokens == nil {
		tokens = map[string]string{}
	}

	gen := s.generateFiles(ctx, agentsClient, req.Provider, req.Model, prompt, tokens)
	if gen == nil || len(gen.Files) == 0 {
		log.Printf("Universal node produced no files")
		return nil
	}

	// Когда нода поиска нашла источники, прикладываем их к решению, чтобы у ответа
	// были проверяемые ссылки (issue #97). Делаем это только для не-кодовых задач —
	// для кода лишний markdown-файл нарушил бы правило «не переусложнять».
	if sources := strings.TrimSpace(req.SearchSources); sources != "" && req.TaskType != "code" && req.TaskType != "github" {
		if gen.Files == nil {
			gen.Files = map[string]string{}
		}
		if _, exists := gen.Files["solution/sources.md"]; !exists {
			gen.Files["solution/sources.md"] = sources
		}
	}

	written := s.writeFiles(req.ProjectPath, gen.Files, emit)
	if len(written) == 0 {
		log.Printf("Universal node files failed validation")
		return nil
	}

	if err := git.Add(req.ProjectPath); err != nil {
		log.Printf("Universal node git add failed: %v", err)
	} else if err := git.Commit(req.ProjectPath, "Universal node: "+req.Title); err != nil {
		log.Printf("Universal node git commit failed: %v", err)
	}

	if len(written) > 0 {
		payload := buildPayload(written, "universal", "universal")
		if payload != "" {
			emit(70, "Universal node produced the solution", map[string]string{
				"task_type":    req.TaskType,
				"current_role": "universal",
				"code_files":   payload,
				"filesCount":   strconv.Itoa(len(written)),
			})
		}
	}

	emit(75, "Universal node finished", map[string]string{
		"task_type":    req.TaskType,
		"current_role": "universal",
	})

	return &SolveResult{Files: written, NeedsDeps: gen.Dependencies}
}

func (s *Solver) generateFiles(ctx context.Context, agentsClient *agents.Client, provider, model, prompt string, tokens map[string]string) *universalResponse {
	for _, p := range config.FallbackChain(provider, model) {
		resp, err := agentsClient.GenerateFromTask(ctx, p.Provider, p.Model, prompt, tokens)
		if err != nil {
			log.Printf("Universal node provider %s/%s failed: %v", p.Provider, p.Model, err)
			continue
		}
		if parsed := ParseResponse(resp); len(parsed.Files) > 0 {
			return parsed
		}
		log.Printf("Universal node: no files parsed from %s/%s response", p.Provider, p.Model)
	}
	return nil
}

func (s *Solver) writeFiles(projectPath string, files map[string]string, emit ProgressFunc) map[string]string {
	written := make(map[string]string, len(files))
	for path, content := range files {
		fullPath, err := util.ValidateFilePath(projectPath, path)
		if err != nil {
			log.Printf("Universal node: path validation failed for %s: %v", path, err)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Printf("Universal node: mkdir for %s: %v", path, err)
			continue
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			log.Printf("Universal node: write %s: %v", path, err)
			continue
		}
		written[path] = content
		if emit != nil {
			emit(60, "Writing file: "+path, map[string]string{
				"file":         path,
				"type":         "write",
				"current_role": "universal",
			})
		}
	}
	return written
}

// ParseResponse extracts the file map and dependencies flag from the
// universal node's JSON response.
func ParseResponse(resp string) *universalResponse {
	jsonStr := util.ExtractJSONFromMarkdown(resp)
	var parsed universalResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("Universal node JSON parse error: %v", err)
		return nil
	}
	files := make(map[string]string, len(parsed.Files))
	for path, content := range parsed.Files {
		path = strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
		for strings.HasPrefix(path, "./") {
			path = strings.TrimPrefix(path, "./")
		}
		if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
			continue
		}
		if strings.TrimSpace(content) == "" {
			continue
		}
		files[path] = content
	}
	parsed.Files = files
	return &parsed
}

func buildPayload(files map[string]string, workerRole, managerRole string) string {
	if len(files) == 0 {
		return ""
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		if strings.TrimSpace(path) != "" && !util.IsIgnoredPath(path) {
			paths = append(paths, path)
		}
	}
	if len(paths) == 0 {
		return ""
	}

	now := time.Now().Unix()
	payload := make([]streamedFile, 0, len(paths))
	for _, path := range paths {
		payload = append(payload, streamedFile{
			Path:        path,
			Content:     files[path],
			Language:    util.LanguageForPath(path),
			WorkerRole:  workerRole,
			ManagerRole: managerRole,
			Status:      "ready",
			UpdatedAt:   now,
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}
