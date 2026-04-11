package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Workflow represents a published workflow template in the library
type Workflow struct {
	ID          uuid.UUID  `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`

	// Owner info
	UserID     uuid.UUID `json:"user_id" gorm:"column:user_id;type:uuid;index;not null"`
	AuthorName string    `json:"author_name" gorm:"column:author_name;not null"`

	// Workflow metadata
	Name        string   `json:"name" gorm:"column:name;not null"`
	Description string   `json:"description" gorm:"column:description"`
	Category    string   `json:"category" gorm:"column:category;index"`
	Tags        []string `json:"tags" gorm:"column:tags;type:text[]"`

	// Workflow structure (JSON representation of nodes and edges)
	Nodes string `json:"nodes" gorm:"column:nodes;type:text;not null"`
	Edges string `json:"edges" gorm:"column:edges;type:text;not null"`

	// Visibility
	IsPublic bool `json:"is_public" gorm:"column:is_public;default:false;index"`

	// Statistics
	Downloads int `json:"downloads" gorm:"column:downloads;default:0"`
}

// MarshalJSON customizes JSON serialization to ensure UUIDs are strings
func (w Workflow) MarshalJSON() ([]byte, error) {
	type Alias Workflow
	return json.Marshal(&struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
		*Alias
	}{
		ID:     w.ID.String(),
		UserID: w.UserID.String(),
		Alias:  (*Alias)(&w),
	})
}
