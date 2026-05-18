package boss

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"

	gh "nodes/internal/service/github"
)

// detectGitHubIssueTask — включает PR-режим только для конкретных GitHub issue URL.
func (s *Service) detectGitHubIssueTask(ctx context.Context, req *CreateTaskRequest, taskID string) *gh.IssueTarget {
	ref, found := gh.ParseIssueReference(issueDetectionText(req))
	if !found {
		return nil
	}

	target := &gh.IssueTarget{
		IssueReference: *ref,
		BranchName:     gh.NewIssueBranchName(ref.Number, taskID),
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

	if s.githubClient != nil {
		if issue, err := s.githubClient.GetIssue(ctx, ref.Owner, ref.Repo, ref.Number); err != nil {
			log.Printf("Failed to fetch GitHub issue %s: %v", ref.URL, err)
		} else {
			target.IssueTitle = issue.Title
			target.IssueBody = issue.Body
			target.IssueURL = firstNonEmpty(issue.HTMLURL, ref.URL)
		}
		if repo, err := s.githubClient.GetRepository(ctx, ref.Owner, ref.Repo); err != nil {
			log.Printf("Failed to fetch GitHub repository %s/%s: %v", ref.Owner, ref.Repo, err)
		} else {
			target.BaseBranch = firstNonEmpty(repo.DefaultBranch, target.BaseBranch)
			target.RepositoryURL = repo.HTMLURL
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
	var b strings.Builder
	b.WriteString(strings.TrimSpace(description))
	b.WriteString("\n\nGITHUB ISSUE TARGET:\n")
	b.WriteString(fmt.Sprintf("- Repository: %s/%s\n", target.Owner, target.Repo))
	b.WriteString(fmt.Sprintf("- Issue: #%d %s\n", target.Number, target.IssueURL))
	if target.IssueTitle != "" {
		b.WriteString("- Issue title: ")
		b.WriteString(target.IssueTitle)
		b.WriteString("\n")
	}
	if target.IssueBody != "" {
		b.WriteString("\nIssue body:\n")
		b.WriteString(limitText(target.IssueBody, 4000))
		b.WriteString("\n")
	}
	b.WriteString("\nExecution rules:\n")
	b.WriteString("- Treat this as a pull request task for the referenced GitHub issue.\n")
	b.WriteString("- Preserve the existing repository structure and make the smallest focused change that fixes the issue.\n")
	b.WriteString("- Treat ordinary repository, package, documentation, and library URLs as reference material only unless they are concrete GitHub issue URLs.\n")
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

func limitText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "\n[truncated]"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
