package models

import (
	"time"

	"github.com/google/uuid"
)

type ContextEntry struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID     uuid.UUID `gorm:"column:project_id;type:uuid;not null;index" json:"project_id"`
	Scope         string    `gorm:"column:scope;type:varchar(20);not null;default:'global'" json:"scope"`
	SourceAgentID string    `gorm:"column:source_agent_id;type:varchar(100)" json:"source_agent_id"`
	TargetID      string    `gorm:"column:target_id;type:varchar(100);index" json:"target_id"`
	ContextType   string    `gorm:"column:context_type;type:varchar(20);not null;default:'message'" json:"context_type"`
	Content       string    `gorm:"column:content;type:text;not null" json:"content"`
	Timestamp     time.Time `gorm:"column:timestamp;not null;index" json:"timestamp"`
	Important     bool      `gorm:"column:important;default:false" json:"important"`
	Forgotten     bool      `gorm:"column:forgotten;default:false;index" json:"forgotten"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
