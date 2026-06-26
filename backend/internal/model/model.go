// Package model defines the GORM models for the Octra monolith.
//
// The data model follows the "new concept" described in the project docs:
// every user owns a single Agent (their personal MCP environment) and a set of
// installed skills.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CLIType enumerates the AI CLIs Octra can launch inside a user environment.
// An empty CLI means the agent runs in pure LLM-proxy mode.
type CLIType string

// SkillType describes how a skill is provisioned into a Nix environment.
type SkillType string

const (
	SkillBuiltin SkillType = "built-in"
	SkillNixpkgs SkillType = "nixpkgs"
	SkillCustom  SkillType = "custom"
)

// User is a platform account.
type User struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
	Username     string     `gorm:"uniqueIndex" json:"username"`
	Email        string     `gorm:"uniqueIndex" json:"email"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	// APIKey authenticates calls to /api/chat and /environment.
	APIKey string `gorm:"uniqueIndex" json:"api_key"`
	// Subscription status ("free", "pro", ...).
	Subscription string `gorm:"default:free" json:"subscription"`
	// SubscriptionEnd is an optional Unix timestamp (seconds).
	SubscriptionEnd *int64 `json:"subscription_end"`
}

// Agent is a user's personal MCP environment. Each user has at most one.
type Agent struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `gorm:"index" json:"user_id"`

	// LLM connection details used both for the proxy mode and to configure the
	// launched CLI.
	LLMAPIKey  string `gorm:"column:llm_api_key" json:"-"`
	LLMBaseURL string `gorm:"column:llm_base_url" json:"llm_base_url"`
	LLMModel   string `gorm:"column:llm_model" json:"llm_model"`

	// CLI is the agent CLI to launch (e.g. "claude-code"). Empty => proxy mode.
	CLI CLIType `gorm:"column:cli" json:"cli"`

	// Active toggles whether the environment may be used.
	Active bool `gorm:"default:true" json:"active"`
}

// Skill is a reusable capability that can be installed into an environment.
type Skill struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"uniqueIndex" json:"name"`
	Type      SkillType `gorm:"column:type" json:"type"`
	// InstallCmd is the shell command (run inside the Nix env) or nixpkgs
	// attribute used to provision the skill.
	InstallCmd  string `gorm:"column:install_cmd" json:"install_cmd"`
	Description string `json:"description"`
}

// UserSkill links a skill to a user's agent and records its install status.
type UserSkill struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `gorm:"index" json:"user_id"`
	AgentID   uuid.UUID `gorm:"index" json:"agent_id"`
	SkillID   uuid.UUID `gorm:"index" json:"skill_id"`
	// Status is "pending", "installed" or "failed".
	Status string `gorm:"default:pending" json:"status"`
}

// AllModels returns every model for AutoMigrate.
func AllModels() []any {
	return []any{&User{}, &Agent{}, &Skill{}, &UserSkill{}}
}

// BeforeCreate assigns a UUID at the application level so the models work
// across databases (PostgreSQL in production, SQLite in tests) without relying
// on a DB-specific default.
func (m *User) BeforeCreate(*gorm.DB) error      { return ensureID(&m.ID) }
func (m *Agent) BeforeCreate(*gorm.DB) error     { return ensureID(&m.ID) }
func (m *Skill) BeforeCreate(*gorm.DB) error     { return ensureID(&m.ID) }
func (m *UserSkill) BeforeCreate(*gorm.DB) error { return ensureID(&m.ID) }

func ensureID(id *uuid.UUID) error {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
	return nil
}
