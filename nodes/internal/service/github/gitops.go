package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// configureGit — задаёт имя и e-mail пользователя git для коммитов
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

// addRemote — добавляет origin remote, встраивая токен для авторизации
func (c *Client) addRemote(dir, repoURL string) error {
	repoURL = c.authenticatedGitURL(repoURL)
	cmd := exec.Command("git", "remote", "add", "origin", repoURL)
	cmd.Dir = dir
	return cmd.Run()
}

// CloneRepository — клонирует репозиторий issue-задачи в рабочую директорию.
func (c *Client) CloneRepository(ctx context.Context, owner, repo, dir string) (*RepositoryResponse, error) {
	repository, err := c.GetRepository(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	cloneURL := repository.CloneURL
	if cloneURL == "" {
		cloneURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return nil, fmt.Errorf("failed to create project parent dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", c.authenticatedGitURL(cloneURL), dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %w - %s", err, c.sanitize(string(out)))
	}
	return repository, nil
}

// gitAdd — выполняет git add . с подробным логированием
func (c *Client) gitAdd(dir string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("git add output: %s", string(out))
		return fmt.Errorf("git add failed: %w - %s", err, string(out))
	}
	log.Printf("git add output: %s", string(out))
	return nil
}

// gitCommit — коммит изменений; пустой коммит не считается ошибкой
func (c *Client) gitCommit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "nothing to commit") {
			log.Printf("git commit: nothing to commit (already committed)")
			return nil
		}
		log.Printf("git commit output: %s", string(out))
		return fmt.Errorf("git commit failed: %w - %s", err, string(out))
	}
	log.Printf("git commit output: %s", string(out))
	return nil
}

// resolvePushBranch — выбирает имя ветки, которую будем пушить в origin
func (c *Client) resolvePushBranch(dir string) string {
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = dir
	branchOut, branchErr := branchCmd.CombinedOutput()
	currentBranch := "unknown"
	if branchErr == nil {
		currentBranch = strings.TrimSpace(string(branchOut))
	}
	log.Printf("git push: current branch is '%s'", currentBranch)

	branchesCmd := exec.Command("git", "branch", "-a")
	branchesCmd.Dir = dir
	branchesOut, _ := branchesCmd.CombinedOutput()
	log.Printf("git push: available branches: %s", string(branchesOut))

	if currentBranch == "" || currentBranch == "(no branch)" {
		log.Printf("git push: no current branch, checking HEAD")
		headCmd := exec.Command("git", "rev-parse", "HEAD")
		headCmd.Dir = dir
		headOut, headErr := headCmd.CombinedOutput()
		if headErr != nil {
			return ""
		}
		log.Printf("git push: HEAD is %s", strings.TrimSpace(string(headOut)))
		checkoutCmd := exec.Command("git", "checkout", "-b", "main")
		checkoutCmd.Dir = dir
		if err := checkoutCmd.Run(); err == nil {
			currentBranch = "main"
		}
	}

	pushBranch := currentBranch
	if currentBranch == "manager-backend" || currentBranch == "(no branch)" {
		pushBranch = "main"
		checkoutCmd := exec.Command("git", "checkout", "main")
		checkoutCmd.Dir = dir
		if err := checkoutCmd.Run(); err != nil {
			checkoutCmd = exec.Command("git", "checkout", "-b", "main")
			checkoutCmd.Dir = dir
			if err := checkoutCmd.Run(); err != nil {
				log.Printf("git checkout main failed: %v, using current branch", err)
				pushBranch = currentBranch
			}
		}
	}
	return pushBranch
}

// gitPush — пушит локальную ветку в origin, при необходимости с force
func (c *Client) gitPush(dir string) error {
	pushBranch := c.resolvePushBranch(dir)
	if pushBranch == "" {
		return fmt.Errorf("no commits found in repository")
	}
	log.Printf("git push: pushing branch '%s'", pushBranch)

	cmd := exec.Command("git", "push", "-u", "origin", pushBranch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "refusing to merge unrelated histories") {
			log.Printf("git push: merging unrelated histories, forcing push")
			forceCmd := exec.Command("git", "push", "-u", "origin", pushBranch, "--force")
			forceCmd.Dir = dir
			forceOut, forceErr := forceCmd.CombinedOutput()
			if forceErr != nil {
				log.Printf("git push --force output: %s", string(forceOut))
				return fmt.Errorf("git push --force failed: %w - %s", forceErr, string(forceOut))
			}
			log.Printf("git push --force output: %s", string(forceOut))
			return nil
		}
		log.Printf("git push output: %s", string(out))
		return fmt.Errorf("git push failed: %w - %s", err, string(out))
	}
	log.Printf("git push output: %s", string(out))
	return nil
}

// PushBranch — пушит указанную ветку в origin.
func (c *Client) PushBranch(ctx context.Context, dir, branch string) error {
	if branch == "" {
		return fmt.Errorf("branch is required")
	}
	cmd := exec.CommandContext(ctx, "git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("git push branch output: %s", c.sanitize(string(out)))
		return fmt.Errorf("git push branch failed: %w - %s", err, c.sanitize(string(out)))
	}
	log.Printf("git push branch output: %s", c.sanitize(string(out)))
	return nil
}
