package prompts

func PlanFiles(role, desc, task, context, techStack string) string {
	return `You are a ` + role + ` developer. Role: ` + desc + `
Language: ` + techStack + `

TASK: ` + task + context + `

IMPORTANT: Use ONLY ` + techStack + ` language. NOT JavaScript, NOT TypeScript.
Create files appropriate for ` + techStack + ` (e.g., .go files for Go, .py for Python).
Return JSON ONLY:
{"files": ["path1.ext", "path2.ext", "path3.ext"]}`
}

func GenerateFile(filename, task, role, techStack string) string {
	return `Write the FULL content of file: ` + filename + `
Language: ` + techStack + `

TASK: ` + task + `
Role: ` + role + `

IMPORTANT: Write COMPLETE ` + techStack + ` code. No placeholders. No TODOs.
Use appropriate file extension (.go for Go, .py for Python, .js for JS).
Return the file content as PLAIN TEXT. NO JSON. NO markdown. Just the raw code.`
}

func GenerateCommands(role, desc, task, context string) string {
	return `You are a ` + role + ` developer. Role: ` + desc + `

TASK: ` + task + context + `

Based on the files created, provide bash commands to execute in the project root (mkdir, echo, etc.).
Return JSON ONLY: {"commands": ["cmd1", "cmd2"]}`
}

func Task(role, managerRole, task, context string) string {
	return `You are a ` + role + ` developer on a project team.

Your manager (` + managerRole + `) gave you this task:

` + task + `

` + context + `

Analyze the task and create a detailed plan. Write your analysis in a file called TASK.md.
Focus on:
1. What exactly needs to be built
2. Key components and their responsibilities
3. Technical decisions to make
4. Potential challenges`
}

func Review(role, feedback, code string) string {
	return `You are a ` + role + ` developer. Your previous work was reviewed.

FEEDBACK: ` + feedback + `

Your code:
` + code + `

Fix the issues identified in the feedback. Write the corrected file content as PLAIN TEXT.
NO JSON. NO markdown. Just the raw code.`
}