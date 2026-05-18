package worker

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"nodes/internal/prompts"
	"nodes/internal/service/util"
)

// generateCode — N+1 подход: спланировать список файлов, потом сгенерировать каждый
func (s *Service) generateCode(
	ctx context.Context, provider, model string, tokens map[string]string,
	taskMD, role, description, managerRole, basePath, extCtx, techStack string,
) (map[string]string, []string, error) {
	contextSection := ""
	if extCtx != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS:\n" + extCtx
	}
	if techStack == "" {
		techStack = "Go"
	}

	planPrompt := prompts.WorkerPlanFiles(role, description, taskMD, contextSection, techStack)
	planResp, err := s.agentsClient.Generate(ctx, provider, model, planPrompt, tokens, 1024, 0.3)
	if err != nil {
		return nil, nil, err
	}

	var plan struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(util.RepairJSON(util.ExtractJSONFromMarkdown(planResp))), &plan); err != nil {
		plan.Files = []string{"main.go", "README.md"}
	}
	if len(plan.Files) == 0 {
		plan.Files = []string{"main.go"}
	}

	files := make(map[string]string)
	for _, file := range plan.Files {
		contentPrompt := prompts.WorkerGenerateFile(file, taskMD, role, techStack)
		content, err := s.agentsClient.Generate(ctx, provider, model, contentPrompt, tokens, 8192, 0.3)
		if err != nil {
			log.Printf("Error generating file %s: %v", file, err)
			continue
		}
		content = util.StripMarkdownCodeBlock(content)
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

	commandsPrompt := prompts.WorkerGenerateCommands(role, description, taskMD, contextSection)
	commandsResp, err := s.agentsClient.Generate(ctx, provider, model, commandsPrompt, tokens, 1024, 0.3)
	if err != nil {
		log.Printf("Error generating commands: %v", err)
		return files, []string{}, nil
	}
	var commandsStruct struct {
		Commands []string `json:"commands"`
	}
	if err := json.Unmarshal([]byte(util.RepairJSON(util.ExtractJSONFromMarkdown(commandsResp))), &commandsStruct); err != nil {
		log.Printf("Error parsing commands JSON: %v", err)
		return files, []string{}, nil
	}
	log.Printf("Generated %d files and %d commands for role %s", len(files), len(commandsStruct.Commands), role)
	return files, commandsStruct.Commands, nil
}
