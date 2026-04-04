package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

func (s *WorkerService) generateCode(ctx context.Context, provider, model string, tokens map[string]string, taskMD, role, description, managerRole, basePath, context string) (map[string]string, error) {
	contextSection := ""
	if context != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS:\n" + context
	}

	// Шаг 1: LLM определяет список файлов
	planPrompt := fmt.Sprintf(`You are a %s developer. Role: %s

TASK: %s%s

List ONLY the 3-5 most important files to create. Be specific.
Return JSON ONLY:
{"files": ["path1.ext", "path2.ext", "path3.ext"]}`,
		role, description, taskMD, contextSection)

	planResp, err := s.agentsClient.Generate(ctx, provider, model, planPrompt, tokens, 1024, 0.3)
	if err != nil {
		return nil, err
	}

	var plan struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(repairJSON(extractJSONFromMarkdown(planResp))), &plan); err != nil {
		// Fallback: default files
		plan.Files = []string{"main.go", "README.md"}
	}

	if len(plan.Files) == 0 {
		plan.Files = []string{"main.go"}
	}

	// Шаг 2: Для КАЖДОГО файла — отдельный запрос к LLM
	files := make(map[string]string)
	for _, file := range plan.Files {
		contentPrompt := fmt.Sprintf(`Write the FULL content of file: %s

TASK: %s
Role: %s

Write COMPLETE code. No placeholders. No TODOs. No "implement later".
Return the file content as PLAIN TEXT. NO JSON. NO markdown. Just the raw code.`,
			file, taskMD, role)

		content, err := s.agentsClient.Generate(ctx, provider, model, contentPrompt, tokens, 8192, 0.3)
		if err != nil {
			log.Printf("Error generating file %s: %v", file, err)
			continue
		}

		// Strip markdown code blocks if present
		content = stripMarkdownCodeBlock(content)

		if content != "" {
			files[fmt.Sprintf("%s/%s/%s", basePath, role, file)] = content
		}
	}

	log.Printf("Generated %d files for role %s", len(files), role)
	return files, nil
}
