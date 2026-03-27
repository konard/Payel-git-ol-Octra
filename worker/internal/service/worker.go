package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"worker/internal/fetcher/grpc/workerpb"
	"worker/pkg/database"
	"worker/pkg/llm"
	"worker/pkg/models"

	"github.com/google/uuid"
)

// WorkerService — сервис рабочих
type WorkerService struct {
	workerpb.UnimplementedWorkerServiceServer
}

func NewWorkerService() *WorkerService {
	return &WorkerService{}
}

// AssignWorkers принимает задачу от менеджера и создаёт воркеров
func (s *WorkerService) AssignWorkers(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	log.Printf("Получена задача от менеджера %s: %s", req.ManagerId, req.ManagerRole)
	log.Printf("Воркеры: %v", req.WorkerRoles)

	// Парсим токены и настройки модели из метаданных
	tokens := parseTokens(req.Metadata)
	model := req.Metadata["model"]
	modelURL := req.Metadata["modelUrl"]

	// Создаём LLM клиент
	llmClient := llm.NewOpenRouterClient(modelURL, model, tokens)

	workers := make([]*workerpb.WorkerInfo, 0, len(req.WorkerRoles))

	for _, role := range req.WorkerRoles {
		workerID := uuid.New()

		// Создаём воркера в БД
		worker := &models.Worker{
			ID:        workerID,
			TaskID:    uuid.MustParse(req.TaskId),
			ManagerID: uuid.MustParse(req.ManagerId),
			Role:      role.Role,
			Status:    "thinking",
			TaskMD:    req.TaskMd,
		}

		if err := database.Db.Create(worker).Error; err != nil {
			return &workerpb.AssignWorkersResponse{
				TaskId:       req.TaskId,
				Status:       "error",
				ErrorMessage: fmt.Sprintf("failed to create worker: %v", err),
			}, nil
		}

		// Воркер думает над задачей через ИИ
		solution, err := s.thinkAboutTask(ctx, llmClient, req.TaskMd, role.Role)
		if err != nil {
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		// Генерирует код
		files, err := s.generateCode(ctx, llmClient, solution, role.Role, req.ManagerRole)
		if err != nil {
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		// Сохраняем решение
		worker.SolutionMD = solution
		worker.Files = marshalJSON(files)
		worker.Status = "done"
		worker.Success = true
		database.Db.Save(worker)

		// Собираем информацию о воркере
		workers = append(workers, &workerpb.WorkerInfo{
			Id:         workerID.String(),
			Role:       role.Role,
			Status:     "done",
			SolutionMd: solution,
			Files:      getFileNames(files),
		})
	}

	return &workerpb.AssignWorkersResponse{
		TaskId:  req.TaskId,
		Status:  "success",
		Workers: workers,
	}, nil
}

// SubmitResult принимает результат от воркера на проверку
func (s *WorkerService) SubmitResult(ctx context.Context, req *workerpb.WorkerResult) (*workerpb.ReviewResponse, error) {
	log.Printf("Получен результат от воркера %s", req.WorkerId)

	// Находим воркера
	workerID := uuid.MustParse(req.WorkerId)
	var worker models.Worker
	if err := database.Db.First(&worker, "id = ?", workerID).Error; err != nil {
		return &workerpb.ReviewResponse{
			Approved: false,
			Feedback: "Worker not found",
		}, nil
	}

	// ИИ проверяет код
	filesStr := make(map[string]string)
	for k, v := range req.Files {
		filesStr[k] = string(v)
	}

	// Получаем токены и модель из метаданных (передаются менеджером)
	tokens := parseTokens(req.Metadata)
	model := req.Metadata["model"]
	modelURL := req.Metadata["modelUrl"]

	llmClient := llm.NewOpenRouterClient(modelURL, model, tokens)

	approved, feedback := s.reviewCode(ctx, llmClient, filesStr, req.SolutionMd, worker.TaskMD)

	worker.Approved = approved
	worker.Feedback = feedback
	database.Db.Save(&worker)

	return &workerpb.ReviewResponse{
		Approved: approved,
		Feedback: feedback,
	}, nil
}

// thinkAboutTask — воркер думает над задачей
func (s *WorkerService) thinkAboutTask(ctx context.Context, llmClient llm.Client, taskMD, role string) (string, error) {
	prompt := `Ты опытный разработчик (` + role + `). Тебе дали задачу:

` + taskMD + `

Подумай над решением:
1. Какую архитектуру выбрать
2. Какие паттерны использовать
3. Какие могут быть проблемы

Напиши подробное решение в формате SOLUTION.md`

	resp, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	log.Printf("Воркер принял решение: %s", truncate(resp, 100))
	return resp, nil
}

// generateCode — генерирует код на основе решения
func (s *WorkerService) generateCode(ctx context.Context, llmClient llm.Client, solution, role, managerRole string) (map[string]string, error) {
	prompt := `Ты разработчик (` + role + `) в команде под управлением ` + managerRole + `.

У тебя есть решение:
` + truncate(solution, 500) + `

Сгенерируй готовый код. Верни ТОЛЬКО JSON в формате:
{
  "src/main.go": "package main...",
  "src/handler.go": "package handler...",
  "README.md": "# Project..."
}

Никакого markdown, только JSON.`

	resp, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var files map[string]string
	if err := json.Unmarshal([]byte(resp), &files); err != nil {
		return nil, err
	}

	log.Printf("Сгенерировано файлов: %d", len(files))
	return files, nil
}

// reviewCode — ИИ проверка кода
func (s *WorkerService) reviewCode(ctx context.Context, llmClient llm.Client, files map[string]string, solutionMD, taskMD string) (bool, string) {
	filesJSON, _ := json.Marshal(files)

	prompt := `Ты senior разработчик на code review.

Задача:
` + taskMD + `

Решение:
` + truncate(solutionMD, 300) + `

Файлы:
` + string(filesJSON)[:min(len(filesJSON), 1000)] + `

Проверь:
1. Код решает задачу?
2. Нет ли ошибок?
3. Следует ли лучшим практикам?

Ответь в формате JSON:
{
  "approved": true,
  "feedback": "Код хороший"
}

Или если есть проблемы:
{
  "approved": false,
  "feedback": "Нужно исправить..."
}`

	resp, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return false, "Error during review: " + err.Error()
	}

	var result struct {
		Approved bool   `json:"approved"`
		Feedback string `json:"feedback"`
	}

	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return false, "Error parsing review: " + err.Error()
	}

	return result.Approved, result.Feedback
}

// Вспомогательные функции

func marshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func getFileNames(files map[string]string) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func StringsHas(s, substr string) bool {
	return strings.Contains(s, substr)
}

// parseTokens — парсит токены из метаданных
func parseTokens(metadata map[string]string) []string {
	// Ожидаем что токены переданы как JSON массив в metadata["tokens"]
	tokensJSON := metadata["tokens"]
	if tokensJSON == "" {
		// Дефолтные токены для тестирования
		return []string{
			"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",
			"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",
			"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",
		}
	}

	var tokens []string
	if err := json.Unmarshal([]byte(tokensJSON), &tokens); err != nil {
		return []string{tokensJSON} // Если не JSON, считаем что это один токен
	}

	return tokens
}
