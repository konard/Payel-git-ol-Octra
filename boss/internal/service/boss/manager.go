package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager/managerpb"
	"boss/pkg/models"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// assignManagersParallelWithProgress calls AssignManager for each manager IN PARALLEL with progress updates
func (s *BossService) assignManagersParallelWithProgress(ctx context.Context, taskID string, decision *BossDecisionResult, req *bosspb.CreateTaskRequest, stream bosspb.BossService_CreateTaskStreamServer, projectPath string) ([]*managerpb.ManagerResult, error) {
	// Create a thread-safe progress sender
	sendCh := make(chan *bosspb.TaskUpdate, 64)
	doneCh := make(chan struct{})

	// Dedicated goroutine for all stream.Send() calls — prevents HTTP/2 race condition
	go func() {
		defer close(doneCh)
		for update := range sendCh {
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled, stopping stream sender")
				return
			default:
				if err := stream.Send(update); err != nil {
					log.Printf("Boss stream send error: %v", err)
					return
				}
			}
		}
	}()

	callback := func(role string, progress int, message string) {
		select {
		case sendCh <- &bosspb.TaskUpdate{
			TaskId:    taskID,
			Message:   message,
			Progress:  int32(progress),
			Status:    "processing",
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"current_role": role,
			},
		}:
		case <-ctx.Done():
		}
	}

	results, err := s.assignManagersParallel(ctx, taskID, decision, req, callback, projectPath)
	close(sendCh)
	<-doneCh // Wait for all pending sends to complete
	return results, err
}

// assignManagersParallel calls AssignManager for each manager IN PARALLEL
func (s *BossService) assignManagersParallel(ctx context.Context, taskID string, decision *BossDecisionResult, req *bosspb.CreateTaskRequest, progressCallback func(string, int, string), projectPath string) ([]*managerpb.ManagerResult, error) {
	techStack := "go"
	if len(decision.TechStack) > 0 {
		techStack = decision.TechStack[0]
	}
	metadata := map[string]string{
		"tokens":       marshalString(req.Tokens),
		"model":        req.Meta["model"],
		"provider":     req.Meta["provider"],
		"title":        req.Title,
		"grade_weight": "10",
		"tech_stack":   techStack,
	}
	for k, v := range req.Tokens {
		metadata[k] = v
	}

	var (
		mu         sync.Mutex
		allResults []*managerpb.ManagerResult
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

			// Get worker roles for this manager (from predefined workflow or empty)
			var workerRolesProto []*managerpb.WorkerRoleConfig
			if decision.ManagerWorkerRoles != nil {
				if workers, ok := decision.ManagerWorkerRoles[role.Role]; ok {
					for _, w := range workers {
						workerRolesProto = append(workerRolesProto, &managerpb.WorkerRoleConfig{
							Role:         w.Role,
							Description:  w.Description,
							CustomPrompt: w.CustomPrompt,
						})
					}
				}
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
				WorkerRoles:          workerRolesProto, // predefined worker roles
				ProjectPath:          projectPath,
				CustomPrompt:         role.CustomPrompt, // кастомный промт для менеджера
			}

			// Create timeout context for manager call (30 minutes per manager)
			managerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			defer cancel()

			log.Printf("Calling AssignManager #%d: %s", idx+1, role.Role)
			stream, err := s.managerClient.AssignManagerStream(managerCtx, mgrReq)
			if err != nil {
				log.Printf("AssignManagerStream %s error: %v", role.Role, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			// Process streaming updates from manager
			for {
				update, err := stream.Recv()
				if err != nil {
					if err.Error() == "EOF" {
						break
					}
					log.Printf("Manager stream %s error: %v", role.Role, err)
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}

				// Forward progress updates
				if update.Message != "" {
					progressCallback(role.Role, 40+idx*10+int(update.Progress*50/100), update.Message)
				}

				// If manager completed, get final result
				if update.Status == "success" {
					break
				}
			}

			// Get final result via non-streaming call (fallback)
			result, err := s.managerClient.AssignManager(managerCtx, mgrReq)
			if err != nil {
				log.Printf("AssignManager final call %s error: %v", role.Role, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			allResults = append(allResults, result)
			mu.Unlock()

			if progressCallback != nil {
				progressCallback(role.Role, 60+idx*3, "✅ Manager completed: "+role.Role)
			}

			log.Printf("AssignManager #%d %s completed", idx+1, role.Role)

			// Send worker-specific updates from manager results
			// (workers were already completed inside AssignManager call)
			if progressCallback != nil && result != nil {
				for wi, wr := range result.WorkerResults {
					filesCount := len(wr.Files)
					status := "✅"
					if !wr.Success {
						status = "❌"
					}
					progressCallback(wr.Role, 62+wi*2, fmt.Sprintf("%s %s completed: %d files", status, wr.Role, filesCount))
				}
			}
		}(i, role)
	}

	wg.Wait()

	if firstErr != nil && len(allResults) == 0 {
		return nil, firstErr
	}

	return allResults, nil
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
