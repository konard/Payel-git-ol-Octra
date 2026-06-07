package util

import (
	"encoding/json"
	"strings"
)

// MarshalJSON — безопасно сериализует значение в строку (ошибка игнорируется)
func MarshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// ExtractJSONFromMarkdown — вытаскивает JSON из ```json ... ``` или ``` ... ```
func ExtractJSONFromMarkdown(s string) string {
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

// StripMarkdownCodeBlock — убирает markdown-обёртку из текста файла
func StripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	markers := []string{
		"```go\n", "```ts\n", "```js\n", "```tsx\n", "```jsx\n",
		"```json\n", "```yaml\n", "```yml\n", "```sh\n", "```bash\n",
		"```python\n", "```py\n", "```\n", "```",
	}
	for _, marker := range markers {
		if strings.HasPrefix(s, marker) {
			s = s[len(marker):]
			break
		}
	}
	if end := strings.LastIndex(s, "```"); end > 0 {
		s = strings.TrimSpace(s[:end])
	}
	return s
}

// RepairJSON — чинит обрезанный JSON, балансирует скобки, удаляет висящие запятые
func RepairJSON(s string) string {
	s = strings.TrimSpace(s)
	braceCount, bracketCount := 0, 0
	inString, escapeNext := false, false
	cutPos := 0
	for i, ch := range s {
		if escapeNext {
			escapeNext = false
			continue
		}
		if ch == '\\' && inString {
			escapeNext = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '{':
			braceCount++
		case '[':
			bracketCount++
		case '}':
			braceCount--
		case ']':
			bracketCount--
		}
		if braceCount < 0 || bracketCount < 0 {
			cutPos = i
			break
		}
	}
	result := s
	if cutPos > 0 {
		result = s[:cutPos]
	}
	for braceCount > 0 {
		result += "}"
		braceCount--
	}
	for bracketCount > 0 {
		result += "]"
		bracketCount--
	}
	result = strings.ReplaceAll(result, ",}", "}")
	result = strings.ReplaceAll(result, ",]", "]")
	return result
}

