package worker

import "strings"

// repairJSON пытается починить обрезанный JSON (LLM обрезало по max_tokens)
func repairJSON(s string) string {
	s = strings.TrimSpace(s)
	// Если JSON обрезан — добавляем закрывающие скобки
	braceCount := 0
	bracketCount := 0
	inString := false
	escapeNext := false
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

	// Убираем висящие запятые перед закрывающими скобками
	result = strings.ReplaceAll(result, ",}", "}")
	result = strings.ReplaceAll(result, ",]", "]")

	return result
}
