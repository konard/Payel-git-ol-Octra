package util

import "testing"

// TestIsInfraPath — issue #75 п.6: служебные файлы окружения/оркестратора не
// должны утекать во вкладку Solution и в чат.
func TestIsInfraPath(t *testing.T) {
	cases := map[string]bool{
		"flake.nix":           true,
		"flake.lock":          true,
		"./flake.nix":         true,
		"sub/flake.nix":       true,
		".octra/context.json": true,
		".octra":              true,
		".git/config":         true,
		"result/bin/app":      true,
		"index.js":            false,
		"src/main.go":         false,
		"package.json":        false,
		"context.json":        false, // только .octra/context.json — служебный
		"README.md":           false,
	}
	for path, want := range cases {
		if got := IsInfraPath(path); got != want {
			t.Errorf("IsInfraPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestIsBinaryPath — бинарные форматы кодируются в base64 при передаче.
func TestIsBinaryPath(t *testing.T) {
	cases := map[string]bool{
		"deck.pptx":   true,
		"doc.docx":    true,
		"sheet.xlsx":  true,
		"report.pdf":  true,
		"logo.png":    true,
		"photo.JPG":   true,
		"archive.zip": true,
		"index.js":    false,
		"main.go":     false,
		"notes.md":    false,
	}
	for path, want := range cases {
		if got := IsBinaryPath(path); got != want {
			t.Errorf("IsBinaryPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestLanguageForPath — подсветка синтаксиса во вкладке Solution.
func TestLanguageForPath(t *testing.T) {
	cases := map[string]string{
		"main.go":     "go",
		"index.js":    "javascript",
		"app.ts":      "typescript",
		"app.tsx":     "typescript",
		"server.py":   "python",
		"style.css":   "css",
		"page.html":   "html",
		"data.json":   "json",
		"README.md":   "markdown",
		"config.yaml": "yaml",
		"deck.pptx":   "binary",
		"unknown.xyz": "plaintext",
	}
	for path, want := range cases {
		if got := LanguageForPath(path); got != want {
			t.Errorf("LanguageForPath(%q) = %q, want %q", path, got, want)
		}
	}
}
