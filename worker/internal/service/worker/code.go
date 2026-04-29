package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func (s *WorkerService) generateCode(ctx context.Context, provider, model string, tokens map[string]string, taskMD, role, description, managerRole, basePath, context string) (map[string]string, []string, error) {
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
		return nil, nil, err
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

		if content == "" {
			continue
		}
		normalizedPath := strings.TrimSpace(strings.ReplaceAll(file, "\\", "/"))
		for strings.HasPrefix(normalizedPath, "./") {
			normalizedPath = strings.TrimPrefix(normalizedPath, "./")
		}
		if normalizedPath == "" || strings.HasPrefix(normalizedPath, "/") || strings.Contains(normalizedPath, "..") {
			continue
		}
		files[normalizedPath] = content
	}

	// Generate commands
	commandsPrompt := fmt.Sprintf(`You are a %s developer. Role: %s

TASK: %s%s

Based on the files created, provide bash commands to execute in the project root (mkdir, echo, etc.).
Return JSON ONLY: {"commands": ["cmd1", "cmd2"]}`,
		role, description, taskMD, contextSection)

	commandsResp, err := s.agentsClient.Generate(ctx, provider, model, commandsPrompt, tokens, 1024, 0.3)
	if err != nil {
		log.Printf("Error generating commands: %v", err)
		return files, []string{}, nil
	}

	var commandsStruct struct {
		Commands []string `json:"commands"`
	}
	if err := json.Unmarshal([]byte(repairJSON(extractJSONFromMarkdown(commandsResp))), &commandsStruct); err != nil {
		log.Printf("Error parsing commands JSON: %v", err)
		return files, []string{}, nil
	}

	log.Printf("Generated %d files and %d commands for role %s", len(files), len(commandsStruct.Commands), role)
	return files, commandsStruct.Commands, nil
}
