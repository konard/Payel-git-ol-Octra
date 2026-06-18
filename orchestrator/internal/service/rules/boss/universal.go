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
	"orchestrator/internal/service/git"
	gh "orchestrator/internal/service/github"
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
// Важно: это решение принимается ДО boss-планирования. Универсальная нода не
// является веткой Boss → Universal; для тривиальных задач Boss вообще не
// строит архитектуру.
//
// Быстрый путь намеренно не применяется, когда:
//   - порог отключён (OCTRA_DISABLE_UNIVERSAL_NODE) или grade вне диапазона;
//   - пользователь задал собственный workflow (PredefinedManagers) — его надо уважать;
//   - это доработка существующего репозитория (ExistingRepoUrl/IsRefinement);
//   - задача привязана к GitHub issue — у неё свой конвейер публикации PR.
func shouldUseUniversalNode(req *CreateTaskRequest, issueTarget *gh.IssueTarget) bool {
	if req == nil {
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

// universalDecisionFromRequest builds the small amount of metadata the shared
// packaging tail needs after the direct universal path. It intentionally does
// not invent managers or call the Boss planner.
func universalDecisionFromRequest(req *CreateTaskRequest) *DecisionResult {
	if req == nil {
		return &DecisionResult{TaskType: TaskTypeCode}
	}
	taskType := normalizeTaskType(req.Meta["task_type"])
	techStack := detectTechStack(req.Title, req.Description)
	if taskType == "" {
		taskType = classifyTaskType(req.Title, req.Description)
		if taskType == TaskTypeCode && looksLikeDirectAnswer(req.Title, req.Description, len(techStack) > 0) {
			taskType = TaskTypeDocument
		}
	}
	if taskType == "" {
		taskType = TaskTypeCode
	}
	return &DecisionResult{
		TaskType:             taskType,
		ManagersCount:        0,
		TechnicalDescription: strings.TrimSpace(req.Title + "\n" + req.Description),
		TechStack:            techStack,
		ArchitectureNotes:    "Direct universal node execution; boss planning skipped.",
	}
}

func looksLikeDirectAnswer(title, description string, hasTechStack bool) bool {
	text := strings.ToLower(strings.TrimSpace(title + "\n" + description))
	if text == "" {
		return false
	}
	strongCodeSignals := []string{
		"hello world", "script", "program", "function", "server", "app",
		"application", "component", "class", "method", "implement", "bug",
		"fix ", "write in ", "write a ", "write an ", "create a ", "build a ",
		"код", "скрипт", "программ", "функци", "сервер", "приложени",
		"реализ", "создай", "сделай", "напиши",
	}
	if containsAny(text, strongCodeSignals) {
		return false
	}
	stackCodeSignals := []string{"code", "api", "апи"}
	if hasTechStack && containsAny(text, stackCodeSignals) {
		return false
	}
	answerSignals := []string{
		"?", "what is", "why", "how many", "calculate", "solve", "logic",
		"math", "proof", "explain", "сколько", "что такое", "почему", "объясни",
		"реши", "посчитай", "математ", "логичес", "доказ",
	}
	if containsAny(text, answerSignals) {
		return true
	}
	return strings.ContainsAny(text, "+=*/<>")
}

func directUniversalChatSummary(fullDoc string) string {
	fullDoc = strings.TrimSpace(fullDoc)
	if fullDoc == "" {
		return "The universal node finished the direct answer."
	}
	const maxSummary = 1200
	runes := []rune(fullDoc)
	if len(runes) <= maxSummary {
		return fullDoc
	}
	return strings.TrimSpace(string(runes[:maxSummary])) + "\n\n..."
}

// solveUniversal запускает универсальную ноду: один AI-вызов выдаёт минимальный
// корректный результат, который упаковывается в один синтетический результат
// менеджера/воркера. Благодаря этому остальной конвейер ExecuteTask (запись на
// диск, nix, публикация, вкладка Solution) работает без отдельного Boss-плана.
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
