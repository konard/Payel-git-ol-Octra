package worker

import "testing"

// TestIsCodeTask проверяет, для каких типов задач пустой результат воркеров
// считается провалом (issue #85): код (включая дефолтный пустой тип) — да,
// документные/исследовательские — нет.
func TestIsCodeTask(t *testing.T) {
	cases := map[string]bool{
		"":             true,
		"code":         true,
		"document":     false,
		"research":     false,
		"presentation": false,
	}
	for taskType, want := range cases {
		if got := isCodeTask(taskType); got != want {
			t.Errorf("isCodeTask(%q) = %v, want %v", taskType, got, want)
		}
	}
}
