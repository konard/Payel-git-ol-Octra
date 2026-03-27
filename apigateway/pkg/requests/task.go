package requests

type CreateTaskRequest struct {
	UserID      string            `json:"userId"`
	Username    string            `json:"username"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tokens      []string          `json:"tokens"`
	Meta        map[string]string `json:"meta"`
}

type TaskStatusRequest struct {
	TaskID string `json:"taskId"`
}
