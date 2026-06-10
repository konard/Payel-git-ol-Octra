package boss

import (
	"log"
	"os/exec"
	"path/filepath"

	"orchestrator/internal/service/rules"
)

// ensureFlakeLock generates flake.lock by running nix flake lock in the project
// directory. This pins all dependency versions (nixpkgs revision, etc.) for
// reproducible builds. Non-fatal: if nix is unavailable or the command fails,
// the pipeline continues with a warning.
func (s *Service) ensureFlakeLock(projectPath string) {
	if !nixAvailable() {
		log.Printf("Nix not available, skipping flake.lock generation")
		return
	}

	log.Printf("Generating flake.lock for project: %s", projectPath)
	cmd := exec.Command("nix", "flake", "lock",
		"--extra-experimental-features", "nix-command flakes")
	cmd.Dir = projectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: nix flake lock failed (non-fatal): %v\nOutput: %s", err, string(out))
		return
	}
	log.Printf("flake.lock generated at: %s", filepath.Join(projectPath, "flake.lock"))
}

// nixBuild runs nix build to verify the project compiles inside Nix's isolated
// build environment. This is a non-fatal validation step: if it fails, the
// pipeline continues but logs a warning (nix may not be available, or there
// may be environment-specific issues).
func (s *Service) nixBuild(projectPath string, progress rules.ProgressFunc) {
	if !nixAvailable() {
		log.Printf("Nix not available, skipping nix build")
		return
	}

	emit(progress, 81, "Building project...", nil)

	cmd := exec.Command("nix", "build",
		"--extra-experimental-features", "nix-command flakes")
	cmd.Dir = projectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: nix build failed (non-fatal): %v\nOutput: %s", err, string(out))
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
	if !nixAvailable() {
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

// nixAvailable checks whether the nix binary is installed and usable.
func nixAvailable() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}
