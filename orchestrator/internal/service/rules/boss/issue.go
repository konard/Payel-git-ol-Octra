package boss

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"

	gh "orchestrator/internal/service/github"
)

// detectGitHubIssueTask — включает PR-режим только для конкретных GitHub issue URL.
func (s *Service) detectGitHubIssueTask(ctx context.Context, req *CreateTaskRequest, taskID string) *gh.IssueTarget {
	ref, found := gh.ParseIssueReference(issueDetectionText(req))
	if !found {
		return nil
	}

	target := &gh.IssueTarget{
		IssueReference: *ref,
		BranchName:     gh.NewIssueBranchName(taskID),
		BaseBranch:     "main",
		IssueURL:       ref.URL,
	}
	if req.Meta == nil {
		req.Meta = map[string]string{}
	}
	req.Meta["github_mode"] = "pull_request"
	req.Meta["github_issue_url"] = ref.URL
	req.Meta["github_repository"] = ref.Owner + "/" + ref.Repo
	req.Meta["github_issue_number"] = fmt.Sprintf("%d", ref.Number)
	req.Meta["github_branch"] = target.BranchName
	// task_kind помечает задачу как github-тип (план фикса, пункт 1), не мутируя
	// AI-классификацию task_type. Boss затем повышает task_type до "github".
	req.Meta["github_task_kind"] = TaskTypeGitHub

	if s.githubClient != nil {
		// Собираем полный паспорт issue ДО пайплайна (план фикса, пункты 1-2):
		// тело целиком, комментарии, метки, состояние, открытые PR.
		if instruction, err := s.githubClient.AnalyzeIssue(ctx, *ref); err != nil {
			log.Printf("Failed to analyze GitHub issue %s: %v", ref.URL, err)
		} else {
			target.Instruction = instruction
			target.IssueTitle = instruction.Title
			target.IssueBody = instruction.Body
			target.IssueURL = firstNonEmpty(instruction.URL, ref.URL)
			target.BaseBranch = firstNonEmpty(instruction.DefaultBranch, target.BaseBranch)
			target.RepositoryURL = instruction.RepositoryURL

			// Паспорт кладётся в Meta как JSON и передаётся по конвейеру без
			// пересборки и дрейфа (план фикса, пункт 2).
			if encoded, err := json.Marshal(instruction); err == nil {
				req.Meta["github_context"] = string(encoded)
			}
			if gh.IsTrivialIssue(instruction) {
				req.Meta["github_trivial"] = "true"
				log.Printf("GitHub issue %s detected as trivial (pipeline bypass eligible)", ref.URL)
			}
		}
	}

	req.Description = withGitHubIssueContext(req.Description, target)
	return target
}

func issueDetectionText(req *CreateTaskRequest) string {
	var b strings.Builder
	b.WriteString(req.Title)
	b.WriteString("\n")
	b.WriteString(req.Description)
	for _, key := range []string{"github_issue_url", "githubIssueUrl", "issue_url", "issueUrl", "repository_url", "repoUrl"} {
		if req.Meta != nil && req.Meta[key] != "" {
			b.WriteString("\n")
			b.WriteString(req.Meta[key])
		}
	}
	return b.String()
}

func withGitHubIssueContext(description string, target *gh.IssueTarget) string {
	kind := "Issue"
	if target.IsPullRequest {
		kind = "Pull request"
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(description))
	b.WriteString("\n\nGITHUB ISSUE TARGET:\n")
	b.WriteString(fmt.Sprintf("- Repository: %s/%s\n", target.Owner, target.Repo))
	b.WriteString(fmt.Sprintf("- %s: #%d %s\n", kind, target.Number, target.IssueURL))
	if target.IssueTitle != "" {
		b.WriteString(fmt.Sprintf("- %s title: ", kind))
		b.WriteString(target.IssueTitle)
		b.WriteString("\n")
	}
	if target.IssueBody != "" {
		b.WriteString(fmt.Sprintf("\n%s body:\n", kind))
		// Тело передаётся целиком (план фикса, пункт 2): больше не режем до 4000.
		b.WriteString(target.IssueBody)
		b.WriteString("\n")
	}
	if in := target.Instruction; in != nil {
		if in.State != "" {
			b.WriteString(fmt.Sprintf("- State: %s\n", in.State))
		}
		if len(in.Labels) > 0 {
			b.WriteString(fmt.Sprintf("- Labels: %s\n", strings.Join(in.Labels, ", ")))
		}
		if len(in.OpenPRs) > 0 {
			b.WriteString("- Existing open pull requests for this issue:\n")
			for _, pr := range in.OpenPRs {
				b.WriteString(fmt.Sprintf("  - #%d %s (%s) %s\n", pr.Number, pr.Title, pr.State, pr.URL))
			}
		}
		if len(in.Comments) > 0 {
			b.WriteString("\nComments:\n")
			for _, comment := range in.Comments {
				author := firstNonEmpty(comment.Author, "unknown")
				b.WriteString(fmt.Sprintf("--- %s (%s):\n", author, comment.CreatedAt))
				b.WriteString(comment.Body)
				b.WriteString("\n")
			}
		}
	}
	if target.Forked {
		b.WriteString(fmt.Sprintf("- Repository has been forked to %s/%s for pull request creation.\n", target.ForkOwner, target.Repo))
	}
	b.WriteString("\nExecution rules:\n")
	if target.IsPullRequest {
		b.WriteString("- Treat this as a pull request task that continues the work described by the referenced GitHub pull request.\n")
	} else {
		b.WriteString("- Treat this as a pull request task for the referenced GitHub issue.\n")
	}
	b.WriteString("- Preserve the existing repository structure and make the smallest focused change that fixes the issue.\n")
	b.WriteString("- Treat ordinary repository, package, documentation, and library URLs as reference material only unless they are concrete GitHub issue or pull request URLs.\n")
	return b.String()
}

func appendRepositoryContext(decision *DecisionResult, projectPath string) {
	summary := summarizeRepository(projectPath)
	if summary == "" {
		return
	}
	decision.TechnicalDescription += "\n\nEXISTING REPOSITORY CONTEXT:\n" + summary
}

func summarizeRepository(projectPath string) string {
	files := make([]string, 0, 120)
	err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(projectPath, path)
		if relErr != nil || rel == "." {
			return nil
		}
		if d.IsDir() {
			if shouldSkipRepoDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if len(files) < 120 {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return ""
	}
	sort.Strings(files)
	if len(files) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Repository file list (first 120 files):\n")
	for _, file := range files {
		b.WriteString("- ")
		b.WriteString(file)
		b.WriteString("\n")
	}
	return b.String()
}

func shouldSkipRepoDir(rel string) bool {
	base := filepath.Base(rel)
	switch base {
	case ".git", "node_modules", "vendor", "dist", "build", ".next", "target", "bin":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
