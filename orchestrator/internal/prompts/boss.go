package prompts

import (
	"fmt"

	"orchestrator/internal/skills"
)

// PlanArchitecture — промпт для босса: спланировать архитектуру задачи.
// Если skillCategories непустое (comma-separated), в промпт попадают только
// фрагменты из указанных категорий. Если категории не заданы, каталог скиллов
// НЕ добавляется вовсе: раньше сюда падал весь каталог даже для тривиальных задач
// ("hello world"), зашумляя промпт нерелевантными специальностями (issue #79).
// Каталог теперь opt-in — он нужен только когда пользователь явно запросил скиллы.
func PlanArchitecture(title, desc string, grade int, skillCategories string) string {
	skillsSection := ""
	if skillCategories != "" {
		areas := skills.ParseCategories(skillCategories)
		warehouse := skills.NewWarehouse()
		if catalog := warehouse.CatalogFiltered(areas...); catalog != "" {
			skillsSection = fmt.Sprintf(`
The user requested these skill categories: %s

Available expert fragments in those categories:
%s`, skillCategories, catalog)
		}
	}
	return `You are CTO. Analyze the task, classify what KIND of deliverable it is, and decide what manager roles are needed.

Title: ` + title + `
Description: ` + desc + `
MODEL GRADED COMPLEXITY: ` + fmt.Sprintf("%d/10", grade) + ` (higher = more complex)` + skillsSection + `

FIRST, classify the task into one "task_type":
- "code"         → write/modify software (apps, services, libraries, scripts, fixing a GitHub issue).
- "research"     → find and synthesize information from the internet/knowledge into a report.
- "document"     → produce a text document: report, essay, abstract, coursework, table, etc.
- "presentation" → produce slides / a presentation deck (exportable to PPTX).
When in doubt between document and research, pick "research" if the user wants information
gathered/verified, otherwise "document". Only pick "code" when software must be produced.

SCOPE FIDELITY — THE MOST IMPORTANT RULE:
Build EXACTLY what the user asked for — the right KIND of deliverable, and nothing more.
- The deliverable type must literally match the request. "a proxy" → a proxy. "a CLI tool" → a CLI tool.
  Do NOT silently turn a small, specific request into a different, bigger project (e.g. a "proxy"
  must NOT become a REST API with auth/database/users).
- Do NOT add features, layers, dependencies or infrastructure the user did not ask for. No auth/JWT,
  database, user management, CORS, rate limiting, Docker, etc. UNLESS the task explicitly requests them.
- Match the architecture complexity to the literal request, not to the complexity grade. The grade only
  hints at effort; it must NEVER inflate scope beyond what was asked.
- Scope-limiting words mean keep it minimal — prefer the smallest reasonable structure (often a single
  file with the standard library and zero dependencies):
  English: "mini", "minimal", "simple", "small", "basic", "tiny", "just", "only", "quick".
  Russian: "мини", "минимальный", "простой", "небольшой", "базовый", "только", "просто".
- When the task is minimal, plan exactly 1 manager with 1 focused role; put the scope limit in
  architecture_notes (e.g. "Minimal HTTP/HTTPS proxy, single file, stdlib only — no extra features").

IMPORTANT (for code tasks):
- Choose tech_stack based on user description ONLY (1 item, the primary language)
- If description explicitly names a language — use that language. Examples:
  - "golang" / "на golang" / "go" → ["go"]
  - "python" / "django" / "flask" → ["python"]
  - "node" / "js" / "javascript" / "typescript" / "express" / "react" → ["nodejs"]
  - "php" / "laravel" / "symfony" / "на php" / "php сервер" → ["php"]
  - "rust" / "cargo" → ["rust"]
  - "java" / "spring" / "maven" → ["java"]
  - "c#" / "csharp" / "dotnet" / ".net" → ["dotnet"]
  - "c++" / "cpp" / "cmake" → ["cpp"]
  - "ruby" / "rails" → ["ruby"]
  - "flutter" / "dart" → ["flutter"]
  - "kotlin" → ["kotlin"]
  - "swift" → ["swift"]
  - "elixir" / "phoenix" → ["elixir"]
  - "haskell" / "cabal" → ["haskell"]
  - "scala" / "sbt" → ["scala"]
  - "zig" → ["zig"]
  - "r" → ["r"]
- If NO language is mentioned at all, fall back to the most popular choice that fits the task description
- If the task contains a concrete GitHub issue URL (https://github.com/{owner}/{repo}/issues/{number}), plan a focused fix for that existing repository and preserve its current structure
- Treat ordinary repository, package, documentation, and library URLs as references only; do not plan to clone or rewrite them unless they are concrete GitHub issue URLs

For research/document/presentation tasks: tech_stack should describe the output format
(e.g. ["markdown"], ["markdown","pptx"]) and roles should be writers/researchers, not developers.

For "research" tasks (web search): EXPAND and clarify the user's request into a precise topic,
then create one or more SEARCH MANAGERS. Each search manager leads analysts that run a real
web search; in "description" state the slice of the topic that manager owns so their teams
cover different SEARCH RANGES (links, tags/keywords, news, general information).

ALWAYS create at least 1 manager (even for simple tasks). Only use 0 in extremely trivial cases.

Reply ONLY with JSON:
{
  "task_type": "code",
  "managers_count": 1,
  "manager_roles": [{"role": "backend", "description": "Backend development", "priority": 1}],
  "tech_stack": ["CHOOSE_FROM_DESCRIPTION"],
  "architecture_notes": "Simple proxy"
}
`

}

// ValidateSolution — промпт для босса: проверить итоговое решение.
// Учитывает task_type: код проверяется как код, а ресёрч/документы/презентации —
// по полноте, достоверности и оформлению результата.
func ValidateSolution(title, tech, stack, archNotes, summary, fileCount, fileList, taskType string) string {
	checklist := `1. Does the solution meet the requirements?
2. Is the architecture followed?
3. Are all managers completed their work?
4. Is the file structure reasonable?
5. Any critical issues?`
	role := "the CTO (Chief Technology Officer)"
	switch taskType {
	case "research", "document", "presentation":
		role = "the editor-in-chief"
		checklist = `1. Does the deliverable fully answer the user's request?
2. Is the content accurate, well-structured and free of placeholders?
3. Did the managers merge and fact-check the workers' results?
4. Is it delivered in the right format (Markdown, plus PPTX for presentations)?
5. Any critical gaps or quality issues?`
	}
	return `You are ` + role + ` reviewing the final deliverable.

ORIGINAL TASK:
Title: ` + title + `

PLAN / DECISION:
Technical: ` + tech + `
Output format: ` + stack + `
` + archNotes + `

MANAGERS RESULTS:
` + summary + `

PRODUCED FILES (` + fileCount + ` total):
` + fileList + `

Review:
` + checklist + `

Reply ONLY with JSON:
{
  "approved": true/false,
  "feedback": "detailed feedback"
}`
}
