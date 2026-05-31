package prompts

// ManagerThink — промпт для менеджера: спланировать команду воркеров.
// Теперь учитывает task_type, чтобы менеджер нанимал воркеров под нужный вид работы
// (ресёрчеров, авторов документов, дизайнеров презентаций), а не только разработчиков.
func ManagerThink(role, desc, task, gradeWeight, taskType string) string {
	switch taskType {
	case "research":
		return researchManagerThink(role, desc, task, gradeWeight)
	case "document":
		return documentManagerThink(role, desc, task, gradeWeight)
	case "presentation":
		return presentationManagerThink(role, desc, task, gradeWeight)
	default:
		return codeManagerThink(role, desc, task, gradeWeight)
	}
}

// codeManagerThink — исходный промпт для кодовых задач (команда разработчиков).
func codeManagerThink(role, desc, task, gradeWeight string) string {
	return `You are a project manager for a ` + role + ` team ONLY. Your team's focus: ` + desc + `

Task:
` + task + `

Task complexity (grade_weight): ` + gradeWeight + `/100

Use this formula:
- grade 1-20: 1 worker ONLY
- grade 21-40: 1-2 workers
- grade 41-60: 2 workers
- grade 61-80: 2-3 workers
- grade 81-100: max 2 workers

IMPORTANT: Only list workers that belong to YOUR team (` + role + `).

Reply ONLY with JSON:
{
  "worker_roles": [{"role": "developer", "description": "Backend developer"}]
}`
}

// researchManagerThink — менеджер ресёрча нанимает аналитиков, каждый со своим углом
// и способом поиска информации (как описано в issue: «Каждый агент использует
// различные способы поиска информации»).
func researchManagerThink(role, desc, task, gradeWeight string) string {
	return `You are a RESEARCH manager leading a team that investigates a topic from MULTIPLE angles.
Your team's focus: ` + desc + `

RESEARCH TOPIC:
` + task + `

Task complexity (grade_weight): ` + gradeWeight + `/100

Decide how many research analysts to hire and give EACH a DISTINCT search range / method,
so their findings complement (not duplicate) each other. As the manager you MUST define the
SEARCH RANGE for every analyst — the kind of material they should look for, expressed in terms of:
- links / web pages (official docs, reference sites),
- tags / keywords to query,
- news & current events,
- general information & background.

Examples of distinct ranges: official documentation & links, news & current events,
statistics & data, expert opinions, academic sources, opposing viewpoints, primary sources.

Use this formula for the number of analysts:
- grade 1-20: 1 analyst
- grade 21-40: 2 analysts
- grade 41-70: 3 analysts
- grade 71-100: 4 analysts

The "description" field MUST state the unique SEARCH RANGE (links/tags/news/information) and
method for that analyst, because the worker will run a real web search within that range.

Reply ONLY with JSON:
{
  "worker_roles": [{"role": "analyst", "description": "Search official documentation links and reference pages"}]
}`
}

// documentManagerThink — менеджер текстовых документов нанимает авторов, каждый
// отвечает за свою часть документа.
func documentManagerThink(role, desc, task, gradeWeight string) string {
	return `You are an EDITORIAL manager leading writers who produce a finished text document
(report, essay, coursework, abstract, article, table, etc.). Your team's focus: ` + desc + `

ASSIGNMENT:
` + task + `

Task complexity (grade_weight): ` + gradeWeight + `/100

Decide how many writers to hire. Each writer produces a complete, well-structured draft of the
document so the manager can later synthesize the best one. For short documents 1 writer is enough;
for longer/complex documents hire up to 2 so their drafts can be merged and fact-checked.

Use this formula:
- grade 1-40: 1 writer
- grade 41-100: 2 writers

The "description" field MUST describe the writer's focus or perspective on the document.

Reply ONLY with JSON:
{
  "worker_roles": [{"role": "writer", "description": "Write the full document with a clear structure"}]
}`
}

// presentationManagerThink — менеджер презентаций нанимает дизайнера(ов) слайдов.
func presentationManagerThink(role, desc, task, gradeWeight string) string {
	return `You are a PRESENTATION manager leading designers who build a slide deck (exported to PPTX).
Your team's focus: ` + desc + `

PRESENTATION TOPIC:
` + task + `

Task complexity (grade_weight): ` + gradeWeight + `/100

A presentation must be ONE coherent deck, so hire exactly 1 presentation designer who produces
the entire slide deck. Do NOT split a single deck across multiple workers.

Reply ONLY with JSON:
{
  "worker_roles": [{"role": "designer", "description": "Design the full slide deck for the topic"}]
}`
}

// ManagerReviewWork — промпт для менеджера: ревью работы воркера.
// Учитывает task_type: код ревьюится как код, документы/ресёрч/презентации —
// по качеству контента, полноте, достоверности и оформлению.
func ManagerReviewWork(managerRole, workerRole, task, solution, files, taskType string) string {
	switch taskType {
	case "research", "document", "presentation":
		return documentReviewWork(managerRole, workerRole, taskType, task, solution, files)
	default:
		return codeReviewWork(managerRole, workerRole, task, solution, files)
	}
}

// codeReviewWork — исходный код-ориентированный ревью.
func codeReviewWork(managerRole, workerRole, task, solution, files string) string {
	return `You are a ` + managerRole + ` manager reviewing work from a ` + workerRole + ` developer.

TASK:
` + task + `

PROPOSED SOLUTION:
` + solution + `

FILES:
` + files + `

Review the work:
1. Does it fulfill the task requirements?
2. Is the code quality acceptable?
3. Are there obvious bugs or missing pieces?
4. Does it integrate well with the team (if context provided)?

Reply ONLY with JSON:
{
  "approved": true/false,
  "feedback": "detailed feedback if not approved, or praise if approved"
}`
}

// documentReviewWork — ревью для ресёрча/документов/презентаций.
func documentReviewWork(managerRole, workerRole, taskType, task, solution, files string) string {
	criteria := `1. Does it fully cover the topic and fulfill the request?
2. Is it factually accurate and free of unsupported or contradictory claims?
3. Is it well-structured, clearly written and properly formatted in Markdown?
4. Are there placeholders, TODOs or missing sections that must be filled in?`
	if taskType == "presentation" {
		criteria = `1. Does the deck cover the topic with a logical slide flow (title → content → closing)?
2. Are the slides accurate, concise and free of placeholders?
3. Is the slide Markdown well-formed (one "# " deck title, "## " per slide, "- " bullets)?
4. Is the number of slides reasonable (roughly 6-12)?`
	}
	return `You are a ` + managerRole + ` manager reviewing a ` + taskType + ` produced by a ` + workerRole + `.

REQUEST:
` + task + `

WORKER SUMMARY:
` + solution + `

PRODUCED CONTENT:
` + files + `

Review the work as an editor (NOT as code):
` + criteria + `

Reply ONLY with JSON:
{
  "approved": true/false,
  "feedback": "detailed feedback if not approved, or praise if approved"
}`
}
