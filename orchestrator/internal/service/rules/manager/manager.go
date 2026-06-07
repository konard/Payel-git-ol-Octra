package manager

import (
	"orchestrator/internal/service/agents"
	ctxsvc "orchestrator/internal/service/context"
	"orchestrator/internal/service/rules/worker"
)

// Service — узел "manager". Вместо вызова worker по gRPC использует
// прямую ссылку на worker.Service.
type Service struct {
	agentsClient  *agents.Client
	worker        *worker.Service
	contextClient *ctxsvc.Service
}

// NewService — создаёт менеджера, привязанного к конкретному воркеру
func NewService(agentsClient *agents.Client, w *worker.Service, contextClient *ctxsvc.Service) *Service {
	return &Service{agentsClient: agentsClient, worker: w, contextClient: contextClient}
}

