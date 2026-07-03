package nix

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"backend/internal/model"
)

type fakeRunner struct {
	commands []string
	err      error
}

func (f *fakeRunner) Run(_ context.Context, _ string, command string) ([]byte, error) {
	f.commands = append(f.commands, command)
	return []byte("ok"), f.err
}

func TestCreateEnvironmentCreatesDirectory(t *testing.T) {
	r := &fakeRunner{}
	m := NewManager(t.TempDir(), r)

	if err := m.CreateEnvironment(context.Background(), "user1"); err != nil {
		t.Fatalf("CreateEnvironment: %v", err)
	}
	if len(r.commands) != 0 {
		t.Fatalf("expected no install commands, got %v", r.commands)
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

func TestEnvPath(t *testing.T) {
	m := NewManager("/base", nil)
	want := filepath.Join("/base", "abc")
	if got := m.EnvPath("abc"); got != want {
		t.Fatalf("EnvPath = %q, want %q", got, want)
	}
}
