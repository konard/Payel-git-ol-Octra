package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizeProjectName — приводит название проекта к имени папки
func SanitizeProjectName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// IsInfraPath сообщает, что путь — служебный файл сборочного окружения/оркестратора
// (flake.nix, flake.lock, .octra/*, .git/*, result/*), а не результат работы воркера.
// Такие файлы НЕ показываются во вкладке Solution и не стримятся в чат, потому что
// пользователь просил конкретный код, а не инфраструктуру Nix/Octra (issue #75 п.6).
func IsInfraPath(path string) bool {
	clean := strings.TrimPrefix(filepath.ToSlash(path), "./")
	switch strings.ToLower(filepath.Base(clean)) {
	case "flake.nix", "flake.lock":
		return true
	}
	if clean == ".octra" || strings.HasPrefix(clean, ".octra/") {
		return true
	}
	if clean == ".git" || strings.HasPrefix(clean, ".git/") {
		return true
	}
	if clean == "result" || strings.HasPrefix(clean, "result/") {
		return true
	}
	return false
}

// IsBinaryPath сообщает, нужно ли кодировать содержимое файла в base64 при передаче
// во фронтенд (бинарные форматы — презентации, документы, изображения, архивы).
func IsBinaryPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pptx", ".docx", ".xlsx", ".pdf", ".png", ".jpg", ".jpeg", ".gif", ".zip":
		return true
	default:
		return false
	}
}

// LanguageForPath — определяет язык подсветки синтаксиса по расширению файла.
func LanguageForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cc", ".cpp", ".cxx", ".hpp", ".hxx":
		return "cpp"
	case ".css":
		return "css"
	case ".html", ".htm":
		return "html"
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".sh", ".bash":
		return "shell"
	case ".sql":
		return "sql"
	case ".pptx", ".docx", ".xlsx", ".pdf", ".png", ".jpg", ".jpeg", ".gif", ".zip":
		return "binary"
	default:
		return "plaintext"
	}
}

// ValidateFilePath — защищает от path traversal атак и абсолютных путей
func ValidateFilePath(basePath, filename string) (string, error) {
	if filepath.IsAbs(filename) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", filename)
	}
	if strings.Contains(filename, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", filename)
	}
	cleanFilename := filepath.Clean(filename)
	if strings.HasPrefix(cleanFilename, "/") || strings.HasPrefix(cleanFilename, "..") {
		return "", fmt.Errorf("invalid file path: %s", filename)
	}
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}
	fullPath := filepath.Join(absBase, cleanFilename)
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve full path: %w", err)
	}
	if !strings.HasPrefix(absFull, absBase+string(filepath.Separator)) && absFull != absBase {
		return "", fmt.Errorf("path escapes base directory: %s", filename)
	}
	return fullPath, nil
}
