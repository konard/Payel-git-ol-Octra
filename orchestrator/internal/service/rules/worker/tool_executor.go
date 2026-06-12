package worker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"orchestrator/internal/prompts"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
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
// 1. AI планирует команды (scaffolding + депенденси)
// 2. Команды выполняются внутри nix develop со стримингом вывода
// 3. Новые файлы детектятся через git diff
// 4. Результат возвращается как map[string]string (как от generateCode)
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
	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 2048, 0.3)
	if err != nil {
		return nil, nil, fmt.Errorf("tool planning failed: %w", err)
	}

	var plan struct {
		Commands []string `json:"commands"`
	}
	if err := json.Unmarshal([]byte(util.RepairJSON(util.ExtractJSONFromMarkdown(resp))), &plan); err != nil {
		return nil, nil, fmt.Errorf("tool plan JSON parse failed: %w\nRaw: %s", err, resp)
	}

	if len(plan.Commands) == 0 {
		return nil, nil, fmt.Errorf("AI returned empty commands list for tool mode")
	}

	log.Printf("[ToolExecutor] Role=%s: executing %d commands via nix develop", role, len(plan.Commands))
	if progress != nil {
		progress(30, fmt.Sprintf("Scaffolding project for role %s via %s tools...", role, techStack), nil)
	}

	for i, cmdStr := range plan.Commands {
		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue
		}
		if progress != nil {
			progress(30+int32(i)*30/int32(len(plan.Commands)), cmdStr, nil)
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

	return files, plan.Commands, nil
}

// executeToolCommand запускает команду внутри nix develop со стримингом вывода.
// Если nix недоступен — запускает команду напрямую (для локальной разработки).
func (s *Service) executeToolCommand(ctx context.Context, projectPath, command string, progress rules.ProgressFunc) (string, error) {
	nixAvailable := false
	if _, err := exec.LookPath("nix"); err == nil {
		nixAvailable = true
	}

	var cmd *exec.Cmd
	if nixAvailable {
		cmd = exec.CommandContext(ctx, "nix", "develop",
			"--extra-experimental-features", "nix-command flakes",
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
