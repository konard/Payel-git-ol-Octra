package boss

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"orchestrator/internal/config"
	"orchestrator/internal/prompts"
	gh "orchestrator/internal/service/github"
	"orchestrator/internal/service/git"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"

	"github.com/google/uuid"
)

// universalRole — роль синтетического менеджера/воркера, под которой результат
// универсальной ноды проходит остальной конвейер.
const universalRole = "universal"

// shouldUseUniversalNode решает, обрабатывать ли задачу одной «универсальной
// нодой» вместо полного конвейера Boss → Manager → Worker (issue #91).
//
// Решение основано на ЧИСТОМ анализе сложности (оценке модели 1-10), а не на
// триггер-словах: задачу берёт быстрый путь только если модель сама оценила её
// как тривиальную (grade в пределах порога). Сложные задачи как и раньше идут
// через полный конвейер.
//
// Быстрый путь намеренно не применяется, когда:
//   - порог отключён (OCTRA_DISABLE_UNIVERSAL_NODE) или grade вне диапазона;
//   - пользователь задал собственный workflow (PredefinedManagers) — его надо уважать;
//   - это доработка существующего репозитория (ExistingRepoUrl/IsRefinement);
//   - задача привязана к GitHub issue — у неё свой конвейер публикации PR.
func shouldUseUniversalNode(req *CreateTaskRequest, decision *DecisionResult, issueTarget *gh.IssueTarget) bool {
	if req == nil || decision == nil {
		return false
	}
	maxGrade := config.UniversalNodeMaxGradeResolved()
	if maxGrade <= 0 {
		return false
	}
	if req.Grade <= 0 || req.Grade > maxGrade {
		return false
	}
	if len(req.PredefinedManagers) > 0 {
		return false
	}
	if req.ExistingRepoUrl != "" || req.IsRefinement {
		return false
	}
	if issueTarget != nil {
		return false
	}
	return true
}

// solveUniversal запускает универсальную ноду: один AI-вызов выдаёт минимальный
// корректный результат, который упаковывается в один синтетический результат
// менеджера/воркера. Благодаря этому остальной конвейер ExecuteTask (запись на
// диск, nix, валидация, публикация, вкладка Solution) работает без изменений.
//
// Возвращает (nil, nil), если нода не смогла дать осмысленный результат — тогда
// вызывающий код откатывается на полный конвейер.
func (s *Service) solveUniversal(
	ctx context.Context,
	taskID string,
	decision *DecisionResult,
	req *CreateTaskRequest,
	progress rules.ProgressFunc,
	projectPath string,
) []*rules.ManagerResult {
	provider, model := pickProviderModel(req.Meta)
	techStack := ""
	if len(decision.TechStack) > 0 {
		techStack = decision.TechStack[0]
	}

	emit(progress, 45, "Universal node solving the task directly...", map[string]string{
		"task_type":    decision.TaskType,
		"current_role": universalRole,
	})

	prompt := prompts.UniversalNode(req.Title, req.Description, decision.TaskType, techStack)

	tokens := req.Tokens
	if tokens == nil {
		tokens = map[string]string{}
	}

	files := s.generateUniversalFiles(ctx, provider, model, prompt, tokens)
	if len(files) == 0 {
		log.Printf("Universal node produced no files — falling back to full pipeline")
		return nil
	}

	written := s.writeUniversalFiles(projectPath, files, progress)
	if len(written) == 0 {
		log.Printf("Universal node files failed validation — falling back to full pipeline")
		return nil
	}

	// Коммитим прямо в текущую ветку: универсальная нода не создаёт manager-ветки,
	// поэтому слияние веток для неё пропускается (см. ExecuteTask).
	if err := git.Add(projectPath); err != nil {
		log.Printf("Universal node git add failed: %v", err)
	} else if err := git.Commit(projectPath, "Universal node: "+req.Title); err != nil {
		log.Printf("Universal node git commit failed: %v", err)
	}

	worker := &rules.WorkerResult{
		WorkerId:   uuid.New().String(),
		Role:       universalRole,
		TaskMd:     req.Title + "\n" + req.Description,
		SolutionMd: "Solved directly by the universal node (trivial task).",
		Files:      written,
		Success:    true,
		Approved:   true,
	}
	result := &rules.ManagerResult{
		TaskId:        taskID,
		ManagerId:     uuid.New().String(),
		Role:          universalRole,
		Status:        "done",
		WorkerResults: []*rules.WorkerResult{worker},
		ReviewSummary: "Trivial task handled by a single universal node — full Boss/Manager/Worker pipeline skipped.",
	}

	results := []*rules.ManagerResult{result}
	if isCodeLikeTask(decision.TaskType) {
		if payload, count := collectCodeFilesPayload(results); payload != "" {
			emit(progress, 70, "Universal node produced the solution", map[string]string{
				"task_type":    decision.TaskType,
				"current_role": universalRole,
				"code_files":   payload,
				"filesCount":   strconv.Itoa(count),
			})
		}
	}
	emit(progress, 75, "Universal node finished", map[string]string{
		"task_type":    decision.TaskType,
		"current_role": universalRole,
	})
	return results
}

// generateUniversalFiles делает один AI-вызов (с детерминированной цепочкой
// фолбэк-провайдеров) и парсит из ответа карту файлов.
func (s *Service) generateUniversalFiles(ctx context.Context, provider, model, prompt string, tokens map[string]string) map[string]string {
	for _, p := range config.FallbackChain(provider, model) {
		resp, err := s.agentsClient.GenerateFromTask(ctx, p.Provider, p.Model, prompt, tokens)
		if err != nil {
			log.Printf("Universal node provider %s/%s failed: %v", p.Provider, p.Model, err)
			continue
		}
		if files := parseUniversalFiles(resp); len(files) > 0 {
			return files
		}
		log.Printf("Universal node: no files parsed from %s/%s response", p.Provider, p.Model)
	}
	return nil
}

// universalResponse — JSON-контракт ответа универсальной ноды.
type universalResponse struct {
	Files map[string]string `json:"files"`
}

// parseUniversalFiles извлекает карту файлов из JSON-ответа универсальной ноды.
func parseUniversalFiles(resp string) map[string]string {
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
	return files
}

// writeUniversalFiles пишет файлы на диск внутри projectPath и стримит каждый
// файл во вкладку Solution через progress. Возвращает фактически записанные файлы.
func (s *Service) writeUniversalFiles(projectPath string, files map[string]string, progress rules.ProgressFunc) map[string]string {
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
		if progress != nil {
			progress(60, "Writing file: "+path, map[string]string{
				"file":         path,
				"type":         "write",
				"current_role": universalRole,
			})
		}
	}
	return written
}
