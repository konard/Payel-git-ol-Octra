package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"nodes/pkg/models"
)

// Client — клиент GitHub: создаёт репозитории и пушит готовые проекты
type Client struct {
	token    string
	username string
	email    string
}

// CreateRepoRequest — тело запроса для создания репозитория
type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

// CreateRepoResponse — ответ GitHub API с URL созданного репозитория
type CreateRepoResponse struct {
	HTMLURL string `json:"html_url"`
	SSHURL  string `json:"ssh_url"`
}

// NewClient — создаёт клиента GitHub с токеном и git-идентичностью
func NewClient(token, username, email string) *Client {
	return &Client{
		token:    token,
		username: username,
		email:    email,
	}
}

// CreateRepository — создаёт новый публичный репозиторий для задачи
func (c *Client) CreateRepository(ctx context.Context, task *models.Task) (string, error) {
	repoName := fmt.Sprintf("crewai-task-%s", task.ID.String()[:8])
	reqBody := CreateRepoRequest{
		Name:        repoName,
		Description: fmt.Sprintf("CrewAI generated project for task: %s", task.Title),
		Private:     false,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/user/repos", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var repoResp CreateRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	log.Printf("Created GitHub repository: %s", repoResp.HTMLURL)
	return repoResp.HTMLURL, nil
}

// PushToRepository — добавляет файлы, коммитит и пушит проект в GitHub
func (c *Client) PushToRepository(ctx context.Context, task *models.Task, repoPath string, repoURL string) error {
	if err := c.configureGit(repoPath); err != nil {
		log.Printf("Warning: failed to configure git: %v", err)
	}
	if err := c.gitAdd(repoPath); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}
	if err := c.gitCommit(repoPath, "Initial commit - CrewAI generated project"); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	if err := c.addRemote(repoPath, repoURL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}
	if err := c.gitPush(repoPath); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	log.Printf("Successfully pushed to GitHub repository: %s", repoURL)
	return nil
}
