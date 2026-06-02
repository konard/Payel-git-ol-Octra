package worker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"nodes/internal/prompts"
	"nodes/internal/service/document"
	"nodes/internal/service/util"
	"nodes/internal/skills"
)

// generateDocument — путь генерации для не-кодовых задач (research/document/presentation).
// Результат всегда кладётся в папку solution/ в формате Markdown; для презентаций
// дополнительно собирается .pptx через серверный билдер. Возвращает (files, commands, err)
// с той же сигнатурой, что и generateCode, чтобы легко встроиться в runOneWorker.
func (s *Service) generateDocument(
	ctx context.Context, provider, model string, tokens map[string]string,
	taskType, role, description, topic, extCtx, workerID string, emit searchEmitter,
) (map[string]string, []string, error) {
	contextSection := ""
	if extCtx != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS:\n" + extCtx
	}

	files := make(map[string]string)
	slug := slugify(role)
	if slug == "" {
		slug = "result"
	}
	// Несколько воркеров часто получают одинаковую роль (например, "analyst"),
	// поэтому добавляем короткий уникальный суффикс из workerID, чтобы документы
	// разных воркеров не перезаписывали друг друга и все попали в синтез менеджера.
	if suffix := shortID(workerID); suffix != "" {
		slug = slug + "-" + suffix
	}

	switch taskType {
	case "presentation":
		searchBlock, sourcesMd, n := s.gatherSearch(ctx, emit, role, topic, "visual references images charts diagrams screenshots examples")
		if sourcesMd != "" {
			files["solution/visual-sources-"+slug+".md"] = sourcesMd
		}
		presentationContext := contextSection
		if searchBlock != "" {
			presentationContext += "\n\nWEB SEARCH RESULTS FOR PRESENTATION SOURCES AND VISUAL IDEAS:\n" + searchBlock
		}
		skill := skills.Guidance(role+" presentation", "pptx markdown", topic)
		prompt := prompts.PresentationWorker(role, topic, presentationContext, skill)
		resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 8192, 0.4)
		if err != nil {
			return nil, nil, fmt.Errorf("presentation generation failed: %w", err)
		}
		md := util.StripMarkdownCodeBlock(resp)
		deck := document.ParseSlideMarkdown(md)
		if deck.Title == "" {
			deck.Title = firstLine(topic)
		}
		// Ищем в интернете реальные картинки и встраиваем их в слайды. Если поиск
		// что-то нашёл — перезаписываем Markdown ссылками, чтобы .md и .pptx совпадали.
		imgN := s.attachImages(ctx, &deck, topic)
		if imgN > 0 {
			md = document.RenderDeckMarkdown(deck)
		}
		pptxBytes, err := document.BuildPPTX(deck)
		if err != nil {
			return nil, nil, fmt.Errorf("pptx build failed: %w", err)
		}
		files["solution/"+slug+".md"] = md
		files["solution/"+slug+".pptx"] = string(pptxBytes)
		log.Printf("[Worker] Presentation generated: %d slides, %d images, %d bytes pptx, %d web sources", len(deck.Slides), imgN, len(pptxBytes), n)

	case "research":
		angle := description
		if angle == "" {
			angle = "general web research"
		}
		// Воркер сначала выполняет реальный веб-поиск в своём диапазоне, затем
		// LLM пишет отчёт, опираясь на найденные источники (а не выдумывая их).
		searchBlock, sourcesMd, n := s.gatherSearch(ctx, emit, role, topic, angle)
		if sourcesMd != "" {
			files["solution/sources-"+slug+".md"] = sourcesMd
		}
		skill := skills.Guidance(role+" research "+angle, "markdown", topic)
		prompt := prompts.ResearchWorker(role, angle, topic, contextSection, searchBlock, skill)
		resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 8192, 0.4)
		if err != nil {
			return nil, nil, fmt.Errorf("research generation failed: %w", err)
		}
		files["solution/"+slug+".md"] = util.StripMarkdownCodeBlock(resp)
		log.Printf("[Worker] Research document generated: %d chars, %d web sources", len(resp), n)

	default: // "document"
		docType := detectDocType(topic + " " + description)
		skill := skills.Guidance(role+" "+docType, "markdown", topic+" "+description)
		prompt := prompts.DocumentWorker(role, docType, topic, contextSection, skill)
		resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 8192, 0.4)
		if err != nil {
			return nil, nil, fmt.Errorf("document generation failed: %w", err)
		}
		files["solution/"+slug+".md"] = util.StripMarkdownCodeBlock(resp)
		log.Printf("[Worker] Text document (%s) generated: %d chars", docType, len(resp))
	}

	return files, []string{}, nil
}

// detectDocType — определяет тип текстового документа по запросу пользователя.
func detectDocType(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "essay") || strings.Contains(t, "сочинени"):
		return "essay"
	case strings.Contains(t, "coursework") || strings.Contains(t, "курсов"):
		return "coursework"
	case strings.Contains(t, "abstract") || strings.Contains(t, "реферат"):
		return "abstract"
	case strings.Contains(t, "table") || strings.Contains(t, "таблиц"):
		return "table"
	case strings.Contains(t, "report") || strings.Contains(t, "отчёт") || strings.Contains(t, "отчет"):
		return "report"
	default:
		return "document"
	}
}

// shortID — короткий стабильный суффикс из UUID воркера (первый сегмент).
func shortID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if i := strings.IndexByte(id, '-'); i > 0 {
		return id[:i]
	}
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slugify — превращает строку в безопасное имя файла.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if len(s) > 80 {
		s = s[:80]
	}
	if s == "" {
		return "Presentation"
	}
	return s
}
