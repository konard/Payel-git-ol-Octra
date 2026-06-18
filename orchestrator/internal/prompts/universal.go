package prompts

import "fmt"

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
func UniversalNode(title, description, taskType, techStack string) string {
	if taskType == "" {
		taskType = "code"
	}
	deliverable := codeDeliverableRules(techStack)
	if taskType != "code" && taskType != "github" {
		deliverable = textDeliverableRules(taskType)
	}

	return fmt.Sprintf(`You are the UNIVERSAL NODE — a single solver for a task that has already been judged TRIVIAL.

Title: %s
Task: %s
Deliverable type: %s
Tech stack / output format: %s

Your job is to deliver the complete, correct answer in ONE step. Do it like a fast,
no-nonsense expert who respects the user's time.

THE GOLDEN RULE — DO NOT OVER-ENGINEER:
- Produce EXACTLY what was asked, and nothing more. The smallest correct result wins.
- A "hello world" is ONE file with ONE statement — not a project. A math or logic
  question is a direct answer — not a framework.
- Do NOT invent extra files, configuration, dependencies, abstractions, error
  handling, tests, READMEs, or features the user never requested.
- Do NOT turn a small request into a bigger project. "a script" stays a script.

%s

REQUIRED — dependencies flag:
When the code relies on ANY external library, framework, or package NOT in the language's
standard library, set "dependencies": true. When the task uses only the standard library
(or no library at all), set "dependencies": false. This controls whether the build system
pins third-party package versions.

Reply with STRICT JSON ONLY (no markdown fences, no commentary):
{"files": {"relative/path.ext": "full file content"}, "dependencies": false}`,
		title, description, taskType, techStack, deliverable)
}

func codeDeliverableRules(techStack string) string {
	lang := techStack
	if lang == "" {
		lang = "the most fitting popular language for the request"
	}
	return `DELIVERABLE (code):
- Return the minimal set of source files that fully satisfy the request, written in ` + lang + `.
- Prefer a SINGLE file using the standard library and zero dependencies unless the task clearly needs more.
- File content must be complete, runnable code — no placeholders, no TODOs.
- Use the correct source-file extension for the language (never put one language's code in another's extension).`
}

func textDeliverableRules(taskType string) string {
	return `DELIVERABLE (` + taskType + `):
- Return a SINGLE Markdown file at "solution/answer.md" containing the complete, finished answer.
- For math/logic questions, give the direct result first, then a short justification if useful.
- Keep it focused and free of filler — answer the actual question, nothing more.`
}
