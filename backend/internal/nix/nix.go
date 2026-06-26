// Package nix manages per-user isolated environments. Each user gets their own
// directory (and Nix profile) into which the chosen AI CLI and skills are
// installed. The actual command execution is abstracted behind Runner so the
// Manager can be unit-tested without a real Nix toolchain.
package nix

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"backend/internal/model"
)

// Runner executes a shell command inside a working directory and returns its
// combined output.
type Runner interface {
	Run(ctx context.Context, workDir, command string) ([]byte, error)
}

// ExecRunner runs commands via the OS shell, wrapping them in `nix develop`
// when the Nix toolchain is available so they execute inside the project's
// pinned environment.
type ExecRunner struct{}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, workDir, command string) ([]byte, error) {
	var cmd *exec.Cmd
	if nixAvailable() {
		profile := filepath.Join(workDir, ".octra", "nix-profile")
		_ = os.MkdirAll(filepath.Dir(profile), 0o755)
		cmd = exec.CommandContext(ctx, "nix", "develop",
			"--extra-experimental-features", "nix-command flakes",
			"--profile", profile,
			"--command", "sh", "-c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = workDir
	return cmd.CombinedOutput()
}

func nixAvailable() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// Manager creates and provisions per-user environments.
type Manager struct {
	baseDir string
	runner  Runner
}

// NewManager returns a Manager rooted at baseDir. If runner is nil an
// ExecRunner is used.
func NewManager(baseDir string, runner Runner) *Manager {
	if runner == nil {
		runner = ExecRunner{}
	}
	return &Manager{baseDir: baseDir, runner: runner}
}

// EnvPath returns the on-disk path of a user's environment.
func (m *Manager) EnvPath(userID string) string {
	return filepath.Join(m.baseDir, userID)
}

// CreateEnvironment ensures the user's environment directory exists and, when a
// CLI is requested, installs it. Installing the CLI is idempotent.
func (m *Manager) CreateEnvironment(ctx context.Context, userID string, cli model.CLIType) error {
	path := m.EnvPath(userID)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create env dir: %w", err)
	}
	if cli == "" {
		// No CLI: the environment is only used for proxy-mode requests.
		return nil
	}
	pkg := cliPackage(cli)
	if pkg == "" {
		return fmt.Errorf("unknown cli %q", cli)
	}
	cmd := fmt.Sprintf("nix profile install nixpkgs#%s || true", pkg)
	if out, err := m.runner.Run(ctx, path, cmd); err != nil {
		return fmt.Errorf("install cli %s: %w\n%s", cli, err, string(out))
	}
	return nil
}

// InstallSkill provisions a single skill into the user's environment. The
// install strategy depends on the skill type.
func (m *Manager) InstallSkill(ctx context.Context, userID string, skill model.Skill) error {
	path := m.EnvPath(userID)
	cmd := skillInstallCommand(skill)
	if cmd == "" {
		// built-in skills need no provisioning.
		return nil
	}
	if out, err := m.runner.Run(ctx, path, cmd); err != nil {
		return fmt.Errorf("install skill %s: %w\n%s", skill.Name, err, string(out))
	}
	return nil
}

// skillInstallCommand derives the shell command used to install a skill.
func skillInstallCommand(skill model.Skill) string {
	switch skill.Type {
	case model.SkillBuiltin:
		return ""
	case model.SkillNixpkgs:
		attr := skill.InstallCmd
		if attr == "" {
			attr = skill.Name
		}
		return fmt.Sprintf("nix profile install nixpkgs#%s || true", attr)
	case model.SkillCustom:
		return skill.InstallCmd
	default:
		return strings.TrimSpace(skill.InstallCmd)
	}
}

// cliPackage maps a CLI identifier to its nixpkgs attribute.
func cliPackage(cli model.CLIType) string {
	switch cli {
	case "claude-code":
		return "claude-code"
	case "opencode":
		return "opencode"
	case "codex":
		return "codex"
	default:
		// Allow arbitrary CLIs as long as the name is a plausible attribute.
		s := string(cli)
		if s != "" && !strings.ContainsAny(s, " \t\n;|&$`") {
			return s
		}
		return ""
	}
}
