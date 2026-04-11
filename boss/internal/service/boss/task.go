package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/service"
	"boss/pkg/database"
	"boss/pkg/models"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// CreateTaskStream — streaming version that sends real-time updates
func (s *BossService) CreateTaskStream(req *bosspb.CreateTaskRequest, stream bosspb.BossService_CreateTaskStreamServer) error {
	ctx := stream.Context()
	log.Printf("🎯 Received task from %s: %s", req.Username, req.Title)

	taskID := uuid.New()

	// Send initial update
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "📝 Task received and processing started",
		Progress:  5,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 1. Save task to DB
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
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ Database error: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "💾 Task saved to database",
		Progress:  10,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 2. Get model and provider
	model := req.Meta["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}
	provider := "openai"
	if p, ok := req.Meta["provider"]; ok {
		provider = p
	}

	agentsClient, err := service.NewAgentClientWrapper()
	if err != nil {
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ Failed to initialize AI client: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}
	defer agentsClient.Close()

	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "🤖 AI client initialized (" + provider + "/" + model + ")",
		Progress:  15,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 3. Check if we should use AI planning or predefined workflow
	var decision *BossDecisionResult
	var planErr error

	useAI := req.UseAiPlanning || len(req.PredefinedManagers) == 0
	if useAI {
		// Use AI to plan architecture
		_ = stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "🤖 AI is planning architecture...",
			Progress:  15,
			Status:    "processing",
			Timestamp: time.Now().Unix(),
		})

		decision, planErr = s.thinkAboutTask(ctx, agentsClient, provider, model, req)
		if planErr != nil {
			task.Status = "error"
			database.Db.Save(task)
			return stream.Send(&bosspb.TaskUpdate{
				TaskId:    taskID.String(),
				Message:   "❌ AI planning failed: " + planErr.Error(),
				Status:    "error",
				Timestamp: time.Now().Unix(),
			})
		}

		_ = stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "✅ Architecture planned by AI",
			Progress:  30,
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"managers": strconv.Itoa(int(decision.ManagersCount)),
			},
		})
	} else {
		// Use predefined workflow from user
		_ = stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "📋 Using predefined workflow (" + strconv.Itoa(len(req.PredefinedManagers)) + " managers)",
			Progress:  15,
			Status:    "processing",
			Timestamp: time.Now().Unix(),
		})

		decision = s.buildDecisionFromPredefined(req)

		_ = stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "✅ Predefined workflow loaded",
			Progress:  30,
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"managers": strconv.Itoa(int(decision.ManagersCount)),
			},
		})
	}

	// 4. Save boss decision
	bossDecision := &models.BossDecision{
		ID:                   uuid.New(),
		TaskID:               task.ID,
		Status:               "planning",
		ManagersCount:        decision.ManagersCount,
		TechnicalDescription: decision.TechnicalDescription,
		ArchitectureNotes:    decision.ArchitectureNotes,
	}

	rolesJSON, _ := json.Marshal(decision.ManagerRoles)
	stackJSON, _ := json.Marshal(decision.TechStack)
	bossDecision.ManagerRoles = string(rolesJSON)
	bossDecision.TechStack = string(stackJSON)

	if err := database.Db.Create(bossDecision).Error; err != nil {
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ Failed to save decision: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "🏗️ Creating " + strconv.Itoa(int(decision.ManagersCount)) + " managers in parallel",
		Progress:  40,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 5. Call managers in parallel with progress updates
	managerResults, zipData, err := s.assignManagersParallelWithProgress(ctx, task.ID.String(), decision, req, stream)
	if err != nil {
		log.Printf("Error from managers: %v", err)
		task.Status = "error"
		database.Db.Save(task)
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ Manager failed: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "👷 All managers completed code generation",
		Progress:  70,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 6. Boss validation
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "🔍 Boss validating solution...",
		Progress:  80,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	if len(managerResults) > 0 {
		_, _ = s.validateSolution(ctx, agentsClient, provider, model, req.Tokens, decision, managerResults, zipData)
	}

	if len(zipData) == 0 {
		task.Status = "error"
		database.Db.Save(task)
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ No solution generated",
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "📦 Packaging project",
		Progress:  90,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 7. Save ZIP solution
	task.Solution = zipData
	task.Status = "done"
	database.Db.Save(task)

	log.Printf("✅ Task completed! ZIP size: %d bytes", len(zipData))

	techStackBytes, _ := json.Marshal(decision.TechStack)
	return stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "🎉 Project ready! " + task.Title + " created successfully",
		Progress:  100,
		Status:    "success",
		Timestamp: time.Now().Unix(),
		Data: map[string]string{
			"managers":  strconv.Itoa(int(decision.ManagersCount)),
			"techStack": string(techStackBytes),
		},
	})
}

// CreateTask accepts task, executes full cycle and returns ZIP
func (s *BossService) CreateTask(ctx context.Context, req *bosspb.CreateTaskRequest) (*bosspb.BossDecision, error) {
	log.Printf("Received task from %s: %s", req.Username, req.Title)

	// 1. Save task to DB
	task := &models.Task{
		ID:          uuid.New(),
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

	// 2. Get model and provider
	model := req.Meta["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}
	provider := "openai"
	if p, ok := req.Meta["provider"]; ok {
		provider = p
	}

	// Create agents client
	agentsClient, err := service.NewAgentClientWrapper()
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		return &bosspb.BossDecision{
			TaskId:       task.ID.String(),
			Status:       "error",
			ErrorMessage: "Failed to initialize AI client: " + err.Error(),
		}, nil
	}
	defer agentsClient.Close()

	// 3. Boss thinks about task
	decision, err := s.thinkAboutTask(ctx, agentsClient, provider, model, req)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		return &bosspb.BossDecision{
			TaskId:       task.ID.String(),
			Status:       "error",
			ErrorMessage: "Failed to plan task: " + err.Error(),
		}, nil
	}

	// 4. Save boss decision
	bossDecision := &models.BossDecision{
		ID:                   uuid.New(),
		TaskID:               task.ID,
		Status:               "planning",
		ManagersCount:        decision.ManagersCount,
		TechnicalDescription: decision.TechnicalDescription,
		ArchitectureNotes:    decision.ArchitectureNotes,
	}

	rolesJSON, _ := json.Marshal(decision.ManagerRoles)
	stackJSON, _ := json.Marshal(decision.TechStack)
	bossDecision.ManagerRoles = string(rolesJSON)
	bossDecision.TechStack = string(stackJSON)

	if err := database.Db.Create(bossDecision).Error; err != nil {
		return nil, err
	}

	task.Status = "managers_assigned"
	database.Db.Save(task)

	// 5. Call Manager service for EACH manager IN PARALLEL
	log.Printf("Calling Manager service: %d managers in parallel", decision.ManagersCount)

	managerResults, zipData, err := s.assignManagersParallel(ctx, task.ID.String(), decision, req, nil)
	if err != nil {
		log.Printf("Error from managers: %v", err)
		task.Status = "error"
		database.Db.Save(task)
		return &bosspb.BossDecision{
			TaskId:       task.ID.String(),
			Status:       "error",
			ErrorMessage: "Failed to execute task: " + err.Error(),
		}, nil
	}

	if len(zipData) == 0 {
		task.Status = "error"
		database.Db.Save(task)
		return &bosspb.BossDecision{
			TaskId:       task.ID.String(),
			Status:       "error",
			ErrorMessage: "No solution generated",
		}, nil
	}

	// 6. Boss validates the final result
	log.Printf("Boss validating solution...")
	validationResult, err := s.validateSolution(ctx, agentsClient, provider, model, req.Tokens, decision, managerResults, zipData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		// Don't fail on validation error, just log it
	} else if !validationResult.Approved {
		log.Printf("Boss rejected solution: %s", validationResult.Feedback)
		task.Status = "reviewing"
		// TODO: send back for rework
		// For now, accept it but log the feedback
	}

	// 7. Save ZIP solution
	task.Solution = zipData
	task.Status = "done"
	database.Db.Save(task)

	log.Printf("Task completed! ZIP size: %d bytes", len(zipData))

	// 8. Return result with ZIP
	return &bosspb.BossDecision{
		TaskId:               task.ID.String(),
		Status:               "done",
		ManagersCount:        decision.ManagersCount,
		ManagerRoles:         decision.ManagerRolesProto(),
		TechnicalDescription: decision.TechnicalDescription,
		TechStack:            decision.TechStack,
		ArchitectureNotes:    decision.ArchitectureNotes,
		Solution:             zipData,
	}, nil
}

func (s *BossService) thinkAboutTask(ctx context.Context, agentsClient *service.AgentClientWrapper, provider, model string, req *bosspb.CreateTaskRequest) (*BossDecisionResult, error) {
	log.Printf("Boss thinking about task: %s", req.Title)

	prompt := `You are CTO. Analyze the task and decide what manager roles are needed.

Title: ` + req.Title + `
Description: ` + req.Description + `

IMPORTANT RULES:
- Create ONLY the managers that are actually needed for this specific task
- Do NOT create frontend manager if the task is backend-only, API, CLI tool, or infrastructure
- Do NOT add unnecessary managers just for the sake of having them
- Each manager must have a clear purpose related to the task
- Typical projects need 1-3 managers, not more

Reply ONLY with JSON:
{
  "managers_count": 2,
  "manager_roles": [
    {"role": "backend", "description": "Backend API and business logic", "priority": 1},
    {"role": "testing", "description": "Unit and integration tests", "priority": 2}
  ],
  "technical_description": "Technical description...",
  "tech_stack": ["Python", "FastAPI", "Redis"],
  "architecture_notes": "Architecture notes..."
}

Choose the tech stack based on what the task actually needs — don't copy the example.
If the task is a CLI tool: Go, Cobra, etc.
If it's a web app: React, Node.js, etc.
If it's a proxy/API: Go, gorilla/mux, etc.
Always include only the technologies that are genuinely required.`

	log.Printf("Sending request to AI...")
	resp, err := agentsClient.GenerateFromTask(ctx, provider, model, prompt, req.Tokens)
	if err != nil {
		log.Printf("AI error: %v", err)
		return nil, err
	}

	// Извлекаем JSON из markdown (```json ... ``` или просто ``` ... ```)
	jsonStr := extractJSONFromMarkdown(resp)

	var decision BossDecisionResult
	if err := json.Unmarshal([]byte(jsonStr), &decision); err != nil {
		log.Printf("JSON parse error: %v", err)
		log.Printf("Raw response: %s", resp)
		return nil, err
	}

	log.Printf("Boss made decision: %+v", decision)
	return &decision, nil
}

// buildDecisionFromPredefined создаёт BossDecision из пользовательского workflow
func (s *BossService) buildDecisionFromPredefined(req *bosspb.CreateTaskRequest) *BossDecisionResult {
	log.Printf("📋 Building decision from predefined workflow: %d managers", len(req.PredefinedManagers))

	// Преобразуем ManagerConfig в ManagerRole (models)
	managerRoles := make([]models.ManagerRole, 0, len(req.PredefinedManagers))
	workerRolesMap := make(map[string][]models.WorkerRole)

	for _, mc := range req.PredefinedManagers {
		managerRoles = append(managerRoles, models.ManagerRole{
			Role:        mc.Role,
			Description: mc.Description,
			Priority:    mc.Priority,
		})

		// Сохраняем worker roles для этого менеджера
		if len(mc.Workers) > 0 {
			workers := make([]models.WorkerRole, 0, len(mc.Workers))
			for _, w := range mc.Workers {
				workers = append(workers, models.WorkerRole{
					Role:        w.Role,
					Description: w.Description,
				})
			}
			workerRolesMap[mc.Role] = workers
		}
	}

	// Если пользователь не указал архитектуру, используем описание задачи
	techDescription := req.PredefinedArchitecture
	if techDescription == "" {
		techDescription = "Predefined workflow for: " + req.Title + "\n" + req.Description
	}

	// Если пользователь не указал стек, оставляем пустым
	techStack := req.PredefinedTechStack
	if techStack == nil {
		techStack = []string{"custom"}
	}

	return &BossDecisionResult{
		ManagersCount:        int32(len(req.PredefinedManagers)),
		ManagerRoles:         managerRoles,
		TechnicalDescription: techDescription,
		TechStack:            techStack,
		ArchitectureNotes:    "Predefined workflow by user",
		ManagerWorkerRoles:   workerRolesMap,
	}
}

// extractJSONFromMarkdown извлекает JSON из markdown-обёртки
