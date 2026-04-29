package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"boss/pkg/models"
)

type Client struct {
	token    string
	username string
	email    string
}

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

type CreateRepoResponse struct {
	HTMLURL string `json:"html_url"`
	SSHURL  string `json:"ssh_url"`
}

func NewClient(token, username, email string) *Client {
	return &Client{
		token:    token,
		username: username,
		email:    email,
	}
}

func (c *Client) CreateRepository(ctx context.Context, task *models.Task) (string, error) {
	repoName := fmt.Sprintf("crewai-task-%s", task.ID.String()[:8])

	reqBody := CreateRepoRequest{
		Name:        repoName,
		Description: fmt.Sprintf("CrewAI generated project for task: %s", task.Title),
		Private:     false, // Make public for demo purposes
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

// PushToRepository pushes an existing git repository to GitHub
func (c *Client) PushToRepository(ctx context.Context, task *models.Task, repoPath string, repoURL string) error {
	// Configure git user (in case not set)
	if err := c.configureGit(repoPath); err != nil {
		log.Printf("Warning: failed to configure git: %v", err)
	}

	// Add remote origin
	if err := c.addRemote(repoPath, repoURL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Push to GitHub
	if err := c.gitPush(repoPath); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	log.Printf("Successfully pushed to GitHub repository: %s", repoURL)
	return nil
}


func (c *Client) initGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

func (c *Client) configureGit(dir string) error {
	cmds := []*exec.Cmd{
		exec.Command("git", "config", "user.name", c.username),
		exec.Command("git", "config", "user.email", c.email),
	}

	for _, cmd := range cmds {
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) addRemote(dir, repoURL string) error {
	// Convert HTTPS URL to use token for authentication
	if strings.HasPrefix(repoURL, "https://") {
		// Insert token into URL: https://TOKEN@github.com/user/repo
		parts := strings.Split(repoURL, "https://")
		if len(parts) == 2 {
			repoURL = "https://" + c.token + "@" + parts[1]
		}
	}

	cmd := exec.Command("git", "remote", "add", "origin", repoURL)
	cmd.Dir = dir
	return cmd.Run()
}

func (c *Client) gitAdd(dir string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	return nil
}

func (c *Client) gitCommit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	return cmd.Run()
}

func (c *Client) gitPush(dir string) error {
	cmd := exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = dir
	return cmd.Run()
}
