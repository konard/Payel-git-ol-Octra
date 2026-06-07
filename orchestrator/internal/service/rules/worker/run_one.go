package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"orchestrator/internal/service/git"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	"orchestrator/pkg/database"
	"orchestrator/pkg/models"

	"github.com/google/uuid"
)

// runOneWorker — обработка одного воркера: TASK.md → код → файлы → коммит.
// progress/basePct используются, чтобы ресёрч-воркер мог транслировать шаги
// веб-поиска в чат (блок «Searching the web»).
// Если skipGit=true — git-операции не выполняются, caller сам делает git add/commit.
func (s *Service) runOneWorker(
	ctx context.Context, req *rules.AssignWorkersRequest, wr *rules.WorkerRole,
	meta workerMeta, basePath, accumulatedContext string,
	progress rules.ProgressFunc, basePct int32, skipGit bool,
) (*rules.WorkerResult, error) {
	role := wr.Role
	description := wr.Description

	workerID := uuid.New()
	taskUUID, _ := uuid.Parse(req.TaskId)
	managerUUID, _ := uuid.Parse(req.ManagerId)
	workerModel := &models.Worker{
		ID:        workerID,
		TaskID:    taskUUID,
		ManagerID: managerUUID,
		Role:      role,
		Status:    "thinking",
		TaskMD:    req.TaskMd,
	}
	if err := database.Db.Create(workerModel).Error; err != nil {
		log.Printf("Error creating worker: %v", err)
		return nil, err
	}

	if req.ProjectPath != "" {
		branchName := fmt.Sprintf("worker-%s", role)
		if err := git.CreateBranch(req.ProjectPath, branchName); err != nil {
			log.Printf("Failed to create worker branch %s: %v", branchName, err)
		}
	}

	taskMD, err := s.createTaskMD(ctx, meta.provider, meta.model, meta.tokens, role, description, req.TaskMd, accumulatedContext, meta.taskType)
	if err != nil {
		workerModel.Status = "error"
		database.Db.Save(workerModel)
		return nil, fmt.Errorf("worker think: %w", err)
	}
	workerModel.Status = "coding"
	database.Db.Save(workerModel)

	var files map[string]string
	var commands []string

	// Разрешаем skill fragments: используем то, что менеджер выбрал со склада
	skillContent := buildSkillContext(wr.SelectedSkillSlugs, role, meta.techStack, req.TaskMd, description)

	if meta.taskType != "" && meta.taskType != "code" {
		// Не-кодовые задачи (research/document/presentation) идут по документному пути:
		// результат — Markdown (и .pptx для презентаций) в папке solution/.
		topic := strings.TrimSpace(meta.title + "\n\n" + meta.description)
		if strings.TrimSpace(topic) == "" {
			topic = req.TaskMd
		}
		emit := s.searchEmitterFor(progress, basePct, req, role)
		files, commands, err = s.generateDocument(ctx, meta.provider, meta.model, meta.tokens, meta.taskType, role, description, topic, accumulatedContext, workerID.String(), emit, skillContent)
	} else {
		workerMode := os.Getenv("WORKER_MODE")
		useTools := workerMode == "tool" || (isToolMode(meta.techStack) && workerMode != "no-tool")
		if useTools {
			files, commands, err = s.generateViaTools(ctx, meta.provider, meta.model, meta.tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext, meta.techStack)
			if err != nil || len(files) == 0 {
				log.Printf("[ToolExecutor] Fallback to AI generation (tool mode failed: %v, files=%d)", err, len(files))
				files = nil
				commands = nil
				useTools = false
			}
		}
		if !useTools {
			if workerMode == "" || workerMode == "tool" {
				workerMode = "multypass"
			}
			if workerMode == "multypass" {
				files, commands, err = s.generateCodeMultiPass(ctx, meta.provider, meta.model, meta.tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext, meta.techStack, skillContent)
			} else {
				files, commands, err = s.generateCode(ctx, meta.provider, meta.model, meta.tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext, meta.techStack, skillContent)
			}
		}
	}
	if err != nil {
		workerModel.Status = "error"
		database.Db.Save(workerModel)
		return nil, fmt.Errorf("worker generate: %w", err)
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Printf("Warning: failed to ensure project path %s: %v", basePath, err)
	}
	for _, cmd := range commands {
		if cmd == "" {
			continue
		}
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}
		c := exec.Command(parts[0], parts[1:]...)
		c.Dir = basePath
		c.Run()
	}
	for path, content := range files {
		fullPath, err := util.ValidateFilePath(basePath, path)
		if err != nil {
			log.Printf("Warning: path validation failed for %s: %v", path, err)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Printf("Warning: failed to create dir for %s: %v", path, err)
			continue
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			log.Printf("Warning: failed to write file %s: %v", path, err)
		}
	}

	if skipGit {
		log.Printf("Worker (%s) skipping git operations (async mode)", role)
	} else {
		isRepo, err := git.IsGitRepo(basePath)
		if err != nil {
			workerModel.Status = "error"
			database.Db.Save(workerModel)
			return nil, fmt.Errorf("git check at %s: %w", basePath, err)
		}
		if !isRepo {
			workerModel.Status = "error"
			database.Db.Save(workerModel)
			return nil, fmt.Errorf("git repository not found at: %s", basePath)
		}

		commitMessage := fmt.Sprintf("Worker %s: %s", role, taskMD)
		if err := git.Add(basePath); err != nil {
			workerModel.Status = "error"
			database.Db.Save(workerModel)
			return nil, fmt.Errorf("git add: %w", err)
		}
		if err := git.Commit(basePath, commitMessage); err != nil {
			workerModel.Status = "error"
			database.Db.Save(workerModel)
			return nil, fmt.Errorf("git commit: %w", err)
		}
	}

	result := &rules.WorkerResult{
		WorkerId:   workerID.String(),
		Role:       role,
		TaskMd:     taskMD,
		SolutionMd: fmt.Sprintf("Generated %d files for role %s", len(files), role),
		Files:      files,
		Success:    true,
		Commands:   commands,
	}

	workerModel.SolutionMD = result.SolutionMd
	workerModel.Files = util.MarshalJSON(files)
	workerModel.Status = "done"
	workerModel.Success = true
	database.Db.Save(workerModel)

	log.Printf("Worker (%s) completed: %d files", role, len(files))
	return result, nil
}

