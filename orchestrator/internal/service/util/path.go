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

