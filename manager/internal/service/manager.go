package service

import (
	"context"
	"encoding/json"
	"log"

	"manager/internal/fetcher/grpc/managerpb"
	"manager/internal/fetcher/grpc/worker"
	"manager/pkg/database"
	"manager/pkg/llm"
	"manager/pkg/models"

	"github.com/google/uuid"
)

// ManagerService — сервис менеджеров
type ManagerService struct {
	managerpb.UnimplementedManagerServiceServer
	workerClient *worker.Client
}

func NewManagerService() *ManagerService {
	wClient, err := worker.NewClient("worker:50053")
	if err != nil {
		log.Printf("Warning: failed to connect to worker service: %v", err)
	}

	return &ManagerService{
		workerClient: wClient,
	}
}

// AssignManagersAndWait принимает задачу, создаёт менеджеров, ждёт Worker и возвращает ZIP
func (s *ManagerService) AssignManagersAndWait(ctx context.Context, req *managerpb.AssignManagersRequest) (*managerpb.AssignManagersResponse, error) {
	log.Printf("Получена задача от Boss: %s", req.TaskId)
	log.Printf("Роли менеджеров: %v", req.Roles)

	tokens := parseTokens(req.Metadata["tokens"])
	model := req.Metadata["model"]
	modelURL := req.Metadata["modelUrl"]
	llmClient := llm.NewLLMClient(modelURL, model, tokens)

	// Собираем ZIP архивы от всех менеджеров
	var allZipArchives [][]byte

	for _, role := range req.Roles {
		managerID := uuid.New()

		// Создаём менеджера в БД
		manager := &models.Manager{
			ID:       managerID,
			TaskID:   uuid.MustParse(req.TaskId),
			Role:     role,
			Status:   "active",
			TaskDesc: req.TechnicalDescription,
		}

		if err := database.Db.Create(manager).Error; err != nil {
			log.Printf("Ошибка создания менеджера: %v", err)
			continue
		}

		// Менеджер думает каких воркеров нанять
		workerRoles, err := s.managerThink(ctx, llmClient, req.TechnicalDescription, role)
		if err != nil {
			log.Printf("Ошибка ИИ для менеджера: %v", err)
			manager.Status = "error"
			database.Db.Save(manager)
			continue
		}

		// Вызываем Worker сервис и ждём ZIP архив
		if s.workerClient != nil {
			log.Printf("Вызов Worker сервиса для роли %s...", role)

			zipData, err := s.workerClient.AssignWorkersAndWait(
				ctx,
				req.TaskId,
				managerID.String(),
				role,
				req.TechnicalDescription,
				workerRoles,
				tokens,
				model,
				modelURL,
			)
			if err != nil {
				log.Printf("Ошибка вызова Worker: %v", err)
				manager.Status = "error"
			} else {
				log.Printf("Получен ZIP от Worker (%d байт)", len(zipData))
				allZipArchives = append(allZipArchives, zipData)
				manager.Status = "done"
			}
		}

		database.Db.Save(manager)
	}

	// Объединяем все ZIP архивы в один
	finalZip, err := mergeZipArchives(allZipArchives)
	if err != nil {
		log.Printf("Ошибка объединения ZIP: %v", err)
		finalZip = []byte{}
	}

	return &managerpb.AssignManagersResponse{
		TaskId:   req.TaskId,
		Status:   "success",
		Message:  "Project generated",
		Solution: finalZip,
	}, nil
}

func (s *ManagerService) managerThink(ctx context.Context, llmClient llm.Client, taskDesc, role string) ([]string, error) {
	log.Printf("Менеджер (%s) думает над задачей...", role)

	prompt := `Ты менеджер проекта (` + role + `). Задание:

` + taskDesc + `

Каких специалистов нужно нанять? Ответь JSON:
{
  "worker_roles": ["role1", "role2", "role3"]
}`

	resp, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		WorkerRoles []string `json:"worker_roles"`
	}

	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return []string{resp}, nil
	}

	log.Printf("Менеджер решил нанять: %v", result.WorkerRoles)
	return result.WorkerRoles, nil
}

func parseTokens(s string) []string {
	var tokens []string
	json.Unmarshal([]byte(s), &tokens)
	return tokens
}

func mergeZipArchives(zipDatas [][]byte) ([]byte, error) {
	// Если только один ZIP, возвращаем его
	if len(zipDatas) == 1 {
		return zipDatas[0], nil
	}

	// Если пусто, возвращаем пустой ZIP
	if len(zipDatas) == 0 {
		return make([]byte, 0), nil
	}

	// TODO: Реализовать объединение ZIP архивов
	// Пока возвращаем первый
	return zipDatas[0], nil
}
