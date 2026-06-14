package github

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// AnalyzeIssue — собирает полный контекст issue (паспорт) до запуска пайплайна.
// Делает несколько вызовов GitHub API: тело issue + метки + состояние,
// все комментарии и список открытых pull request'ов, относящихся к issue.
//
// Ошибки отдельных вызовов не фатальны — паспорт заполняется тем, что удалось
// получить, чтобы пайплайн мог продолжить работу даже при частичном ответе API.
func (c *Client) AnalyzeIssue(ctx context.Context, ref IssueReference) (*IssueInstruction, error) {
	if c == nil {
		return nil, fmt.Errorf("github client is nil")
	}

	instruction := &IssueInstruction{
		Owner:  ref.Owner,
		Repo:   ref.Repo,
		Number: ref.Number,
		URL:    ref.URL,
	}

	issue, err := c.GetIssue(ctx, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return nil, fmt.Errorf("fetch issue: %w", err)
	}
	instruction.Title = issue.Title
	instruction.Body = issue.Body
	instruction.State = issue.State
	if issue.HTMLURL != "" {
		instruction.URL = issue.HTMLURL
	}
	for _, label := range issue.Labels {
		if name := strings.TrimSpace(label.Name); name != "" {
			instruction.Labels = append(instruction.Labels, name)
		}
	}

	if repo, err := c.GetRepository(ctx, ref.Owner, ref.Repo); err != nil {
		log.Printf("AnalyzeIssue: failed to fetch repository %s/%s: %v", ref.Owner, ref.Repo, err)
	} else {
		instruction.DefaultBranch = repo.DefaultBranch
		instruction.RepositoryURL = repo.HTMLURL
	}

	if comments, err := c.GetIssueComments(ctx, ref.Owner, ref.Repo, ref.Number); err != nil {
		log.Printf("AnalyzeIssue: failed to fetch comments for %s: %v", ref.URL, err)
	} else {
		for _, comment := range comments {
			instruction.Comments = append(instruction.Comments, IssueComment{
				Author:    comment.User.Login,
				Body:      comment.Body,
				CreatedAt: comment.CreatedAt,
			})
		}
	}

	if prs, err := c.ListOpenPullRequests(ctx, ref.Owner, ref.Repo); err != nil {
		log.Printf("AnalyzeIssue: failed to list open PRs for %s/%s: %v", ref.Owner, ref.Repo, err)
	} else {
		for _, pr := range prs {
			if !pullRequestReferencesIssue(pr, ref) {
				continue
			}
			instruction.OpenPRs = append(instruction.OpenPRs, OpenPullRequest{
				Number: pr.Number,
				State:  pr.State,
				Title:  pr.Title,
				URL:    pr.HTMLURL,
			})
		}
	}

	return instruction, nil
}

// pullRequestReferencesIssue — эвристика: PR относится к issue, если его тело или
// заголовок упоминает "#N" или "owner/repo#N". GitHub не отдаёт прямую связь
// issue→PR, поэтому ищем по closing-ключевым словам и номеру.
func pullRequestReferencesIssue(pr PullRequestListItem, ref IssueReference) bool {
	needleShort := fmt.Sprintf("#%d", ref.Number)
	needleFull := fmt.Sprintf("%s/%s#%d", ref.Owner, ref.Repo, ref.Number)
	haystack := strings.ToLower(pr.Title + "\n" + pr.Body)
	if strings.Contains(haystack, strings.ToLower(needleFull)) {
		return true
	}
	// "#N" должен граничить с не-цифрой, чтобы #4 не матчил #42.
	idx := 0
	for {
		pos := strings.Index(haystack[idx:], strings.ToLower(needleShort))
		if pos < 0 {
			return false
		}
		end := idx + pos + len(needleShort)
		if end >= len(haystack) || !isDigit(haystack[end]) {
			return true
		}
		idx = end
	}
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// trivialTitleKeywords — заголовки, явно указывающие на тривиальный фикс.
var trivialTitleKeywords = []string{
	"typo", "fix typo", "rename", "spelling", "grammar",
	"опечатк", "переименова", "спасибо", "please add", "wording",
}

// docOnlyExtensions / docOnlyDirs — изменения, затрагивающие только документацию.
var docOnlyMarkers = []string{
	".md", ".mdx", ".rst", ".txt", "readme", "changelog", "docs/", "license",
}

// IsTrivialIssue — детектор тривиальных issue (план фикса, пункт 7). Такие issue
// не нужно гонять через весь конвейер Boss→Manager→Worker→Review→Validate —
// достаточно одного целевого AI-запроса.
//
// Критерии (issue считается тривиальным, если выполнены ВСЕ применимые условия):
//   - тело issue короче 500 символов;
//   - нет меток enhancement / feature / epic;
//   - либо заголовок содержит тривиальный ключевик (typo, rename, ...),
//     либо контекст указывает только на документацию.
func IsTrivialIssue(in *IssueInstruction) bool {
	if in == nil {
		return false
	}
	if len(strings.TrimSpace(in.Body)) >= 500 {
		return false
	}
	for _, blocking := range []string{"enhancement", "feature", "epic"} {
		if in.HasLabel(blocking) {
			return false
		}
	}

	title := strings.ToLower(in.Title)
	for _, kw := range trivialTitleKeywords {
		if strings.Contains(title, kw) {
			return true
		}
	}

	text := strings.ToLower(in.Title + "\n" + in.Body)
	if mentionsOnlyDocs(text) {
		return true
	}
	return false
}

// mentionsOnlyDocs — true, если текст упоминает документационные артефакты и не
// упоминает явных кодовых маркеров (расширения исходников). Грубая эвристика.
func mentionsOnlyDocs(text string) bool {
	mentionsDocs := false
	for _, marker := range docOnlyMarkers {
		if strings.Contains(text, marker) {
			mentionsDocs = true
			break
		}
	}
	if !mentionsDocs {
		return false
	}
	codeMarkers := []string{".go", ".py", ".js", ".ts", ".rs", ".java", ".php", ".rb", ".c", ".cpp"}
	for _, marker := range codeMarkers {
		if strings.Contains(text, marker) {
			return false
		}
	}
	return true
}
