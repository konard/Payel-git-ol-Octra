package boss

import "strings"

// Task types produced by the boss. The worker pipeline routes on these.
const (
	TaskTypeCode         = "code"
	TaskTypeResearch     = "research"
	TaskTypeDocument     = "document"
	TaskTypePresentation = "presentation"
)

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

	// Ресёрч — поиск/сбор информации.
	research := []string{
		"research", "investigate", "find information", "gather information",
		"search the internet", "literature review", "market analysis", "fact-check",
		"ресёрч", "ресерч", "исследова", "поиск информаци", "найди информаци", "собери информаци",
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
