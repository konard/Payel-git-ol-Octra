package fetcher

import (
	"strings"
	"testing"
)

// TestNewStreamIDIsolatesConcurrentStreams guards the fix for cross-tab history
// mixing (issue #31): two concurrent task/chat streams from the SAME user must
// receive distinct Redis stream ids. Previously the user id itself was used as
// the stream id, so two browser tabs shared the STREAM:<userID> key and one
// tab's updates (e.g. a presentation) leaked into the other tab (code only).
func TestNewStreamIDIsolatesConcurrentStreams(t *testing.T) {
	const userID = "user-123"

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := newStreamID(userID)

		if !strings.HasPrefix(id, userID+":") {
			t.Fatalf("newStreamID(%q) = %q, want prefix %q", userID, id, userID+":")
		}
		if id == userID {
			t.Fatalf("newStreamID(%q) returned the bare user id, tabs would collide", userID)
		}
		if seen[id] {
			t.Fatalf("newStreamID(%q) produced a duplicate id %q, tabs would collide", userID, id)
		}
		seen[id] = true
	}

	// Different users must never collide either.
	if a, b := newStreamID("u-a"), newStreamID("u-b"); a == b {
		t.Fatalf("newStreamID produced identical ids for different users: %q == %q", a, b)
	}
}

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
