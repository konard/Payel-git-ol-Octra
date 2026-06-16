package github

import "strings"

// IssueInstruction — структурированный «паспорт» GitHub issue, который
// собирается ДО запуска пайплайна и передаётся от Boss к Manager к Worker без
// пересборки и дрейфа (план фикса, пункты 1 и 2).
//
// В отличие от старого подхода, когда тело issue обрезалось до 4000 символов и
// приклеивалось к Description, паспорт хранит полный контекст: тело целиком,
// комментарии, метки, состояние и список уже открытых pull request'ов.
type IssueInstruction struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	State  string `json:"state"`

	Title string `json:"title"`
	Body  string `json:"body"`

	Labels   []string          `json:"labels,omitempty"`
	Comments []IssueComment    `json:"comments,omitempty"`
	OpenPRs  []OpenPullRequest `json:"open_prs,omitempty"`

	DefaultBranch string `json:"default_branch,omitempty"`
	RepositoryURL string `json:"repository_url,omitempty"`
}

// IssueComment — один комментарий issue внутри паспорта.
type IssueComment struct {
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// OpenPullRequest — открытый pull request, потенциально относящийся к issue.
type OpenPullRequest struct {
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// ForkResponse — ответ GitHub API на создание форка.
type ForkResponse struct {
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// HasLabel — true, если у issue есть метка с таким именем (регистронезависимо).
func (in *IssueInstruction) HasLabel(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, l := range in.Labels {
		if strings.ToLower(strings.TrimSpace(l)) == name {
			return true
		}
	}
	return false
}
