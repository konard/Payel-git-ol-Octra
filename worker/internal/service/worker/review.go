package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"worker/internal/fetcher/grpc/workerpb"
)

// ReviewWorker handles manager feedback and fixes code
func (s *WorkerService) ReviewWorker(ctx context.Context, req *workerpb.ReviewRequest) (*workerpb.ReviewResponse, error) {
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

	// Fix each file individually (plain text approach)
	fixedFiles := make(map[string]string)
	for filePath, oldContent := range req.OriginalFiles {
		prompt := fmt.Sprintf(`You are a %s developer. Your previous work was reviewed.

FILE: %s

MANAGER FEEDBACK:
%s

PREVIOUS CODE:
%s

FIX the code based on the feedback. Return the FULL corrected file as PLAIN TEXT. NO JSON. NO markdown.`,
			req.WorkerRole, filePath, req.Feedback, oldContent)

		fixedContent, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 8192, 0.3)
		if err != nil {
			log.Printf("Error fixing file %s: %v", filePath, err)
			continue
		}

		fixedContent = stripMarkdownCodeBlock(fixedContent)
		if fixedContent != "" {
			fixedFiles[filePath] = fixedContent
		}
	}

	if len(fixedFiles) == 0 {
		return &workerpb.ReviewResponse{
			TaskId:   req.TaskId,
			WorkerId: req.WorkerId,
			Status:   "failed",
			Feedback: "No files were fixed during review",
		}, nil
	}

	log.Printf("Worker %s reviewed: %d files fixed", req.WorkerId, len(fixedFiles))

	return &workerpb.ReviewResponse{
		TaskId:     req.TaskId,
		WorkerId:   req.WorkerId,
		Status:     "fixed",
		SolutionMd: fmt.Sprintf("Reviewed and fixed based on manager feedback"),
		Files:      fixedFiles,
		Feedback:   "",
	}, nil
}
