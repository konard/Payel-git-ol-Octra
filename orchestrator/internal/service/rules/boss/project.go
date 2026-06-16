package boss

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"orchestrator/internal/service/git"
	gh "orchestrator/internal/service/github"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	"orchestrator/pkg/models"
)

// setupProject — создаёт workspace-директорию проекта и инициализирует git
func (s *Service) setupProject(ctx context.Context, taskID, title string, issueTarget *gh.IssueTarget) (string, error) {
	projectsDir := os.Getenv("PROJECTS_DIR")
	if projectsDir == "" {
		projectsDir = "/workspace/projects"
	}

	if info, err := os.Stat(projectsDir); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("projects dir exists but is a file: %s", projectsDir)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check projects dir: %w", err)
	} else if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create projects dir: %w", err)
	}

	projectPath := filepath.Join(projectsDir, taskID)
	if title != "" {
		projectPath = filepath.Join(projectPath, util.SanitizeProjectName(title))
	}

	if issueTarget != nil && s.githubClient != nil {
		if err := ensureCloneTarget(projectPath); err != nil {
			return "", err
		}

		// Проверяем, есть ли предыдущий task для этого issue
		existingTask := s.findExistingTaskByIssue(ctx, issueTarget.IssueURL)
		if existingTask != nil {
			log.Printf("Found existing task %s for issue %s — reusing fork", existingTask.ID.String(), issueTarget.IssueURL)

			// Извлекаем fork-инфу из Meta предыдущей задачи
			existingMeta := s.parseTaskMeta(existingTask)
			if forkOwner, ok := existingMeta["github_fork_owner"]; ok && forkOwner != "" {
				issueTarget.Forked = true
				issueTarget.ForkOwner = forkOwner
				issueTarget.ForkCloneURL = existingMeta["github_fork_clone_url"]
				log.Printf("Reusing existing fork: %s/%s", forkOwner, issueTarget.Repo)
			}

			targetOwner := issueTarget.Owner
			targetRepo := issueTarget.Repo
			if issueTarget.ForkOwner != "" {
				targetOwner = issueTarget.ForkOwner
			}

			// Клонируем форк (или оригинал, если прав хватало)
			repository, cloneErr := s.githubClient.CloneRepository(ctx, targetOwner, targetRepo, projectPath)
			if cloneErr != nil {
				return "", fmt.Errorf("clone repo: %w", cloneErr)
			}
			issueTarget.BaseBranch = firstNonEmpty(repository.DefaultBranch, issueTarget.BaseBranch, "main")
			issueTarget.RepositoryURL = firstNonEmpty(repository.HTMLURL, issueTarget.RepositoryURL)

			if err := git.SetUser(projectPath, envOrDefault("GIT_USER_NAME", "CrewAI Bot"), envOrDefault("GIT_USER_EMAIL", "bot@crewai.local")); err != nil {
				return "", fmt.Errorf("failed to configure git user: %w", err)
			}

			// Добавляем upstream если форк
			if issueTarget.ForkOwner != "" {
				upstreamURL := fmt.Sprintf("https://github.com/%s/%s.git", issueTarget.Owner, issueTarget.Repo)
				if err := s.githubClient.AddUpstream(projectPath, upstreamURL); err != nil {
					return "", fmt.Errorf("add upstream: %w", err)
				}
				if err := s.githubClient.SyncWithUpstream(ctx, projectPath, issueTarget.BaseBranch); err != nil {
					log.Printf("Warning: upstream sync failed: %v", err)
				}
			}

			// Читаем новые комментарии с даты последней задачи
			newComments, commentErr := s.githubClient.GetIssueCommentsSince(ctx, issueTarget.Owner, issueTarget.Repo, issueTarget.Number, existingTask.UpdatedAt)
			if commentErr != nil {
				log.Printf("Warning: failed to fetch new comments: %v", commentErr)
			} else {
				issueTarget.NewComments = newComments
				if len(newComments) > 0 {
					log.Printf("Found %d new comments since %s", len(newComments), existingTask.UpdatedAt)
				}
			}

			if err := git.CreateBranch(projectPath, issueTarget.BranchName); err != nil {
				return "", fmt.Errorf("failed to create branch: %w", err)
			}
			issueTarget.Cloned = true
			log.Printf("Project restored for existing issue at: %s (branch %s)", projectPath, issueTarget.BranchName)
			return projectPath, nil
		}

		// Новый issue — проверяем права и форкаем при необходимости
		forkOwner, forkCloneURL, forkErr := s.ensureFork(ctx, issueTarget)
		if forkErr != nil {
			return "", fmt.Errorf("fork setup: %w", forkErr)
		}

		targetOwner := issueTarget.Owner
		targetRepo := issueTarget.Repo
		if forkOwner != "" {
			targetOwner = forkOwner
		}

		repository, err := s.githubClient.CloneRepository(ctx, targetOwner, targetRepo, projectPath)
		if err != nil {
			return "", fmt.Errorf("failed to clone GitHub issue repository: %w", err)
		}
		issueTarget.BaseBranch = firstNonEmpty(repository.DefaultBranch, issueTarget.BaseBranch, "main")
		issueTarget.RepositoryURL = firstNonEmpty(repository.HTMLURL, issueTarget.RepositoryURL)
		if err := git.SetUser(projectPath, envOrDefault("GIT_USER_NAME", "CrewAI Bot"), envOrDefault("GIT_USER_EMAIL", "bot@crewai.local")); err != nil {
			return "", fmt.Errorf("failed to configure git user: %w", err)
		}

		// Если используем форк — добавляем upstream remote
		if forkOwner != "" {
			upstreamURL := fmt.Sprintf("https://github.com/%s/%s.git", issueTarget.Owner, issueTarget.Repo)
			if err := s.githubClient.AddUpstream(projectPath, upstreamURL); err != nil {
				return "", fmt.Errorf("failed to add upstream remote: %w", err)
			}
			if err := s.githubClient.SyncWithUpstream(ctx, projectPath, issueTarget.BaseBranch); err != nil {
				log.Printf("Warning: upstream sync failed: %v", err)
			}
			issueTarget.Forked = true
			issueTarget.ForkOwner = forkOwner
			issueTarget.ForkCloneURL = forkCloneURL
			log.Printf("Fork %s/%s in use, upstream added: %s/%s", forkOwner, issueTarget.Repo, issueTarget.Owner, issueTarget.Repo)
		}

		if err := git.CreateBranch(projectPath, issueTarget.BranchName); err != nil {
			return "", fmt.Errorf("failed to create pull request branch: %w", err)
		}
		issueTarget.Cloned = true
		log.Printf("GitHub issue repository cloned at: %s (branch %s)", projectPath, issueTarget.BranchName)
		return projectPath, nil
	}

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create project dir: %w", err)
	}
	if err := s.initGitRepo(projectPath); err != nil {
		return "", fmt.Errorf("failed to init git: %w", err)
	}
	return projectPath, nil
}

// findExistingTaskByIssue — ищет предыдущий task с таким же issue URL.
// Возвращает самую свежую выполненную задачу с NixStorePath.
func (s *Service) findExistingTaskByIssue(ctx context.Context, issueURL string) *models.Task {
	if issueURL == "" || s.db == nil {
		return nil
	}
	var tasks []models.Task
	// Ищем по JSONB полю Meta: {"github_issue_url": "<url>"}
	likeQuery := `%{"github_issue_url":"` + issueURL + `"}%`
	if err := s.db.Where("meta LIKE ? AND status = 'done' AND nix_store_path != ''", likeQuery).
		Order("updated_at DESC").
		Limit(1).
		Find(&tasks).Error; err != nil {
		log.Printf("findExistingTaskByIssue: %v", err)
		return nil
	}
	if len(tasks) == 0 {
		return nil
	}
	return &tasks[0]
}

// parseTaskMeta — достаёт map[string]string из JSONB поля Meta задачи.
func (s *Service) parseTaskMeta(task *models.Task) map[string]string {
	if task == nil || task.Meta == "" {
		return nil
	}
	var meta map[string]string
	if err := json.Unmarshal([]byte(task.Meta), &meta); err != nil {
		log.Printf("parseTaskMeta: %v", err)
		return nil
	}
	return meta
}

// ensureFork — проверяет права на запись в репозиторий и, при необходимости,
// создаёт форк под аккаунтом бота. Возвращает owner и clone URL форка
// (или пустые строки, если форк не нужен).
func (s *Service) ensureFork(ctx context.Context, target *gh.IssueTarget) (forkOwner, forkCloneURL string, err error) {
	canWrite, permErr := s.githubClient.CheckWritePermission(ctx, target.Owner, target.Repo)
	if permErr != nil {
		log.Printf("Warning: cannot check write permission for %s/%s: %v — proceeding without fork", target.Owner, target.Repo, permErr)
		return "", "", nil
	}
	if canWrite {
		log.Printf("Write permission confirmed for %s/%s — no fork needed", target.Owner, target.Repo)
		return "", "", nil
	}

	log.Printf("No write permission for %s/%s — creating fork", target.Owner, target.Repo)
	fork, err := s.githubClient.ForkRepository(ctx, target.Owner, target.Repo)
	if err != nil {
		return "", "", fmt.Errorf("fork %s/%s: %w", target.Owner, target.Repo, err)
	}

	if err := s.githubClient.WaitForkReady(ctx, fork.Owner.Login, target.Repo); err != nil {
		return "", "", fmt.Errorf("wait for fork %s/%s: %w", fork.Owner.Login, target.Repo, err)
	}

	log.Printf("Repository forked to %s/%s", fork.Owner.Login, target.Repo)
	return fork.Owner.Login, fork.CloneURL, nil
}

func ensureCloneTarget(projectPath string) error {
	if entries, err := os.ReadDir(projectPath); err == nil {
		if len(entries) > 0 {
			return fmt.Errorf("project dir already exists and is not empty: %s", projectPath)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check project dir: %w", err)
	}
	return nil
}

// initGitRepo — инициализирует git-репозиторий с настройкой пользователя
func (s *Service) initGitRepo(repoPath string) error {
	if err := git.InitRepo(repoPath); err != nil {
		return err
	}
	if err := git.WriteGitignore(repoPath); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	// Ensure default branch name is "main" regardless of git version
	exec.Command("git", "-C", repoPath, "branch", "-m", "main").Run()
	userName := envOrDefault("GIT_USER_NAME", "CrewAI Bot")
	userEmail := envOrDefault("GIT_USER_EMAIL", "bot@crewai.local")
	if err := git.SetUser(repoPath, userName, userEmail); err != nil {
		return err
	}
	if err := git.InitialCommit(repoPath, "Initial commit"); err != nil {
		return err
	}
	log.Printf("Git repository initialized at: %s", repoPath)
	return nil
}

// mergeManagerBranches — сливает manager-{role} ветки в итоговую ветку.
// Если targetBranch не указан или не существует, использует текущую ветку.
func (s *Service) mergeManagerBranches(repoPath string, roles []models.ManagerRole, targetBranch string) {
	if targetBranch == "" {
		targetBranch = "main"
	}
	if err := git.CheckoutBranch(repoPath, targetBranch); err != nil {
		// Возможно ветка называется master или мы уже на нужной — пробуем определить текущую
		current, _ := git.GetCurrentBranch(repoPath)
		if current != "" {
			targetBranch = current
			log.Printf("Falling back to current branch: %s", current)
		} else {
			log.Printf("Failed to checkout %s: %v", targetBranch, err)
			return
		}
	}
	for _, role := range roles {
		branchName := fmt.Sprintf("manager-%s", role.Role)
		msg := fmt.Sprintf("Merge %s manager branch", role.Role)
		if err := git.MergeBranch(repoPath, branchName, msg); err != nil {
			log.Printf("Failed to merge branch %s: %v", branchName, err)
		}
	}
	log.Printf("Merged all manager branches into %s", targetBranch)
}

// generateFlake — записывает flake.nix в корень проекта для Nix-совместимости.
// Использует FlakeBuilder для генерации богатого flake.nix с зависимостями
// на основе techStack, определённого AI на этапе планирования.
// После записи flake.nix:
//   1. Генерирует flake.lock для закрепления версий зависимостей.
//   2. Коммитит flake.nix + flake.lock в git, чтобы они не потерялись
//      при последующих git-операциях (ветвление/мерж воркеров).
func (s *Service) generateFlake(projectPath, taskID, title string, techStack []string, progress rules.ProgressFunc) {
	packages := NewFlakeBuilder().ResolveFromTechStacks(techStack)
	s.WriteFlake(projectPath, taskID, title, packages)
	s.ensureFlakeLock(projectPath, progress)
	s.nixFlakeCheck(projectPath)

	// Коммитим flake.nix и flake.lock, чтобы они не потерялись при git-операциях
	if err := git.Add(projectPath, "flake.nix", "flake.lock"); err != nil {
		log.Printf("Warning: git add flake.nix failed: %v", err)
		return
	}
	if err := git.Commit(projectPath, "Add Nix flake configuration"); err != nil {
		log.Printf("Warning: git commit flake.nix failed (may already be committed): %v", err)
	}
}

// detectNixSystem — определяет систему Nix на основе архитектуры хоста
func detectNixSystem() string {
	switch runtime.GOARCH {
	case "arm64":
		return "aarch64-linux"
	case "amd64":
		return "x86_64-linux"
	default:
		return "x86_64-linux"
	}
}

// snapshotProject — сохраняет проект в Nix store и возвращает store path
func (s *Service) snapshotProject(projectPath, taskID string) (storePath, flakeContent string, err error) {
	log.Printf("Snapshoting project to Nix store: %s", projectPath)

	stagingDir, err := prepareSnapshotDir(projectPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare snapshot: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	cmd := exec.Command("nix-store", "--add", stagingDir)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("nix snapshot failed (nix-store not available?): %w", err)
	}
	storePath = strings.TrimSpace(string(out))
	log.Printf("Project snapshoted to Nix store: %s", storePath)

	registerGCRoot(storePath, taskID)

	flakePath := filepath.Join(projectPath, "flake.nix")
	flakeBytes, rerr := os.ReadFile(flakePath)
	if rerr == nil {
		flakeContent = string(flakeBytes)
	}

	if storePath != "" && flakeContent != "" {
		annotated := fmt.Sprintf("%s\n# NixStorePath: %s\n# SnapshotTime: %s\n", flakeContent, storePath, time.Now().UTC().Format(time.RFC3339))
		if werr := os.WriteFile(flakePath, []byte(annotated), 0644); werr != nil {
			log.Printf("Failed to annotate flake.nix: %v", werr)
		}
	}

	return storePath, flakeContent, nil
}

// prepareSnapshotDir — создаёт временную директорию только с отслеживаемыми git-файлами
// (уважает .gitignore, исключает node_modules, .git и т.д.)
func prepareSnapshotDir(projectPath string) (string, error) {
	stagingDir, err := os.MkdirTemp("", "octra-snapshot-")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "-C", projectPath, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	out, err := cmd.Output()
	if err != nil {
		os.RemoveAll(stagingDir)
		return "", fmt.Errorf("git ls-files failed: %w", err)
	}

	files := strings.Split(strings.TrimRight(string(out), "\x00"), "\x00")
	for _, file := range files {
		if file == "" {
			continue
		}
		src := filepath.Join(projectPath, file)
		dst := filepath.Join(stagingDir, file)

		// Пропускаем директории (git ls-files может вернуть их, а copyFile копирует только файлы)
		if info, statErr := os.Stat(src); statErr != nil {
			log.Printf("Warning: cannot stat %s: %v", src, statErr)
			continue
		} else if info.IsDir() {
			continue
		}

		if err := copyFile(src, dst); err != nil {
			os.RemoveAll(stagingDir)
			return "", fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	flakeSrc := filepath.Join(projectPath, "flake.nix")
	if _, err := os.Stat(flakeSrc); err == nil {
		if err := copyFile(flakeSrc, filepath.Join(stagingDir, "flake.nix")); err != nil {
			os.RemoveAll(stagingDir)
			return "", fmt.Errorf("failed to copy flake.nix: %w", err)
		}
	}

	return stagingDir, nil
}

// copyFile — копирует файл с созданием родительских директорий
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// registerGCRoot — регистрирует store path как GC root, чтобы Nix не удалил его
func registerGCRoot(storePath, taskID string) {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Failed to get current user for GC root: %v", err)
		return
	}
	gcrootsDir := filepath.Join("/nix/var/nix/gcroots/per-user", currentUser.Username)
	if err := os.MkdirAll(gcrootsDir, 0755); err != nil {
		log.Printf("Failed to create GC roots dir %s: %v", gcrootsDir, err)
		return
	}
	rootPath := filepath.Join(gcrootsDir, fmt.Sprintf("octra-project-%s", taskID))
	os.Remove(rootPath)
	if err := os.Symlink(storePath, rootPath); err != nil {
		log.Printf("Failed to create GC root %s -> %s: %v", rootPath, storePath, err)
		return
	}
	log.Printf("GC root registered: %s -> %s", rootPath, storePath)
}

// restoreProjectFromStore — восстанавливает проект из Nix store по store path
func (s *Service) restoreProjectFromStore(storePath, destPath string) error {
	log.Printf("Restoring project from Nix store: %s -> %s", storePath, destPath)

	resolvedPath := resolveStorePath(storePath)
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return fmt.Errorf("Nix store path not found (may have been garbage collected): %s", storePath)
	}

	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("failed to clean dest path: %w", err)
	}
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create dest dir: %w", err)
	}

	cmd := exec.Command("cp", "-r", resolvedPath+"/.", destPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy from Nix store: %w", err)
	}

	// Fix permissions: Nix store files are read-only (0444), делаем writable для git
	if err := exec.Command("chmod", "-R", "u+w", destPath).Run(); err != nil {
		log.Printf("Failed to set writable permissions: %v", err)
	}

	if err := git.SetUser(destPath,
		envOrDefault("GIT_USER_NAME", "CrewAI Bot"),
		envOrDefault("GIT_USER_EMAIL", "bot@crewai.local")); err != nil {
		log.Printf("Failed to set git user after restore: %v", err)
	}

	log.Printf("Project restored successfully to: %s", destPath)
	return nil
}

// ReadProjectFiles — читает файлы проекта из директории, фильтруя игнорируемые
func ReadProjectFiles(projectPath string) ([]streamedSolutionFile, error) {
	var files []streamedSolutionFile
	seen := map[string]bool{}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		if util.IsIgnoredPath(relPath) || seen[relPath] {
			return nil
		}
		seen[relPath] = true

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		encoding := ""
		contentStr := string(content)
		if util.IsBinaryPath(relPath) {
			contentStr = base64.StdEncoding.EncodeToString(content)
			encoding = "base64"
		}

		files = append(files, streamedSolutionFile{
			Path:     relPath,
			Content:  contentStr,
			Language: util.LanguageForPath(relPath),
			Encoding: encoding,
			Status:   "ready",
		})
		return nil
	})

	return files, err
}

// RestoreProjectFromNix — публичный метод для восстановления из Nix store
func (s *Service) RestoreProjectFromNix(nixStorePath, destPath string) error {
	return s.restoreProjectFromStore(nixStorePath, destPath)
}

// resolveStorePath — учитывает NIX_STORE env, если store настроен на нестандартный путь
func resolveStorePath(storePath string) string {
	nixStore := os.Getenv("NIX_STORE")
	if nixStore == "" {
		return storePath
	}
	if strings.HasPrefix(storePath, "/nix/store") {
		return filepath.Join(nixStore, strings.TrimPrefix(storePath, "/nix/store"))
	}
	return storePath
}

// RestoreProject — публичный метод для восстановления проекта по taskID
func (s *Service) RestoreProject(taskID string) (string, error) {
	task := &models.Task{}
	if err := s.db.First(task, "id = ?", taskID).Error; err != nil {
		return "", fmt.Errorf("task not found: %w", err)
	}

	if task.NixStorePath == "" {
		return "", fmt.Errorf("task %s has no Nix store snapshot", taskID)
	}

	projectsDir := os.Getenv("PROJECTS_DIR")
	if projectsDir == "" {
		projectsDir = "/workspace/projects"
	}

	projectPath := filepath.Join(projectsDir, taskID, util.SanitizeProjectName(task.Title))

	if err := s.restoreProjectFromStore(task.NixStorePath, projectPath); err != nil {
		return "", fmt.Errorf("restore failed: %w", err)
	}

	return projectPath, nil
}

// cleanupProject — снапшотит проект в Nix store, затем удаляет workspace-директорию
func (s *Service) cleanupProject(projectPath, taskID string) {
	if projectPath == "" {
		return
	}

	log.Printf("Cleaning up project directory: %s", projectPath)

	storePath, flakeContent, err := s.snapshotProject(projectPath, taskID)
	if err != nil {
		log.Printf("WARNING: snapshot failed (%v), keeping project directory at %s to avoid data loss", err, projectPath)
		return
	}

	if storePath != "" && taskID != "" {
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"nix_store_path": storePath,
			"nix_flake":      flakeContent,
		})
	}

	os.RemoveAll(projectPath)
	log.Printf("Project directory removed. Nix store path: %s", storePath)
}

// sanitizeNixComment — заменяет недопустимые символы для Nix комментария
func sanitizeNixComment(s string) string {
	r := strings.NewReplacer(
		"\n", " ",
		"\r", "",
		"\t", " ",
		"\"", "'",
		"\\", "/",
	)
	return r.Replace(s)
}

// envOrDefault — простая утилита для чтения env с дефолтом
func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

