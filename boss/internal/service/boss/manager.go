package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager/managerpb"
	"boss/pkg/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

// assignManagersParallelWithProgress calls AssignManager for each manager IN PARALLEL with progress updates
func (s *BossService) assignManagersParallelWithProgress(ctx context.Context, taskID string, decision *BossDecisionResult, req *bosspb.CreateTaskRequest, stream bosspb.BossService_CreateTaskStreamServer) ([]*managerpb.ManagerResult, []byte, error) {
	return s.assignManagersParallel(ctx, taskID, decision, req, func(role string, progress int, message string) {
		_ = stream.Send(&bosspb.TaskUpdate{
			TaskId:    taskID,
			Message:   message,
			Progress:  int32(progress),
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"current_role": role,
			},
		})
	})
}

// assignManagersParallel calls AssignManager for each manager IN PARALLEL
func (s *BossService) assignManagersParallel(ctx context.Context, taskID string, decision *BossDecisionResult, req *bosspb.CreateTaskRequest, progressCallback func(string, int, string)) ([]*managerpb.ManagerResult, []byte, error) {
	metadata := map[string]string{
		"tokens":   marshalString(req.Tokens),
		"model":    req.Meta["model"],
		"provider": req.Meta["provider"],
	}
	for k, v := range req.Tokens {
		metadata[k] = v
	}

	var (
		mu         sync.Mutex
		allResults []*managerpb.ManagerResult
		allZipData [][]byte
		firstErr   error
	)

	var wg sync.WaitGroup
	for i, role := range decision.ManagerRoles {
		wg.Add(1)
		go func(idx int, role models.ManagerRole) {
			defer wg.Done()

			if progressCallback != nil {
				progressCallback(role.Role, 40+idx*10, "👷 Starting manager: "+role.Role)
			}

			managerID := uuid.New().String()

			// Wait for higher-priority managers to start producing context
			// (small delay to get some initial results)
			if idx > 0 {
				select {
				case <-time.After(time.Duration(idx) * 5 * time.Second):
				case <-ctx.Done():
					return
				}
			}

			// Collect context from already-completed managers
			mu.Lock()
			var contextResults []*managerpb.WorkerResult
			for _, mr := range allResults {
				contextResults = append(contextResults, mr.WorkerResults...)
			}
			mu.Unlock()

			if progressCallback != nil {
				progressCallback(role.Role, 45+idx*5, "👷 Manager working: "+role.Role)
			}

			mgrReq := &managerpb.AssignManagerRequest{
				TaskId:               taskID,
				ManagerId:            managerID,
				Role:                 role.Role,
				Description:          role.Description,
				TechnicalDescription: decision.TechnicalDescription,
				Priority:             role.Priority,
				Metadata:             metadata,
				OtherWorkersResults:  contextResults,
			}

			log.Printf("Calling AssignManager #%d: %s", idx+1, role.Role)
			result, err := s.managerClient.AssignManager(ctx, mgrReq)
			if err != nil {
				log.Printf("AssignManager %s error: %v", role.Role, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			allResults = append(allResults, result)
			if len(result.Solution) > 0 {
				allZipData = append(allZipData, result.Solution)
			}
			mu.Unlock()

			if progressCallback != nil {
				progressCallback(role.Role, 60+idx*3, "✅ Manager completed: "+role.Role)
			}

			log.Printf("AssignManager #%d %s completed", idx+1, role.Role)
		}(i, role)
	}

	wg.Wait()

	if firstErr != nil && len(allResults) == 0 {
		return nil, nil, firstErr
	}

	if progressCallback != nil {
		progressCallback("merger", 68, "🔧 Merger combining all projects...")
	}

	// Use merger service to combine all manager outputs
	projectsMap := make(map[string][]byte)
	for i, result := range allResults {
		roleName := result.Role
		if roleName == "" {
			roleName = fmt.Sprintf("project_%d", i)
		}
		if len(result.Solution) > 0 {
			projectsMap[roleName] = result.Solution
		}
	}

	var finalZip []byte
	if s.mergerClient != nil && len(projectsMap) > 0 {
		var err error
		finalZip, err = s.mergerClient.MergeProjects(ctx, projectsMap)
		if err != nil {
			log.Printf("Merger service error: %v, falling back to simple merge", err)
			finalZip, _ = mergeZipArchives(allZipData)
		}
	} else {
		finalZip, _ = mergeZipArchives(allZipData)
	}

	if progressCallback != nil {
		progressCallback("merger", 70, "✅ All projects merged successfully")
	}

	return allResults, finalZip, nil
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
