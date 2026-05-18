package boss

import (
	"context"
	"log"
	"os"

	"nodes/pkg/database"
	"nodes/pkg/models"
)

// pushToGitHub — создаёт репозиторий и пушит результат, если есть GITHUB_TOKEN.
// Возвращает URL созданного репозитория или пустую строку.
func (s *Service) pushToGitHub(ctx context.Context, task *models.Task, projectPath string) string {
	if os.Getenv("GITHUB_TOKEN") == "" || s.githubClient == nil {
		return ""
	}
	if task.ProjectJSON != "" {
		return task.ProjectJSON
	}

	log.Printf("Pushing results to GitHub...")
	repoURL, err := s.githubClient.CreateRepository(ctx, task)
	if err != nil {
		log.Printf("Failed to create GitHub repository: %v", err)
		return ""
	}
	if err := s.githubClient.PushToRepository(ctx, task, projectPath, repoURL); err != nil {
		log.Printf("Failed to push to GitHub: %v", err)
		return ""
	}
	log.Printf("Successfully pushed to GitHub: %s", repoURL)
	task.ProjectJSON = repoURL
	database.Db.Save(task)
	return repoURL
}
