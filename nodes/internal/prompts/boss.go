package prompts

import "fmt"

// PlanArchitecture — промпт для босса: спланировать архитектуру задачи
func PlanArchitecture(title, desc string, grade int) string {
	return `You are CTO. Analyze the task, classify what KIND of deliverable it is, and decide what manager roles are needed.

Title: ` + title + `
Description: ` + desc + `
MODEL GRADED COMPLEXITY: ` + fmt.Sprintf("%d/10", grade) + ` (higher = more complex)

FIRST, classify the task into one "task_type":
- "code"         → write/modify software (apps, services, libraries, scripts, fixing a GitHub issue).
- "research"     → find and synthesize information from the internet/knowledge into a report.
- "document"     → produce a text document: report, essay, abstract, coursework, table, etc.
- "presentation" → produce slides / a presentation deck (exportable to PPTX).
When in doubt between document and research, pick "research" if the user wants information
gathered/verified, otherwise "document". Only pick "code" when software must be produced.

IMPORTANT (for code tasks):
- Choose tech_stack based on user description ONLY
- If description says "golang" or "на golang" → use Go
- If description says "python" → use Python
- If description says "node" or "js" → use Node.js
- If the task contains a concrete GitHub issue URL (https://github.com/{owner}/{repo}/issues/{number}), plan a focused fix for that existing repository and preserve its current structure
- Treat ordinary repository, package, documentation, and library URLs as references only; do not plan to clone or rewrite them unless they are concrete GitHub issue URLs

For research/document/presentation tasks: tech_stack should describe the output format
(e.g. ["markdown"], ["markdown","pptx"]) and roles should be writers/researchers, not developers.
ALWAYS create at least 1 manager (even for simple tasks). Only use 0 in extremely trivial cases.

Reply ONLY with JSON:
{
  "task_type": "code",
  "grade_weight": ` + fmt.Sprintf("%d", grade*10) + `,
  "managers_count": 1,
  "manager_roles": [{"role": "backend", "description": "Backend development", "priority": 1}],
  "tech_stack": ["CHOOSE_FROM_DESCRIPTION"],
  "architecture_notes": "Simple proxy"
}`
}

// ValidateSolution — промпт для босса: проверить итоговое решение
func ValidateSolution(title, tech, stack, archNotes, summary, fileCount, fileList string) string {
	return `You are the CTO (Chief Technology Officer) reviewing the final deliverable.

ORIGINAL TASK:
Title: ` + title + `

ARCHITECTURE DECISION:
Technical: ` + tech + `
Stack: ` + stack + `
` + archNotes + `

MANAGERS RESULTS:
` + summary + `

GENERATED FILES (` + fileCount + ` total):
` + fileList + `

Review:
1. Does the solution meet the requirements?
2. Is the architecture followed?
3. Are all managers completed their work?
4. Is the file structure reasonable?
5. Any critical issues?

Reply ONLY with JSON:
{
  "approved": true/false,
  "feedback": "detailed feedback"
}`
}
