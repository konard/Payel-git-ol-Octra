package github

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// issueURLPattern распознаёт как issue-ссылки (.../issues/N), так и
// pull request-ссылки (.../pull/N). Раньше принимались только issue-URL, из-за
// чего ссылка на pull request не запускала PR-режим (issue #44).
var issueURLPattern = regexp.MustCompile(`(?i)\b(?:https?://)?(?:www\.)?github\.com/([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)/(issues|pull|pulls)/([0-9]+)(?:[/?#\s\]).,;:!]|$)`)

// IssueReference — ссылка на конкретный GitHub issue или pull request.
type IssueReference struct {
	Owner         string
	Repo          string
	Number        int
	URL           string
	IsPullRequest bool
}

// IssueTarget — контекст задачи, которую нужно оформить pull request'ом.
type IssueTarget struct {
	IssueReference
	BranchName    string
	BaseBranch    string
	IssueTitle    string
	IssueBody     string
	IssueURL      string
	RepositoryURL string
	Cloned        bool

	// Поля результата — заполняются после создания pull request, чтобы фронтенд
	// мог показать обзор PR без перехода на GitHub (issue #44, часть 2).
	PullRequestNumber int
	PullRequestTitle  string
	Commits           int
	Additions         int
	Deletions         int
	ChangedFiles      []string
}

// IssueResponse — минимальные поля issue из GitHub API.
type IssueResponse struct {
	HTMLURL string `json:"html_url"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
}

// RepositoryResponse — минимальные поля repository из GitHub API.
type RepositoryResponse struct {
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

type PullRequestRequest struct {
	Owner string `json:"-"`
	Repo  string `json:"-"`
	Title string `json:"title"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Body  string `json:"body"`
	Draft bool   `json:"draft,omitempty"`
}

type PullRequestResponse struct {
	HTMLURL string `json:"html_url"`
	Number  int    `json:"number"`
}

func ParseIssueReference(text string) (*IssueReference, bool) {
	match := issueURLPattern.FindStringSubmatch(text)
	if len(match) != 5 {
		return nil, false
	}
	number, err := strconv.Atoi(match[4])
	if err != nil || number < 1 {
		return nil, false
	}
	isPullRequest := strings.EqualFold(match[3], "pull") || strings.EqualFold(match[3], "pulls")
	pathSegment := "issues"
	if isPullRequest {
		pathSegment = "pull"
	}
	ref := &IssueReference{
		Owner:         match[1],
		Repo:          match[2],
		Number:        number,
		URL:           fmt.Sprintf("https://github.com/%s/%s/%s/%d", match[1], match[2], pathSegment, number),
		IsPullRequest: isPullRequest,
	}
	return ref, true
}

func NewIssueBranchName(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		taskID = "task"
	}
	return fmt.Sprintf("Issue-%s", taskID)
}

func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*IssueResponse, error) {
	var issue IssueResponse
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	if err := c.doJSON(ctx, "GET", path, nil, &issue, 200); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*RepositoryResponse, error) {
	var repository RepositoryResponse
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	if err := c.doJSON(ctx, "GET", path, nil, &repository, 200); err != nil {
		return nil, err
	}
	return &repository, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, req PullRequestRequest) (*PullRequestResponse, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}
	if req.Title == "" || req.Head == "" || req.Base == "" {
		return nil, fmt.Errorf("title, head, and base are required")
	}
	var pr PullRequestResponse
	path := fmt.Sprintf("/repos/%s/%s/pulls", req.Owner, req.Repo)
	if err := c.doJSON(ctx, "POST", path, req, &pr, 201); err != nil {
		return nil, err
	}
	return &pr, nil
}

