package fetcher

import "testing"

func TestShouldLaunchWorkflowFromChatSearchRequests(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "explicit English search request",
			message: "Find how to install httpx on Python",
			want:    true,
		},
		{
			name:    "natural install question",
			message: "How to install httpx on Python",
			want:    true,
		},
		{
			name:    "documentation lookup",
			message: "Look up the httpx documentation reference",
			want:    true,
		},
		{
			name:    "Russian search request",
			message: "Найди как установить httpx на python",
			want:    true,
		},
		{
			name:    "regular greeting stays in chat",
			message: "How are you?",
			want:    false,
		},
		{
			name:    "Russian presentation change request launches workflow",
			message: "Введи в презентацию 3 самых высоких горы в мире и когда их покорили",
			want:    true,
		},
		{
			name:    "Russian code generation launches workflow",
			message: "Сделай мини прокси на Go",
			want:    true,
		},
		{
			name:    "Russian casual message stays in chat",
			message: "Спасибо, понял",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldLaunchWorkflowFromChat(tt.message); got != tt.want {
				t.Fatalf("shouldLaunchWorkflowFromChat(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}
