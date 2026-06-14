package boss

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
)

// ensureFlakeLock generates flake.lock by running nix flake lock in the project
// directory. This pins all dependency versions (nixpkgs revision, etc.) for
// reproducible builds. Non-fatal: if nix is unavailable or the command fails,
// the pipeline continues with a warning.
func (s *Service) ensureFlakeLock(projectPath string, progress rules.ProgressFunc) {
	if !util.NixAvailable() {
		log.Printf("Nix not available, skipping flake.lock generation")
		return
	}

	// flake.nix must be tracked by git for nix commands to see it
	exec.Command("git", "-C", projectPath, "add", "flake.nix").Run()

	log.Printf("Generating flake.lock for project: %s", projectPath)
	if progress != nil {
		progress(65, "Pinning dependency versions...", nil)
	}

	cmd := exec.Command("nix", "flake", "lock",
		"--extra-experimental-features", "nix-command flakes")
	cmd.Dir = projectPath

	if err := streamCommandOutput(cmd, progress, "Resolving dependencies..."); err != nil {
		log.Printf("Warning: nix flake lock failed (non-fatal): %v", err)
		emit(progress, 66, "flake.lock generation failed (non-fatal)", map[string]string{
			"warning": err.Error(),
		})
		return
	}
	log.Printf("flake.lock generated at: %s", filepath.Join(projectPath, "flake.lock"))
}

// nixBuild runs nix build to verify the project compiles inside Nix's isolated
// build environment. This is a non-fatal validation step: if it fails, the
// pipeline continues but logs a warning (nix may not be available, or there
// may be environment-specific issues).
func (s *Service) nixBuild(projectPath string, progress rules.ProgressFunc) {
	if !util.NixAvailable() {
		log.Printf("Nix not available, skipping nix build")
		return
	}

	emit(progress, 81, "Building project...", nil)

	cmd := exec.Command("nix", "build",
		"--extra-experimental-features", "nix-command flakes")
	cmd.Dir = projectPath

	if err := streamCommandOutput(cmd, progress, "Compiling project..."); err != nil {
		log.Printf("Warning: nix build failed (non-fatal): %v", err)
		emit(progress, 82, "Build completed with warnings", map[string]string{
			"build": "failed",
		})
		return
	}
	log.Printf("nix build succeeded: %s", projectPath)
	emit(progress, 82, "Build passed", map[string]string{
		"build": "passed",
	})
}

// nixFlakeCheck validates the flake health and structure.
// Non-fatal: only logs warnings on failure.
func (s *Service) nixFlakeCheck(projectPath string) {
	if !util.NixAvailable() {
		return
	}

	cmd := exec.Command("nix", "flake", "check",
		"--extra-experimental-features", "nix-command flakes")
	cmd.Dir = projectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: nix flake check failed (non-fatal): %v\nOutput: %s", err, string(out))
		return
	}
	log.Printf("nix flake check passed: %s", projectPath)
}

// streamCommandOutput запускает команду, стримит stdout+stderr построчно в лог,
// и отправляет прогресс при загрузке nix-пакетов.
func streamCommandOutput(cmd *exec.Cmd, progress rules.ProgressFunc, defaultMsg string) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[nix] %s", line)

		if progress != nil && strings.Contains(line, "copying path") {
			if m := util.NixPkgRe.FindStringSubmatch(line); len(m) > 1 {
				progress(50, "Installing "+strings.TrimSpace(m[1])+"...", nil)
			}
		}
	}

	return cmd.Wait()
}
