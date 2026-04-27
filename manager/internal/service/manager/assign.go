package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"manager/internal/fetcher/grpc/managerpb"
	"time"
)

// AssignManager — назначить ОДНОГО менеджера (вызывается Boss для каждого менеджера отдельно)
func (s *ManagerService) AssignManager(ctx context.Context, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
	log.Printf("=== Manager assigned: %s (%s), priority %d ===", req.ManagerId, req.Role, req.Priority)

	return s.processManager(ctx, req)
}

// AssignManagerStream — streaming version that sends progress updates
func (s *ManagerService) AssignManagerStream(req *managerpb.AssignManagerRequest, stream managerpb.ManagerService_AssignManagerStreamServer) error {
	log.Printf("=== Manager stream assigned: %s (%s), priority %d ===", req.ManagerId, req.Role, req.Priority)

	ctx := stream.Context()

	// Send initial progress
	stream.Send(&managerpb.TaskUpdate{
		TaskId:    req.TaskId,
		Message:   fmt.Sprintf("Manager %s assigned and starting work", req.Role),
		Progress:  0,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
		Data: map[string]string{
			"current_role": req.Role,
		},
	})

	// Process manager with progress callback
	_, err := s.processManagerWithProgress(ctx, req, func(progress int, message string) {
		stream.Send(&managerpb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   message,
			Progress:  int32(progress),
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"current_role": req.Role,
			},
		})
	})

	if err != nil {
		stream.Send(&managerpb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   fmt.Sprintf("Manager %s failed: %v", req.Role, err),
			Progress:  0,
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
		return err
	}

	// Send final completion update
	stream.Send(&managerpb.TaskUpdate{
		TaskId:    req.TaskId,
		Message:   fmt.Sprintf("Manager %s task finalized", req.Role),
		Progress:  100,
		Status:    "completed",
		Timestamp: time.Now().Unix(),
		Data: map[string]string{
			"current_role": req.Role,
			"manager_id":   req.ManagerId,
		},
	})

	return nil
}

// AssignManagersAndWait — legacy, назначает всех менеджеров последовательно
func (s *ManagerService) AssignManagersAndWait(ctx context.Context, req *managerpb.AssignManagersRequest) (*managerpb.AssignManagersResponse, error) {
	log.Printf("=== Legacy AssignManagers: %d roles ===", len(req.Roles))

	metadata := req.Metadata
	tokens := make(map[string]string)
	if tokensJSON, ok := metadata["tokens"]; ok {
		json.Unmarshal([]byte(tokensJSON), &tokens)
	}
	provider := metadata["provider"]
	if provider == "" {
		provider = "openai"
	}
	model := metadata["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Collect ZIP archives from all managers
	var allZipArchives [][]byte
	managerResults := make([]*managerpb.ManagerResult, 0)

	for i, role := range req.Roles {
		managerID := uuid.New()

		// Build context from previous manager results
		var prevResults []*managerpb.WorkerResult
		for _, mr := range managerResults {
			prevResults = append(prevResults, mr.WorkerResults...)
		}

		mgrReq := &managerpb.AssignManagerRequest{
			TaskId:               req.TaskId,
			ManagerId:            managerID.String(),
			Role:                 role,
			Description:          fmt.Sprintf("%s team manager", role),
			TechnicalDescription: req.TechnicalDescription,
			Priority:             int32(i + 1),
			Metadata:             metadata,
			OtherWorkersResults:  prevResults,
		}

		result, err := s.processManager(ctx, mgrReq)
		if err != nil {
			log.Printf("Error processing manager %s: %v", role, err)
			continue
		}

		managerResults = append(managerResults, result)
		if len(result.Solution) > 0 {
			allZipArchives = append(allZipArchives, result.Solution)
		}
	}

	finalZip, _ := mergeZipArchives(allZipArchives)

	return &managerpb.AssignManagersResponse{
		TaskId:         req.TaskId,
		Status:         "success",
		Message:        "Project generated",
		Solution:       finalZip,
		ManagerResults: managerResults,
	}, nil
}
