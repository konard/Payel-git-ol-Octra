package prompts

func PlanArchitecture(title, desc string) string {
	return `You are CTO. Analyze the task and decide what manager roles are needed.

Title: ` + title + `
Description: ` + desc + `

Grade the task complexity from 1 to 100:
- 1-20: Simple (single file, CLI, mini proxy)
- 21-40: Medium (API with few endpoints)
- 41-60: Complex (full system)

IMPORTANT:
- Choose tech_stack based on user description ONLY
- If description says "golang" or "на golang" → use Go
- If description says "python" → use Python
- If description says "node" or "js" → use Node.js
- Create ONLY the managers that are actually needed

Reply ONLY with JSON:
{
  "grade_weight": 10,
  "managers_count": 1,
  "manager_roles": [{"role": "backend", "description": "Backend development", "priority": 1}],
  "tech_stack": ["CHOOSE_FROM_DESCRIPTION"],
  "architecture_notes": "Simple proxy"
}`
}

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