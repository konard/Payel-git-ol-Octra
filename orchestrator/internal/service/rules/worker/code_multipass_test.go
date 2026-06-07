package worker

import (
	"strings"
	"testing"
)

func TestParseMultiFileResponseStopsFileBeforeCommands(t *testing.T) {
	content := `=== FILE: cmd/proxy/main.go ===
package main

func main() {}
=== COMMANDS ===
mkdir -p cmd config fallback docs
go mod tidy
`

	files, commands := parseMultiFileResponse(content)

	mainFile, ok := files["cmd/proxy/main.go"]
	if !ok {
		t.Fatalf("cmd/proxy/main.go was not parsed")
	}
	if strings.Contains(mainFile, "=== COMMANDS ===") || strings.Contains(mainFile, "go mod tidy") {
		t.Fatalf("commands leaked into file content:\n%s", mainFile)
	}
	if len(commands) != 2 {
		t.Fatalf("commands = %#v, want 2 commands", commands)
	}
	if commands[0] != "mkdir -p cmd config fallback docs" || commands[1] != "go mod tidy" {
		t.Fatalf("commands = %#v, want mkdir/go mod tidy", commands)
	}
}

