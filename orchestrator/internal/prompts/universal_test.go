package prompts

import (
	"strings"
	"testing"
)

// TestUniversalNode_CodeTask — для кодовой задачи промпт требует минимальный
// результат и строгий JSON-контракт файлов (issue #91: hello world должен быть
// одним файлом, а не деревом непонятных файлов).
func TestUniversalNode_CodeTask(t *testing.T) {
	p := UniversalNode("hello world", "напиши hello world на python", "code", "python")
	mustContain(t, p, "UNIVERSAL NODE")
	mustContain(t, p, "DO NOT OVER-ENGINEER")
	mustContain(t, p, "python")
	mustContain(t, p, `{"files":`)
	mustContain(t, p, "DELIVERABLE (code)")
	if strings.Contains(p, "solution/answer.md") {
		t.Errorf("code task prompt should not force a markdown answer file")
	}
}

// TestUniversalNode_TextTask — research/document/presentation отдают единый
// Markdown в solution/answer.md, который подхватывает существующий конвейер.
func TestUniversalNode_TextTask(t *testing.T) {
	for _, tt := range []string{"research", "document", "presentation"} {
		p := UniversalNode("q", "what is 2+2", tt, "markdown")
		mustContain(t, p, "solution/answer.md")
		mustContain(t, p, "DELIVERABLE ("+tt+")")
	}
}

// TestUniversalNode_DefaultsToCode — пустой task_type трактуется как код.
func TestUniversalNode_DefaultsToCode(t *testing.T) {
	p := UniversalNode("t", "d", "", "")
	mustContain(t, p, "DELIVERABLE (code)")
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("prompt missing %q", needle)
	}
}
