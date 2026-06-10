package fetcher

import (
	"strings"
	"testing"

	"apigateway/pkg/requests"
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
		{
			name:    "pasted GitHub issue link launches workflow",
			message: "https://github.com/Payel-git-ol/Octra/issues/44",
			want:    true,
		},
		{
			name:    "pasted GitHub pull request link launches workflow",
			message: "https://github.com/Payel-git-ol/Octra/pull/45",
			want:    true,
		},
		{
			name:    "GitHub issue link with surrounding text launches workflow",
			message: "Please take a look at https://github.com/octra-labs/app/issues/7 thanks",
			want:    true,
		},
		{
			name:    "plain GitHub repo link stays in chat",
			message: "https://github.com/gin-gonic/gin",
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

// TestBuildBossChatReplyConversational guards issue #70: casual messages that
// are not workflow/search requests must get a real conversational answer, in
// the user's language, instead of silence or a generic English line. The key
// regression is «привет» — it used to fall through to an English fallback, so
// the chat felt dead.
func TestBuildBossChatReplyConversational(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		wantSubstrings []string // reply must contain at least one of these
		wantRussian    bool     // reply must be in Russian (Cyrillic) ...
		wantNotRussian bool     // ... or must NOT contain Cyrillic
	}{
		{
			name:           "Russian greeting replies in Russian",
			message:        "привет",
			wantSubstrings: []string{"Привет"},
			wantRussian:    true,
		},
		{
			name:           "English greeting replies in English",
			message:        "hello",
			wantSubstrings: []string{"Hi"},
			wantNotRussian: true,
		},
		{
			name:           "Russian thanks",
			message:        "спасибо",
			wantSubstrings: []string{"пожалуйста"},
			wantRussian:    true,
		},
		{
			name:           "English thanks",
			message:        "thanks!",
			wantSubstrings: []string{"welcome"},
			wantNotRussian: true,
		},
		{
			name:           "Russian capability question",
			message:        "что ты умеешь?",
			wantSubstrings: []string{"Octra"},
			wantRussian:    true,
		},
		{
			name:           "English who are you",
			message:        "who are you?",
			wantSubstrings: []string{"Octra"},
			wantNotRussian: true,
		},
		{
			name:           "Russian fallback stays Russian",
			message:        "ну ладно",
			wantSubstrings: []string{"задач"},
			wantRussian:    true,
		},
		{
			name:           "English fallback stays English",
			message:        "ok then",
			wantSubstrings: []string{"task"},
			wantNotRussian: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reply := buildBossChatReply(tt.message)
			if reply == "" {
				t.Fatalf("buildBossChatReply(%q) returned empty reply", tt.message)
			}

			matched := false
			for _, sub := range tt.wantSubstrings {
				if strings.Contains(reply, sub) {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("buildBossChatReply(%q) = %q, want it to contain one of %v", tt.message, reply, tt.wantSubstrings)
			}

			if tt.wantRussian && !isRussian(reply) {
				t.Fatalf("buildBossChatReply(%q) = %q, want a Russian reply", tt.message, reply)
			}
			if tt.wantNotRussian && isRussian(reply) {
				t.Fatalf("buildBossChatReply(%q) = %q, want an English reply", tt.message, reply)
			}
		})
	}
}

// TestSendCompletionReportContent guards the issue #70 requirement that the boss
// reports back in chat when a task finishes — including the boss's own answer
// (chatSummary), a result link, and the request's language.
func TestSendCompletionReportContent(t *testing.T) {
	t.Run("Russian task with PR link and summary", func(t *testing.T) {
		report := buildCompletionReport(
			requests.CreateTaskRequest{Title: "Мини прокси на Go"},
			map[string]string{
				"chatSummary":    "Сделал прокси с маршрутизацией.",
				"pullRequestUrl": "https://github.com/o/r/pull/7",
			},
		)
		for _, want := range []string{"Готово", "Мини прокси на Go", "Сделал прокси", "https://github.com/o/r/pull/7", "Результат"} {
			if !strings.Contains(report, want) {
				t.Fatalf("report = %q, want it to contain %q", report, want)
			}
		}
	})

	t.Run("English task falls back to repo link", func(t *testing.T) {
		report := buildCompletionReport(
			requests.CreateTaskRequest{Title: "PHP server"},
			map[string]string{"repoUrl": "https://github.com/o/r"},
		)
		for _, want := range []string{"Done", "PHP server", "https://github.com/o/r", "Result"} {
			if !strings.Contains(report, want) {
				t.Fatalf("report = %q, want it to contain %q", report, want)
			}
		}
		if isRussian(report) {
			t.Fatalf("report = %q, want an English report", report)
		}
	})
}
