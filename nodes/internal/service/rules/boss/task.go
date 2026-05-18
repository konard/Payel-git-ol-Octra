package boss

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	emit(progress, 30, "Architecture planned by AI", map[string]string{
		"managers": strconv.Itoa(int(decision.ManagersCount)),
	})
	s.saveBossDecision(task.ID, decision)

	projectPath, err := s.setupProject(taskID.String(), req.Title)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, err.Error(), errorData())
		return err
	}
	defer s.cleanupProject(projectPath)

	emit(progress, 40, fmt.Sprintf("Creating %d managers in parallel", decision.ManagersCount), nil)
	managerResults, err := s.assignManagersParallel(ctx, taskID.String(), decision, req, progress, projectPath)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, "Managers failed: "+err.Error(), errorData())
		return err
	}
	if len(managerResults) == 0 {
		task.Status = "error"
		database.Db.Save(task)
		emit(progress, 0, "No solution generated", errorData())
		return fmt.Errorf("no solution generated")
	}

	s.mergeManagerBranches(projectPath, decision.ManagerRoles)
	emit(progress, 80, "Boss validating solution...", nil)
	tokens := req.Tokens
	if tokens == nil {
		tokens = map[string]string{}
	}
	tokens["title"] = req.Title
	s.validateSolution(ctx, provider, model, tokens, decision, managerResults)

	emit(progress, 90, "Packaging project", nil)
	repoURL := s.pushToGitHub(ctx, task, projectPath)

	task.Status = "done"
	database.Db.Save(task)

	data := map[string]string{
		"managers":  strconv.Itoa(int(decision.ManagersCount)),
		"techStack": util_stack(decision.TechStack),
	}
	if repoURL != "" && strings.HasPrefix(repoURL, "https://") {
		data["repoUrl"] = repoURL
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

// gradeTask calls the HTTP grader and returns complexity 1-10 (default 5 on error)
func gradeTask(taskText string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	body := map[string]string{"task": taskText}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "http://octra-grader:50055/grade", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 5
	}
	defer resp.Body.Close()
	var res struct {
		Grade int `json:"grade"`
	}
	if json.NewDecoder(resp.Body).Decode(&res) == nil && res.Grade >= 1 && res.Grade <= 10 {
		return res.Grade
	}
	// fallback: try to read raw int from body
	raw, _ := io.ReadAll(resp.Body)
	if len(raw) > 0 {
		var g int
		if _, err := fmt.Sscanf(string(raw), "%d", &g); err == nil && g >= 1 && g <= 10 {
			return g
		}
	}
	return 5
}
