package groupchat

import "strings"

type RoundRobinSelector struct {
	index int
}

func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{}
}

func (s *RoundRobinSelector) SelectNext(conv *Conversation, agents []*AgentInfo) string {
	if len(agents) == 0 {
		return ""
	}
	// Skip done/error agents
	for i := 0; i < len(agents); i++ {
		idx := (s.index + i) % len(agents)
		a := agents[idx]
		if a.Status == AgentIdle || a.Status == AgentResponded {
			s.index = (idx + 1) % len(agents)
			return a.ID
		}
	}
	return ""
}

type AllAtOnceSelector struct{}

func NewAllAtOnceSelector() *AllAtOnceSelector {
	return &AllAtOnceSelector{}
}

func (s *AllAtOnceSelector) SelectNext(conv *Conversation, agents []*AgentInfo) string {
	return "all"
}

type AgentBasedSelector struct {
	SelectFunc func(conv *Conversation, agents []*AgentInfo) string
}

func (s *AgentBasedSelector) SelectNext(conv *Conversation, agents []*AgentInfo) string {
	if s.SelectFunc == nil {
		return ""
	}
	id := s.SelectFunc(conv, agents)
	// verify agent exists
	for _, a := range agents {
		if a.ID == id && (a.Status == AgentIdle || a.Status == AgentResponded) {
			return id
		}
	}
	return ""
}

func FirstResponderSelector(conv *Conversation, agents []*AgentInfo) string {
	// Find first agent that hasn't responded yet
	for _, a := range agents {
		if a.Status == AgentIdle {
			return a.ID
		}
	}
	return ""
}

func RoleBasedSelector(role string) func(conv *Conversation, agents []*AgentInfo) string {
	return func(conv *Conversation, agents []*AgentInfo) string {
		for _, a := range agents {
			if strings.EqualFold(a.Role, role) && (a.Status == AgentIdle || a.Status == AgentResponded) {
				return a.ID
			}
		}
		return ""
	}
}
