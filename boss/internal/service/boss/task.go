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
	taskID := uuid.New()
	log.Printf("🎯 Received task from %s: %s (task_id=%s)", req.Username, req.Title, taskID.String())

	return s.executeTaskFlow(ctx, stream, taskID.String(), req.UserId, req)
}

// ResumeTaskStream — resumes streaming updates for an existing task
func (s *BossService) ResumeTaskStream(req *bosspb.ResumeTaskStreamRequest, stream bosspb.BossService_CreateTaskStreamServer) error {
	log.Printf("🔄 Resuming task stream for task_id=%s", req.TaskId)

	ctx := stream.Context()

	// Check if task exists and is still running
	var task models.Task
	if err := database.Db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   "❌ Task not found: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	// If task is already completed, send final status
	if task.Status == "done" || task.Status == "error" {
		log.Printf("✅ Task %s already completed with status: %s", req.TaskId, task.Status)

		// Send historical updates from Redis
		s.sendHistoricalUpdates(stream, req.TaskId)

		status := "success"
		if task.Status == "error" {
			status = "error"
		}

		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   "Task " + task.Title + " " + task.Status,
			Progress:  100,
			Status:    status,
			Timestamp: time.Now().Unix(),
		})
	}

	// If task is still running, resume streaming
	if task.Status == "boss_planning" || task.Status == "managers_assigned" || task.Status == "processing" {
		log.Printf("🔄 Task %s is still running (status: %s), resuming stream", req.TaskId, task.Status)

		// Send historical updates first
		s.sendHistoricalUpdates(stream, req.TaskId)

		// Resume the task flow from the beginning
		// The streamSender will persist progress to Redis, so we can continue
		return s.executeTaskFlow(ctx, stream, req.TaskId, req.UserId, &bosspb.CreateTaskRequest{
			UserId:        req.UserId,
			Username:      task.Username,
			Title:         task.Title,
			Description:   task.Description,
			Tokens:        parseJSONToMap(task.Tokens),
			Meta:          parseJSONToMap(task.Meta),
			UseAiPlanning: true, // Default to AI planning for resume
		})
	}

	// Unknown status
	return stream.Send(&bosspb.TaskUpdate{
		TaskId:    req.TaskId,
		Message:   "❌ Unknown task status: " + task.Status,
		Status:    "error",
		Timestamp: time.Now().Unix(),
	})
}

// sendHistoricalUpdates sends all historical updates from Redis
func (s *BossService) sendHistoricalUpdates(stream bosspb.BossService_CreateTaskStreamServer, taskID string) {
	if s.redisClient == nil || !s.redisClient.IsEnabled() {
		return
	}

	updates, err := s.redisClient.GetStreamUpdates(stream.Context(), taskID)
	if err != nil {
		log.Printf("Warning: failed to get historical updates for task %s: %v", taskID, err)
		return
	}

	for _, update := range updates {
		protoUpdate := &bosspb.TaskUpdate{
			TaskId:    update.TaskID,
			Message:   update.Message,
			Progress:  update.Progress,
			Status:    update.Status,
			Timestamp: update.Timestamp,
		}

		// Convert data back
		if update.Data != nil {
			if dataMap, ok := update.Data.(map[string]any); ok {
				protoUpdate.Data = make(map[string]string)
				for k, v := range dataMap {
					if str, ok := v.(string); ok {
						protoUpdate.Data[k] = str
					}
				}
			}
		}

		if err := stream.Send(protoUpdate); err != nil {
			log.Printf("Warning: failed to send historical update: %v", err)
			return
		}
	}

	if len(updates) > 0 {
		log.Printf("📜 Sent %d historical updates for task %s", len(updates), taskID)
	}
}

// executeTaskFlow executes the full task flow with streaming updates
func (s *BossService) executeTaskFlow(
	ctx context.Context,
	stream bosspb.BossService_CreateTaskStreamServer,
	taskID string,
	userID string,
	req *bosspb.CreateTaskRequest,
) error {
	sender := newStreamSender(stream, taskID, userID, s.redisClient)

	// Check if task already exists (for resume scenarios)
	var task models.Task
	existingTaskID := parseUUID(taskID)
	if err := database.Db.First(&task, "id = ?", existingTaskID).Error; err == nil {
		// Task exists, this is a resume - skip initial update and start from current progress
		log.Printf("🔄 Resuming existing task %s (status: %s)", taskID, task.Status)
		// For now, we'll restart from the beginning but with existing task
		// In a full implementation, we'd need to track progress more granularly
	} else {
		// Send initial update
		if err := sender.sendInitial("📝 Task received and processing started", 5); err != nil {
			return err
		}

		// 1. Save task to DB
		task = models.Task{
			ID:          existingTaskID,
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

		if err := database.Db.Create(&task).Error; err != nil {
			return sender.sendError("❌ Database error: " + err.Error())
		}
	}

	if err := sender.sendInitial("💾 Task saved to database", 10); err != nil {
		return err
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

	agentsClient, err := service.NewAgentClientWrapper()
	if err != nil {
		return sender.sendError("❌ Failed to initialize AI client: " + err.Error())
	}
	defer agentsClient.Close()

	msg := "🤖 AI client initialized (" + provider + "/" + model + ")"
	if err := sender.sendInitial(msg, 15); err != nil {
		return err
	}

	// 3. Check if we should use AI planning or predefined workflow
	var decision *BossDecisionResult
	var planErr error

	useAI := req.UseAiPlanning || len(req.PredefinedManagers) == 0
	if useAI {
		if err := sender.sendInitial("🤖 AI is planning architecture...", 15); err != nil {
			return err
		}

		decision, planErr = s.thinkAboutTask(ctx, agentsClient, provider, model, req)
		if planErr != nil {
			task.Status = "error"
			database.Db.Save(task)
			return sender.sendError("❌ AI planning failed: " + planErr.Error())
		}

		if err := sender.send(&bosspb.TaskUpdate{
			TaskId:    taskID,
			Message:   "✅ Architecture planned by AI",
			Progress:  30,
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"managers": strconv.Itoa(int(decision.ManagersCount)),
			},
		}); err != nil {
			return err
		}
	} else {
		if err := sender.sendInitial("📋 Using predefined workflow ("+strconv.Itoa(len(req.PredefinedManagers))+" managers)", 15); err != nil {
			return err
		}

		decision = s.buildDecisionFromPredefined(req)

		if err := sender.send(&bosspb.TaskUpdate{
			TaskId:    taskID,
			Message:   "✅ Predefined workflow loaded",
			Progress:  30,
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"managers": strconv.Itoa(int(decision.ManagersCount)),
			},
		}); err != nil {
			return err
		}
	}

	// 4. Save boss decision
	bossDecision := &models.BossDecision{
		ID:                   uuid.New(),
		TaskID:               parseUUID(taskID),
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
		return sender.sendError("❌ Failed to save decision: " + err.Error())
	}

	msg = "🏗️ Creating " + strconv.Itoa(int(decision.ManagersCount)) + " managers in parallel"
	if err := sender.sendInitial(msg, 40); err != nil {
		return err
	}

	// 4.5. Restore project if exists
	projectPath, err := s.restoreProject(taskID)
	if err != nil {
		log.Printf("No project to restore for task %s: %v", taskID, err)
		projectPath = "" // proceed without
	}

	// 5. Call managers in parallel with progress updates
	managerResults, zipData, err := s.assignManagersParallelWithProgress(ctx, taskID, decision, req, stream, projectPath)
	if err != nil {
		log.Printf("Error from managers: %v", err)
		task.Status = "error"
		database.Db.Save(task)
		return sender.sendError("❌ Manager failed: " + err.Error())
	}

	if err := sender.sendInitial("👷 All managers completed code generation", 70); err != nil {
		return err
	}

	// 6. Boss validation
	if err := sender.sendInitial("🔍 Boss validating solution...", 80); err != nil {
		return err
	}

	if len(managerResults) > 0 {
		_, _ = s.validateSolution(ctx, agentsClient, provider, model, req.Tokens, decision, managerResults, zipData)
	}

	if len(zipData) == 0 {
		task.Status = "error"
		database.Db.Save(task)
		return sender.sendError("❌ No solution generated")
	}

	if err := sender.sendInitial("📦 Packaging project", 90); err != nil {
		return err
	}

	// 7. Save ZIP solution
	task.Solution = zipData
	task.Status = "done"
	database.Db.Save(task)

	log.Printf("✅ Task %s completed! ZIP size: %d bytes", taskID, len(zipData))

	techStackBytes, _ := json.Marshal(decision.TechStack)
	return sender.sendSuccess("🎉 Project ready! "+task.Title+" created successfully", 100, map[string]string{
		"managers":  strconv.Itoa(int(decision.ManagersCount)),
		"techStack": string(techStackBytes),
	})
}

// parseUUID parses a string to UUID, returns nil on error
func parseUUID(id string) uuid.UUID {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.New()
	}
	return uid
}

// parseJSONToMap parses JSON string to map
func parseJSONToMap(jsonStr string) map[string]string {
	result := make(map[string]string)
	if jsonStr == "" {
		return result
	}
	_ = json.Unmarshal([]byte(jsonStr), &result)
	return result
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
