package requests

// CreateWorkflowRequest represents the request to create a workflow
type CreateWorkflowRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Nodes       string   `json:"nodes" binding:"required"`
	Edges       string   `json:"edges" binding:"required"`
	IsPublic    bool     `json:"is_public"`
}

// UpdateWorkflowRequest represents the request to update a workflow
type UpdateWorkflowRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	IsPublic    bool     `json:"is_public"`
}
