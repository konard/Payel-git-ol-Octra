package boss

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"orchestrator/internal/service/git"
	gh "orchestrator/internal/service/github"
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

// mergeManagerBranches — сливает manager-{role} ветки в итоговую ветку
func (s *Service) mergeManagerBranches(repoPath string, roles []models.ManagerRole, targetBranch string) {
	if targetBranch == "" {
		targetBranch = "main"
	}
	if err := git.CheckoutBranch(repoPath, targetBranch); err != nil {
		log.Printf("Failed to checkout %s: %v", targetBranch, err)
		return
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

// generateFlake — записывает flake.nix в корень проекта для Nix-совместимости
func (s *Service) generateFlake(projectPath, taskID, title string) {
	flakeContent := fmt.Sprintf(`{
  description = "Octra project: %s - %s";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages.${system}.default = pkgs.stdenv.mkDerivation {
        name = "octra-project-%s";
        src = ./.;
        installPhase = ''
          mkdir -p "$out"
          cp -r . "$out/"
        '';
      };
    };
}
`, taskID, sanitizeNixComment(title), taskID)
	if err := os.WriteFile(filepath.Join(projectPath, "flake.nix"), []byte(flakeContent), 0644); err != nil {
		log.Printf("Failed to write flake.nix: %v", err)
		return
	}
	log.Printf("Generated flake.nix at: %s", filepath.Join(projectPath, "flake.nix"))
}

// snapshotProject — сохраняет проект в Nix store и возвращает store path
func (s *Service) snapshotProject(projectPath, taskID string) (string, string) {
	log.Printf("Snapshoting project to Nix store: %s", projectPath)

	cmd := exec.Command("nix-store", "--add", projectPath)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Nix snapshot failed (nix-store not available?): %v", err)
		return "", ""
	}
	storePath := strings.TrimSpace(string(out))
	log.Printf("Project snapshoted to Nix store: %s", storePath)

	flakePath := filepath.Join(projectPath, "flake.nix")
	flakeBytes, err := os.ReadFile(flakePath)
	flakeContent := ""
	if err == nil {
		flakeContent = string(flakeBytes)
	}

	if storePath != "" && flakeContent != "" {
		annotated := fmt.Sprintf("%s\n# NixStorePath: %s\n# SnapshotTime: %s\n", flakeContent, storePath, time.Now().UTC().Format(time.RFC3339))
		if err := os.WriteFile(flakePath, []byte(annotated), 0644); err != nil {
			log.Printf("Failed to annotate flake.nix: %v", err)
		}
	}

	return storePath, flakeContent
}

// restoreProjectFromStore — восстанавливает проект из Nix store по store path
func (s *Service) restoreProjectFromStore(storePath, destPath string) error {
	log.Printf("Restoring project from Nix store: %s -> %s", storePath, destPath)

	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return fmt.Errorf("Nix store path not found (may have been garbage collected): %s", storePath)
	}

	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("failed to clean dest path: %w", err)
	}
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create dest dir: %w", err)
	}

	cmd := exec.Command("cp", "-r", storePath+"/.", destPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy from Nix store: %w", err)
	}

	if err := git.SetUser(destPath,
		envOrDefault("GIT_USER_NAME", "CrewAI Bot"),
		envOrDefault("GIT_USER_EMAIL", "bot@crewai.local")); err != nil {
		log.Printf("Failed to set git user after restore: %v", err)
	}

	log.Printf("Project restored successfully to: %s", destPath)
	return nil
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

	storePath, flakeContent := s.snapshotProject(projectPath, taskID)
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

