package worker

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// TestContainsSourceCode проверяет корневую причину issue #75 п.6: tool-режим,
// который наскаффолдил только манифесты/локи (package.json, package-lock.json,
// flake.nix), НЕ должен считаться «есть код» — иначе реальный код фичи (express
// index.js) никогда не сгенерируется, и фронтенду уедут только служебные файлы.
func TestContainsSourceCode(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  bool
	}{
		{
			name: "only manifests and infra (npm init -y result)",
			files: map[string]string{
				"package.json":      `{"name":"app"}`,
				"package-lock.json": `{}`,
				"flake.nix":         "{ }",
				".gitignore":        "node_modules",
			},
			want: false,
		},
		{
			name: "has real express source",
			files: map[string]string{
				"package.json": `{"name":"app"}`,
				"index.js":     "const express = require('express'); express().listen(3000)",
			},
			want: true,
		},
		{
			name: "source file present but empty does not count",
			files: map[string]string{
				"index.js": "   ",
			},
			want: false,
		},
		{
			name:  "no files",
			files: map[string]string{},
			want:  false,
		},
		{
			name: "go source",
			files: map[string]string{
				"main.go": "package main\nfunc main() {}",
			},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := containsSourceCode(tc.files); got != tc.want {
				t.Fatalf("containsSourceCode(%v) = %v, want %v", tc.files, got, tc.want)
			}
		})
	}
}

// TestLiveCodeFilesPayloadSkipsInfra проверяет issue #75 п.1+п.6: инфраструктурные
// файлы (flake.nix, .octra/context.json) не должны стримиться во вкладку Solution.
func TestLiveCodeFilesPayloadSkipsInfra(t *testing.T) {
	infra := []string{"flake.nix", "flake.lock", ".octra/context.json", "result/bin/app", ""}
	for _, p := range infra {
		if got := liveCodeFilesPayload("backend", p, "whatever"); got != "" {
			t.Fatalf("liveCodeFilesPayload(%q) = %q, want empty (infra must be skipped)", p, got)
		}
	}
}

// TestLiveCodeFilesPayloadSourceFile проверяет, что для реального исходника
// формируется корректный инкрементальный code_files-JSON со status "streaming"
// (issue #75 п.1: файлы появляются во вкладке Solution по мере создания).
func TestLiveCodeFilesPayloadSourceFile(t *testing.T) {
	out := liveCodeFilesPayload("backend", "index.js", "console.log('hi')")
	if out == "" {
		t.Fatal("expected non-empty payload for source file")
	}
	var files []liveCodeFile
	if err := json.Unmarshal([]byte(out), &files); err != nil {
		t.Fatalf("payload is not valid JSON: %v\n%s", err, out)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.Path != "index.js" {
		t.Errorf("Path = %q, want index.js", f.Path)
	}
	if f.Content != "console.log('hi')" {
		t.Errorf("Content = %q, want raw source", f.Content)
	}
	if f.Language != "javascript" {
		t.Errorf("Language = %q, want javascript", f.Language)
	}
	if f.Status != "streaming" {
		t.Errorf("Status = %q, want streaming", f.Status)
	}
	if f.WorkerRole != "backend" {
		t.Errorf("WorkerRole = %q, want backend", f.WorkerRole)
	}
	if f.Encoding != "" {
		t.Errorf("Encoding = %q, want empty for text file", f.Encoding)
	}
}

// TestLiveCodeFilesPayloadBinaryBase64 проверяет, что бинарное содержимое
// (например .pptx-презентация) кодируется в base64 для безопасной передачи.
func TestLiveCodeFilesPayloadBinaryBase64(t *testing.T) {
	raw := "\x00\x01binary\xff"
	out := liveCodeFilesPayload("designer", "slides.pptx", raw)
	if out == "" {
		t.Fatal("expected non-empty payload for binary file")
	}
	var files []liveCodeFile
	if err := json.Unmarshal([]byte(out), &files); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	f := files[0]
	if f.Encoding != "base64" {
		t.Fatalf("Encoding = %q, want base64", f.Encoding)
	}
	decoded, err := base64.StdEncoding.DecodeString(f.Content)
	if err != nil {
		t.Fatalf("content is not valid base64: %v", err)
	}
	if string(decoded) != raw {
		t.Errorf("decoded content = %q, want %q", string(decoded), raw)
	}
}
