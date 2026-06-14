package prompts

import (
	"strings"
	"testing"
)

// Issue #77: a "server in Express.js" task was generated with the Node http
// stdlib and saved into main.go. The root cause included prompts that used the
// tech-stack name itself as a file extension (".nodejs") and hardcoded a
// Go-only bias. These tests lock in the language metadata that fixes that.
func TestLangExtensionAndMainFile(t *testing.T) {
	cases := []struct {
		stack    string
		ext      string
		mainFile string
	}{
		{"go", "go", "main.go"},
		{"golang", "go", "main.go"},
		{"Go", "go", "main.go"}, // case-insensitive
		{"nodejs", "js", "index.js"},
		{"node", "js", "index.js"},
		{"express", "js", "index.js"},
		{"javascript", "js", "index.js"},
		{"typescript", "ts", "index.ts"},
		{"python", "py", "main.py"},
		{"flask", "py", "app.py"},
		{"rust", "rs", "src/main.rs"},
		{"php", "php", "index.php"},
		{"java", "java", "Main.java"},
		{"dotnet", "cs", "Program.cs"},
	}
	for _, tc := range cases {
		t.Run(tc.stack, func(t *testing.T) {
			if got := LangExtension(tc.stack); got != tc.ext {
				t.Errorf("LangExtension(%q) = %q, want %q", tc.stack, got, tc.ext)
			}
			if got := LangMainFile(tc.stack); got != tc.mainFile {
				t.Errorf("LangMainFile(%q) = %q, want %q", tc.stack, got, tc.mainFile)
			}
		})
	}
}

func TestLangUnknownStack(t *testing.T) {
	if got := LangExtension("brainfuck"); got != "" {
		t.Errorf("LangExtension(unknown) = %q, want empty", got)
	}
	if got := LangMainFile("brainfuck"); got != "main.txt" {
		t.Errorf("LangMainFile(unknown) = %q, want main.txt", got)
	}
	if got := LangDisplay("brainfuck"); got != "brainfuck" {
		t.Errorf("LangDisplay(unknown) = %q, want passthrough", got)
	}
}

// The worker file-planning prompt must never tell a Node.js worker to use a
// ".nodejs" extension nor put JavaScript into a .go file.
func TestWorkerPlanFilesNodeExtension(t *testing.T) {
	p := WorkerPlanFiles("backend", "API", "Express.js hello world server", "", "nodejs", "")
	if strings.Contains(p, ".nodejs") {
		t.Fatalf("plan prompt leaks .nodejs extension:\n%s", p)
	}
	if !strings.Contains(p, ".js") {
		t.Fatalf("plan prompt should hint the .js extension:\n%s", p)
	}
}

// The multi-pass codegen prompt for a Node.js task must use real JS conventions
// and must NOT carry the old Go-only bias ("NOT JavaScript") or the bogus
// ".nodejs" extension that caused issue #77.
func TestWorkerMultiPassCodeNode(t *testing.T) {
	p := WorkerMultiPassCode("backend", "API", "Express.js hello world server", "", "nodejs")
	if strings.Contains(p, "NOT JavaScript") {
		t.Fatalf("multipass prompt still forbids JavaScript for a Node.js task:\n%s", p)
	}
	if strings.Contains(p, ".nodejs") {
		t.Fatalf("multipass prompt leaks .nodejs extension:\n%s", p)
	}
	if !strings.Contains(p, "index.js") {
		t.Fatalf("multipass prompt should reference a .js entrypoint:\n%s", p)
	}
	if !strings.Contains(p, "Express") {
		t.Fatalf("multipass prompt should keep the requested framework guidance:\n%s", p)
	}
}

// Go tasks must still get Go conventions after the refactor.
func TestWorkerMultiPassCodeGo(t *testing.T) {
	p := WorkerMultiPassCode("backend", "service", "Build a TCP proxy", "", "go")
	if !strings.Contains(p, "main.go") {
		t.Fatalf("multipass prompt should reference main.go for a Go task:\n%s", p)
	}
}
