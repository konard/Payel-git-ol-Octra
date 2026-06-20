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
// сразу выдаёт корректный результат.
//
// Для кодовых задач действует принцип минимальности («напиши hello world на
// python» → один файл, одна строка). Для фактуальных запросов с результатами
// веб-поиска нода пишет развёрнутый ответ с цитированием источников.
//
// Формат ответа — строгий JSON {"files": {...}}, тот же контракт о файлах, что и
// у обычного воркера, поэтому остальной конвейер (запись на диск, публикация,
// вкладка Solution) работает без изменений.
func UniversalNode(title, description, taskType, techStack, searchBlock string) string {
	searchSection := ""
	isSearchTask := false
	if s := strings.TrimSpace(searchBlock); s != "" {
		isSearchTask = true
		searchSection = `

WEB SEARCH RESULTS (use these REAL sources to answer — do not invent facts; cite them when relevant):
` + s + `
Base your answer on the search results above, citing sources by number [1], [2], etc.
If the search results don't cover something, say so instead of guessing.
`
	}

	var answerFormat string
	if isSearchTask {
		answerFormat = `- Write "solution/answer.md" as a complete, well-structured Markdown document.
- Include a clear answer to the question, supporting details, and citations [1], [2], etc.
- Use sections (headings, lists, paragraphs) to organize the information.
- The more relevant detail you include from the search results, the better.
- Do NOT add extra files or code — just the answer document.`
	} else {
		answerFormat = `- For code tasks: use the correct source-file extension for the language.
- For questions or research: return a SINGLE Markdown file at "solution/answer.md".`
	}

	return fmt.Sprintf(`You are the UNIVERSAL NODE — a single solver.

Title: %s
Task: %s
Tech stack / language hint: %s
%s
Your job is to deliver the complete, correct answer in ONE step.

DECIDE THE OUTPUT FORMAT YOURSELF:
- If the user asks for code, a script, a function, or implementation — write source files (.py, .js, .go, etc.).
- If the user asks a question, research, or factual information — write solution/answer.md.
- If web search results are provided above — answer factually and comprehensively based on those sources.

GUIDELINES:
%s
For code tasks: keep it minimal. A "hello world" is ONE file — not a project.
Do NOT invent extra files, configuration, dependencies, abstractions, error
handling, tests, READMEs, or features the user never requested.

REQUIRED — dependencies flag:
When the code relies on ANY external library, framework, or package NOT in the language's
standard library, set "dependencies": true. When the task uses only the standard library
(or no library at all), set "dependencies": false. This controls whether the build system
pins third-party package versions.

Reply with STRICT JSON ONLY (no markdown fences, no commentary):
{"files": {"relative/path.ext": "full file content"}, "dependencies": false}`,
		title, description, techStack, searchSection, answerFormat)
}
