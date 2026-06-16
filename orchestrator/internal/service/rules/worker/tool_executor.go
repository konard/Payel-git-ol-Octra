package worker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"orchestrator/internal/config"
	"orchestrator/internal/prompts"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	_ "orchestrator/pkg/instralutions"
	instcore "orchestrator/pkg/instralutions/core"
)

// toolTechStacks — tech stack'и, для которых используется ToolExecutor вместо AI-генерации.
// Для этих стеков воркер запускает реальные команды (npm init, flutter create, cargo init, …)
// внутри nix develop, а файлы детектит через git.
var toolTechStacks = map[string]bool{
	"node":       true,
	"nodejs":     true,
	"typescript": true,
	"javascript": true,
	"flutter":    true,
	"dart":       true,
	"rust":       true,
	"php":        true,
	"c++":        true,
	"cpp":        true,
	"dotnet":     true,
	"csharp":     true,
	"ruby":       true,
	"elixir":     true,
	"haskell":    true,
	"scala":      true,
	"kotlin":     true,
	"java":       true,
}

// generateViaTools — генерирует код через реальные тулы внутри nix develop.
// 1. AI планирует: возвращает install-флаги + post-generation команды
// 2. install-флаги разрешаются в команды через instralutions registry
// 3. Все команды выполняются внутри nix develop со стримингом вывода
// 4. Новые файлы детектятся через git diff
// 5. Результат возвращается как map[string]string (как от generateCode)
func (s *Service) generateViaTools(
	ctx context.Context, provider, model string, tokens map[string]string,
	taskMD, role, description, managerRole, basePath, extCtx, techStack string,
	progress rules.ProgressFunc,
) (map[string]string, []string, error) {
	contextSection := ""
	if extCtx != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS:\n" + extCtx
	}

	prompt := prompts.WorkerToolCommands(role, description, taskMD, contextSection, techStack)
	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 2048, config.Temperature)
	if err != nil {
		return nil, nil, fmt.Errorf("tool planning failed: %w", err)
	}

	var plan struct {
		Install  []string `json:"install"`
		Commands []string `json:"commands"`
	}
	if err := json.Unmarshal([]byte(util.RepairJSON(util.ExtractJSONFromMarkdown(resp))), &plan); err != nil {
		return nil, nil, fmt.Errorf("tool plan JSON parse failed: %w\nRaw: %s", err, resp)
	}

	// Шаг 1: resolve install-флагов в команды через instralutions registry
	var allCommands []string
	if len(plan.Install) > 0 {
		log.Printf("[ToolExecutor] Role=%s: resolving install flags: %v", role, plan.Install)
		if progress != nil {
			progress(30, fmt.Sprintf("Installing %s via nix develop...", strings.Join(plan.Install, ", ")), nil)
		}
		installCmds := instcore.Resolve(plan.Install)

		// Апдейтим flake.nix с nix-пакетами, нужными для install-флагов
		nixPkgs := instcore.ResolvePackages(plan.Install)
		if len(nixPkgs) > 0 {
			log.Printf("[ToolExecutor] Ensuring nix packages in flake: %v", nixPkgs)
			instcore.EnsureNixPackages(basePath, nixPkgs)
		}

		allCommands = append(allCommands, installCmds...)
	}

	// Шаг 2: AI-generated post-generation команды
	allCommands = append(allCommands, plan.Commands...)

	if len(allCommands) == 0 {
		return nil, nil, fmt.Errorf("AI returned empty commands (no install flags, no commands)")
	}

	log.Printf("[ToolExecutor] Role=%s: executing %d commands (install:%d, commands:%d) via nix develop",
		role, len(allCommands), len(plan.Install), len(plan.Commands))

	for i, cmdStr := range allCommands {
		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue
		}
		if progress != nil {
			progress(30+int32(i)*30/int32(len(allCommands)), cmdStr, nil)
		}
		output, execErr := s.executeToolCommand(ctx, basePath, cmdStr, progress)
		if execErr != nil {
			log.Printf("[ToolExecutor] Command failed: %q\nError: %v\nOutput: %s", cmdStr, execErr, output)
		} else {
			log.Printf("[ToolExecutor] Command succeeded: %q", cmdStr)
		}
	}

	files := detectNewFiles(basePath)
	log.Printf("[ToolExecutor] Role=%s: detected %d new files after tool execution", role, len(files))

	if len(files) == 0 {
		log.Printf("[ToolExecutor] No files detected via git diff. Attempting fallback: read all non-git files")
		files = readProjectFiles(basePath)
	}

	// Шлём каждый обнаруженный файл на фронтенд
	i := 0
	for path, content := range files {
		if progress != nil {
			pct := 60 + int32(i)*20/int32(len(files))
			if pct > 80 {
				pct = 80
			}
			data := map[string]string{
				"file": path,
				"type": "scaffold",
				"size": fmt.Sprintf("%d", len(content)),
			}
			// Стримим наскаффоленный файл во вкладку Solution вживую (issue #75 п.1).
			if cf := liveCodeFilesPayload(role, path, content); cf != "" {
				data["code_files"] = cf
			}
			progress(pct, "Scaffolded: "+path, data)
			i++
		}
	}

	return files, plan.Commands, nil
}

// executeToolCommand запускает команду внутри nix develop со стримингом вывода.
// Использует --profile .octra/nix-profile для кэширования собранного окружения
// между вызовами — nix пересоберёт только если flake.nix изменился.
// Если nix недоступен — запускает команду напрямую (для локальной разработки).
func (s *Service) executeToolCommand(ctx context.Context, projectPath, command string, progress rules.ProgressFunc) (string, error) {
	nixAvailable := false
	if _, err := exec.LookPath("nix"); err == nil {
		nixAvailable = true
	}

	var cmd *exec.Cmd
	if nixAvailable {
		profilePath := util.NixProfilePath(projectPath)
		os.MkdirAll(filepath.Dir(profilePath), 0755)
		cmd = exec.CommandContext(ctx, "nix", "develop",
			"--extra-experimental-features", "nix-command flakes",
			"--profile", profilePath,
			"--command", "bash", "-c", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}
	cmd.Dir = projectPath

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	var output strings.Builder
	reader := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*64), 1024*64)

	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line + "\n")
		log.Printf("[nix] %s", line)

		if progress != nil && strings.Contains(line, "copying path") {
			if m := util.NixPkgRe.FindStringSubmatch(line); len(m) > 1 {
				pkg := strings.TrimSpace(m[1])
				progress(50, "Installing "+pkg+"...", nil)
			}
		}
	}

	err := cmd.Wait()
	return output.String(), err
}

// detectNewFiles находит файлы, созданные инструментами, через git status --porcelain.
// Парсит staged (A/M), unstaged modified (M) и untracked (??) файлы.
func detectNewFiles(projectPath string) map[string]string {
	files := make(map[string]string)

	cmd := exec.Command("git", "-C", projectPath, "status", "--porcelain", "-u")
	out, err := cmd.Output()
	if err != nil {
		return files
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Формат: XY filename
		// XY может быть: "??", " M", "M ", "A ", "AM", etc.
		// Нас интересуют все, кроме пробелов (неизменённые)
		if len(line) < 3 {
			continue
		}
		x := line[0]
		y := line[1]
		if x == ' ' && y == ' ' {
			continue
		}
		path := strings.TrimSpace(line[2:])
		if path == "" {
			continue
		}
		// Для переименованных файлов: "R  old -> new"
		if x == 'R' {
			if parts := strings.SplitN(path, " -> ", 2); len(parts) == 2 {
				path = strings.TrimSpace(parts[1])
			}
		}
		content, readErr := os.ReadFile(filepath.Join(projectPath, path))
		if readErr != nil {
			continue
		}
		files[path] = string(content)
	}

	return files
}

// readProjectFiles — fallback: читает все файлы проекта (исключая .git).
func readProjectFiles(projectPath string) map[string]string {
	files := make(map[string]string)
	_ = filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(projectPath, path)
		if err != nil {
			return nil
		}
		if strings.HasPrefix(rel, ".git") || strings.HasPrefix(rel, ".octra") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		files[rel] = string(content)
		return nil
	})
	return files
}

// isToolMode проверяет, нужно ли использовать ToolExecutor для данного tech stack.
func isToolMode(techStack string) bool {
	return toolTechStacks[strings.ToLower(strings.TrimSpace(techStack))]
}

// sourceExtensions — расширения, которые считаются «настоящим» исходным кодом
// (в отличие от манифестов/локов/инфраструктуры).
var sourceExtensions = map[string]bool{
	".go": true, ".js": true, ".mjs": true, ".cjs": true, ".jsx": true,
	".ts": true, ".tsx": true, ".py": true, ".java": true, ".kt": true,
	".rs": true, ".rb": true, ".php": true, ".c": true, ".h": true,
	".cc": true, ".cpp": true, ".cxx": true, ".hpp": true, ".hxx": true,
	".cs": true, ".scala": true, ".ex": true, ".exs": true, ".hs": true,
	".dart": true, ".html": true, ".htm": true, ".css": true, ".scss": true,
	".vue": true, ".svelte": true, ".swift": true, ".sh": true, ".sql": true,
	".zig": true, ".clj": true, ".erl": true,
}

// liveCodeFile — форма одного файла для инкрементального code_files-апдейта.
// Совпадает с StreamedCodeFile на фронтенде (hooks/useWebSocket.ts).
type liveCodeFile struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	Language   string `json:"language"`
	Encoding   string `json:"encoding,omitempty"`
	WorkerRole string `json:"worker_role"`
	Status     string `json:"status"`
	UpdatedAt  int64  `json:"updated_at"`
}

// liveCodeFilesPayload собирает code_files-JSON для ОДНОГО только что записанного
// файла, чтобы вкладка Solution обновлялась вживую по мере создания файлов
// (issue #75 п.1). Инфраструктурные файлы (flake.nix/.octra/...) пропускаются,
// бинарное содержимое кодируется в base64. Возвращает "" если файл показывать не надо.
func liveCodeFilesPayload(role, path, content string) string {
	if strings.TrimSpace(path) == "" || util.IsIgnoredPath(path) {
		return ""
	}
	encoding := ""
	if util.IsBinaryPath(path) {
		content = base64.StdEncoding.EncodeToString([]byte(content))
		encoding = "base64"
	}
	entry := liveCodeFile{
		Path:       path,
		Content:    content,
		Language:   util.LanguageForPath(path),
		Encoding:   encoding,
		WorkerRole: role,
		Status:     "streaming",
		UpdatedAt:  time.Now().Unix(),
	}
	data, err := json.Marshal([]liveCodeFile{entry})
	if err != nil {
		return ""
	}
	return string(data)
}

// содержит ли файл реальный код, а не только TODO-заглушки и комментарии.
func hasRealCode(content string) bool {
	cleaned := strings.TrimSpace(content)
	if cleaned == "" {
		return false
	}
	lines := strings.Split(cleaned, "\n")
	nonStubLines := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		nonStubLines++
	}
	// Файл считается реальным кодом, если в нём есть хотя бы одна строка кода,
	// а не только пустые строки и комментарии/TODO-заглушки (issue #98). Порог
	// «>= 3 строки» был ошибкой: он отбраковывал валидный короткий код вроде
	// `package main` + `func main() {}` или однострочного express-сервера, из-за
	// чего такой код не считался исходником (см. TestContainsSourceCode, issue #75).
	return nonStubLines >= 1
}

// containsSourceCode сообщает, есть ли среди файлов хотя бы один непустой файл с
// исходным кодом. Манифесты и инфраструктура (package.json, flake.nix, .gitignore,
// lock-файлы) НЕ считаются. Когда tool-режим наскаффолдил только их (например
// `npm init -y` создал лишь package.json), воркер обязан уйти в AI-генерацию,
// чтобы реальный код фичи (index.js express-сервера и т.п.) всё-таки появился.
// Это корневая причина issue #75 п.6: фронтенду уезжали только flake.nix/context.json.
// Дополнительно проверяется, что код не состоит из одних TODO-заглушек (issue #98).
func containsSourceCode(files map[string]string) bool {
	for path, content := range files {
		if !sourceExtensions[strings.ToLower(filepath.Ext(path))] {
			continue
		}
		if hasRealCode(content) {
			return true
		}
	}
	return false
}
