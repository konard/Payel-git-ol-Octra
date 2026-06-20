package prompts

import (
	"fmt"
	"strings"
)

// UniversalNode — промпт «универсальной ноды» (issue #91).
//
// Когда модель оценивает задачу как тривиальную, нет смысла запускать
// boss-планирование и весь конвейер Boss → Manager → Worker. Универсальная нода
// — это один AI-воркер с менее ограниченными возможностями: она читает задачу и
// сразу выдаёт минимальный корректный результат. Главное правило — НЕ переусложнять:
// «напиши hello world на python» должно вернуть один файл с одной строкой, а не
// дерево файлов с непонятной логикой.
//
// Формат ответа — строгий JSON {"files": {...}}, тот же контракт о файлах, что и
// у обычного воркера, поэтому остальной конвейер (запись на диск, публикация,
// вкладка Solution) работает без изменений.
func UniversalNode(title, description, taskType, techStack, searchBlock string) string {
	searchSection := ""
	if s := strings.TrimSpace(searchBlock); s != "" {
		searchSection = `

WEB SEARCH RESULTS (use these REAL sources to answer — do not invent facts; cite them when relevant):
` + s + `
Base your answer on the search results above. If they don't cover something, say so instead of guessing.
`
	}

	sourceFilesHint := ""
	if searchBlock == "" {
		sourceFilesHint = `- For code tasks: use the correct source-file extension for the language.
- For questions or research: return a SINGLE Markdown file at "solution/answer.md".`
	} else {
		sourceFilesHint = `- Return a SINGLE Markdown file at "solution/answer.md" containing the complete, finished answer with citations.`
	}

	return fmt.Sprintf(`You are the UNIVERSAL NODE — a single solver for a task that has already been judged TRIVIAL.

Title: %s
Task: %s
Tech stack / language hint: %s
%s
Your job is to deliver the complete, correct answer in ONE step. Do it like a fast,
no-nonsense expert who respects the user's time.

DECIDE THE OUTPUT FORMAT YOURSELF based on what the task asks:
- If the user asks for code, a script, a function, or implementation — write source files (.py, .js, .go, etc.).
- If the user asks a question, research, or factual information — write solution/answer.md.
- If web search results are provided above — answer factually based on those sources.

THE GOLDEN RULE — DO NOT OVER-ENGINEER:
- Produce EXACTLY what was asked, and nothing more. The smallest correct result wins.
- A "hello world" is ONE file — not a project. A question is a direct answer — not a framework.
- Do NOT invent extra files, configuration, dependencies, abstractions, error
  handling, tests, READMEs, or features the user never requested.
- %s

REQUIRED — dependencies flag:
When the code relies on ANY external library, framework, or package NOT in the language's
standard library, set "dependencies": true. When the task uses only the standard library
(or no library at all), set "dependencies": false. This controls whether the build system
pins third-party package versions.

Reply with STRICT JSON ONLY (no markdown fences, no commentary):
{"files": {"relative/path.ext": "full file content"}, "dependencies": false}`,
		title, description, techStack, searchSection, sourceFilesHint)
}
