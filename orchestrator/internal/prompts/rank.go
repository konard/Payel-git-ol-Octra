package prompts

import "strings"

// Prompt ranks (план фикса, пункт 9). Слабые модели плохо следуют длинным
// промптам с примерами и каталогами скиллов, поэтому для них промпт сокращается.
const (
	RankFull  = "full"
	RankLight = "light"
)

// lightModelMarkers — подстроки в имени модели, указывающие на «слабую» модель,
// которой нужен сокращённый промпт.
var lightModelMarkers = []string{
	"mini", "haiku", "flash", "small", "8b", "7b", "3.5-turbo", "gemini-pro",
}

// PromptRank — выбирает ранг промпта по имени модели. При fallback на
// gpt-4o-mini / haiku / gemini-pro и подобные возвращает RankLight.
func PromptRank(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" {
		return RankFull
	}
	for _, marker := range lightModelMarkers {
		if strings.Contains(m, marker) {
			return RankLight
		}
	}
	return RankFull
}
