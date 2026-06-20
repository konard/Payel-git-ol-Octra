package boss

import "testing"

// TestGradeWeight — перевод оценки босса (1-10) в шкалу менеджера (0-100).
// Регрессия на issue #79: раньше grade_weight был жёстко "10", из-за чего босс и
// менеджер видели рассогласованные сигналы сложности одной и той же задачи.
func TestGradeWeight(t *testing.T) {
	cases := []struct {
		grade int
		want  string
	}{
		{grade: 0, want: "10"},  // неоценено → минимум
		{grade: -3, want: "10"}, // защита от отрицательных
		{grade: 1, want: "10"},
		{grade: 5, want: "50"},
		{grade: 8, want: "80"},
		{grade: 10, want: "100"},
		{grade: 42, want: "100"}, // защита от выхода за диапазон
	}
	for _, c := range cases {
		if got := gradeWeight(c.grade); got != c.want {
			t.Errorf("gradeWeight(%d) = %q, want %q", c.grade, got, c.want)
		}
	}
}

// TestBuildManagerMetadataPropagatesGrade — метадата менеджера должна нести реальную
// оценку босса, а не хардкод "10" (issue #79). Босс видит grade/10, менеджер —
// grade*10/100: один и тот же сигнал сложности на обоих слоях.
func TestBuildManagerMetadataPropagatesGrade(t *testing.T) {
	req := &CreateTaskRequest{
		Title:       "hello world",
		Description: "напиши hello world на express js",
		Grade:       8,
		Tokens:      map[string]string{},
		Meta: map[string]string{
			"search": `{"provider":"apodex","model":"apodex-1-0-deepresearch-mini","base-url":"https://api.apodex.ai/v1/responses","api-key":"sk-test","striming":true}`,
		},
	}
	decision := &DecisionResult{TaskType: "code", TechStack: []string{"nodejs"}}

	meta := buildManagerMetadata(req, decision)
	if got := meta["grade_weight"]; got != "80" {
		t.Errorf("grade_weight = %q, want %q (grade 8 → 80/100)", got, "80")
	}
	if got := meta["search"]; got != req.Meta["search"] {
		t.Errorf("search metadata = %q, want %q", got, req.Meta["search"])
	}

	// Неоценённая задача сохраняет безопасный минимум.
	req.Grade = 0
	if got := buildManagerMetadata(req, decision)["grade_weight"]; got != "10" {
		t.Errorf("ungraded grade_weight = %q, want %q", got, "10")
	}
}
