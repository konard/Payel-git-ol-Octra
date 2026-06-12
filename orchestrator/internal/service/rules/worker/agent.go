package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orchestrator/internal/service/groupchat"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	"orchestrator/pkg/database"
	"orchestrator/pkg/models"

	"github.com/google/uuid"
)

// WorkerAgent implements groupchat.Agent — воркер, который живёт в групповом чате,
// видит всю историю беседы, и может генерировать/улучшать код по ходу раундов.
type WorkerAgent struct {
	id                 string
	role               string
	description        string
	service            *Service
	req                *rules.AssignWorkersRequest
	meta               workerMeta
	basePath           string
	accumulatedContext string
	selectedSlugs      []string // skill fragments, выбранные менеджером со склада
	progress           rules.ProgressFunc
}

func NewWorkerAgent(s *Service, req *rules.AssignWorkersRequest, wr *rules.WorkerRole, meta workerMeta, basePath, accCtx string, progress rules.ProgressFunc) *WorkerAgent {
	return &WorkerAgent{
		id:                 uuid.New().String(),
		role:               wr.Role,
		description:        wr.Description,
		service:            s,
		req:                req,
		meta:               meta,
		basePath:           basePath,
		accumulatedContext: accCtx,
		selectedSlugs:      wr.SelectedSkillSlugs,
		progress:           progress,
	}
}

func (a *WorkerAgent) ID() string  { return a.id }
func (a *WorkerAgent) Role() string { return a.role }

// Process вызывается оркестратором, когда этот агент выбран для выступления.
// Он получает полную историю беседы, генерирует код, и возвращает файлы.
func (a *WorkerAgent) Process(ctx context.Context, conv *groupchat.Conversation) ([]groupchat.Message, error) {
	log.Printf("[WorkerAgent %s] Processing round, conversation has %d messages", a.role, len(conv.Messages))

	// Собираем контекст из чата: что другие воркеры уже нагенерили
	chatContext := a.buildChatContext(conv)
	fullContext := chatContext + "\n" + a.accumulatedContext

	// Разрешаем skill fragments: используем то, что менеджер выбрал со склада
	skillContent := buildSkillContext(a.selectedSlugs, a.role, a.meta.techStack, a.req.TaskMd, a.description)

	// TASK.md (AI)
	taskMD, err := a.service.createTaskMD(ctx, a.meta.provider, a.meta.model, a.meta.tokens,
		a.role, a.description, a.req.TaskMd, fullContext, a.meta.taskType)
	if err != nil {
		log.Printf("[WorkerAgent %s] createTaskMD failed: %v", a.role, err)
		return nil, fmt.Errorf("agent think: %w", err)
	}
	log.Printf("[WorkerAgent %s] Starting code generation (taskType=%s, techStack=%s)...",
		a.role, a.meta.taskType, a.meta.techStack)

	// Генерация кода
	workerMode := os.Getenv("WORKER_MODE")
	var files map[string]string
	var commands []string

	if a.meta.taskType != "" && a.meta.taskType != "code" {
		// Документный путь
		topic := strings.TrimSpace(a.meta.title + "\n\n" + a.meta.description)
		if strings.TrimSpace(topic) == "" {
			topic = a.req.TaskMd
		}
		files, commands, err = a.service.generateDocument(ctx, a.meta.provider, a.meta.model, a.meta.tokens,
				a.meta.taskType, a.role, a.description, topic, fullContext, a.id, nil, skillContent)
	} else {
		useTools := workerMode == "tool" || (isToolMode(a.meta.techStack) && workerMode != "no-tool")
		if useTools {
			files, commands, err = a.service.generateViaTools(ctx, a.meta.provider, a.meta.model, a.meta.tokens,
				taskMD, a.role, a.description, a.req.ManagerRole, a.basePath, fullContext, a.meta.techStack, a.progress)
			if err != nil || len(files) == 0 {
				log.Printf("[WorkerAgent %s] Tool fallback to AI: %v", a.role, err)
				files = nil
				commands = nil
			}
		}
		if !useTools || len(files) == 0 {
			if workerMode == "" || workerMode == "tool" {
				workerMode = "multypass"
			}
			if workerMode == "multypass" {
				if a.progress != nil {
					a.progress(40, "Generating code via AI multi-pass...", nil)
				}
				files, commands, err = a.service.generateCodeMultiPass(ctx, a.meta.provider, a.meta.model, a.meta.tokens,
					taskMD, a.role, a.description, a.req.ManagerRole, a.basePath, fullContext, a.meta.techStack, skillContent, a.progress)
			} else {
				if a.progress != nil {
					a.progress(40, "Generating code via AI...", nil)
				}
				files, commands, err = a.service.generateCode(ctx, a.meta.provider, a.meta.model, a.meta.tokens,
					taskMD, a.role, a.description, a.req.ManagerRole, a.basePath, fullContext, a.meta.techStack, skillContent, a.progress)
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("agent generate: %w", err)
	}

	// Выполняем команды (npm init, go mod init и т.д.) через shell (с nix develop если доступен)
	for _, cmd := range commands {
		if cmd == "" {
			continue
		}
		// Если команда делает cd в несуществующую директорию — убираем cd-часть
		// (AI часто генерирует cd app после mvn archetype:generate, но проект может
		// оказаться в корне)
		resolvedCmd := resolveCD(a.basePath, cmd)
		if a.progress != nil {
			a.progress(45, "Running: "+cmd, nil)
		}
		c := util.NixDevelopCmd(a.basePath, resolvedCmd)
		if output, err := c.CombinedOutput(); err != nil {
			log.Printf("[WorkerAgent %s] cmd %q (resolved from %q) failed: %v\n%s", a.role, resolvedCmd, cmd, err, string(output))
		}
	}

	// Если есть что писать — пишем на диск (для последующего git add/commit) и шлём на фронтенд
	if len(files) > 0 {
		if err := os.MkdirAll(a.basePath, 0755); err != nil {
			log.Printf("[WorkerAgent %s] mkdir %s: %v", a.role, a.basePath, err)
		}
		i := 0
		for path, content := range files {
			fullPath, pathErr := util.ValidateFilePath(a.basePath, path)
			if pathErr != nil {
				log.Printf("[WorkerAgent %s] invalid path %s: %v", a.role, path, pathErr)
				continue
			}
			if dirErr := os.MkdirAll(filepath.Dir(fullPath), 0755); dirErr != nil {
				log.Printf("[WorkerAgent %s] mkdir for %s: %v", a.role, path, dirErr)
				continue
			}
			if writeErr := os.WriteFile(fullPath, []byte(content), 0644); writeErr != nil {
				log.Printf("[WorkerAgent %s] write %s: %v", a.role, path, writeErr)
			} else if a.progress != nil {
				pct := 50 + int32(i)*30/int32(len(files))
				if pct > 80 {
					pct = 80
				}
				a.progress(pct, "Writing file: "+path, map[string]string{
					"file": path,
					"type": "write",
					"size": fmt.Sprintf("%d", len(content)),
				})
				i++
			}
		}
	}

	summary := fmt.Sprintf("Generated %d files for role %s", len(files), a.role)
	msg := groupchat.Message{
		AgentID:   a.id,
		Role:      a.role,
		Content:   summary,
		Files:     files,
		Type:      groupchat.MsgResponse,
		Timestamp: time.Now().UTC(),
	}

	log.Printf("[WorkerAgent %s] done: %s", a.role, summary)
	return []groupchat.Message{msg}, nil
}

// buildChatContext собирает из истории чата контекст для AI:
// что другие воркеры уже сгенерировали, какие файлы создали.
func (a *WorkerAgent) buildChatContext(conv *groupchat.Conversation) string {
	var b strings.Builder
	b.WriteString("=== GROUP CHAT HISTORY ===\n")

	for _, msg := range conv.Messages {
		if msg.AgentID == a.id {
			continue // свои сообщения не добавляем в контекст
		}
		b.WriteString(fmt.Sprintf("[%s - %s] %s\n", msg.Role, msg.Type, msg.Content))
		if len(msg.Files) > 0 {
			b.WriteString(fmt.Sprintf("  Files generated (%d):\n", len(msg.Files)))
			for path := range msg.Files {
				b.WriteString(fmt.Sprintf("    - %s\n", path))
			}
			// Показываем содержимое файлов других воркеров для контекста
			for path, content := range msg.Files {
				b.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", path, content))
			}
		}
	}

	b.WriteString("=== END HISTORY ===")
	return b.String()
}

// persistWorkerDB создаёт/обновляет запись воркера в БД.
// Вызывается после того как все агенты завершили работу.
func persistWorkerDB(req *rules.AssignWorkersRequest, agentID, role, taskMD string, files map[string]string, success bool) {
	workerID, err := uuid.Parse(agentID)
	if err != nil {
		workerID = uuid.New()
	}
	taskUUID, _ := uuid.Parse(req.TaskId)
	managerUUID, _ := uuid.Parse(req.ManagerId)

	workerModel := &models.Worker{
		ID:        workerID,
		TaskID:    taskUUID,
		ManagerID: managerUUID,
		Role:      role,
		Status:    "done",
		TaskMD:    taskMD,
		Files:     util.MarshalJSON(files),
		Success:   success,
	}
	if success {
		workerModel.SolutionMD = fmt.Sprintf("Generated %d files for role %s", len(files), role)
	}
	if err := database.Db.Create(workerModel).Error; err != nil {
		// Update if exists
		database.Db.Model(&models.Worker{}).
			Where("id = ?", workerID).
			Updates(map[string]interface{}{
				"status":     "done",
				"success":    success,
				"files":      util.MarshalJSON(files),
				"solution_md": workerModel.SolutionMD,
			})
	}
}

// resolveCD проверяет, существует ли директория, в которую делает cd команда.
// Если нет — убирает cd-часть, чтобы остаток команды выполнился из корня проекта.
// Пример: "cd app && mvn package" → "mvn package" (если app/ не существует)
func resolveCD(basePath, cmd string) string {
	if !strings.HasPrefix(cmd, "cd ") {
		return cmd
	}
	parts := strings.SplitN(cmd, "&&", 2)
	if len(parts) < 2 {
		return cmd
	}
	cdPart := strings.TrimSpace(parts[0]) // "cd app"
	rest := strings.TrimSpace(parts[1])   // "mvn package"

	// Извлекаем имя директории из cd
	dir := strings.TrimSpace(strings.TrimPrefix(cdPart, "cd "))
	dir = strings.Trim(dir, `"'`) // убираем возможные кавычки

	fullDir := filepath.Join(basePath, dir)
	if _, err := os.Stat(fullDir); os.IsNotExist(err) {
		log.Printf("[resolveCD] directory %q does not exist, stripping cd (cmd: %q)", fullDir, cmd)
		return rest
	}
	return cmd
}
