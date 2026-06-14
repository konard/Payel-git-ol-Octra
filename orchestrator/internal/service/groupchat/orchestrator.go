package groupchat

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Orchestrator struct {
	agents      map[string]*AgentInfo
	agentImpls  map[string]Agent
	conv        *Conversation
	selector    SpeakerSelector
	maxRounds   int
	terminateFn TerminationCondition
	mu          sync.RWMutex
	eventCh     chan Event

	errsMu sync.Mutex
	errs   []error
}

func NewOrchestrator(maxRounds int) *Orchestrator {
	return &Orchestrator{
		agents:     make(map[string]*AgentInfo),
		agentImpls: make(map[string]Agent),
		conv:       &Conversation{},
		selector:   NewRoundRobinSelector(),
		maxRounds:  maxRounds,
		eventCh:    make(chan Event, 100),
	}
}

// recordError — сохраняет ошибку агента, чтобы вызывающий код мог понять,
// что воркер реально провалился (раньше ошибки только эмитились в events и
// молча терялись, из-за чего менеджер репортил success при пустом результате — issue #85).
func (o *Orchestrator) recordError(err error) {
	if err == nil {
		return
	}
	o.errsMu.Lock()
	o.errs = append(o.errs, err)
	o.errsMu.Unlock()
}

// Errors — возвращает ошибки агентов, накопленные за время прогона.
func (o *Orchestrator) Errors() []error {
	o.errsMu.Lock()
	defer o.errsMu.Unlock()
	return append([]error(nil), o.errs...)
}

func (o *Orchestrator) SetSelector(s SpeakerSelector) {
	o.selector = s
}

func (o *Orchestrator) SetTerminationCondition(fn TerminationCondition) {
	o.terminateFn = fn
}

func (o *Orchestrator) RegisterAgent(a Agent) {
	o.mu.Lock()
	defer o.mu.Unlock()

	info := &AgentInfo{
		ID:     a.ID(),
		Role:   a.Role(),
		Status: AgentIdle,
	}
	o.agents[a.ID()] = info
	o.agentImpls[a.ID()] = a
}

func (o *Orchestrator) Events() <-chan Event {
	return o.eventCh
}

func (o *Orchestrator) emit(evt Event) {
	select {
	case o.eventCh <- evt:
	default:
	}
}

func (o *Orchestrator) AddSystemMessage(content string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.conv.Add(Message{
		AgentID: "system",
		Role:    "system",
		Content: content,
		Type:    MsgSystem,
	})
}

func (o *Orchestrator) SnapshotConversation() *Conversation {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return &Conversation{
		Task:     o.conv.Task,
		Messages: append([]Message{}, o.conv.Messages...),
	}
}

func (o *Orchestrator) GetAgentInfo(id string) *AgentInfo {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.agents[id]
}

func (o *Orchestrator) activeAgents() []*AgentInfo {
	o.mu.RLock()
	defer o.mu.RUnlock()
	out := make([]*AgentInfo, 0, len(o.agents))
	for _, a := range o.agents {
		out = append(out, a)
	}
	return out
}

func (o *Orchestrator) agentReady(id string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	a, ok := o.agents[id]
	if !ok {
		return false
	}
	return a.Status == AgentIdle || a.Status == AgentResponded
}

func (o *Orchestrator) setAgentStatus(id string, status AgentStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if a, ok := o.agents[id]; ok {
		a.Status = status
	}
}

func (o *Orchestrator) allAgentsDone() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, a := range o.agents {
		if a.Status != AgentDone && a.Status != AgentError {
			return false
		}
	}
	return true
}

// Run запускает Group Chat: все агенты бегут конкурентно, результат каждого
// broadcast'ится в общий чат. После каждого раунда оркестратор проверяет
// условие завершения. Если не завершено — запускает следующий раунд.
// Агенты "видят" полную историю через conv.Format().
func (o *Orchestrator) Run(ctx context.Context, task string) error {
	if task != "" {
		o.AddSystemMessage("Task: " + task)
	}
	o.conv.Task = task

	if len(o.agents) == 0 {
		return fmt.Errorf("group chat: no agents registered")
	}

	for round := 1; round <= o.maxRounds; round++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Проверка завершения
		if o.terminateFn != nil {
			conv := o.SnapshotConversation()
			if o.terminateFn(conv, round) {
				log.Printf("[GroupChat] Terminated at round %d by condition", round)
				o.emit(Event{Type: EventWorkflowDone, Round: round})
				return nil
			}
		}
		if o.allAgentsDone() {
			o.emit(Event{Type: EventWorkflowDone, Round: round})
			return nil
		}

		// Определяем, кто говорит в этом раунде
		agents := o.activeAgents()
		next := o.selector.SelectNext(o.SnapshotConversation(), agents)
		if next == "" {
			log.Printf("[GroupChat] No speaker selected, terminating")
			o.emit(Event{Type: EventWorkflowDone, Round: round})
			return nil
		}

		if next == "all" {
			// Конкурентный раунд: все агенты говорят одновременно
			o.runConcurrentRound(ctx, round)
		} else {
			// Один агент говорит
			o.runSingle(ctx, next, round)
		}

		o.emit(Event{Type: EventRoundComplete, Round: round})
	}

	o.emit(Event{Type: EventWorkflowDone, Round: o.maxRounds})
	return nil
}

// runSingle — один агент говорит, результат broadcast'ится
func (o *Orchestrator) runSingle(ctx context.Context, agentID string, round int) {
	o.setAgentStatus(agentID, AgentThinking)

	agent, ok := o.agentImpls[agentID]
	if !ok {
		return
	}

	conv := o.SnapshotConversation()
	messages, err := agent.Process(ctx, conv)
	if err != nil {
		o.recordError(fmt.Errorf("agent %s: %w", agentID, err))
		o.setAgentStatus(agentID, AgentError)
		o.emit(Event{Type: EventError, AgentID: agentID, Error: err.Error(), Round: round})
		return
	}

	o.mu.Lock()
	for _, msg := range messages {
		o.conv.Add(msg)
	}
	o.mu.Unlock()

	o.setAgentStatus(agentID, AgentResponded)
	for _, msg := range messages {
		o.emit(Event{Type: EventAgentResponse, Message: &msg, AgentID: agentID, Round: round})
		if msg.Type == MsgTerminate {
			o.setAgentStatus(agentID, AgentDone)
		}
	}
	o.emit(Event{Type: EventAgentComplete, AgentID: agentID, Round: round})
}

// runConcurrentRound — все агенты говорят конкурентно в одном раунде
func (o *Orchestrator) runConcurrentRound(ctx context.Context, round int) {
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	results := make(map[string][]Message)

	for id, agent := range o.agentImpls {
		if !o.agentReady(id) {
			continue
		}
		wg.Add(1)
		go func(agentID string, a Agent) {
			defer wg.Done()

			o.setAgentStatus(agentID, AgentThinking)
			conv := o.SnapshotConversation()

			messages, err := a.Process(ctx, conv)
			if err != nil {
				o.recordError(fmt.Errorf("agent %s: %w", agentID, err))
				o.setAgentStatus(agentID, AgentError)
				o.emit(Event{Type: EventError, AgentID: agentID, Error: err.Error(), Round: round})
				return
			}

			mu.Lock()
			results[agentID] = messages
			mu.Unlock()
		}(id, agent)
	}

	wg.Wait()

	// Broadcast all results
	for agentID, messages := range results {
		o.mu.Lock()
		for _, msg := range messages {
			o.conv.Add(msg)
		}
		o.mu.Unlock()

		o.setAgentStatus(agentID, AgentResponded)
		for _, msg := range messages {
			o.emit(Event{Type: EventAgentResponse, Message: &msg, AgentID: agentID, Round: round})
			if msg.Type == MsgTerminate {
				o.setAgentStatus(agentID, AgentDone)
			}
		}
		o.emit(Event{Type: EventAgentComplete, AgentID: agentID, Round: round})
	}
}
