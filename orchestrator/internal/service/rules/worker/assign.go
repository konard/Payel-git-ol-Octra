package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"orchestrator/internal/service/git"
	"orchestrator/internal/service/groupchat"
	"orchestrator/internal/service/rules"

	"github.com/google/uuid"
)

// AssignWorkersAndWait — запускает всех воркеров через Group Chat паттерн.
// Оркестратор регистрирует каждого воркера как агента, запускает раунды,
// воркеры видят полную историю беседы (проверяют чат).
// После завершения — единый git add/commit.
func (s *Service) AssignWorkersAndWait(ctx context.Context, req *rules.AssignWorkersRequest, progress rules.ProgressFunc) (*rules.AssignWorkersResponse, error) {
	log.Printf("Received task from manager %s (%s): %s", req.ManagerId, req.ManagerRole, req.TaskId)

	meta := parseWorkerMetadata(req.Metadata)
	contextSummary := buildContextSummary(req.OtherWorkersResults)
	basePath := resolveBasePath(req)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory %s: %w", basePath, err)
	}
	writeContextFile(basePath, req.TaskId, req.ManagerId, req.ManagerRole, req.WorkerRoles, contextSummary)

	if progress != nil {
		roleNames := make([]string, 0, len(req.WorkerRoles))
		for _, wr := range req.WorkerRoles {
			if wr != nil {
				roleNames = append(roleNames, wr.Role)
			}
		}
		progress(5, "Starting Group Chat with all workers", map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
			"workers":      fmt.Sprintf("%d", len(req.WorkerRoles)),
			"worker_roles": strings.Join(roleNames, ","),
		})
	}

	// Build project-level context once — shared by all workers
	var projectCtx string
	if taskUUID, err := uuid.Parse(req.TaskId); err == nil && s.contextClient != nil {
		projectCtx = s.contextClient.GetForPrompt(ctx, taskUUID, "workers-"+req.ManagerRole, req.ManagerId)
	}

	acc := buildAccumulatedContext(nil, contextSummary) + projectCtx

	// Fast path: single worker — no group chat overhead
	if len(req.WorkerRoles) == 1 {
		wr := req.WorkerRoles[0]
		if progress != nil {
			progress(10, "Running single worker: "+wr.Role, map[string]string{
				"manager_id":   req.ManagerId,
				"manager_role": req.ManagerRole,
				"worker_role":  wr.Role,
			})
		}

		agent := NewWorkerAgent(s, req, wr, meta, basePath, acc, progress)
		messages, err := agent.Process(ctx, &groupchat.Conversation{})
		if err != nil {
			return nil, fmt.Errorf("worker %s failed: %w", wr.Role, err)
		}

		allFiles := make(map[string]string)
		for _, msg := range messages {
			for k, v := range msg.Files {
				allFiles[k] = v
			}
		}

		if progress != nil {
			progress(80, "Worker completed, writing files and committing", map[string]string{
				"files": fmt.Sprintf("%d", len(allFiles)),
			})
		}

		if len(allFiles) > 0 {
			// Create worker branch so manager can merge it
			branchName := fmt.Sprintf("worker-%s", wr.Role)
			if err := git.CreateBranch(basePath, branchName); err != nil {
				log.Printf("Warning: failed to create worker branch %s: %v", branchName, err)
			}

			if err := git.Add(basePath); err != nil {
				log.Printf("Warning: git add failed: %v", err)
			}
			if err := git.Commit(basePath, fmt.Sprintf("Worker: %s", wr.Role)); err != nil {
				log.Printf("Warning: git commit failed: %v", err)
			}
		}

		persistWorkerDB(req, agent.id, wr.Role, "", allFiles, true)

		workerResults := []*rules.WorkerResult{{
			WorkerId:   agent.id,
			Role:       wr.Role,
			Files:      allFiles,
			Success:    true,
			SolutionMd: fmt.Sprintf("Generated %d files for role %s", len(allFiles), wr.Role),
		}}

		if progress != nil {
			progress(100, "Single worker completed successfully", map[string]string{
				"files_count": strconv.Itoa(len(allFiles)),
			})
		}

		log.Printf("Single worker %s completed: %d files", wr.Role, len(allFiles))
		return &rules.AssignWorkersResponse{
			TaskId:        req.TaskId,
			Status:        "success",
			WorkerResults: workerResults,
		}, nil
	}

	// Создаём Group Chat оркестратор (2+ workers)
	orch := groupchat.NewOrchestrator(2)
	orch.SetSelector(groupchat.NewAllAtOnceSelector())

	// Регистрируем агентов-воркеров
	agents := make([]*WorkerAgent, len(req.WorkerRoles))
	for i, wr := range req.WorkerRoles {
		agent := NewWorkerAgent(s, req, wr, meta, basePath, acc, progress)
		agents[i] = agent
		orch.RegisterAgent(agent)
	}

	// Условие завершения: после 1-го раунда (все сгенерировали) останавливаемся
	orch.SetTerminationCondition(func(conv *groupchat.Conversation, round int) bool {
		return round > 1
	})

	if progress != nil {
		progress(10, "Running Group Chat conversation", map[string]string{
			"agents": fmt.Sprintf("%d", len(agents)),
			"rounds": "2",
		})
	}

	// Запускаем Group Chat
	if err := orch.Run(ctx, req.TaskMd); err != nil {
		return nil, fmt.Errorf("group chat failed: %w", err)
	}

	// Собираем результаты из чата
	conv := orch.SnapshotConversation()
	allFiles := conv.AllFiles()

	if progress != nil {
		progress(80, "All agents completed, writing files and committing", map[string]string{
			"files": fmt.Sprintf("%d", len(allFiles)),
		})
	}

	// Single git add/commit for all workers after all complete
	if len(allFiles) > 0 {
		if err := git.Add(basePath); err != nil {
			log.Printf("Warning: git add failed: %v", err)
		}
		roles := make([]string, 0, len(agents))
		for _, a := range agents {
			roles = append(roles, a.role)
		}
		msg := fmt.Sprintf("Group Chat Workers: %s", strings.Join(roles, ", "))
		if err := git.Commit(basePath, msg); err != nil {
			log.Printf("Warning: git commit failed: %v", err)
		}
	}

	// Персистим в БД
	for _, a := range agents {
		persistWorkerDB(req, a.id, a.role, "", allFiles, true)
	}

	// Строим ответ
	workerResults := make([]*rules.WorkerResult, 0, len(agents))
	for _, a := range agents {
		wr := &rules.WorkerResult{
			WorkerId: a.id,
			Role:     a.role,
			TaskMd:   "",
			Files:    allFiles,
			Success:  true,
		}
		// Фильтруем только файлы этого воркера
		wr.Files = make(map[string]string)
		for _, msg := range conv.Messages {
			if msg.AgentID == a.id && msg.Files != nil {
				for k, v := range msg.Files {
					wr.Files[k] = v
				}
			}
		}
		wr.SolutionMd = fmt.Sprintf("Generated %d files for role %s", len(wr.Files), a.role)
		workerResults = append(workerResults, wr)
	}

	if progress != nil {
		fileList := make([]string, 0, len(allFiles))
		for k := range allFiles {
			fileList = append(fileList, k)
		}
		progress(100, fmt.Sprintf("Group Chat completed: %d agents, %d files", len(agents), len(allFiles)), map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
			"files":        strings.Join(fileList, ","),
			"files_count":  strconv.Itoa(len(allFiles)),
		})
	}

	return &rules.AssignWorkersResponse{
		TaskId:        req.TaskId,
		Status:        "success",
		WorkerResults: workerResults,
	}, nil
}

// buildAccumulatedContext — собирает контекст из уже завершённых воркеров команды
func buildAccumulatedContext(workerResults []*rules.WorkerResult, externalCtx string) string {
	acc := ""
	for _, prevResult := range workerResults {
		acc += fmt.Sprintf("\n=== Previous worker %s (%s) ===\n", prevResult.WorkerId, prevResult.Role)
		if prevResult.Success {
			for path, content := range prevResult.Files {
				acc += fmt.Sprintf("\n--- File: %s ---\n%s\n", path, content)
			}
		}
		if prevResult.Feedback != "" {
			acc += fmt.Sprintf("\nFeedback: %s\n", prevResult.Feedback)
		}
	}
	if externalCtx != "" {
		acc += "\n=== Context from other worker groups ===\n" + externalCtx
	}
	return acc
}
