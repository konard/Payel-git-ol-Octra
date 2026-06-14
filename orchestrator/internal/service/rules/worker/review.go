package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"orchestrator/internal/config"
	"orchestrator/internal/service/document"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
)

// ReviewWorker — менеджер прислал фидбэк, воркер чинит каждый файл индивидуально
func (s *Service) ReviewWorker(ctx context.Context, req *rules.ReviewRequest) (*rules.ReviewResponse, error) {
	log.Printf("Review request for worker %s (%s): %s", req.WorkerId, req.WorkerRole, req.Feedback)

	metadata := req.Metadata
	provider := metadata["provider"]
	if provider == "" {
		provider = "openai"
	}
	model := metadata["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}

	tokens := make(map[string]string)
	if tokensJSON, ok := metadata["tokens"]; ok {
		json.Unmarshal([]byte(tokensJSON), &tokens)
	}
	if apiKey, ok := metadata[provider]; ok {
		tokens[provider] = apiKey
	}
	taskType := metadata["task_type"]
	if taskType == "" {
		taskType = "code"
	}
	isDoc := taskType != "code"

	fixedFiles := make(map[string]string)
	for filePath, oldContent := range req.OriginalFiles {
		// Бинарные артефакты (например, .pptx) собираются из Markdown билдером и не
		// правятся построчно — иначе LLM повредит файл. Пропускаем их при фиксе;
		// они будут пересобраны из исправленного Markdown.
		if isBinaryPath(filePath) {
			log.Printf("Skipping binary file during review fix: %s", filePath)
			continue
		}

		var prompt string
		if isDoc {
			prompt = fmt.Sprintf(`You are a %s. Your previous %s was reviewed.

FILE: %s

MANAGER FEEDBACK:
%s

PREVIOUS CONTENT:
%s

Revise the content to address the feedback. Keep it well-structured GitHub-Flavored Markdown,
factually accurate and free of placeholders. Return the FULL corrected document as PLAIN TEXT.
No commentary, no surrounding code fences.`,
				req.WorkerRole, taskType, filePath, req.Feedback, oldContent)
		} else {
			prompt = fmt.Sprintf(`You are a %s developer. Your previous work was reviewed.

FILE: %s

MANAGER FEEDBACK:
%s

PREVIOUS CODE:
%s

FIX the code based on the feedback. Return the FULL corrected file as PLAIN TEXT. NO JSON. NO markdown.`,
				req.WorkerRole, filePath, req.Feedback, oldContent)
		}

		fixedContent, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 8192, config.Temperature)
		if err != nil {
			log.Printf("Error fixing file %s: %v", filePath, err)
			continue
		}
		fixedContent = util.StripMarkdownCodeBlock(fixedContent)
		if fixedContent != "" {
			fixedFiles[filePath] = fixedContent
		}
	}

	// Для презентаций пересобираем .pptx из исправленного slide-Markdown,
	// чтобы бинарный файл соответствовал обновлённому содержимому.
	if taskType == "presentation" {
		s.rebuildPresentationArtifacts(ctx, req.OriginalFiles, fixedFiles)
	}

	if len(fixedFiles) == 0 {
		return &rules.ReviewResponse{
			TaskId:   req.TaskId,
			WorkerId: req.WorkerId,
			Status:   "failed",
			Feedback: "No files were fixed during review",
		}, nil
	}

	log.Printf("Worker %s reviewed: %d files fixed", req.WorkerId, len(fixedFiles))
	return &rules.ReviewResponse{
		TaskId:     req.TaskId,
		WorkerId:   req.WorkerId,
		Status:     "fixed",
		SolutionMd: "Reviewed and fixed based on manager feedback",
		Files:      fixedFiles,
		Feedback:   "",
	}, nil
}

// rebuildPresentationArtifacts — для каждого .pptx из исходных файлов, у которого
// есть исправленный парный .md (slide-Markdown), пересобирает .pptx из обновлённого
// Markdown, чтобы бинарник соответствовал отредактированному содержимому.
func (s *Service) rebuildPresentationArtifacts(ctx context.Context, originalFiles, fixedFiles map[string]string) {
	for path := range originalFiles {
		if !strings.HasSuffix(strings.ToLower(path), ".pptx") {
			continue
		}
		mdPath := strings.TrimSuffix(path, ".pptx") + ".md"
		md, ok := fixedFiles[mdPath]
		if !ok || strings.TrimSpace(md) == "" {
			continue
		}
		deck := document.ParseSlideMarkdown(md)
		// После правок Markdown повторно подтягиваем картинки (парсер восстанавливает
		// только их URL, без байтов), иначе встроенные изображения пропали бы.
		if s.attachImages(ctx, &deck, deck.Title) > 0 {
			fixedFiles[mdPath] = document.RenderDeckMarkdown(deck)
		}
		pptxBytes, err := document.BuildPPTX(deck)
		if err != nil {
			log.Printf("Failed to rebuild pptx %s during review: %v", path, err)
			continue
		}
		fixedFiles[path] = string(pptxBytes)
		log.Printf("Rebuilt presentation %s from revised Markdown (%d bytes)", path, len(pptxBytes))
	}
}

