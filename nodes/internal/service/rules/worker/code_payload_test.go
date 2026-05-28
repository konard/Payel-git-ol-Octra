package worker

import (
	"encoding/json"
	"testing"
)

func TestBuildCodeFilesPayloadIncludesGeneratedFiles(t *testing.T) {
	payload := buildCodeFilesPayload(map[string]string{
		"cmd/api/main.go": "package main\n\nfunc main() {}\n",
		"README.md":       "# Demo\n",
	}, "Go Backend Developer", "Platform Manager")

	if payload == "" {
		t.Fatal("expected non-empty code_files payload")
	}

	var files []streamedCodeFile
	if err := json.Unmarshal([]byte(payload), &files); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	if files[0].Path != "README.md" {
		t.Fatalf("expected files to be sorted by path, got first path %q", files[0].Path)
	}
	if files[1].Path != "cmd/api/main.go" {
		t.Fatalf("expected Go file path, got %q", files[1].Path)
	}
	if files[1].Content != "package main\n\nfunc main() {}\n" {
		t.Fatalf("unexpected Go file content: %q", files[1].Content)
	}
	if files[1].Language != "go" {
		t.Fatalf("expected go language, got %q", files[1].Language)
	}
	if files[1].WorkerRole != "Go Backend Developer" || files[1].ManagerRole != "Platform Manager" {
		t.Fatalf("roles were not preserved: %#v", files[1])
	}
	if files[1].Status != "streaming" {
		t.Fatalf("expected streaming status, got %q", files[1].Status)
	}
}
