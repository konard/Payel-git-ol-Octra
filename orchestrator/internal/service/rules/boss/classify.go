package boss

import "strings"

// Task types produced by the boss. The worker pipeline routes on these.
const (
	TaskTypeCode         = "code"
	TaskTypeResearch     = "research"
	TaskTypeDocument     = "document"
	TaskTypePresentation = "presentation"
	// TaskTypeGitHub — задача, привязанная к конкретному GitHub issue/PR. Ведёт
	// себя как код (тот же конвейер и публикация), но Boss получает отдельный
	// «паспорт» issue в Meta["github_context"] вместо засорённого Description
	// (план фикса, пункт 1).
	TaskTypeGitHub = "github"
)

// isCodeLikeTask — типы задач, которые проходят кодовый конвейер и публикуются в
// GitHub: обычный код и github-issue задачи.
func isCodeLikeTask(taskType string) bool {
	return taskType == "" || taskType == TaskTypeCode || taskType == TaskTypeGitHub
}

// clampToSingleManager — сводит решение к одному менеджеру с одной ролью. Нужно
// для тривиальных issue (план фикса, пункт 7): один целевой проход вместо всего
// конвейера. Сохраняет первую AI-выбранную роль, если она есть.
func clampToSingleManager(decision *DecisionResult) {
	if decision == nil {
		return
	}
	decision.ManagersCount = 1
	if len(decision.ManagerRoles) > 1 {
		decision.ManagerRoles = decision.ManagerRoles[:1]
	}
	if len(decision.ManagerWorkflows) > 1 {
		decision.ManagerWorkflows = decision.ManagerWorkflows[:1]
	}
}

// classifyTaskType — детерминированный fallback-классификатор по ключевым словам.
// Используется, когда AI не вернул task_type (или JSON не распарсился). Поддерживает
// русский и английский, так как пользователи Octra пишут на обоих языках.
func classifyTaskType(title, description string) string {
	text := strings.ToLower(title + "\n" + description)

	// Презентации — самый специфичный тип, проверяем первым.
	presentation := []string{
		"presentation", "slide deck", "slides", "slideshow", "pptx", "powerpoint", "keynote",
		"презентаци", "слайд", "доклад со слайдами",
	}
	if containsAny(text, presentation) {
		return TaskTypePresentation
	}

	// Ресёрч — поиск/сбор информации в вебе.
	research := []string{
		"research", "investigate", "find information", "gather information",
		"search the internet", "search the web", "web search", "search online",
		"look up", "google", "literature review", "market analysis", "fact-check",
		"ресёрч", "ресерч", "исследова", "поиск информаци", "найди информаци", "собери информаци",
		"поиск в интернете", "найди в интернете", "погугли", "поищи в интернете", "поиск в сети",
	}
	if containsAny(text, research) {
		return TaskTypeResearch
	}

	// Текстовые документы.
	document := []string{
		"report", "essay", "abstract", "coursework", "term paper", "article", "whitepaper",
		"summary", "document", "table", "spreadsheet", "memo",
		"отчёт", "отчет", "реферат", "курсов", "сочинени", "статья", "документ", "таблиц", "доклад",
	}
	if containsAny(text, document) {
		return TaskTypeDocument
	}

	// Явные сигналы кода.
	code := []string{
		"code", "implement", "bug", "refactor", "api", "function", "library", "service",
		"github.com", "программ", "напиши код", "приложени", "сервис", "функци", "багу",
	}
	if containsAny(text, code) {
		return TaskTypeCode
	}

	return TaskTypeCode
}

// detectTechStack — детектит tech stack из заголовка/описания задачи,
// когда AI не смог определить (stack=[]).
func detectTechStack(title, description string) []string {
	text := strings.ToLower(title + " " + description)

	type stackMatch struct {
		keywords []string
		stack    string
	}
	known := []stackMatch{
		{[]string{"python", "django", "flask", "fastapi", "pytest"}, "python"},
		{[]string{"php", "laravel", "symfony", "composer"}, "php"},
		{[]string{"node", "nodejs", "express", "nestjs", "typescript", "javascript", "react", "vue", "angular", "npm", "yarn"}, "nodejs"},
		{[]string{"golang", "go ", " go-", "go "}, "go"},
		{[]string{"rust", "cargo"}, "rust"},
		{[]string{"java", "maven", "gradle", "spring", "kotlin"}, "java"},
		{[]string{"c#", "csharp", "dotnet", "asp.net", ".net"}, "dotnet"},
		{[]string{"c++", "cpp", "cmake"}, "cpp"},
		{[]string{"ruby", "rails", "gem"}, "ruby"},
		{[]string{"flutter", "dart"}, "flutter"},
		{[]string{"swift", "xcode"}, "swift"},
		{[]string{"elixir", "phoenix"}, "elixir"},
		{[]string{"haskell", "cabal"}, "haskell"},
		{[]string{"scala", "sbt"}, "scala"},
		{[]string{"r ", "rstats"}, "r"},
		{[]string{"zig"}, "zig"},
	}

	matched := make(map[string]bool)
	var result []string
	for _, s := range known {
		for _, kw := range s.keywords {
			if strings.Contains(text, kw) {
				if !matched[s.stack] {
					matched[s.stack] = true
					result = append(result, s.stack)
				}
				break
			}
		}
	}
	return result
}

// normalizeTaskType — приводит значение к одному из поддерживаемых типов.
func normalizeTaskType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case TaskTypeResearch:
		return TaskTypeResearch
	case TaskTypeDocument, "text", "report":
		return TaskTypeDocument
	case TaskTypePresentation, "slides", "pptx", "powerpoint":
		return TaskTypePresentation
	case TaskTypeCode, "software", "dev", "development":
		return TaskTypeCode
	case TaskTypeGitHub, "issue", "pull_request", "pr":
		return TaskTypeGitHub
	default:
		return ""
	}
}

func containsAny(text string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(text, n) {
			return true
		}
	}
	return false
}
