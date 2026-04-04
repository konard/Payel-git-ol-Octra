package worker

import (
	"context"
	"fmt"
	"log"
)

func (s *WorkerService) createTaskMD(ctx context.Context, provider, model string, tokens map[string]string, role, description, taskMD, context, basePath string) (string, error) {
	contextSection := ""
	if context != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS (use for coordination):\n" + context
	}

	prompt := fmt.Sprintf(`You are a %s developer on a project team.

YOUR ROLE: %s
TASK:
%s
%s

Create TASK.md file with detailed task breakdown for your role. Include:
1. Files to create
2. Functionality to implement
3. Dependencies
4. How your work integrates with other workers (if context provided)

Return ONLY the content of TASK.md file.`, role, description, taskMD, contextSection)

	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 2048, 0.7)
	if err != nil {
		return "", err
	}

	log.Printf("Worker created TASK.md (%s)", role)
	return resp, nil
}
