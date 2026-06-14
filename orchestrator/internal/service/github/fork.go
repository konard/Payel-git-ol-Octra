package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ForkResponse — минимальные поля форка из GitHub API.
type ForkResponse struct {
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// IsPermissionError — распознаёт ошибку отсутствия прав на запись в репозиторий
// (push возвращает HTTP 403 / "Permission to ... denied"). Используется, чтобы
// переключиться на создание PR из форка (issue #85).
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "403") ||
		strings.Contains(msg, "permission to") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "denied to") ||
		strings.Contains(msg, "write access to repository not granted")
}

// ForkRepository — создаёт форк upstream-репозитория в аккаунт бота.
// GitHub отвечает 202 Accepted; форк создаётся асинхронно (см. WaitForRepository).
func (c *Client) ForkRepository(ctx context.Context, owner, repo string) (*ForkResponse, error) {
	var fork ForkResponse
	path := fmt.Sprintf("/repos/%s/%s/forks", owner, repo)
	if err := c.doJSON(ctx, "POST", path, struct{}{}, &fork, http.StatusAccepted); err != nil {
		return nil, fmt.Errorf("failed to fork repository: %w", err)
	}
	log.Printf("Forked %s/%s -> %s", owner, repo, fork.FullName)
	return &fork, nil
}

// WaitForRepository — ждёт, пока репозиторий (форк) станет доступен через API.
// Форк создаётся асинхронно, поэтому push сразу после fork может упасть.
func (c *Client) WaitForRepository(ctx context.Context, owner, repo string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for attempt := 0; ; attempt++ {
		if _, err := c.GetRepository(ctx, owner, repo); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("repository %s/%s not ready after %s", owner, repo, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// PushBranchToRemote — пушит ветку в произвольный remote URL (например, форк),
// используя временный remote. Токен встраивается в URL для авторизации.
func (c *Client) PushBranchToRemote(ctx context.Context, dir, branch, remoteURL, remoteName string) error {
	if branch == "" {
		return fmt.Errorf("branch is required")
	}
	if remoteName == "" {
		remoteName = "octra-fork"
	}
	// Снимаем старый remote с тем же именем, если остался от прошлого прогона.
	rm := exec.CommandContext(ctx, "git", "remote", "remove", remoteName)
	rm.Dir = dir
	rm.Run() // ошибку игнорируем — remote мог не существовать

	auth := c.authenticatedGitURL(remoteURL)
	add := exec.CommandContext(ctx, "git", "remote", "add", remoteName, auth)
	add.Dir = dir
	if out, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add failed: %w - %s", err, c.sanitize(string(out)))
	}

	push := exec.CommandContext(ctx, "git", "push", "-u", remoteName, branch, "--force")
	push.Dir = dir
	out, err := push.CombinedOutput()
	if err != nil {
		log.Printf("git push to fork output: %s", c.sanitize(string(out)))
		return fmt.Errorf("git push to fork failed: %w - %s", err, c.sanitize(string(out)))
	}
	log.Printf("git push to fork output: %s", c.sanitize(string(out)))
	return nil
}
