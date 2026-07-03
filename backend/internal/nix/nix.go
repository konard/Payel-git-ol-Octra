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

// ExecRunner runs commands via the OS shell inside an isolated HOME. Nix
// profile bin directories are prepended to PATH so provisioned packages are
// visible without requiring a flake in the working directory.
type ExecRunner struct{}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, workDir, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir

	homeDir := filepath.Join(workDir, "home")
	os.MkdirAll(filepath.Join(homeDir, ".config"), 0o755)
	os.MkdirAll(filepath.Join(homeDir, ".local", "share"), 0o755)
	cmd.Env = prependPath(os.Environ(), profileBinPaths(workDir))
	cmd.Env = append(cmd.Env,
		"HOME="+homeDir,
		"XDG_CONFIG_HOME="+filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(homeDir, ".local", "share"),
	)

	return cmd.CombinedOutput()
}

// Available reports whether the Nix toolchain is present on $PATH.
func Available() bool {
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

// CreateEnvironment ensures the user's environment directory exists.
func (m *Manager) CreateEnvironment(ctx context.Context, userID string) error {
	path := m.EnvPath(userID)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create env dir: %w", err)
	}
	return nil
}

// InstallCLI installs a CLI binary into the user's Nix profile. If cmd is
// non-empty it is run as a shell command; otherwise attr is installed from
// nixpkgs via `nix profile install`.
func (m *Manager) InstallCLI(ctx context.Context, userID, attr, cmd string) error {
	path := m.EnvPath(userID)
	profile := filepath.Join(path, ".octra", "nix-profile")
	var shellCmd string
	if cmd != "" {
		shellCmd = cmd
	} else if attr != "" {
		shellCmd = fmt.Sprintf("nix --extra-experimental-features %s profile install --profile %s %s", shellQuote("nix-command flakes"), shellQuote(profile), shellQuote("nixpkgs#"+attr))
	} else {
		return nil
	}
	if out, err := m.runner.Run(ctx, path, shellCmd); err != nil {
		return fmt.Errorf("install cli: %w\n%s", err, string(out))
	}
	return nil
}

// InstallSkill provisions a single skill into the user's environment. The
// install strategy depends on the skill type.
func (m *Manager) InstallSkill(ctx context.Context, userID string, skill model.Skill) error {
	path := m.EnvPath(userID)
	profile := filepath.Join(path, ".octra", "nix-profile")
	cmd := skillInstallCommand(skill, profile)
	if cmd == "" {
		return nil
	}
	if out, err := m.runner.Run(ctx, path, cmd); err != nil {
		return fmt.Errorf("install skill %s: %w\n%s", skill.Name, err, string(out))
	}
	return nil
}

// skillInstallCommand derives the shell command used to install a skill.
func skillInstallCommand(skill model.Skill, profile string) string {
	switch skill.Type {
	case model.SkillBuiltin:
		return ""
	case model.SkillNixpkgs:
		attr := skill.InstallCmd
		if attr == "" {
			attr = skill.Name
		}
		return fmt.Sprintf("nix --extra-experimental-features %s profile install --profile %s %s", shellQuote("nix-command flakes"), shellQuote(profile), shellQuote("nixpkgs#"+attr))
	case model.SkillCustom:
		return skill.InstallCmd
	default:
		return strings.TrimSpace(skill.InstallCmd)
	}
}


func profileBinPaths(workDir string) []string {
	baseDir := filepath.Dir(workDir)
	return []string{
		filepath.Join(workDir, ".octra", "nix-profile", "bin"),
		filepath.Join(workDir, "home", ".nix-profile", "bin"),
		filepath.Join(baseDir, ".system", "nix-profile", "bin"),
		filepath.Join(baseDir, ".system", "home", ".nix-profile", "bin"),
	}
}

func prependPath(env []string, dirs []string) []string {
	currentPath := os.Getenv("PATH")
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			currentPath = strings.TrimPrefix(entry, "PATH=")
			break
		}
	}
	pathValue := strings.Join(append(dirs, currentPath), string(os.PathListSeparator))
	next := make([]string, 0, len(env)+1)
	added := false
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			next = append(next, "PATH="+pathValue)
			added = true
			continue
		}
		next = append(next, entry)
	}
	if !added {
		next = append(next, "PATH="+pathValue)
	}
	return next
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
