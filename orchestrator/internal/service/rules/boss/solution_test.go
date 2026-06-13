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

// TestCollectCodeFilesPayloadSkipsInfra — корневой тест issue #75 п.6: фронтенду
// должны уходить РЕАЛЬНЫЕ файлы кода, а не служебное окружение (flake.nix,
// flake.lock, .octra/context.json). Пользователь спросил: «зачем мне flake.nix и
// context.json, если я просил express hello world».
func TestCollectCodeFilesPayloadSkipsInfra(t *testing.T) {
	results := []*rules.ManagerResult{
		{
			Role: "backend-manager",
			WorkerResults: []*rules.WorkerResult{
				{
					Role: "backend",
					Files: map[string]string{
						"flake.nix":           "{ }",
						"flake.lock":          "{}",
						".octra/context.json": `{"x":1}`,
						"result/bin/app":      "elf",
						"index.js":            "const express = require('express')",
						"package.json":        `{"name":"app"}`,
					},
				},
			},
		},
	}

	payload, total := collectCodeFilesPayload(results)
	if payload == "" {
		t.Fatal("expected non-empty payload (real code files present)")
	}

	var files []streamedSolutionFile
	if err := json.Unmarshal([]byte(payload), &files); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}

	got := map[string]bool{}
	for _, f := range files {
		got[f.Path] = true
	}
	for _, infra := range []string{"flake.nix", "flake.lock", ".octra/context.json", "result/bin/app"} {
		if got[infra] {
			t.Errorf("infra file %q must NOT be in payload", infra)
		}
	}
	if !got["index.js"] || !got["package.json"] {
		t.Errorf("expected real code files in payload, got %v", got)
	}
	// total считает только показываемые файлы (инфраструктура исключена).
	if total != 2 {
		t.Errorf("total = %d, want 2 (index.js + package.json)", total)
	}
}

// TestCollectCodeFilesPayloadOnlyInfra — если воркер выдал ТОЛЬКО инфраструктуру,
// payload пустой (показывать нечего), и это сигнал, что код не сгенерировался.
func TestCollectCodeFilesPayloadOnlyInfra(t *testing.T) {
	results := []*rules.ManagerResult{
		{
			Role: "m",
			WorkerResults: []*rules.WorkerResult{
				{Role: "backend", Files: map[string]string{
					"flake.nix":           "{ }",
					".octra/context.json": "{}",
				}},
			},
		},
	}
	payload, total := collectCodeFilesPayload(results)
	if payload != "" {
		t.Errorf("expected empty payload when only infra files present, got %q", payload)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

