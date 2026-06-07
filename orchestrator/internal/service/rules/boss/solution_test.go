package boss

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"orchestrator/internal/service/rules"
)

func TestBridgePreservesWorkerData(t *testing.T) {
	service := &Service{}
	var got map[string]string

	bridge := service.bridge(func(_ int32, _ string, data map[string]string) {
		got = data
	})
	bridge("Research Manager", 50, "Worker completed", map[string]string{
		"code_files": `[{"path":"solution/result.md","content":"done"}]`,
		"task_type":  TaskTypeResearch,
	})

	if got["current_role"] != "Research Manager" {
		t.Fatalf("current_role = %q, want Research Manager", got["current_role"])
	}
	if got["code_files"] == "" {
		t.Fatalf("code_files was dropped from bridged progress data")
	}
	if got["task_type"] != TaskTypeResearch {
		t.Fatalf("task_type = %q, want %q", got["task_type"], TaskTypeResearch)
	}
}

func TestCollectCodeFilesPayloadIncludesWorkerFiles(t *testing.T) {
	results := []*rules.ManagerResult{
		{
			Role: "Research Manager",
			WorkerResults: []*rules.WorkerResult{
				{
					Role: "Research Worker",
					Files: map[string]string{
						"solution/report.md": "# Report\n",
						"solution/deck.pptx": "binary",
					},
				},
			},
		},
	}

	payload, count := collectCodeFilesPayload(results)
	if count != 2 {
		t.Fatalf("files count = %d, want 2 including filtered binary files", count)
	}
	if payload == "" {
		t.Fatalf("payload is empty")
	}

	var files []streamedSolutionFile
	if err := json.Unmarshal([]byte(payload), &files); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("payload files = %d, want 2 including binary presentation", len(files))
	}

	byPath := map[string]streamedSolutionFile{}
	for _, file := range files {
		byPath[file.Path] = file
	}

	report, ok := byPath["solution/report.md"]
	if !ok {
		t.Fatalf("solution/report.md missing from payload: %#v", files)
	}
	if report.Language != "markdown" {
		t.Fatalf("report language = %q, want markdown", report.Language)
	}
	if report.Encoding != "" {
		t.Fatalf("report encoding = %q, want empty text encoding", report.Encoding)
	}
	if report.ManagerRole != "Research Manager" || report.WorkerRole != "Research Worker" {
		t.Fatalf("roles = manager %q worker %q, want Research Manager/Research Worker", report.ManagerRole, report.WorkerRole)
	}

	deck, ok := byPath["solution/deck.pptx"]
	if !ok {
		t.Fatalf("solution/deck.pptx missing from payload: %#v", files)
	}
	if deck.Language != "binary" {
		t.Fatalf("deck language = %q, want binary", deck.Language)
	}
	if deck.Encoding != "base64" {
		t.Fatalf("deck encoding = %q, want base64", deck.Encoding)
	}
	decoded, err := base64.StdEncoding.DecodeString(deck.Content)
	if err != nil {
		t.Fatalf("deck content is not base64: %v", err)
	}
	if string(decoded) != "binary" {
		t.Fatalf("decoded deck = %q, want original binary content", string(decoded))
	}
}

