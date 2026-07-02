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

// MarginMode controls whether hosting charges may put a user into debt.
type MarginMode string

// AutoPayInterval is the user-selected cadence for margin-call billing.
type AutoPayInterval string

// SkillType describes how a skill is provisioned into a Nix environment.
type SkillType string

// TransactionType classifies balance ledger rows.
type TransactionType string

// TransactionReason records why credits moved.
type TransactionReason string

const (
	DefaultRegistrationCredits = 100
	DefaultAgentPriority       = 100

	MarginUnlimited MarginMode = "unlimited"
	MarginSafe      MarginMode = "safe"

	AutoPayDaily    AutoPayInterval = "day"
	AutoPayWeekly   AutoPayInterval = "week"
	AutoPayMonthly  AutoPayInterval = "month"
	AutoPayHalfYear AutoPayInterval = "half_year"
	AutoPayYearly   AutoPayInterval = "year"

	SkillBuiltin SkillType = "built-in"
	SkillNixpkgs SkillType = "nixpkgs"
	SkillCustom  SkillType = "custom"

	TransactionCredit TransactionType = "credit"
	TransactionDebit  TransactionType = "debit"

	TransactionReasonRegistration TransactionReason = "registration"
	TransactionReasonHosting      TransactionReason = "hosting"
	TransactionReasonLefineReward TransactionReason = "lefine_reward"
	TransactionReasonTopUp        TransactionReason = "topup"
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
	APIKey       string     `gorm:"uniqueIndex" json:"api_key"`
	Subscription string     `gorm:"default:free" json:"subscription"`

	Balance         int             `gorm:"column:balance;default:100" json:"balance"`
	MarginMode      MarginMode      `gorm:"column:margin_mode;default:unlimited" json:"margin_mode"`
	SafeMarginLimit int             `gorm:"column:safe_margin_limit;default:0" json:"safe_margin_limit"`
	AutoPayInterval AutoPayInterval `gorm:"column:auto_pay_interval;default:month" json:"auto_pay_interval"`
	AutoPayDay      int             `gorm:"column:auto_pay_day;default:1" json:"auto_pay_day"`
	SubscriptionEnd *int64          `json:"subscription_end"`
}

// Agent is a user's personal MCP environment. Each user has at most one.
type Agent struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `gorm:"index" json:"user_id"`

	// LLM connection details used both for the proxy mode and to configure the
	// launched CLI.
	LLMProvider string `gorm:"column:llm_provider" json:"llm_provider"`
	LLMAPIKey   string `gorm:"column:llm_api_key" json:"-"`
	LLMBaseURL  string `gorm:"column:llm_base_url" json:"llm_base_url"`
	LLMModel    string `gorm:"column:llm_model" json:"llm_model"`

	// CLI is the agent CLI to launch (e.g. "claude-code"). Empty => proxy mode.
	CLI CLIType `gorm:"column:cli" json:"cli"`

	// Active toggles whether the environment may be used.
	Active bool `gorm:"default:true" json:"active"`

	// Priority is used by safe margin mode to decide which agents should keep
	// running when there are not enough credits for all environments.
	Priority int `gorm:"column:priority;default:100" json:"priority"`
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
	// SkillID is the short identifier from skills.sh (e.g. "claude-api").
	SkillID string `gorm:"column:skill_id;index" json:"skill_id"`
	// Source is the source repo from skills.sh (e.g. "anthropics/skills").
	Source string `gorm:"column:source" json:"source"`
}

// CLI is a known AI CLI that Octra can provision and launch.
type CLI struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `gorm:"uniqueIndex" json:"name"`
	NixAttr    string    `gorm:"column:nix_attr" json:"nix_attr"`
	InstallCmd string    `gorm:"column:install_cmd" json:"install_cmd,omitempty"`
}

// Provider is a known model provider shown in search and used to configure
// proxy-mode or CLI-backed environments.
type Provider struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Key          string    `gorm:"column:key;uniqueIndex" json:"key"`
	Name         string    `gorm:"column:name" json:"name"`
	BaseURL      string    `gorm:"column:base_url" json:"base_url"`
	AuthEnv      string    `gorm:"column:auth_env" json:"auth_env"`
	DefaultModel string    `gorm:"column:default_model" json:"default_model"`
	Description  string    `gorm:"column:description" json:"description"`
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

// Transaction is an append-only balance ledger entry.
type Transaction struct {
	ID        uuid.UUID         `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UserID    uuid.UUID         `gorm:"column:user_id;type:uuid;index;not null" json:"user_id"`
	Type      TransactionType   `gorm:"column:type;not null" json:"type"`
	Amount    int               `gorm:"column:amount;not null" json:"amount"`
	Reason    TransactionReason `gorm:"column:reason;not null" json:"reason"`
	AgentID   *uuid.UUID        `gorm:"column:agent_id;type:uuid;index" json:"agent_id,omitempty"`
	// BalanceAfter is denormalised to make audit trails stable even if later
	// transactions are edited during manual incident repair.
	BalanceAfter int `gorm:"column:balance_after;not null" json:"balance_after"`
}

// UserAPIKey is a user-generated API key with a name and optional expiry.
type UserAPIKey struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UserID    uuid.UUID  `gorm:"column:user_id;type:uuid;index;not null" json:"user_id"`
	Name      string     `gorm:"column:name;not null" json:"name"`
	Key       string     `gorm:"column:key;uniqueIndex;not null" json:"-"`
	ExpiresAt *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`
}

func (m *UserAPIKey) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt)
}

// DashboardEnvironment is a named environment visible on the dashboard.
type DashboardEnvironment struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UserID     uuid.UUID `gorm:"column:user_id;type:uuid;index;not null" json:"user_id"`
	Name       string    `gorm:"column:name;not null" json:"name"`
	Visibility string    `gorm:"column:visibility;default:private" json:"visibility"`
	Active     bool      `gorm:"column:active;default:true" json:"active"`
	Building   bool      `gorm:"column:building;default:false" json:"building"`
}

// UsageMetric stores daily resource usage that drives hosting charges.
type UsageMetric struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UserID        uuid.UUID `gorm:"column:user_id;type:uuid;index;not null" json:"user_id"`
	Date          time.Time `gorm:"column:date;index;not null" json:"date"`
	CPUSeconds    int64     `gorm:"column:cpu_seconds;not null" json:"cpu_seconds"`
	MemoryMBHours int64     `gorm:"column:memory_mb_hours;not null" json:"memory_mb_hours"`
	DiskMB        int64     `gorm:"column:disk_mb;not null" json:"disk_mb"`
	LoadPercent   int       `gorm:"column:load_percent;not null" json:"load_percent"`
}

// RequestMetric records a single chat/completion request routed through Octra.
// These rows drive the request-count metrics shown on the dashboard, both
// globally for a user and broken down per dashboard environment.
type RequestMetric struct {
	ID uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	// CreatedAt is the moment the request completed. It is the time axis used
	// when bucketing metrics into an hourly or daily series.
	CreatedAt time.Time `gorm:"column:created_at;index" json:"created_at"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;index;not null" json:"user_id"`
	// EnvironmentID links the request to a dashboard environment. It is nil for
	// agent (proxy-mode) chats that are not tied to a named environment.
	EnvironmentID *uuid.UUID `gorm:"column:environment_id;type:uuid;index" json:"environment_id,omitempty"`
	// Model is the resolved "provider/model" (or "cli/<name>") string.
	Model string `gorm:"column:model" json:"model,omitempty"`
	// Success is true when the upstream provider answered with a 2xx response.
	Success bool `gorm:"column:success" json:"success"`
	// LatencyMs is the wall-clock duration of the upstream request.
	LatencyMs int64 `gorm:"column:latency_ms" json:"latency_ms"`
}

// CanvasNode stores a single node on the workflow canvas linked to a dashboard
// environment. The entire set of nodes for an environment is replaced atomically.
type CanvasNode struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	EnvironmentID uuid.UUID `gorm:"column:environment_id;type:uuid;index;not null" json:"environment_id"`
	UserID        uuid.UUID `gorm:"column:user_id;type:uuid;index;not null" json:"user_id"`
	ItemID        string    `gorm:"column:item_id;not null" json:"item_id"`
	Kind          string    `gorm:"column:kind;not null" json:"kind"`
	Name          string    `gorm:"column:name;not null" json:"name"`
	Detail        string    `gorm:"column:detail" json:"detail,omitempty"`
	Description   string    `gorm:"column:description" json:"description,omitempty"`
	Meta          string    `gorm:"column:meta;type:jsonb" json:"meta,omitempty"`
	PositionX     float64   `gorm:"column:position_x" json:"position_x"`
	PositionY     float64   `gorm:"column:position_y" json:"position_y"`
	SortOrder     int       `gorm:"column:sort_order;default:0" json:"sort_order"`
}

// AllModels returns every model for AutoMigrate.
func AllModels() []any {
	return []any{&User{}, &Agent{}, &Skill{}, &UserSkill{}, &Transaction{}, &UserAPIKey{}, &DashboardEnvironment{}, &UsageMetric{}, &CLI{}, &Provider{}, &CanvasNode{}, &RequestMetric{}}
}

// BeforeCreate assigns a UUID at the application level so the models work
// across databases (PostgreSQL in production, SQLite in tests) without relying
// on a DB-specific default.
func (m *User) BeforeCreate(*gorm.DB) error {
	if err := ensureID(&m.ID); err != nil {
		return err
	}
	if m.Balance == 0 {
		m.Balance = DefaultRegistrationCredits
	}
	m.ApplyBillingDefaults()
	return nil
}

func (m *Agent) BeforeCreate(*gorm.DB) error {
	if err := ensureID(&m.ID); err != nil {
		return err
	}
	if m.Priority == 0 {
		m.Priority = DefaultAgentPriority
	}
	return nil
}

func (m *Skill) BeforeCreate(*gorm.DB) error                { return ensureID(&m.ID) }
func (m *UserSkill) BeforeCreate(*gorm.DB) error            { return ensureID(&m.ID) }
func (m *Transaction) BeforeCreate(*gorm.DB) error          { return ensureID(&m.ID) }
func (m *UserAPIKey) BeforeCreate(*gorm.DB) error           { return ensureID(&m.ID) }
func (m *DashboardEnvironment) BeforeCreate(*gorm.DB) error { return ensureID(&m.ID) }
func (m *CanvasNode) BeforeCreate(*gorm.DB) error           { return ensureID(&m.ID) }
func (m *CLI) BeforeCreate(*gorm.DB) error                  { return ensureID(&m.ID) }
func (m *Provider) BeforeCreate(*gorm.DB) error             { return ensureID(&m.ID) }
func (m *UsageMetric) BeforeCreate(*gorm.DB) error          { return ensureID(&m.ID) }
func (m *RequestMetric) BeforeCreate(*gorm.DB) error        { return ensureID(&m.ID) }

// ApplyBillingDefaults fills missing billing preference fields.
func (m *User) ApplyBillingDefaults() {
	if m.MarginMode == "" {
		m.MarginMode = MarginUnlimited
	}
	if m.AutoPayInterval == "" {
		m.AutoPayInterval = AutoPayMonthly
	}
	if m.AutoPayDay == 0 {
		m.AutoPayDay = 1
	}
}

func ensureID(id *uuid.UUID) error {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
	return nil
}
