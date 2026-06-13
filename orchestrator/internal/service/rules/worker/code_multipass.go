package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"orchestrator/internal/prompts"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
)

// generateCodeMultiPass — single-pass генерация: все файлы за один LLM-запрос (-45% токенов)
// skill — уже разрешённый контент скиллов (может быть пустым, используется только при fallback).
// Каждый распарсенный файл сразу пишется на диск и шлётся на фронтенд через progress.
func (s *Service) generateCodeMultiPass(
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

	prompt := prompts.WorkerMultiPassCode(role, description, taskMD, contextSection, techStack)

	log.Printf("[Worker] Multi-pass generating code for role %s (provider=%s, model=%s, max_tokens=16384, tech=%s)...", role, provider, model, techStack)
	response, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 16384, 0.3)
	if err != nil {
		return nil, nil, fmt.Errorf("multi-pass generation failed: %w", err)
	}
	log.Printf("[Worker] Multi-pass response length: %d chars for role %s", len(response), role)

	files, commands := parseMultiFileResponse(response)
	if len(files) == 0 {
		log.Printf("[Worker] Multi-pass parsing failed, falling back to N+1")
		return s.generateCode(ctx, provider, model, tokens, taskMD, role, description, managerRole, basePath, extCtx, techStack, skill, progress)
	}

	finalFiles := make(map[string]string, len(files))
	i := 0
	for path, content := range files {
		normalizedPath := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
		for strings.HasPrefix(normalizedPath, "./") {
			normalizedPath = strings.TrimPrefix(normalizedPath, "./")
		}
		if normalizedPath == "" || strings.HasPrefix(normalizedPath, "/") || strings.Contains(normalizedPath, "..") {
			continue
		}
		finalFiles[normalizedPath] = content

		// Пишем файл на диск сразу — без ожидания остальных
		fullPath := filepath.Join(basePath, normalizedPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Printf("[Worker] mkdir for %s: %v", normalizedPath, err)
		} else if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			log.Printf("[Worker] write %s: %v", normalizedPath, err)
		} else if progress != nil {
			pct := 50 + int32(i)*30/int32(len(files))
			if pct > 80 {
				pct = 80
			}
			progress(pct, "Writing file: "+normalizedPath, map[string]string{
				"file": normalizedPath,
				"type": "write",
				"size": fmt.Sprintf("%d", len(content)),
			})
		}
		i++
	}
	log.Printf("[Worker] Multi-pass generated %d files and %d commands for role %s", len(finalFiles), len(commands), role)
	return finalFiles, commands, nil
}

// parseMultiFileResponse — разбор формата `=== FILE: path === ... === COMMANDS === ...`
// Также умеет извлекать файлы из heredoc-команд (cat > path << 'EOF' ... EOF) в секции COMMANDS.
func parseMultiFileResponse(content string) (map[string]string, []string) {
	files := make(map[string]string)
	commandsMarker := "=== COMMANDS ==="
	filesSection := content
	commandsSection := ""
	if idx := strings.Index(content, commandsMarker); idx != -1 {
		filesSection = content[:idx]
		commandsSection = content[idx+len(commandsMarker):]
	}

	re := regexp.MustCompile(`(?m)^=== FILE:\s+(.+?)\s+===\s*$`)
	matches := re.FindAllStringSubmatchIndex(filesSection, -1)
	for i, match := range matches {
		path := strings.TrimSpace(filesSection[match[2]:match[3]])
		if path == "" {
			continue
		}
		start := match[1]
		end := len(filesSection)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		fileContent := strings.TrimSpace(filesSection[start:end])
		if fileContent == "" {
			continue
		}
		fileContent = util.StripMarkdownCodeBlock(fileContent)
		files[path] = fileContent
	}

	commands := []string{}
	commandsText := strings.TrimSpace(commandsSection)
	if commandsText != "" {
		commands = strings.Split(commandsText, "\n")
		for i, cmd := range commands {
			commands[i] = strings.TrimSpace(cmd)
		}
	}

	// Если FILE-секция пуста — парсим heredoc-команды из COMMANDS как файлы
	if len(files) == 0 && len(commands) > 0 {
		heredocFiles, usedIndices := parseHeredocCommands(commands)
		if len(heredocFiles) > 0 {
			log.Printf("[Worker] Extracted %d files from heredoc commands", len(heredocFiles))
		}
		for k, v := range heredocFiles {
			files[k] = v
		}
		// Убираем строки, вошедшие в heredoc-блоки
		used := make(map[int]bool)
		for _, idx := range usedIndices {
			used[idx] = true
		}
		var clean []string
		for i, cmd := range commands {
			if !used[i] {
				clean = append(clean, cmd)
			}
		}
		commands = clean
	}

	return files, commands
}

// heredocRe ищет cat > path << ['"]?DELIMITER['"]?
var heredocRe = regexp.MustCompile(`^cat\s+>\s+(\S+)\s+<<\s+['"]?(\w+)['"]?\s*$`)

// parseHeredocCommands извлекает файлы из списка heredoc-команд:
//
//	cat > path/to/file << 'EOF'
//	content...
//	EOF
//
// Возвращает файлы и индексы строк, вошедших в heredoc-блоки.
func parseHeredocCommands(commands []string) (map[string]string, []int) {
	files := make(map[string]string)
	var used []int
	for i := 0; i < len(commands); i++ {
		cmd := strings.TrimSpace(commands[i])
		m := heredocRe.FindStringSubmatch(cmd)
		if m == nil {
			continue
		}
		path := m[1]
		delim := m[2]

		// Ищем закрывающий разделитель
		var contentLines []string
		contentStart := i + 1
		j := contentStart
		for ; j < len(commands); j++ {
			line := strings.TrimSpace(commands[j])
			if line == delim {
				break
			}
			contentLines = append(contentLines, commands[j])
		}
		end := j
		if j >= len(commands) {
			end = len(commands) - 1
		}
		for k := i; k <= end; k++ {
			used = append(used, k)
		}
		content := strings.Join(contentLines, "\n")
		if content != "" {
			files[path] = content
		}
		i = end
	}
	return files, used
}

