package boss

import (
	"context"
	"log"
	"strconv"
	"strings"

	"orchestrator/internal/config"
	gh "orchestrator/internal/service/github"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/rules/universal"

	"github.com/google/uuid"
)

const universalRole = "universal"

func shouldUseUniversalNode(req *CreateTaskRequest, issueTarget *gh.IssueTarget) bool {
	if req == nil {
		return false
	}
	if req.ExistingRepoUrl != "" || req.IsRefinement {
		return false
	}
	if issueTarget != nil {
		return false
	}

	if hasExplicitUniversalManager(req) {
		return true
	}

	if len(req.PredefinedManagers) > 0 {
		return false
	}

	maxGrade := config.UniversalNodeMaxGradeResolved()
	if maxGrade <= 0 {
		return false
	}
	if req.Grade <= 0 || req.Grade > maxGrade {
		return false
	}
	return true
}

func hasExplicitUniversalManager(req *CreateTaskRequest) bool {
	for _, m := range req.PredefinedManagers {
		if strings.EqualFold(m.Role, universalRole) {
			return true
		}
	}
	return false
}

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

	solverReq := &universal.SolveRequest{
		Title:       req.Title,
		Description: req.Description,
		TaskType:    decision.TaskType,
		TechStack:   techStack,
		Provider:    provider,
		Model:       model,
		Tokens:      req.Tokens,
		ProjectPath: projectPath,
		Progress: func(p int32, msg string, data map[string]string) {
			if progress != nil {
				progress(p, msg, data)
			}
		},
	}

	result := s.universalSolver.Solve(ctx, s.agentsClient, solverReq)
	if result == nil || len(result.Files) == 0 {
		log.Printf("Universal node produced no solution")
		return nil
	}

	written := result.Files
	worker := &rules.WorkerResult{
		WorkerId:   uuid.New().String(),
		Role:       universalRole,
		TaskMd:     req.Title + "\n" + req.Description,
		SolutionMd: "Solved directly by the universal node (trivial task).",
		Files:      written,
		Success:    true,
		Approved:   true,
	}
	mgrResult := &rules.ManagerResult{
		TaskId:        taskID,
		ManagerId:     uuid.New().String(),
		Role:          universalRole,
		Status:        "done",
		WorkerResults: []*rules.WorkerResult{worker},
		ReviewSummary: "Trivial task handled by a single universal node — full Boss/Manager/Worker pipeline skipped.",
	}
	results := []*rules.ManagerResult{mgrResult}

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
