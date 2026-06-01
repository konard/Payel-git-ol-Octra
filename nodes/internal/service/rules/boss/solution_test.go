package boss

import (
	"encoding/json"
	"testing"

	"nodes/internal/service/rules"
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
	if len(files) != 1 {
		t.Fatalf("payload files = %d, want 1 non-binary file", len(files))
	}
	if files[0].Path != "solution/report.md" {
		t.Fatalf("path = %q, want solution/report.md", files[0].Path)
	}
	if files[0].Language != "markdown" {
		t.Fatalf("language = %q, want markdown", files[0].Language)
	}
	if files[0].ManagerRole != "Research Manager" || files[0].WorkerRole != "Research Worker" {
		t.Fatalf("roles = manager %q worker %q, want Research Manager/Research Worker", files[0].ManagerRole, files[0].WorkerRole)
	}
}
