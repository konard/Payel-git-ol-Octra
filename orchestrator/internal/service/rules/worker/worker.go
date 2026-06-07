package worker

import (
	"sync"

	"orchestrator/internal/service/agents"
	"orchestrator/internal/service/search"
)

// Service — узел "worker". Не делает сетевых вызовов между узлами:
// его методы вызываются менеджером напрямую как Go-функции.
type Service struct {
	agentsClient *agents.Client
	searchClient *search.Client
	mu           sync.Mutex
}

// NewService — создаёт нового воркера, используя общий agents-клиент.
// Воркер также получает клиент веб-поиска для ресёрч-задач.
func NewService(agentsClient *agents.Client) *Service {
	return &Service{
		agentsClient: agentsClient,
		searchClient: search.NewClient(),
	}
}

