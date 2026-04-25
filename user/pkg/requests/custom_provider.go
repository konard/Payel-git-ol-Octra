package requests

import "github.com/google/uuid"

type CreateCustomProviderRequest struct {
	Name           string `json:"name"`
	BaseURL        string `json:"base_url"`
	APIKey         string `json:"api_key,omitempty"`
	RequiresApiKey bool   `json:"requires_api_key"`
}

type UpdateCustomProviderRequest struct {
	Name           string `json:"name,omitempty"`
	BaseURL        string `json:"base_url,omitempty"`
	APIKey         string `json:"api_key,omitempty"`
	RequiresApiKey *bool  `json:"requires_api_key,omitempty"`
}

type CreateCustomModelRequest struct {
	Name       string     `json:"name"`
	ProviderID *uuid.UUID `json:"provider_id,omitempty"`
}

type UpdateCustomModelRequest struct {
	Name       string     `json:"name,omitempty"`
	ProviderID *uuid.UUID `json:"provider_id,omitempty"`
}
