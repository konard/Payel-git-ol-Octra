package util

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// NixPkgRe извлекает имя пакета из nix store path: /nix/store/hash-name-version' from
var NixPkgRe = regexp.MustCompile(`/nix/store/[^-]+-(.+?)' from`)

// NixAvailable проверяет доступность nix в системе
func NixAvailable() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// NixProfilePath returns the path to the project's nix profile.
// The profile caches the built dev shell so subsequent nix develop calls
// skip flake evaluation and reuse the cached derivation.
func NixProfilePath(projectPath string) string {
	return filepath.Join(projectPath, ".octra", "nix-profile")
}

// NixDevelopCmd оборачивает команду в nix develop, если nix доступен.
// Использует --profile для кэширования собранного окружения между вызовами.
// Возвращает готовый к запуску *exec.Cmd с установленным Dir.
func NixDevelopCmd(workDir, shellCmd string) *exec.Cmd {
	if !NixAvailable() {
		cmd := exec.Command("sh", "-c", shellCmd)
		cmd.Dir = workDir
		return cmd
	}

	profilePath := NixProfilePath(workDir)
	ensureProfileDir(profilePath)

	cmd := exec.Command("nix", "develop",
		"--extra-experimental-features", "nix-command flakes",
		"--profile", profilePath,
		"--command", "sh", "-c", shellCmd)
	cmd.Dir = workDir
	return cmd
}

// NixBatchedCmd runs multiple shell commands in a single nix develop session.
// All commands are joined with && and executed together.
func NixBatchedCmd(workDir string, commands []string) *exec.Cmd {
	if len(commands) == 0 {
		return nil
	}
	// Filter empties
	var filtered []string
	for _, c := range commands {
		if strings.TrimSpace(c) != "" {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	shellCmd := strings.Join(filtered, " && ")
	return NixDevelopCmd(workDir, shellCmd)
}

// NixBatchedCmdStr runs a list of commands joined by && in a single nix develop session.
func ensureProfileDir(profilePath string) {
	dir := filepath.Dir(profilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Warning: failed to create profile dir %s: %v", dir, err)
	}
}

// EnsureNixProfile ensures the project's nix profile exists and is up to date.
// Returns the profile path or an error.
func EnsureNixProfile(projectPath string) (string, error) {
	profilePath := NixProfilePath(projectPath)
	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return "", fmt.Errorf("mkdir profile dir: %w", err)
	}

	// If profile already exists, nix develop --profile will check
	// if it's stale and rebuild only if needed
	cmd := exec.Command("nix", "develop",
		"--extra-experimental-features", "nix-command flakes",
		"--profile", profilePath,
		"--command", "true")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("nix develop --profile failed: %w\n%s", err, string(out))
	}
	return profilePath, nil
}
