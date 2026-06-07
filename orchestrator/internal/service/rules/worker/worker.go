package worker

import (
	"sync"

	"orchestrator/internal/service/agents"
	ctxsvc "orchestrator/internal/service/context"
	"orchestrator/internal/service/search"
)

// Service — узел "worker". Не делает сетевых вызовов между узлами:
// его методы вызываются менеджером напрямую как Go-функции.
type Service struct {
	agentsClient  *agents.Client
	searchClient  *search.Client
	contextClient *ctxsvc.Service
	mu            sync.Mutex
}

// NewService — создаёт нового воркера, используя общий agents-клиент.
// Воркер также получает клиент веб-поиска для ресёрч-задач.
func NewService(agentsClient *agents.Client, contextClient *ctxsvc.Service) *Service {
	return &Service{
		agentsClient:  agentsClient,
		searchClient:  search.NewClient(),
		contextClient: contextClient,
	}
}

