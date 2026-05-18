package manager

import (
	"nodes/internal/service/agents"
	"nodes/internal/service/rules/worker"
)

// Service — узел "manager". Вместо вызова worker по gRPC использует
// прямую ссылку на worker.Service.
type Service struct {
	agentsClient *agents.Client
	worker       *worker.Service
}

// NewService — создаёт менеджера, привязанного к конкретному воркеру
func NewService(agentsClient *agents.Client, w *worker.Service) *Service {
	return &Service{agentsClient: agentsClient, worker: w}
}
