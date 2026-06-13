package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/prompts"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	"orchestrator/internal/skills"
)

// generateCode — N+1 подход: спланировать список файлов, потом сгенерировать каждый
// skill — уже разрешённый контент скиллов (из Warehouse.Select() или Guidance()), может быть пустым.
// Каждый сгенерированный файл сразу пишется на диск и шлётся на фронтенд через progress.
func (s *Service) generateCode(
	ctx context.Context, provider, model string, tokens map[string]string,
	taskMD, role, description, managerRole, basePath, extCtx, techStack, skill string,
	progress rules.ProgressFunc,
) (map[string]string, []string, error) {
	contextSection := ""
	if extCtx != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS:\n" + extCtx
	}
	if techStack == "" {
		techStack = "Go"
	}
	if skill == "" {
		skill = skills.Guidance(role+" "+description, techStack, taskMD)
	}

	planPrompt := prompts.WorkerPlanFiles(role, description, taskMD, contextSection, techStack, skill)
	log.Printf("[Worker] Planning files for role %s (provider=%s, model=%s)...", role, provider, model)
	planResp, err := s.agentsClient.Generate(ctx, provider, model, planPrompt, tokens, 1024, 0.3)
	if err != nil {
		return nil, nil, err
	}

	var plan struct {
		Files    []string `json:"files"`
		Commands []string `json:"commands"`
	}
	// Fallback-файл должен соответствовать стеку: для Node.js это index.js, а не
	// main.go. Раньше fallback всегда был "main.go" — из-за этого JS-код мог
	// оказаться в .go-файле, когда планирование файлов проваливалось.
	mainFile := prompts.LangMainFile(techStack)
	if err := json.Unmarshal([]byte(util.RepairJSON(util.ExtractJSONFromMarkdown(planResp))), &plan); err != nil {
		plan.Files = []string{mainFile, "README.md"}
	}
	if len(plan.Files) == 0 {
		plan.Files = []string{mainFile}
	}

	files := make(map[string]string)
	for i, file := range plan.Files {
		log.Printf("[Worker] Generating file %d/%d: %s for role %s", i+1, len(plan.Files), file, role)
		contentPrompt := prompts.WorkerGenerateFile(file, taskMD, role, techStack, "")
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

		// Пишем файл на диск сразу — не ждём остальных
		fullPath := filepath.Join(basePath, normalizedPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Printf("[Worker] mkdir for %s: %v", normalizedPath, err)
		} else if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			log.Printf("[Worker] write %s: %v", normalizedPath, err)
		} else if progress != nil {
			pct := 50 + int32(i)*30/int32(len(plan.Files))
			if pct > 80 {
				pct = 80
			}
			progress(pct, "Writing file: "+normalizedPath, map[string]string{
				"file": normalizedPath,
				"type": "write",
				"size": fmt.Sprintf("%d", len(content)),
			})
		}
	}

	log.Printf("Generated %d files and %d commands for role %s", len(files), len(plan.Commands), role)
	return files, plan.Commands, nil
}

