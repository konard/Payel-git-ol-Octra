package worker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// generateCodeMultiPass generates all files in a single LLM request (optimized, -45% tokens)
func (s *WorkerService) generateCodeMultiPass(ctx context.Context, provider, model string, tokens map[string]string, taskMD, role, description, managerRole, basePath, context string) (map[string]string, error) {
	contextSection := ""
	if context != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS:\n" + context
	}

	// Single request: generate all files at once
	prompt := fmt.Sprintf(`You are a %s developer. Role: %s

TASK: %s%s

Create 3-5 most important files for this project.

RETURN FORMAT (STRICT - follow exactly):
=== FILE: path/to/file1.ext ===
<complete code for file1.ext - no placeholders, no TODOs>
=== FILE: path/to/file2.ext ===
<complete code for file2.ext - no placeholders, no TODOs>
=== FILE: path/to/file3.ext ===
<complete code for file3.ext - no placeholders, no TODOs>

RULES:
1. Each file MUST start with "=== FILE: <path> ===" on its own line
2. File content MUST be complete code - no placeholders, no TODOs, no "implement later"
3. Use proper imports and exports
4. Keep code compact but functional (300-500 lines max per file)
5. Do NOT include markdown code fences around the entire response
6. If you can't create a file, skip it and move to the next one
7. Return ONLY the files, no explanations`,
		role, description, taskMD, contextSection)

	response, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 16384, 0.3)
	if err != nil {
		return nil, fmt.Errorf("multi-pass generation failed: %w", err)
	}

	// Parse multi-file response
	files := parseMultiFileResponse(response)

	if len(files) == 0 {
		// Fallback to N+1 approach if parsing failed
		log.Printf("[Worker] Multi-pass parsing failed, falling back to N+1")
		return s.generateCode(ctx, provider, model, tokens, taskMD, role, description, managerRole, basePath, context)
	}

	// Prepend basePath to all files
	finalFiles := make(map[string]string)
	for path, content := range files {
		finalFiles[fmt.Sprintf("%s/%s/%s", basePath, role, path)] = content
	}

	log.Printf("[Worker] Multi-pass generated %d files for role %s", len(finalFiles), role)
	return finalFiles, nil
}

// parseMultiFileResponse parses a multi-file response
// Expected format:
// === FILE: path/to/file.go ===
// <code>
// === FILE: path/to/file2.go ===
// <code>
func parseMultiFileResponse(content string) map[string]string {
	files := make(map[string]string)

	// Regex to match "=== FILE: <path> ==="
	re := regexp.MustCompile(`(?m)^=== FILE:\s+(.+?)\s+===\s*$`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		return files
	}

	for i, match := range matches {
		// Extract file path
		path := strings.TrimSpace(content[match[2]:match[3]])
		if path == "" {
			continue
		}

		// Extract content: from end of marker to start of next marker (or end of string)
		start := match[1]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}

		fileContent := strings.TrimSpace(content[start:end])
		if fileContent == "" {
			continue
		}

		// Strip markdown code blocks if present
		fileContent = stripMarkdownCodeBlock(fileContent)

		files[path] = fileContent
	}

	return files
}
