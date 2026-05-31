package prompts

// Этот файл содержит промпты для не-кодовых задач: ресёрч, текстовые документы
// (отчёты, рефераты, курсовые) и презентации. Они переписаны так, чтобы система
// Воркеры → Менеджер → Босс выдавала результат в формате Markdown (а презентации —
// дополнительно в PPTX), а не исходный код.

// ResearchWorker — промпт для воркера-ресёрчера.
// Каждый воркер исследует тему под своим углом и своим способом поиска.
func ResearchWorker(role, angle, topic, context string) string {
	return `You are a research analyst on a multi-agent research team. Your role: ` + role + `.
Your search angle / method: ` + angle + `

RESEARCH TOPIC:
` + topic + context + `

Investigate the topic from YOUR angle only. Be specific, factual and cite concrete
facts, figures, dates and named sources whenever possible. Distinguish clearly between
established facts and your own inferences.

Return a well-structured Markdown report with:
- A short "## Key findings" section (3-7 bullet points).
- A "## Details" section with supporting evidence.
- A "## Sources" section listing the sources or reasoning you relied on.

Return ONLY GitHub-Flavored Markdown. No code fences around the whole answer.`
}

// DocumentWorker — промпт для воркера, пишущего текстовый документ.
// docType: "report" | "essay" | "coursework" | "abstract" | "table" | "document".
func DocumentWorker(role, docType, topic, context string) string {
	return `You are a professional writer producing a ` + docType + `. Your role: ` + role + `.

ASSIGNMENT:
` + topic + context + `

Write a complete, well-structured ` + docType + ` in GitHub-Flavored Markdown:
- Start with a single H1 title.
- Use clear section headings (##, ###), short paragraphs and lists.
- Use Markdown tables where data is naturally tabular.
- Keep a professional, coherent tone suitable for small businesses and students.
- Do NOT leave placeholders or TODOs — deliver finished, usable content.

Return ONLY the Markdown document. No commentary, no code fences around the whole answer.`
}

// PresentationWorker — промпт для воркера, готовящего презентацию.
// Воркер возвращает слайды в простом Markdown-формате, который затем
// конвертируется в PPTX серверным билдером.
func PresentationWorker(role, topic, context string) string {
	return `You are a presentation designer. Your role: ` + role + `.

TOPIC:
` + topic + context + `

Produce a slide deck in the following STRICT Markdown slide format:
- The deck title is a single line starting with "# " at the very top.
- Each slide starts with "## " followed by the slide title on its own line.
- Under each slide title, list slide bullet points as "- " lines.
- Optionally add a short "> speaker notes" line per slide (starts with "> ").
- Produce between 6 and 12 slides, including a title slide and a closing slide.

Keep bullets concise (max ~12 words). No code fences around the whole answer.
Return ONLY the slide Markdown.`
}

// ManagerSynthesis — менеджер объединяет результаты воркеров, проводит собственный
// ресёрч и проверяет достоверность, выдавая один связный документ.
func ManagerSynthesis(managerRole, taskType, topic, workerOutputs string) string {
	return `You are the ` + managerRole + ` manager coordinating a ` + taskType + ` team.

ORIGINAL REQUEST:
` + topic + `

YOUR WORKERS PRODUCED:
` + workerOutputs + `

Now act as the senior editor:
1. Merge the workers' results into ONE coherent document.
2. Add your own analysis and fill gaps the workers missed.
3. Fact-check: remove unsupported or contradictory claims, keep what is reliable.
4. Improve structure, flow and formatting.

Return ONE polished GitHub-Flavored Markdown document. No code fences around the whole answer.`
}

// BossChatSummary — босс готовит КОРОТКИЙ текстовый ответ для чата
// (полная развёрнутая версия уходит во вкладку Solution).
func BossChatSummary(taskType, topic, fullDocument string) string {
	return `You are the Boss delivering a ` + taskType + ` result to the user in chat.

USER REQUEST:
` + topic + `

FULL SOLUTION (already shown in the Solution tab):
` + fullDocument + `

Write a SHORT chat reply (3-6 sentences, plain text) that summarizes the result and
tells the user the full version is available in the Solution tab. Do not repeat the
whole document. Return plain text only.`
}
