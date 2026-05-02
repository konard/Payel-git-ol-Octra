package prompts

func Think(role, desc, task, gradeWeight string) string {
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

func ReviewWork(managerRole, workerRole, task, solution, files string) string {
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