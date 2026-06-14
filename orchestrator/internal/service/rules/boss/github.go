package boss

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"orchestrator/internal/service/git"
	gh "orchestrator/internal/service/github"
	"orchestrator/pkg/database"
	"orchestrator/pkg/models"
)

// pushToGitHub — создаёт репозиторий и пушит результат, если есть GITHUB_TOKEN.
// Возвращает URL созданного репозитория или пустую строку.
// meta может содержать флаги "publish_repositories" и "create_pull_requests"
// для управления поведением из UI интеграций.
func (s *Service) pushToGitHub(ctx context.Context, task *models.Task, projectPath string, issueTarget *gh.IssueTarget, meta map[string]string) string {
	if os.Getenv("GITHUB_TOKEN") == "" || s.githubClient == nil {
		return ""
	}

	// Если флаги явно установлены в "false" — пропускаем соответствующую операцию
	publishRepos := meta["publish_repositories"] != "false"
	createPRs := meta["create_pull_requests"] != "false"

	if task.ProjectJSON != "" {
		return task.ProjectJSON
	}
	if issueTarget != nil && issueTarget.Cloned {
		if !createPRs {
			log.Printf("Skipping pull request creation (create_pull_requests=false)")
			// Всё равно пушим код в ветку, чтобы пользователь мог создать PR вручную
			if err := git.CheckoutBranch(projectPath, issueTarget.BranchName); err != nil {
				log.Printf("Failed to checkout branch: %v", err)
				return ""
			}
			if err := s.githubClient.PushBranch(ctx, projectPath, issueTarget.BranchName); err != nil {
				log.Printf("Failed to push branch: %v", err)
				return ""
			}
			return issueTarget.RepositoryURL
		}
		return s.createPullRequest(ctx, task, projectPath, issueTarget)
	}

	if !publishRepos {
		log.Printf("Skipping repository publish (publish_repositories=false)")
		return ""
	}

	log.Printf("Pushing results to GitHub...")
	// If the repo already has an "origin" remote (cloned or refinement), push update.
	// NOTE: we check for the remote, not just .git — initGitRepo also creates .git
	// for new projects but doesn't configure a remote.
	existingRemote := extractRemoteURL(projectPath)
	if existingRemote != "" {
		repoURL := task.ProjectJSON
		if repoURL == "" {
			repoURL = existingRemote
		}
		if err := s.githubClient.PushToRepository(ctx, task, projectPath, repoURL); err != nil {
			log.Printf("Failed to push update to existing: %v", err)
			return ""
		}
		return repoURL
	}
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

func (s *Service) createPullRequest(ctx context.Context, task *models.Task, projectPath string, target *gh.IssueTarget) string {
	log.Printf("Creating GitHub pull request for %s/%s#%d", target.Owner, target.Repo, target.Number)
	if err := git.CheckoutBranch(projectPath, target.BranchName); err != nil {
		log.Printf("Failed to checkout pull request branch: %v", err)
		return ""
	}
	headRef := target.BranchName
	if err := s.githubClient.PushBranch(ctx, projectPath, target.BranchName); err != nil {
		// 403 → нет прав на запись в upstream. Раньше пайплайн просто падал и PR
		// не создавался (issue #85). Теперь форкаем репозиторий и пушим в форк.
		if gh.IsPermissionError(err) {
			log.Printf("Push denied (no write access to %s/%s) — falling back to fork-based pull request", target.Owner, target.Repo)
			forkHead := s.pushBranchToFork(ctx, projectPath, target)
			if forkHead == "" {
				return ""
			}
			headRef = forkHead
		} else {
			log.Printf("Failed to push pull request branch: %v", err)
			return ""
		}
	}
	prTitle := pullRequestTitle(task, target)
	pr, err := s.githubClient.CreatePullRequest(ctx, gh.PullRequestRequest{
		Owner: target.Owner,
		Repo:  target.Repo,
		Title: prTitle,
		Head:  headRef,
		Base:  firstNonEmpty(target.BaseBranch, "main"),
		Body:  pullRequestBody(task, target),
	})
	if err != nil {
		log.Printf("Failed to create GitHub pull request: %v", err)
		return ""
	}
	log.Printf("Successfully created GitHub pull request: %s", pr.HTMLURL)

	// Capture the PR overview (number, title, diff stats) so the frontend can
	// render the Solution-pane summary without a round trip to GitHub.
	target.PullRequestNumber = pr.Number
	target.PullRequestTitle = prTitle
	if stats, statErr := git.DiffStats(projectPath, firstNonEmpty(target.BaseBranch, "main"), target.BranchName); statErr != nil {
		log.Printf("Failed to compute pull request diff stats: %v", statErr)
	} else {
		target.Commits = stats.Commits
		target.Additions = stats.Additions
		target.Deletions = stats.Deletions
		target.ChangedFiles = stats.ChangedFiles
	}

	task.ProjectJSON = pr.HTMLURL
	database.Db.Save(task)
	return pr.HTMLURL
}

// pushBranchToFork — форкает upstream-репозиторий и пушит ветку в форк.
// Возвращает head-ссылку вида "forkOwner:branch" для кросс-репозиторного PR,
// либо "" при ошибке. Вызывается, когда у бота нет прав на запись в upstream.
func (s *Service) pushBranchToFork(ctx context.Context, projectPath string, target *gh.IssueTarget) string {
	fork, err := s.githubClient.ForkRepository(ctx, target.Owner, target.Repo)
	if err != nil {
		log.Printf("Fork fallback failed: %v", err)
		return ""
	}
	if err := s.githubClient.WaitForRepository(ctx, fork.Owner.Login, target.Repo, 60*time.Second); err != nil {
		log.Printf("Fork not ready: %v", err)
		return ""
	}
	forkURL := fork.CloneURL
	if forkURL == "" {
		forkURL = fmt.Sprintf("https://github.com/%s/%s.git", fork.Owner.Login, target.Repo)
	}
	if err := s.githubClient.PushBranchToRemote(ctx, projectPath, target.BranchName, forkURL, "octra-fork"); err != nil {
		log.Printf("Failed to push branch to fork: %v", err)
		return ""
	}
	log.Printf("Pushed branch to fork %s/%s, opening cross-repo pull request", fork.Owner.Login, target.Repo)
	return fmt.Sprintf("%s:%s", fork.Owner.Login, target.BranchName)
}

func pullRequestTitle(task *models.Task, target *gh.IssueTarget) string {
	title := target.IssueTitle
	if title == "" {
		title = task.Title
	}
	if title == "" {
		title = "Octra generated fix"
	}
	if target.IsPullRequest {
		// Linking a PR to another PR with "Fix #N" would be inaccurate, so we
		// describe it as a continuation of the referenced pull request.
		return fmt.Sprintf("Continue #%d: %s", target.Number, title)
	}
	return fmt.Sprintf("Fix #%d: %s", target.Number, title)
}

func pullRequestBody(task *models.Task, target *gh.IssueTarget) string {
	// "Fixes #N" auto-closes the referenced issue on merge, which is the desired
	// behaviour for issue targets but wrong for pull request targets, so we use a
	// neutral reference for PRs.
	reference := "Fixes"
	if target.IsPullRequest {
		reference = "Related to"
	}
	return fmt.Sprintf(`%s %s/%s#%d

Generated by Octra for task %s.
	`, reference, target.Owner, target.Repo, target.Number, task.ID.String())
}

// extractRemoteURL — получает origin URL из локального .git
func extractRemoteURL(projectPath string) string {
	cmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
