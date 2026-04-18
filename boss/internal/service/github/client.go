package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func (c *Client) PushToRepository(ctx context.Context, task *models.Task, zipData []byte, repoURL string) error {
	// Create temporary directory for the project
	tempDir, err := os.MkdirTemp("", "crewai-push-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract ZIP data to temp directory
	if err := c.extractZip(zipData, tempDir); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	// Initialize git repository
	if err := c.initGitRepo(tempDir); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// Configure git user
	if err := c.configureGit(tempDir); err != nil {
		return fmt.Errorf("failed to configure git: %w", err)
	}

	// Add remote origin
	if err := c.addRemote(tempDir, repoURL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Add all files
	if err := c.gitAdd(tempDir); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Commit changes
	commitMsg := fmt.Sprintf("CrewAI Task: %s\n\n%s", task.Title, task.Description)
	if err := c.gitCommit(tempDir, commitMsg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push to GitHub
	if err := c.gitPush(tempDir); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	log.Printf("Successfully pushed to GitHub repository: %s", repoURL)
	return nil
}

func (c *Client) extractZip(zipData []byte, destDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		filePath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()
		if err != nil {
			return err
		}
	}

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
	return cmd.Run()
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
