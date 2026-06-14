package worker

import (
	"strings"
	"testing"

	"orchestrator/internal/service/groupchat"
)

// TestBuildChatContextEmptyForSingleWorker — regression на issue #79: одиночному
// воркеру в промпт падала пустая секция "GROUP CHAT HISTORY", которая затем
// оборачивалась в "CONTEXT FROM OTHER WORKERS" и зашумляла запрос. Теперь при
// отсутствии чужих сообщений контекст пуст.
func TestBuildChatContextEmptyForSingleWorker(t *testing.T) {
	a := &WorkerAgent{id: "self", role: "backend"}

	// Пустой разговор → пустой контекст.
	if got := a.buildChatContext(&groupchat.Conversation{}); got != "" {
		t.Errorf("expected empty context for empty conversation, got:\n%q", got)
	}

	// Разговор только с собственными сообщениями → тоже пусто (свои не считаются).
	own := &groupchat.Conversation{Messages: []groupchat.Message{
		{AgentID: "self", Role: "backend", Content: "done", Type: groupchat.MsgResponse},
	}}
	if got := a.buildChatContext(own); got != "" {
		t.Errorf("expected empty context when only own messages exist, got:\n%q", got)
	}
}

// TestBuildChatContextIncludesOtherWorkers — при наличии сообщений от других
// воркеров история собирается с маркерами и содержимым файлов.
func TestBuildChatContextIncludesOtherWorkers(t *testing.T) {
	a := &WorkerAgent{id: "self", role: "backend"}
	conv := &groupchat.Conversation{Messages: []groupchat.Message{
		{
			AgentID: "other",
			Role:    "frontend",
			Content: "built UI",
			Type:    groupchat.MsgResponse,
			Files:   map[string]string{"index.html": "<html></html>"},
		},
	}}
	got := a.buildChatContext(conv)
	for _, want := range []string{"GROUP CHAT HISTORY", "frontend", "index.html", "END HISTORY"} {
		if !strings.Contains(got, want) {
			t.Errorf("chat context missing %q:\n%s", want, got)
		}
	}
}
