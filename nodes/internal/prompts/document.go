package prompts

// Этот файл содержит промпты для не-кодовых задач: ресёрч, текстовые документы
// (отчёты, рефераты, курсовые) и презентации. Они переписаны так, чтобы система
// Воркеры → Менеджер → Босс выдавала результат в формате Markdown (а презентации —
// дополнительно в PPTX), а не исходный код.

// ResearchWorker — промпт для воркера-ресёрчера.
// Каждый воркер исследует тему под своим углом и своим способом поиска.
// searchResults — это блок реальных результатов веб-поиска (заголовок, URL,
// сниппет), уже переранжированных алгоритмом BM25. Если он пуст, воркер
// опирается только на собственные знания и честно это отмечает.
func ResearchWorker(role, angle, topic, context, searchResults string) string {
	sourcesSection := ""
	if searchResults != "" {
		sourcesSection = `

WEB SEARCH RESULTS (ranked by relevance — use these as your primary sources):
` + searchResults + `

Ground your report in these results. Cite the specific [number] and the URL of each
source you rely on. Do NOT invent URLs or facts that are not supported by the results
above or by well-established knowledge.`
	} else {
		sourcesSection = `

NOTE: No live web search results are available. Rely on your own knowledge and clearly
state that the findings are not backed by a fresh web search.`
	}

	return `You are a research analyst on a multi-agent research team. Your role: ` + role + `.
Your search angle / method: ` + angle + `

RESEARCH TOPIC:
` + topic + context + sourcesSection + `

Investigate the topic from YOUR angle only. Be specific, factual and cite concrete
facts, figures, dates and named sources whenever possible. Distinguish clearly between
established facts and your own inferences.

Return a well-structured Markdown report with:
- A short "## Key findings" section (3-7 bullet points).
- A "## Details" section with supporting evidence.
- A "## Sources" section listing the URLs/sources you relied on.

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

Design a UNIQUE, memorable deck — not a generic template. Find a specific angle, story or
through-line for THIS topic so the deck feels custom-made, not interchangeable with any other.

REAL IMAGES: the system automatically searches the web for real, free-license photos and embeds
them into the slides. To help it pick great pictures, give every visual-heavy slide a vivid,
concrete "Visual: " line (subject, setting, mood) — specific subjects beat abstract ones. If the
context's web search results contain a DIRECT image URL (ending in .jpg/.jpeg/.png), you may add
an "Image: <url>" line or a Markdown image "![short alt](url)" so that exact picture is embedded.
Only use image URLs that literally appear in the web search results; never invent URLs.

If the context includes web search results, use them to ground factual claims. Cite only URLs that
appear in those web search results. If no web search results are present, do not invent URLs.

Produce a slide deck in the following STRICT Markdown slide format:
- The deck title is a single line starting with "# " at the very top.
- Each slide starts with "## " followed by the slide title on its own line.
- Under each slide title, list slide bullet points as "- " lines.
- Optionally add one "Visual: " line describing the image, chart, diagram or screenshot
  that should appear on the slide.
- Optionally add one "Image: <url>" line or "![alt](url)" with a DIRECT image URL from the
  web search results when you want that exact picture embedded.
- Optionally add one "Source: " line with a title and URL from the web search results
  when a slide uses a factual claim or visual reference from that source.
- Optionally add a short "> speaker notes" line per slide (starts with "> ").
- Produce between 6 and 12 slides, including a title slide and a closing slide.

VARY THE STRUCTURE so slides do NOT all look identical:
- Mix slide shapes — some with 2 punchy bullets, some with 4-5, some that are a single bold
  statement, some built around one big image, some comparing two things side by side.
- Alternate which slides lean on a Visual/Image and which are text-only, so rhythm changes.
- Give the title slide and the closing slide distinct, non-bullet phrasing.

Keep bullets concise (max ~12 words). Avoid generic placeholders and filler.
No code fences around the whole answer.
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
