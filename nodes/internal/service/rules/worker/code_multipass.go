package worker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"nodes/internal/service/util"
)

// generateCodeMultiPass — single-pass генерация: все файлы за один LLM-запрос (-45% токенов)
func (s *Service) generateCodeMultiPass(
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

	prompt := fmt.Sprintf(`You are a %s developer. Role: %s
Language: %s

TASK: %s%s

IMPORTANT: Write code in %s ONLY. NOT JavaScript, NOT TypeScript.
Create 3-5 most important files for this project (use .%s extension).

AFTER files, provide COMMANDS to execute in the project (mkdir, echo, etc.).

RETURN FORMAT (STRICT - follow exactly):
=== FILE: path/to/file1.%s ===
<complete code for file1.%s - no placeholders, no TODOs>
=== FILE: path/to/file2.%s ===
<complete code for file2.%s - no placeholders, no TODOs>
=== FILE: path/to/file3.%s ===
<complete code for file3.%s - no placeholders, no TODOs>
=== COMMANDS ===
mkdir -p dir
echo 'content' > file.txt
# other bash commands

RULES:
1. Each file MUST start with "=== FILE: <path> ===" on its own line
2. File paths should be relative to project root (e.g., main.%s, cmd/server/main.%s) - DO NOT include project name in path
3. File content MUST be complete code - no placeholders, no TODOs, no "implement later"
4. Use proper imports and exports
5. Keep code compact but functional (300-500 lines max per file)
6. Do NOT include markdown code fences around the entire response
7. If you can't create a file, skip it and move to the next one
8. COMMANDS: List bash commands to run in project root, one per line
9. Return ONLY the files and commands, no explanations`,
		role, description, techStack,
		taskMD, contextSection,
		techStack, techStack,
		techStack, techStack, techStack, techStack, techStack, techStack, techStack, techStack)

	response, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 16384, 0.3)
	if err != nil {
		return nil, nil, fmt.Errorf("multi-pass generation failed: %w", err)
	}
	log.Printf("[Worker] Multi-pass response length: %d chars", len(response))

	files, commands := parseMultiFileResponse(response)
	if len(files) == 0 {
		log.Printf("[Worker] Multi-pass parsing failed, falling back to N+1")
		return s.generateCode(ctx, provider, model, tokens, taskMD, role, description, managerRole, basePath, extCtx, techStack)
	}

	finalFiles := make(map[string]string, len(files))
	for path, content := range files {
		normalizedPath := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
		for strings.HasPrefix(normalizedPath, "./") {
			normalizedPath = strings.TrimPrefix(normalizedPath, "./")
		}
		if normalizedPath == "" || strings.HasPrefix(normalizedPath, "/") || strings.Contains(normalizedPath, "..") {
			continue
		}
		finalFiles[normalizedPath] = content
	}
	log.Printf("[Worker] Multi-pass generated %d files and %d commands for role %s", len(finalFiles), len(commands), role)
	return finalFiles, commands, nil
}

// parseMultiFileResponse — разбор формата `=== FILE: path === ... === COMMANDS === ...`
func parseMultiFileResponse(content string) (map[string]string, []string) {
	files := make(map[string]string)
	re := regexp.MustCompile(`(?m)^=== FILE:\s+(.+?)\s+===\s*$`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return files, []string{}
	}
	for i, match := range matches {
		path := strings.TrimSpace(content[match[2]:match[3]])
		if path == "" {
			continue
		}
		start := match[1]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		fileContent := strings.TrimSpace(content[start:end])
		if fileContent == "" {
			continue
		}
		fileContent = util.StripMarkdownCodeBlock(fileContent)
		files[path] = fileContent
	}
	commands := []string{}
	commandsMarker := "=== COMMANDS ==="
	if idx := strings.Index(content, commandsMarker); idx != -1 {
		commandsText := strings.TrimSpace(content[idx+len(commandsMarker):])
		if commandsText != "" {
			commands = strings.Split(commandsText, "\n")
			for i, cmd := range commands {
				commands[i] = strings.TrimSpace(cmd)
			}
		}
	}
	return files, commands
}
