package worker

import "strings"

// extractJSONFromMarkdown извлекает JSON из markdown блоков ```json ... ``` или ``` ... ```
func extractJSONFromMarkdown(s string) string {
	start := -1
	for _, marker := range []string{"```json\n", "```\n", "```"} {
		if idx := strings.Index(s, marker); idx >= 0 {
			start = idx + len(marker)
			break
		}
	}
	if start >= 0 {
		end := strings.Index(s[start:], "```")
		if end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
		return strings.TrimSpace(s[start:])
	}
	return s
}
