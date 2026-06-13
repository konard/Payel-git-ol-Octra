package prompts

import "strings"

// langInfo описывает соглашение об именовании исходных файлов для tech stack.
type langInfo struct {
	display  string // человекочитаемое имя языка для промптов
	ext      string // расширение исходного файла БЕЗ ведущей точки
	mainFile string // типичное имя файла-точки входа
}

// langByStack сопоставляет нормализованный (lowercase) tech stack с языковыми
// метаданными. Ключи включают как базовые языки, так и популярные алиасы
// фреймворков — AI/детектор часто возвращает имя фреймворка ("express", "django"),
// а файлы всё равно должны получать расширение базового языка (.js, .py).
//
// КРИТИЧНО: раньше промпты воркера подставляли сам techStack как расширение
// (".%s") и хардкодили "NOT JavaScript" — из-за этого Node.js-задача
// сохранялась в main.go. Эта карта — единый источник правды о расширениях.
var langByStack = map[string]langInfo{
	// Go
	"go":     {"Go", "go", "main.go"},
	"golang": {"Go", "go", "main.go"},

	// Node.js / JS / TS и фреймворки поверх них
	"node":       {"Node.js", "js", "index.js"},
	"nodejs":     {"Node.js", "js", "index.js"},
	"javascript": {"JavaScript", "js", "index.js"},
	"js":         {"JavaScript", "js", "index.js"},
	"express":    {"Node.js (Express)", "js", "index.js"},
	"expressjs":  {"Node.js (Express)", "js", "index.js"},
	"nestjs":     {"Node.js (NestJS)", "ts", "main.ts"},
	"next":       {"Node.js (Next.js)", "js", "index.js"},
	"nextjs":     {"Node.js (Next.js)", "js", "index.js"},
	"react":      {"JavaScript (React)", "jsx", "App.jsx"},
	"vue":        {"JavaScript (Vue)", "vue", "App.vue"},
	"angular":    {"TypeScript (Angular)", "ts", "main.ts"},
	"svelte":     {"JavaScript (Svelte)", "svelte", "App.svelte"},
	"typescript": {"TypeScript", "ts", "index.ts"},
	"ts":         {"TypeScript", "ts", "index.ts"},

	// Python и фреймворки
	"python":  {"Python", "py", "main.py"},
	"django":  {"Python (Django)", "py", "manage.py"},
	"flask":   {"Python (Flask)", "py", "app.py"},
	"fastapi": {"Python (FastAPI)", "py", "main.py"},

	// Rust
	"rust": {"Rust", "rs", "src/main.rs"},

	// PHP и фреймворки
	"php":     {"PHP", "php", "index.php"},
	"laravel": {"PHP (Laravel)", "php", "index.php"},
	"symfony": {"PHP (Symfony)", "php", "index.php"},

	// JVM
	"java":   {"Java", "java", "Main.java"},
	"spring": {"Java (Spring)", "java", "Application.java"},
	"kotlin": {"Kotlin", "kt", "Main.kt"},
	"scala":  {"Scala", "scala", "Main.scala"},

	// .NET
	"dotnet": {"C#", "cs", "Program.cs"},
	"csharp": {"C#", "cs", "Program.cs"},
	"c#":     {"C#", "cs", "Program.cs"},

	// C / C++
	"c":   {"C", "c", "main.c"},
	"c++": {"C++", "cpp", "main.cpp"},
	"cpp": {"C++", "cpp", "main.cpp"},

	// Прочее
	"ruby":    {"Ruby", "rb", "main.rb"},
	"rails":   {"Ruby (Rails)", "rb", "main.rb"},
	"flutter": {"Dart (Flutter)", "dart", "lib/main.dart"},
	"dart":    {"Dart", "dart", "bin/main.dart"},
	"swift":   {"Swift", "swift", "main.swift"},
	"elixir":  {"Elixir", "ex", "lib/main.ex"},
	"phoenix": {"Elixir (Phoenix)", "ex", "lib/main.ex"},
	"haskell": {"Haskell", "hs", "Main.hs"},
	"zig":     {"Zig", "zig", "main.zig"},
	"r":       {"R", "R", "main.R"},
}

func lookupLang(techStack string) (langInfo, bool) {
	key := strings.ToLower(strings.TrimSpace(techStack))
	info, ok := langByStack[key]
	return info, ok
}

// LangDisplay возвращает человекочитаемое имя языка для tech stack
// (например "Node.js" для "nodejs"). Для неизвестного стека возвращает
// сам techStack без изменений.
func LangDisplay(techStack string) string {
	if info, ok := lookupLang(techStack); ok {
		return info.display
	}
	return techStack
}

// LangExtension возвращает типичное расширение исходного файла (без точки) для
// tech stack, или пустую строку для неизвестного стека.
func LangExtension(techStack string) string {
	if info, ok := lookupLang(techStack); ok {
		return info.ext
	}
	return ""
}

// LangMainFile возвращает типичное имя файла-точки входа для tech stack.
// Для неизвестного стека возвращает "main.txt", чтобы fallback всё равно мог
// создать файл, не подставляя ошибочное расширение чужого языка.
func LangMainFile(techStack string) string {
	if info, ok := lookupLang(techStack); ok {
		return info.mainFile
	}
	return "main.txt"
}
