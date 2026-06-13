package prompts

import (
	"fmt"
	"strings"

	instcore "orchestrator/pkg/instralutions/core"
)

// WorkerPlanFiles — промпт для воркера: спланировать список файлов.
// skill — опциональный блок экспертных рекомендаций из системных скиллов
// (skills.Guidance), подобранный под роль/стек воркера. Если пусто — не влияет.
func WorkerPlanFiles(role, desc, task, context, techStack, skill string) string {
	cheat := ToolCheatSheet(techStack)
	structureLine := ""
	if cheat != "" {
		structureLine = "\nPROJECT STRUCTURE for " + techStack + " (typical files):" + cheat
	}
	extHint := ""
	if ext := LangExtension(techStack); ext != "" {
		extHint = " (e.g. ." + ext + ")"
	}

	return `You are a ` + role + ` developer. Role: ` + desc + `
Language: ` + techStack + `

TASK: ` + task + context + skill + structureLine + `

SCOPE FIDELITY: plan the MINIMUM set of files that builds EXACTLY what the task asks for — nothing more.
Do NOT add files for features the user did not request (no auth, database, user management, extra
services, configs). If the task is minimal ("mini", "simple", "small", "basic", "мини", "минимальный",
"простой"), a single source file plus the language manifest is usually enough.

IMPORTANT: Write code in ` + techStack + ` ONLY. Use the standard source file extension for ` + techStack + extHint + `
— do NOT use another language's extension.
Use the SPECIFIC frameworks/libraries named in the task (e.g. if it says "Express", use Express — not the bare stdlib).
Return JSON ONLY:
{"files": ["path1.ext", "path2.ext", "path3.ext"]}`
}

// WorkerGenerateFile — промпт для воркера: сгенерировать содержимое одного файла.
// skill — опциональный блок экспертных рекомендаций из системных скиллов.
func WorkerGenerateFile(filename, task, role, techStack, skill string) string {
	return `Write the FULL content of file: ` + filename + `
Language: ` + techStack + `

TASK: ` + task + `
Role: ` + role + skill + `

IMPORTANT: Write COMPLETE ` + techStack + ` code. No placeholders. No TODOs.
Use the standard file extension for ` + techStack + `.
Return the file content as PLAIN TEXT. NO JSON. NO markdown. Just the raw code.`
}

// WorkerMultiPassCode — промпт single-pass генерации: все файлы за один LLM-запрос.
// Раньше этот промпт жил инлайном в worker и был жёстко заточен под Go: подставлял
// сам techStack как расширение (".%s" → ".nodejs") и содержал хардкод
// "NOT JavaScript, NOT TypeScript". Из-за этого задача про Express.js сохранялась
// в main.go. Теперь язык/расширение/точку входа берём из единой карты Lang*.
func WorkerMultiPassCode(role, description, task, context, techStack string) string {
	langDisplay := LangDisplay(techStack)
	ext := LangExtension(techStack)
	if ext == "" {
		ext = "txt"
	}
	mainFile := LangMainFile(techStack)

	return fmt.Sprintf(`You are a %s developer. Role: %s
Language: %s

TASK: %s%s

SCOPE FIDELITY: Build EXACTLY what the task asks for — nothing more.
Do NOT add files for features the user did not request (no auth, database, user management, configs).
If the task is simple or basic, use the MINIMUM number of files (1-2 source files + manifest).

IMPORTANT:
- Write all code in %s. Use the SPECIFIC frameworks/libraries the task names (e.g. if the task says
  "Express", use Express — do not substitute the language's bare standard library).
- Use the correct source file extension for this language: .%s — do NOT rename files to another
  language's extension (e.g. never put JavaScript into a .go file).

AFTER files, provide COMMANDS to execute in the project (mkdir, echo, etc.).

RETURN FORMAT (STRICT - follow exactly):
=== FILE: %s ===
<complete code - no placeholders, no TODOs>
=== FILE: path/to/another.%s ===
<complete code - no placeholders, no TODOs>
=== COMMANDS ===
mkdir -p dir
echo 'content' > file.txt
# other bash commands

RULES:
1. Each file MUST start with "=== FILE: <path> ===" on its own line
2. File paths are relative to project root (e.g., %s) - DO NOT include project name in path
3. File content MUST be complete code - no placeholders, no TODOs, no "implement later"
4. Use proper imports and exports
5. Keep code compact but functional (300-500 lines max per file)
6. Do NOT include markdown code fences around the entire response
7. If you can't create a file, skip it and move to the next one
8. COMMANDS: List bash commands to run in project root, one per line
9. Return ONLY the files and commands, no explanations`,
		role, description, langDisplay,
		task, context,
		langDisplay,
		ext,
		mainFile, ext,
		mainFile)
}

// WorkerGenerateCommands — промпт для воркера: bash-команды для инициализации проекта
func WorkerGenerateCommands(role, desc, task, context string) string {
	return `You are a ` + role + ` developer. Role: ` + desc + `

TASK: ` + task + context + `

Based on the files created, provide bash commands to execute in the project root (mkdir, echo, etc.).
Return JSON ONLY: {"commands": ["cmd1", "cmd2"]}`
}

// WorkerTask — промпт для воркера: написать TASK.md
func WorkerTask(role, managerRole, task, context string) string {
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

// WorkerToolCommands — промпт для воркера: спланировать scaffolding через install-флаги.
// AI возвращает install-флаги (выбирает из известных установщиков), а не полные команды.
// Это надёжнее: установщики пишутся Go-разработчиком, AI не ошибается в синтаксисе тулов.
// После install AI может добавить post-generation команды (сборка, миграции и т.д.).
func WorkerToolCommands(role, desc, task, context, techStack string) string {
	// Собираем список доступных install-флагов из instralutions registry
	available := strings.Builder{}
	available.WriteString("\n\nAVAILABLE INSTALL FLAGS (choose from these):\n")
	for _, s := range instcore.All() {
		available.WriteString(fmt.Sprintf("  - %s (requires: %s)\n", s.Name, strings.Join(s.Requires, ", ")))
	}

	return `You are a ` + role + ` developer using the ` + techStack + ` toolchain.

ROLE: ` + desc + `
TASK: ` + task + context + `

You have a BLANK project directory. Your job is to scaffold the project using
REAL tools — NOT by writing code manually. The tools will be available via
` + "`nix develop`" + ` (they are already in the environment).

HOW IT WORKS:
1. Pick the RIGHT install flag(s) from the AVAILABLE INSTALL FLAGS list below that match the task.
   For example: ["java", "spring"] for a Spring Boot project.
2. Optionally add post-generation commands (e.g. "mvn package") in "commands".
   Do NOT create scaffolding commands yourself — use install flags for that.

RULES:
- Use install flags for ALL scaffolding (project creation, dependency installation).
- Post-generation commands (build, test, run) go in "commands" array.
- Keep it MINIMAL — only what the task asks for. No extra features.
- If the task is simple enough for a single source file, omit install and use echo/cat in commands.

Return JSON ONLY:
{"install": ["flag1", "flag2"], "commands": ["cmd1", "cmd2"]}` + available.String()
}

// WorkerReview — промпт для воркера: переписать файл с учётом фидбэка
func WorkerReview(role, feedback, code string) string {
	return `You are a ` + role + ` developer. Your previous work was reviewed.

FEEDBACK: ` + feedback + `

Your code:
` + code + `

Fix the issues identified in the feedback. Write the corrected file content as PLAIN TEXT.
NO JSON. NO markdown. Just the raw code.`
}

