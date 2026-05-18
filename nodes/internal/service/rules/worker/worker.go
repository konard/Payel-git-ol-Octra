package worker

import (
	"sync"

	"nodes/internal/service/agents"
)

// Service — узел "worker". Не делает сетевых вызовов между узлами:
// его методы вызываются менеджером напрямую как Go-функции.
type Service struct {
	agentsClient *agents.Client
	mu           sync.Mutex
}

// NewService — создаёт нового воркера, используя общий agents-клиент
func NewService(agentsClient *agents.Client) *Service {
	return &Service{agentsClient: agentsClient}
}
