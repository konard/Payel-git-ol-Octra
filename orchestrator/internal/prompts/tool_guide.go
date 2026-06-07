package prompts

import (
	"strings"

	"orchestrator/pkg/guids"
)

// techStackGuideKeys maps lowercase tech-stack strings to one or more guide
// names in the guids registry. Unknown stacks map to nothing.
var techStackGuideKeys = map[string][]string{
	"go":         {"Go"},
	"golang":     {"Go"},
	"node":       {"npm"},
	"nodejs":     {"npm"},
	"javascript": {"npm"},
	"typescript": {"npm"},
	"rust":       {"Cargo"},
	"php":        {"Composer", "Artisan"},
	"flutter":    {"Flutter"},
	"dart":       {"Flutter"},
	"python":     {"pip"},
	"c++":        {"CMake"},
	"cpp":        {"CMake"},
	"c":          {"CMake"},
	"java":       {"Maven", "Gradle"},
	"kotlin":     {"Kotlin"},
	"ruby":       {"Bundler"},
	"dotnet":     {"dotnet"},
	"csharp":     {"dotnet"},
	"elixir":     {"Mix"},
	"haskell":    {"Cabal"},
	"zig":        {"Zig"},
	"swift":      {"SwiftPM"},
}

// resolveGuides возвращает список Guide для tech stack (nil если неизвестен).
func resolveGuides(techStack string) []*guids.Guide {
	key := strings.ToLower(strings.TrimSpace(techStack))
	names, ok := techStackGuideKeys[key]
	if !ok {
		return nil
	}
	out := make([]*guids.Guide, 0, len(names))
	for _, name := range names {
		if g := guids.Get(name); g != nil {
			out = append(out, g)
		}
	}
	return out
}

// ToolCheatSheet returns a concise reference of common tool commands for the
// given tech stack. The AI reads this instead of guessing commands, saving
// tokens and avoiding mistakes. Returns empty string for unknown stacks.
func ToolCheatSheet(techStack string) string {
	guides := resolveGuides(techStack)
	if len(guides) == 0 {
		return ""
	}
	var parts []string
	for _, g := range guides {
		parts = append(parts, g.FormatCheat())
	}
	return strings.Join(parts, "\n\n")
}

// ToolCheatSheetBrief returns a shorter one-liner description for use in
// manager or boss prompts where full detail isn't needed.
func ToolCheatSheetBrief(techStack string) string {
	guides := resolveGuides(techStack)
	if len(guides) == 0 {
		return ""
	}
	var parts []string
	for _, g := range guides {
		parts = append(parts, g.FormatBrief())
	}
	return strings.Join(parts, " / ")
}
