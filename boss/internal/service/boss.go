package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager"
	"boss/pkg/database"
	"boss/pkg/models"

	"github.com/google/uuid"
)

// BossService — boss service
type BossService struct {
	bosspb.UnimplementedBossServiceServer
	managerClient *manager.Client
}

func NewBossService() *BossService {
	mgrClient, err := manager.NewClient("manager:50052")
	if err != nil {
		log.Printf("Warning: failed to connect to manager service: %v", err)
	}

	return &BossService{
		managerClient: mgrClient,
	}
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

	// 2. Get model and modelUrl
	model := req.Meta["model"]
	if model == "" {
		model = "gpt-4-mini"
	}

	// Determine provider from model or use default
	provider := "openai"
	if len(req.Meta) > 0 {
		if p, ok := req.Meta["provider"]; ok {
			provider = p
		}
	}

	// Create agents client
	agentsClient, err := NewAgentClientWrapper()
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

	// 5. Call Manager service (wait for full cycle)
	log.Printf("Calling Manager service for full cycle...")

	roles := make([]string, len(decision.ManagerRoles))
	for i, r := range decision.ManagerRoles {
		roles[i] = r.Role
	}

	// Call Manager and get ZIP archive
	zipData, err := s.managerClient.AssignManagersAndWait(
		ctx,
		task.ID.String(),
		decision.TechnicalDescription,
		roles,
		req.Tokens,
		model,
		provider,
	)
	if err != nil {
		log.Printf("Error calling Manager: %v", err)
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

	// 6. Save ZIP solution
	task.Solution = zipData
	task.Status = "done"
	database.Db.Save(task)

	log.Printf("Task completed! ZIP size: %d bytes", len(zipData))

	// 7. Return result with ZIP
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

// BossDecisionResult — result of boss thinking
type BossDecisionResult struct {
	ManagersCount        int32                `json:"managers_count"`
	ManagerRoles         []models.ManagerRole `json:"manager_roles"`
	TechnicalDescription string               `json:"technical_description"`
	TechStack            []string             `json:"tech_stack"`
	ArchitectureNotes    string               `json:"architecture_notes"`
}

func (r *BossDecisionResult) ManagerRolesProto() []*bosspb.ManagerRole {
	result := make([]*bosspb.ManagerRole, len(r.ManagerRoles))
	for i, role := range r.ManagerRoles {
		result[i] = &bosspb.ManagerRole{
			Role:        role.Role,
			Description: role.Description,
			Priority:    role.Priority,
		}
	}
	return result
}

func (s *BossService) thinkAboutTask(ctx context.Context, agentsClient *AgentClientWrapper, provider, model string, req *bosspb.CreateTaskRequest) (*BossDecisionResult, error) {
	log.Printf("Boss thinking about task: %s", req.Title)

	prompt := `You are CTO. Task:

Title: ` + req.Title + `
Description: ` + req.Description + `

Reply ONLY with JSON:
{
  "managers_count": 3,
  "manager_roles": [
    {"role": "frontend", "description": "Frontend part", "priority": 1},
    {"role": "backend", "description": "Backend API", "priority": 2},
    {"role": "testing", "description": "Testing", "priority": 3}
  ],
  "technical_description": "Technical description...",
  "tech_stack": ["Go", "React", "PostgreSQL", "Docker"],
  "architecture_notes": "Architecture notes..."
}`

	log.Printf("Sending request to AI...")
	resp, err := agentsClient.GenerateFromTask(ctx, provider, model, prompt, req.Tokens)
	if err != nil {
		log.Printf("AI error: %v", err)
		return nil, err
	}

	var decision BossDecisionResult
	if err := json.Unmarshal([]byte(resp), &decision); err != nil {
		log.Printf("JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("Boss made decision: %+v", decision)
	return &decision, nil
}

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

	// Send progress
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "💾 Task saved to database",
		Progress:  10,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 2. Get model and modelUrl
	model := req.Meta["model"]
	if model == "" {
		model = "gpt-4-mini"
	}

	// Determine provider
	provider := "openai"
	if len(req.Meta) > 0 {
		if p, ok := req.Meta["provider"]; ok {
			provider = p
		}
	}

	// Create agents client
	agentsClient, err := NewAgentClientWrapper()
	if err != nil {
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ Failed to initialize AI client: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}
	defer agentsClient.Close()

	// Send progress
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "🤖 AI client initialized (" + provider + "/" + model + ")",
		Progress:  15,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 3. Boss thinks about task
	decision, err := s.thinkAboutTask(ctx, agentsClient, provider, model, req)
	if err != nil {
		task.Status = "error"
		database.Db.Save(task)
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ AI planning failed: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	// Send progress
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

	// Send progress
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "🏗️ Creating " + strconv.Itoa(int(decision.ManagersCount)) + " managers",
		Progress:  40,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 5. Call Manager service (wait for full cycle)
	log.Printf("Calling Manager service for full cycle...")

	roles := make([]string, len(decision.ManagerRoles))
	for i, r := range decision.ManagerRoles {
		roles[i] = r.Role
	}

	// Call Manager and get ZIP archive
	zipData, err := s.managerClient.AssignManagersAndWait(
		ctx,
		task.ID.String(),
		decision.TechnicalDescription,
		roles,
		req.Tokens,
		model,
		provider,
	)
	if err != nil {
		log.Printf("Error calling Manager: %v", err)
		task.Status = "error"
		database.Db.Save(task)
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID.String(),
			Message:   "❌ Manager failed: " + err.Error(),
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}

	// Send progress
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "👷 Managers completed code generation",
		Progress:  70,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

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

	// Send progress
	_ = stream.Send(&bosspb.TaskUpdate{
		TaskId:    taskID.String(),
		Message:   "📦 Packaging project (" + string(rune(len(zipData)/1024)) + "KB)",
		Progress:  90,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})

	// 6. Save ZIP solution
	task.Solution = zipData
	task.Status = "done"
	database.Db.Save(task)

	log.Printf("✅ Task completed! ZIP size: %d bytes", len(zipData))

	// Send final success update
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

// GetTaskStatus returns task status
func (s *BossService) GetTaskStatus(ctx context.Context, req *bosspb.TaskStatusRequest) (*bosspb.TaskStatusResponse, error) {
	var task models.Task
	if err := database.Db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		return nil, err
	}

	return &bosspb.TaskStatusResponse{
		TaskId:   task.ID.String(),
		Status:   task.Status,
		Progress: "50%",
	}, nil
}

// CreateZipArchive creates ZIP from files
func CreateZipArchive(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for path, content := range files {
		f, err := w.Create(path)
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
