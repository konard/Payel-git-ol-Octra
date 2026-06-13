package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizeProjectName — приводит название проекта к имени папки
func SanitizeProjectName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = transliterateCyrillic(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func transliterateCyrillic(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'а':
			b.WriteString("a")
		case 'б':
			b.WriteString("b")
		case 'в':
			b.WriteString("v")
		case 'г':
			b.WriteString("g")
		case 'д':
			b.WriteString("d")
		case 'е', 'э':
			b.WriteString("e")
		case 'ё':
			b.WriteString("yo")
		case 'ж':
			b.WriteString("zh")
		case 'з':
			b.WriteString("z")
		case 'и':
			b.WriteString("i")
		case 'й':
			b.WriteString("y")
		case 'к':
			b.WriteString("k")
		case 'л':
			b.WriteString("l")
		case 'м':
			b.WriteString("m")
		case 'н':
			b.WriteString("n")
		case 'о':
			b.WriteString("o")
		case 'п':
			b.WriteString("p")
		case 'р':
			b.WriteString("r")
		case 'с':
			b.WriteString("s")
		case 'т':
			b.WriteString("t")
		case 'у':
			b.WriteString("u")
		case 'ф':
			b.WriteString("f")
		case 'х':
			b.WriteString("kh")
		case 'ц':
			b.WriteString("ts")
		case 'ч':
			b.WriteString("ch")
		case 'ш':
			b.WriteString("sh")
		case 'щ':
			b.WriteString("shch")
		case 'ы':
			b.WriteString("y")
		case 'ю':
			b.WriteString("yu")
		case 'я':
			b.WriteString("ya")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsIgnoredPath сообщает, что файл/папку не надо показывать во вкладке Solution,
// стримить в чат или коммитить в git — это артефакты сборки, менеджеры пакетов,
// кэш, бинарники и т.п. Пользователю они не нужны.
func IsIgnoredPath(path string) bool {
	clean := strings.TrimPrefix(filepath.ToSlash(path), "./")

	// Исключаем всю инфраструктуру Octra/Nix
	if IsInfraPath(clean) {
		return true
	}

	// node_modules — рай для npm-мусора
	if clean == "node_modules" || strings.HasPrefix(clean, "node_modules/") {
		return true
	}

	// Папки сборки
	buildDirs := []string{
		"dist", "build", "out", "target", "bin", "obj",
		".next", ".nuxt", ".output", ".cache",
		"__pycache__", ".pytest_cache", "venv", ".venv",
	}
	for _, d := range buildDirs {
		if clean == d || strings.HasPrefix(clean, d+"/") {
			return true
		}
	}

	// Lock-файлы (их содержимое не нужно пользователю)
	base := strings.ToLower(filepath.Base(clean))
	switch base {
	case "package-lock.json", "yarn.lock", "pnpm-lock.yaml", "pnpm-lock.yml":
		return true
	case "go.sum", "go.work.sum":
		return true
	case "cargo.lock", "gemfile.lock", "poetry.lock", "composer.lock":
		return true
	case ".ds_store", "thumbs.db":
		return true
	}

	// Расширения файлов, которые не несут смысла для пользователя
	switch strings.ToLower(filepath.Ext(clean)) {
	case ".pyc", ".pyo", ".pyd":
		return true
	case ".exe", ".dll", ".so", ".dylib", ".o", ".a", ".lib", ".obj":
		return true
	case ".jar", ".class", ".war", ".ear":
		return true
	case ".swp", ".swo", ".swn", ".bak", ".orig":
		return true
	case ".log":
		return true
	}

	return false
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
