// Package nix manages per-user isolated environments. Each user gets their own
// directory (and Nix profile) into which the chosen AI CLI and skills are
// installed. The actual command execution is abstracted behind Runner so the
// Manager can be unit-tested without a real Nix toolchain.
package nix

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"backend/internal/cli"
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

// CreateEnvironment ensures the user's environment directory exists. CLI
// packages are provisioned into the shared system profile by ProvisionSystem so
// each environment stays focused on per-profile dependencies like skills.
func (m *Manager) CreateEnvironment(ctx context.Context, userID string, cli model.CLIType) error {
	path := m.EnvPath(userID)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create env dir: %w", err)
	}
	if cli == "" {
		// No CLI: the environment is only used for proxy-mode requests.
		return nil
	}
	if cliPackage(cli) == "" {
		return fmt.Errorf("unknown cli %q", cli)
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
		// built-in skills need no provisioning.
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

// cliPackage maps a CLI identifier to its nixpkgs attribute.
func cliPackage(ct model.CLIType) string {
	if attr := cli.NixpkgsAttr(string(ct)); attr != "" {
		return attr
	}
	// Allow arbitrary CLIs as long as the name is a plausible attribute.
	s := string(ct)
	if s != "" && !strings.ContainsAny(s, " \t\n;|&$`") {
		return s
	}
	return ""
}

// ProvisionSystem installs every built-in CLI into the default Nix user profile
// so they are available globally (no per-environment install needed).
func (m *Manager) ProvisionSystem(ctx context.Context) error {
	workDir := filepath.Join(m.baseDir, ".system")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("provision workdir: %w", err)
	}
	profile := filepath.Join(workDir, "nix-profile")

	var lastErr error
	for _, pkg := range cli.BuiltinCLIs() {
		cmd := pkg.InstallCmd
		if attr := pkg.NixAttr; attr != "" {
			cmd = fmt.Sprintf("nix --extra-experimental-features %s profile install --profile %s %s", shellQuote("nix-command flakes"), shellQuote(profile), shellQuote("nixpkgs#"+attr))
		}
		if cmd == "" {
			continue
		}
		if out, err := m.runner.Run(ctx, workDir, cmd); err != nil {
			err = fmt.Errorf("provision %s: %w\n%s", pkg.Name, err, string(out))
			log.Printf("nix provision (non-fatal): %v", err)
			lastErr = err
		}
	}
	return lastErr
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
