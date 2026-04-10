package requests

// WorkerRole — конфигурация воркера в workflow
type WorkerRole struct {
	Role        string `json:"role"`
	Description string `json:"description"`
}

// ManagerWorkflow — конфигурация менеджера с воркерами
type ManagerWorkflow struct {
	Role        string       `json:"role"`
	Description string       `json:"description"`
	Priority    int32        `json:"priority"`
	Workers     []WorkerRole `json:"workers"`
}

// WorkflowConfig — полная конфигурация workflow от пользователя
type WorkflowConfig struct {
	UseAIPlanning bool              `json:"useAiPlanning"` // true = AI планирует, false = использовать predefined
	Managers      []ManagerWorkflow `json:"managers"`
	Architecture  string            `json:"architecture"`
	TechStack     []string          `json:"techStack"`
}

type CreateTaskRequest struct {
	UserID      string            `json:"userId"`
	Username    string            `json:"username"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tokens      map[string]string `json:"tokens"`
	Meta        map[string]string `json:"meta"`
	Workflow    *WorkflowConfig   `json:"workflow,omitempty"` // optional predefined workflow
}

type TaskStatusRequest struct {
	TaskID string `json:"taskId"`
}
