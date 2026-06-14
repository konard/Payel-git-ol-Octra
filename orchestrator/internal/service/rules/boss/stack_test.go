package boss

import "testing"

// Issue #77: the user asked for an "Express.js" server but the boss/AI guessed
// the "go" stack, so the worker wrote JavaScript into main.go. detectTechStack
// must recognize Express as Node.js, and stackMentioned must report that a
// "go" guess conflicts with an explicitly requested Node.js stack so the boss
// can override it.
func TestDetectTechStackExpress(t *testing.T) {
	got := detectTechStack("Сделать сервер на Express.js который отвечает hello world", "")
	if len(got) == 0 || got[0] != "nodejs" {
		t.Fatalf("detectTechStack(express) = %v, want [nodejs ...]", got)
	}
}

func TestStackMentioned(t *testing.T) {
	cases := []struct {
		name     string
		detected []string
		ai       string
		want     bool
	}{
		{"express vs go -> conflict", []string{"nodejs"}, "go", false},
		{"express vs nodejs -> match", []string{"nodejs"}, "nodejs", true},
		{"express vs node alias -> match", []string{"nodejs"}, "express", true},
		{"go vs golang alias -> match", []string{"go"}, "golang", true},
		{"python vs flask alias -> match", []string{"python"}, "flask", true},
		{"multi detected, one matches", []string{"python", "go"}, "go", true},
		{"multi detected, none match", []string{"python", "nodejs"}, "rust", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stackMentioned(tc.detected, tc.ai); got != tc.want {
				t.Fatalf("stackMentioned(%v, %q) = %v, want %v", tc.detected, tc.ai, got, tc.want)
			}
		})
	}
}
