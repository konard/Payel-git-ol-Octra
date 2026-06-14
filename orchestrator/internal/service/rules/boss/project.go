package boss

import (
	"context"
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
		repository, err := s.githubClient.CloneRepository(ctx, issueTarget.Owner, issueTarget.Repo, projectPath)
		if err != nil {
			return "", fmt.Errorf("failed to clone GitHub issue repository: %w", err)
		}
		issueTarget.BaseBranch = firstNonEmpty(repository.DefaultBranch, issueTarget.BaseBranch, "main")
		issueTarget.RepositoryURL = firstNonEmpty(repository.HTMLURL, issueTarget.RepositoryURL)
		if err := git.SetUser(projectPath, envOrDefault("GIT_USER_NAME", "CrewAI Bot"), envOrDefault("GIT_USER_EMAIL", "bot@crewai.local")); err != nil {
			return "", fmt.Errorf("failed to configure git user: %w", err)
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
//  1. Генерирует flake.lock для закрепления версий зависимостей.
//  2. Коммитит flake.nix + flake.lock в git, чтобы они не потерялись
//     при последующих git-операциях (ветвление/мерж воркеров).
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
		if _, err := copyFile(src, dst); err != nil {
			os.RemoveAll(stagingDir)
			return "", fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	flakeSrc := filepath.Join(projectPath, "flake.nix")
	if _, err := os.Stat(flakeSrc); err == nil {
		if _, err := copyFile(flakeSrc, filepath.Join(stagingDir, "flake.nix")); err != nil {
			os.RemoveAll(stagingDir)
			return "", fmt.Errorf("failed to copy flake.nix: %w", err)
		}
	}

	return stagingDir, nil
}

// copyFile — копирует файл с созданием родительских директорий.
// Возвращает (skipped=true, nil) для записей, которые нельзя/не нужно копировать
// как обычный файл: симлинки в Nix store (например, `result` от `nix build`),
// директории и битые симлинки. Раньше попытка скопировать такой `result` падала
// с `copy_file_range: is a directory` и срывала весь снапшот проекта (issue #85).
func copyFile(src, dst string) (skipped bool, err error) {
	info, err := os.Lstat(src)
	if err != nil {
		return false, err
	}

	// Симлинки: не идём по ссылке (это и вызывало падение на `result`,
	// указывающем на директорию в /nix/store), а воссоздаём саму ссылку.
	if info.Mode()&os.ModeSymlink != 0 {
		target, rerr := os.Readlink(src)
		if rerr != nil {
			return false, rerr
		}
		// Артефакты сборки Nix (`result`, `result-*`) не должны попадать в снапшот.
		if strings.HasPrefix(target, "/nix/store") {
			return true, nil
		}
		// Битый симлинк или указывает на директорию — пропускаем, чтобы не падать.
		if resolved, serr := os.Stat(src); serr != nil || resolved.IsDir() {
			return true, nil
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return false, err
		}
		return false, os.Symlink(target, dst)
	}

	// git ls-files не должен возвращать директории, но на всякий случай пропускаем.
	if info.IsDir() {
		return true, nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return false, err
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return false, err
	}
	return false, nil
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
