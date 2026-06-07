package groupchat

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type AgentStatus int

const (
	AgentIdle     AgentStatus = iota
	AgentThinking
	AgentResponded
	AgentDone
	AgentError
)

type AgentInfo struct {
	ID          string
	Role        string
	Description string
	Status      AgentStatus
}

type MessageType string

const (
	MsgSystem    MessageType = "system"
	MsgThought   MessageType = "thought"
	MsgResponse  MessageType = "response"
	MsgContext   MessageType = "context"
	MsgTerminate MessageType = "terminate"
)

type Message struct {
	AgentID   string            `json:"agent_id"`
	Role      string            `json:"role"`
	Content   string            `json:"content"`
	Files     map[string]string `json:"files,omitempty"` // filename → content
	Type      MessageType       `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
}

type Conversation struct {
	Messages []Message
	Task     string
}

func (c *Conversation) Add(msg Message) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	c.Messages = append(c.Messages, msg)
}

func (c *Conversation) Last() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

func (c *Conversation) MessagesByAgent(agentID string) []Message {
	var out []Message
	for _, m := range c.Messages {
		if m.AgentID == agentID {
			out = append(out, m)
		}
	}
	return out
}

func (c *Conversation) Format() string {
	var b strings.Builder
	b.WriteString("=== GROUP CHAT CONVERSATION ===\n")
	if c.Task != "" {
		b.WriteString("Task: " + c.Task + "\n")
	}
	b.WriteString(fmt.Sprintf("Total messages: %d\n", len(c.Messages)))
	for _, msg := range c.Messages {
		prefix := "[" + msg.Role + "]"
		if msg.Type == MsgSystem {
			prefix = "[SYSTEM]"
		}
		b.WriteString(prefix + " " + msg.Content + "\n")
		if len(msg.Files) > 0 {
			b.WriteString(fmt.Sprintf("  files: %d generated\n", len(msg.Files)))
		}
	}
	b.WriteString("=== END CONVERSATION ===")
	return b.String()
}

func (c *Conversation) FormatWithoutFiles() string {
	var b strings.Builder
	b.WriteString("=== CONVERSATION HISTORY ===\n")
	for _, msg := range c.Messages {
		prefix := "[" + msg.Role + "]"
		if msg.Type == MsgSystem {
			prefix = "[SYSTEM]"
		}
		b.WriteString(prefix + " " + msg.Content + "\n")
	}
	b.WriteString("=== END HISTORY ===")
	return b.String()
}

func (c *Conversation) AllFiles() map[string]string {
	all := make(map[string]string)
	for _, msg := range c.Messages {
		for path, content := range msg.Files {
			if _, exists := all[path]; !exists {
				all[path] = content
			}
		}
	}
	return all
}

type EventType string

const (
	EventAgentResponse EventType = "agent_response"
	EventAgentComplete EventType = "agent_complete"
	EventWorkflowDone  EventType = "workflow_done"
	EventRoundComplete EventType = "round_complete"
	EventError         EventType = "error"
)

type Event struct {
	Type      EventType `json:"type"`
	Message   *Message  `json:"message,omitempty"`
	AgentID   string    `json:"agent_id,omitempty"`
	Round     int       `json:"round,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type SpeakerSelector interface {
	SelectNext(conv *Conversation, agents []*AgentInfo) string
}

type Agent interface {
	ID() string
	Role() string
	Process(ctx context.Context, conv *Conversation) ([]Message, error)
}

type TerminationCondition func(conv *Conversation, round int) bool
