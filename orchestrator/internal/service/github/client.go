package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"orchestrator/pkg/models"
)

// Client — клиент GitHub: создаёт репозитории и пушит готовые проекты
type Client struct {
	token      string
	username   string
	email      string
	apiBaseURL string
	httpClient *http.Client
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
	if username == "" {
		username = "Octra Bot"
	}
	if email == "" {
		email = "bot@octra.local"
	}
	return &Client{
		token:      token,
		username:   username,
		email:      email,
		apiBaseURL: "https://api.github.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateRepository — создаёт новый публичный репозиторий для задачи
func (c *Client) CreateRepository(ctx context.Context, task *models.Task) (string, error) {
	repoName := fmt.Sprintf("octra-task-%s", task.ID.String()[:8])
	reqBody := CreateRepoRequest{
		Name:        repoName,
		Description: fmt.Sprintf("CrewAI generated project for task: %s", task.Title),
		Private:     false,
	}
	var repoResp CreateRepoResponse
	if err := c.doJSON(ctx, "POST", "/user/repos", reqBody, &repoResp, http.StatusCreated); err != nil {
		return "", fmt.Errorf("failed to create repository: %w", err)
	}
	log.Printf("Created GitHub repository: %s", repoResp.HTMLURL)
	return repoResp.HTMLURL, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any, expectedStatus int) error {
	var reader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), reader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(responseBody))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

func (c *Client) apiURL(path string) string {
	base := strings.TrimRight(c.apiBaseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
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

