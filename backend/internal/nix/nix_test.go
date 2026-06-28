package nix

import (
	"context"
	"strings"
	"testing"

	"backend/internal/model"
)

// fakeRunner records the commands it is asked to run.
type fakeRunner struct {
	commands []string
	err      error
}

func (f *fakeRunner) Run(_ context.Context, _ string, command string) ([]byte, error) {
	f.commands = append(f.commands, command)
	return []byte("ok"), f.err
}

func TestCreateEnvironmentWithoutCLI(t *testing.T) {
	r := &fakeRunner{}
	m := NewManager(t.TempDir(), r)

	if err := m.CreateEnvironment(context.Background(), "user1", ""); err != nil {
		t.Fatalf("CreateEnvironment: %v", err)
	}
	if len(r.commands) != 0 {
		t.Fatalf("expected no install commands for empty CLI, got %v", r.commands)
	}
}

func TestCreateEnvironmentUsesGlobalCLIProfile(t *testing.T) {
	r := &fakeRunner{}
	m := NewManager(t.TempDir(), r)

	if err := m.CreateEnvironment(context.Background(), "user1", "claude-code"); err != nil {
		t.Fatalf("CreateEnvironment: %v", err)
	}
	if len(r.commands) != 0 {
		t.Fatalf("expected no per-environment CLI install, got %v", r.commands)
	}
}

func TestCreateEnvironmentRejectsUnknownCLI(t *testing.T) {
	r := &fakeRunner{}
	m := NewManager(t.TempDir(), r)

	err := m.CreateEnvironment(context.Background(), "user1", "bad cli; rm -rf /")
	if err == nil {
		t.Fatal("expected unknown CLI error")
	}
}

func TestInstallSkillByType(t *testing.T) {
	cases := []struct {
		name     string
		skill    model.Skill
		wantRun  bool
		contains string
	}{
		{"builtin", model.Skill{Name: "fs", Type: model.SkillBuiltin}, false, ""},
		{"nixpkgs", model.Skill{Name: "ripgrep", Type: model.SkillNixpkgs}, true, "nixpkgs#ripgrep"},
		{"nixpkgs-attr", model.Skill{Name: "rg", Type: model.SkillNixpkgs, InstallCmd: "ripgrep"}, true, "nixpkgs#ripgrep"},
		{"custom", model.Skill{Name: "x", Type: model.SkillCustom, InstallCmd: "echo hi"}, true, "echo hi"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRunner{}
			m := NewManager(t.TempDir(), r)
			if err := m.InstallSkill(context.Background(), "u", tc.skill); err != nil {
				t.Fatalf("InstallSkill: %v", err)
			}
			if tc.wantRun {
				if len(r.commands) != 1 {
					t.Fatalf("expected 1 command, got %v", r.commands)
				}
				if !strings.Contains(r.commands[0], tc.contains) {
					t.Fatalf("command %q does not contain %q", r.commands[0], tc.contains)
				}
				if tc.skill.Type == model.SkillNixpkgs {
					if !strings.Contains(r.commands[0], "--profile") {
						t.Fatalf("expected nixpkgs skill command to install into explicit profile: %q", r.commands[0])
					}
					if strings.Contains(r.commands[0], "|| true") {
						t.Fatalf("skill install command must not mask failures: %q", r.commands[0])
					}
				}
			} else if len(r.commands) != 0 {
				t.Fatalf("expected no command for builtin, got %v", r.commands)
			}
		})
	}
}

func TestProvisionSystemInstallsCLIsIntoPersistentProfile(t *testing.T) {
	r := &fakeRunner{}
	baseDir := t.TempDir()
	m := NewManager(baseDir, r)

	if err := m.ProvisionSystem(context.Background()); err != nil {
		t.Fatalf("ProvisionSystem: %v", err)
	}
	if len(r.commands) == 0 {
		t.Fatal("expected built-in CLI provisioning commands")
	}
	foundNixProfileInstall := false
	for _, cmd := range r.commands {
		if strings.Contains(cmd, "profile install") {
			foundNixProfileInstall = true
			if !strings.Contains(cmd, ".system/nix-profile") {
				t.Fatalf("nix CLI install must target persistent system profile: %q", cmd)
			}
		}
		if strings.Contains(cmd, "|| true") {
			t.Fatalf("provisioning command must not mask failures: %q", cmd)
		}
	}
	if !foundNixProfileInstall {
		t.Fatal("expected at least one nix profile install command")
	}
}

func TestEnvPath(t *testing.T) {
	m := NewManager("/base", nil)
	if got := m.EnvPath("abc"); got != "/base/abc" {
		t.Fatalf("EnvPath = %q", got)
	}
}
