package boss

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nodes/internal/service/git"
	gh "nodes/internal/service/github"
	"nodes/internal/service/util"
	"nodes/pkg/models"
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

// cleanupProject — удаляет workspace-директорию после завершения задачи
func (s *Service) cleanupProject(projectPath string) {
	if projectPath == "" {
		return
	}
	log.Printf("Cleaning up project directory: %s", projectPath)
	os.RemoveAll(projectPath)
}

// envOrDefault — простая утилита для чтения env с дефолтом
func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
