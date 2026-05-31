package boss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"nodes/internal/service/rules"
	"nodes/pkg/database"
	"nodes/pkg/models"

	"github.com/google/uuid"
)

// ExecuteTask — основной поток обработки задачи от apigateway.
// Раньше это был streaming gRPC-эндпоинт; теперь это обычная функция,
// которая отдаёт прогресс через rules.ProgressFunc — а сервер уже
// сам решает, как этот прогресс отправлять во внешний gRPC stream.
func (s *Service) ExecuteTask(ctx context.Context, req *CreateTaskRequest, progress rules.ProgressFunc) error {
	taskID := uuid.New()
	log.Printf("Received task from %s (user_id=%s): %s (task_id=%s)", req.Username, req.UserId, req.Title, taskID.String())
	issueTarget := s.detectGitHubIssueTask(ctx, req, taskID.String())

	task, err := s.persistTask(taskID, req)
	if err != nil {
		emit(progress, 0, "Database error: "+err.Error(), errorData())
		return err
	}
	emit(progress, 10, "Task saved to database", nil)

	req.Grade = gradeTask(req.Title + "\n" + req.Description)
	emit(progress, 12, fmt.Sprintf("Task graded: %d/10", req.Grade), nil)

	provider, model := pickProviderModel(req.Meta)
	emit(progress, 15, fmt.Sprintf("AI client initialized (%s/%s)", provider, model), nil)

	emit(progress, 20, "Boss thinking about architecture...", nil)
	decision, err := s.thinkAboutTask(ctx, provider, model, req)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, "AI planning failed: "+err.Error(), errorData())
		return err
	}
	// Safeguard: AI sometimes returns 0 managers or empty roles
	if decision.ManagersCount <= 0 {
		decision.ManagersCount = 1
		log.Printf("Boss decision corrected: managers_count was 0, set to 1")
	}

	// === Hard fallback for bad AI responses ===
	// If AI returns 0 managers or empty roles, create a default one
	if len(decision.ManagerRoles) == 0 || decision.ManagersCount == 0 {
		decision.ManagersCount = 1
		decision.ManagerRoles = []models.ManagerRole{
			{
				Role:        "development",
				Description: "Implement the requested functionality",
				Priority:    1,
			},
		}
		log.Printf("Boss: Applied fallback - created default 'development' manager")
	}

	emit(progress, 30, "Architecture planned by AI", map[string]string{
		"managers": strconv.Itoa(int(decision.ManagersCount)),
	})
	s.saveBossDecision(task.ID, decision)

	projectPath, err := s.setupProject(ctx, taskID.String(), req.Title, issueTarget)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, err.Error(), errorData())
		return err
	}

	// Если это доработка существующего проекта — клонируем репозиторий с GitHub
	if req.ExistingRepoUrl != "" {
		log.Printf("Refinement mode: cloning existing repo %s", req.ExistingRepoUrl)
		os.RemoveAll(projectPath)
		if err := os.MkdirAll(projectPath, 0755); err != nil {
			return err
		}
		cmd := exec.Command("git", "clone", req.ExistingRepoUrl, projectPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to clone existing repo: %v", err)
		}
	}
	defer s.cleanupProject(projectPath)
	if issueTarget != nil && issueTarget.Cloned {
		appendRepositoryContext(decision, projectPath)
	}

	// При доработке не восстанавливаем проект из старых данных
	if req.IsRefinement {
		log.Printf("Refinement mode enabled — skipping project restore")
	}

	emit(progress, 40, fmt.Sprintf("Creating %d managers in parallel", decision.ManagersCount), nil)
	managerResults, err := s.assignManagersParallel(ctx, taskID.String(), decision, req, progress, projectPath)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, "Managers failed: "+err.Error(), errorData())
		return err
	}
	if len(managerResults) == 0 {
		log.Printf("Boss: No manager results. decision.ManagersCount=%d, ManagerRoles=%v", 
			decision.ManagersCount, decision.ManagerRoles)
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, "No solution generated", errorData())
		return fmt.Errorf("no solution generated")
	}

	integrationBranch := "main"
	if issueTarget != nil && issueTarget.Cloned && issueTarget.BranchName != "" {
		integrationBranch = issueTarget.BranchName
	}
	s.mergeManagerBranches(projectPath, decision.ManagerRoles, integrationBranch)
	emit(progress, 80, "Boss validating solution...", nil)
	tokens := req.Tokens
	if tokens == nil {
		tokens = map[string]string{}
	}
	tokens["title"] = req.Title
	validation := s.validateSolution(ctx, provider, model, tokens, decision, managerResults)
	if !validation.Approved {
		log.Printf("Boss validation rejected: %s", validation.Feedback)
	}

	emit(progress, 90, "Packaging project", nil)
	// Для не-кодовых задач (research/document/presentation) публикация в GitHub
	// не требуется (см. issue): результат отдаётся во вкладку Solution и в чат.
	// Исключение — задача привязана к конкретному GitHub issue.
	repoURL := ""
	isCodeTask := decision.TaskType == "" || decision.TaskType == TaskTypeCode
	if isCodeTask || issueTarget != nil {
		repoURL = s.pushToGitHub(ctx, task, projectPath, issueTarget)
	} else {
		log.Printf("Skipping GitHub publish for %s task (delivered to Solution tab)", decision.TaskType)
	}

	task.Status = "done"
	database.Db.Save(task)

	data := map[string]string{
		"managers":  strconv.Itoa(int(decision.ManagersCount)),
		"techStack": util_stack(decision.TechStack),
		"taskType":  decision.TaskType,
	}
	if validation != nil && validation.Feedback != "" {
		data["bossReview"] = validation.Feedback
	}

	// Для не-кодовых задач (research/document/presentation) босс отдаёт короткий
	// ответ в чат, а полный результат уже лежит в папке solution/ (вкладка Solution).
	if decision.TaskType != "" && decision.TaskType != TaskTypeCode {
		fullDoc, docCount := collectSolutionMarkdown(managerResults)
		// Когда воркеров несколько, менеджерский синтез сливает их результаты
		// в один документ, проверяя факты (как описано в issue: «менеджер
		// объединяет результаты, делает свой ресёрч и проверяет достоверность»).
		if docCount > 1 {
			if synth := s.synthesizeSolution(ctx, provider, model, tokens, decision, req.Title+"\n"+req.Description, fullDoc); synth != "" {
				fullDoc = synth
			}
		}
		if fullDoc != "" {
			data["solutionMarkdown"] = fullDoc
			data["chatSummary"] = s.buildChatSummary(ctx, provider, model, tokens, decision.TaskType, req.Title+"\n"+req.Description, fullDoc)
		}
	}
	if repoURL != "" && strings.HasPrefix(repoURL, "https://") {
		data["repoUrl"] = repoURL
	}
	if issueTarget != nil && repoURL != "" {
		data["githubMode"] = "pull_request"
		data["pullRequestUrl"] = repoURL
		data["githubIssueUrl"] = issueTarget.IssueURL
	}
	emit(progress, 100, "Project ready! "+task.Title+" created successfully", data)
	return nil
}

// persistTask — сохраняет первоначальное состояние задачи в БД
func (s *Service) persistTask(taskID uuid.UUID, req *CreateTaskRequest) (*models.Task, error) {
	task := &models.Task{
		ID:          taskID,
		UserID:      req.UserId,
		Username:    req.Username,
		Title:       req.Title,
		Description: req.Description,
		Status:      "boss_planning",
	}
	tokensJSON, _ := json.Marshal(req.Tokens)
	metaJSON, _ := json.Marshal(req.Meta)
	task.Tokens = string(tokensJSON)
	task.Meta = string(metaJSON)
	if err := database.Db.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

// saveBossDecision — сохраняет решение босса в БД
func (s *Service) saveBossDecision(taskID uuid.UUID, decision *DecisionResult) {
	d := &models.BossDecision{
		ID:                   uuid.New(),
		TaskID:               taskID,
		Status:               "planning",
		TaskType:             decision.TaskType,
		ManagersCount:        decision.ManagersCount,
		TechnicalDescription: decision.TechnicalDescription,
		ArchitectureNotes:    decision.ArchitectureNotes,
	}
	rolesJSON, _ := json.Marshal(decision.ManagerRoles)
	stackJSON, _ := json.Marshal(decision.TechStack)
	d.ManagerRoles = string(rolesJSON)
	d.TechStack = string(stackJSON)
	if err := database.Db.Create(d).Error; err != nil {
		log.Printf("Failed to save boss decision: %v", err)
	}
}

// pickProviderModel — извлекает provider/model из metadata запроса
func pickProviderModel(meta map[string]string) (provider, model string) {
	model = meta["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}
	provider = meta["provider"]
	if provider == "" {
		provider = "openai"
	}
	return
}

// emit — отправляет апдейт прогресса, если callback задан
func emit(progress rules.ProgressFunc, p int32, msg string, data map[string]string) {
	if progress != nil {
		progress(p, msg, data)
	}
}

// errorData — общая метка для ошибочных апдейтов
func errorData() map[string]string {
	return map[string]string{"status": "error"}
}

// util_stack — компактное представление tech stack для metadata апдейта
func util_stack(stack []string) string {
	b, _ := json.Marshal(stack)
	return string(b)
}
