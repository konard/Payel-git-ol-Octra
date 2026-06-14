package boss

import "testing"

func TestNormalizeTaskTypeGitHub(t *testing.T) {
	for _, in := range []string{"github", "GitHub", "issue", "pull_request", "pr"} {
		if got := normalizeTaskType(in); got != TaskTypeGitHub {
			t.Errorf("normalizeTaskType(%q) = %q, want %q", in, got, TaskTypeGitHub)
		}
	}
}

func TestIsCodeLikeTask(t *testing.T) {
	codeLike := []string{"", TaskTypeCode, TaskTypeGitHub}
	for _, tt := range codeLike {
		if !isCodeLikeTask(tt) {
			t.Errorf("isCodeLikeTask(%q) = false, want true", tt)
		}
	}
	notCode := []string{TaskTypeResearch, TaskTypeDocument, TaskTypePresentation}
	for _, tt := range notCode {
		if isCodeLikeTask(tt) {
			t.Errorf("isCodeLikeTask(%q) = true, want false", tt)
		}
	}
}
