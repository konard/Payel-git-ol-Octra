// Package repository provides persistence for the Octra models. Interfaces are
// declared here so the service layer can be unit-tested with in-memory fakes;
// GORM-backed implementations live alongside in the same package.
package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a lookup matches no row.
var ErrNotFound = errors.New("not found")

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	Update(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, limit int) ([]model.User, error)
}

// AgentRepository persists agents (user environments).
type AgentRepository interface {
	Upsert(ctx context.Context, a *model.Agent) error
	Update(ctx context.Context, a *model.Agent) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Agent, error)
}

// SkillRepository persists the skill catalogue.
type SkillRepository interface {
	GetByName(ctx context.Context, name string) (*model.Skill, error)
	GetBySkillID(ctx context.Context, skillID string) (*model.Skill, error)
	List(ctx context.Context) ([]model.Skill, error)
	Upsert(ctx context.Context, s *model.Skill) error
}

// UserSkillRepository persists per-environment skill installations.
type UserSkillRepository interface {
	Add(ctx context.Context, us *model.UserSkill) error
	ListByAgent(ctx context.Context, agentID uuid.UUID) ([]model.UserSkill, error)
}

// TransactionRepository persists balance ledger entries.
type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Transaction, error)
}

// APIKeyRepository persists user-generated API keys.
type APIKeyRepository interface {
	Create(ctx context.Context, k *model.UserAPIKey) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserAPIKey, error)
	GetByKey(ctx context.Context, key string) (*model.UserAPIKey, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// DashboardEnvironmentRepository persists dashboard environments.
type DashboardEnvironmentRepository interface {
	Create(ctx context.Context, env *model.DashboardEnvironment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.DashboardEnvironment, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.DashboardEnvironment, error)
	ListByBuilding(ctx context.Context, building bool) ([]model.DashboardEnvironment, error)
	Update(ctx context.Context, env *model.DashboardEnvironment) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	SetVisibility(ctx context.Context, id uuid.UUID, visibility string) error
	SetBuilding(ctx context.Context, id uuid.UUID, building bool) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// CLIRepository persists the CLI catalogue.
type CLIRepository interface {
	Upsert(ctx context.Context, c *model.CLI) error
	List(ctx context.Context) ([]model.CLI, error)
	GetByName(ctx context.Context, name string) (*model.CLI, error)
}

// ProviderRepository persists the provider catalogue.
type ProviderRepository interface {
	Upsert(ctx context.Context, p *model.Provider) error
	List(ctx context.Context) ([]model.Provider, error)
	GetByKey(ctx context.Context, key string) (*model.Provider, error)
}

// UsageMetricsRepository persists resource usage snapshots.
type UsageMetricsRepository interface {
	Create(ctx context.Context, metric *model.UsageMetric) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.UsageMetric, error)
}

// RequestMetricsRepository persists per-request telemetry used to build the
// request-count metrics on the dashboard.
type RequestMetricsRepository interface {
	Create(ctx context.Context, metric *model.RequestMetric) error
	// ListByUserSince returns request metrics for a user newer than since,
	// ordered oldest-first. When envID is non-nil only requests tied to that
	// dashboard environment are returned.
	ListByUserSince(ctx context.Context, userID uuid.UUID, envID *uuid.UUID, since time.Time) ([]model.RequestMetric, error)
}

// CanvasNodeRepository persists workflow canvas nodes per dashboard environment.
type CanvasNodeRepository interface {
	Replace(ctx context.Context, envID uuid.UUID, nodes []model.CanvasNode) error
	ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]model.CanvasNode, error)
	DeleteByEnvironment(ctx context.Context, envID uuid.UUID) error
}

// --- GORM implementations ---------------------------------------------------

type userRepo struct{ db *gorm.DB }

// NewUserRepository returns a GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return firstResult(&u, err)
}

func (r *userRepo) GetByAPIKey(ctx context.Context, apiKey string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&u).Error
	return firstResult(&u, err)
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return firstResult(&u, err)
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	return firstResult(&u, err)
}

func (r *userRepo) List(ctx context.Context, limit int) ([]model.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var users []model.User
	err := r.db.WithContext(ctx).Order("created_at asc").Limit(limit).Find(&users).Error
	return users, err
}

type agentRepo struct{ db *gorm.DB }

// NewAgentRepository returns a GORM-backed AgentRepository.
func NewAgentRepository(db *gorm.DB) AgentRepository { return &agentRepo{db: db} }

func (r *agentRepo) Upsert(ctx context.Context, a *model.Agent) error {
	existing, err := r.GetByUserID(ctx, a.UserID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if existing != nil {
		a.ID = existing.ID
		return r.db.WithContext(ctx).Save(a).Error
	}
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *agentRepo) Update(ctx context.Context, a *model.Agent) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *agentRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Agent, error) {
	var a model.Agent
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&a).Error
	return firstResult(&a, err)
}

type skillRepo struct{ db *gorm.DB }

// NewSkillRepository returns a GORM-backed SkillRepository.
func NewSkillRepository(db *gorm.DB) SkillRepository { return &skillRepo{db: db} }

func (r *skillRepo) GetByName(ctx context.Context, name string) (*model.Skill, error) {
	var s model.Skill
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&s).Error
	return firstResult(&s, err)
}

func (r *skillRepo) GetBySkillID(ctx context.Context, skillID string) (*model.Skill, error) {
	var s model.Skill
	err := r.db.WithContext(ctx).Where("skill_id = ?", skillID).First(&s).Error
	return firstResult(&s, err)
}

func (r *skillRepo) Upsert(ctx context.Context, s *model.Skill) error {
	existing, err := r.GetBySkillID(ctx, s.SkillID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if existing != nil {
		s.ID = existing.ID
		s.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(s).Error
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *skillRepo) List(ctx context.Context) ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.WithContext(ctx).Find(&skills).Error
	return skills, err
}

type userSkillRepo struct{ db *gorm.DB }

// NewUserSkillRepository returns a GORM-backed UserSkillRepository.
func NewUserSkillRepository(db *gorm.DB) UserSkillRepository { return &userSkillRepo{db: db} }

func (r *userSkillRepo) Add(ctx context.Context, us *model.UserSkill) error {
	return r.db.WithContext(ctx).Create(us).Error
}

func (r *userSkillRepo) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]model.UserSkill, error) {
	var list []model.UserSkill
	err := r.db.WithContext(ctx).Where("agent_id = ?", agentID).Find(&list).Error
	return list, err
}

type transactionRepo struct{ db *gorm.DB }

// NewTransactionRepository returns a GORM-backed TransactionRepository.
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var list []model.Transaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	return list, err
}

type cliRepo struct{ db *gorm.DB }

// NewCLIRepository returns a GORM-backed CLIRepository.
func NewCLIRepository(db *gorm.DB) CLIRepository { return &cliRepo{db: db} }

func (r *cliRepo) Upsert(ctx context.Context, c *model.CLI) error {
	existing, err := r.GetByName(ctx, c.Name)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if existing != nil {
		c.ID = existing.ID
		c.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(c).Error
	}
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *cliRepo) List(ctx context.Context) ([]model.CLI, error) {
	var list []model.CLI
	err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error
	return list, err
}

func (r *cliRepo) GetByName(ctx context.Context, name string) (*model.CLI, error) {
	var c model.CLI
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error
	return firstResult(&c, err)
}

type providerRepo struct{ db *gorm.DB }

// NewProviderRepository returns a GORM-backed ProviderRepository.
func NewProviderRepository(db *gorm.DB) ProviderRepository { return &providerRepo{db: db} }

func (r *providerRepo) Upsert(ctx context.Context, p *model.Provider) error {
	existing, err := r.GetByKey(ctx, p.Key)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if existing != nil {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(p).Error
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *providerRepo) List(ctx context.Context) ([]model.Provider, error) {
	var list []model.Provider
	err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error
	return list, err
}

func (r *providerRepo) GetByKey(ctx context.Context, key string) (*model.Provider, error) {
	var p model.Provider
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&p).Error
	return firstResult(&p, err)
}

type usageMetricsRepo struct{ db *gorm.DB }

// NewUsageMetricsRepository returns a GORM-backed UsageMetricsRepository.
func NewUsageMetricsRepository(db *gorm.DB) UsageMetricsRepository {
	return &usageMetricsRepo{db: db}
}

func (r *usageMetricsRepo) Create(ctx context.Context, metric *model.UsageMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

func (r *usageMetricsRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.UsageMetric, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var list []model.UsageMetric
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date desc, created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	return list, err
}

type requestMetricsRepo struct{ db *gorm.DB }

// NewRequestMetricsRepository returns a GORM-backed RequestMetricsRepository.
func NewRequestMetricsRepository(db *gorm.DB) RequestMetricsRepository {
	return &requestMetricsRepo{db: db}
}

func (r *requestMetricsRepo) Create(ctx context.Context, metric *model.RequestMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

func (r *requestMetricsRepo) ListByUserSince(ctx context.Context, userID uuid.UUID, envID *uuid.UUID, since time.Time) ([]model.RequestMetric, error) {
	q := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ?", userID, since)
	if envID != nil {
		q = q.Where("environment_id = ?", *envID)
	}
	var list []model.RequestMetric
	// Cap the scan so a very busy account cannot blow up memory while still
	// covering the widest supported window at a generous request rate.
	err := q.Order("created_at asc").Limit(20000).Find(&list).Error
	return list, err
}

type dashboardEnvRepo struct{ db *gorm.DB }

// NewDashboardEnvironmentRepository returns a GORM-backed DashboardEnvironmentRepository.
func NewDashboardEnvironmentRepository(db *gorm.DB) DashboardEnvironmentRepository {
	return &dashboardEnvRepo{db: db}
}

func (r *dashboardEnvRepo) Create(ctx context.Context, env *model.DashboardEnvironment) error {
	return r.db.WithContext(ctx).Create(env).Error
}

func (r *dashboardEnvRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.DashboardEnvironment, error) {
	var env model.DashboardEnvironment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&env).Error
	return firstResult(&env, err)
}

func (r *dashboardEnvRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.DashboardEnvironment, error) {
	var list []model.DashboardEnvironment
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *dashboardEnvRepo) Update(ctx context.Context, env *model.DashboardEnvironment) error {
	return r.db.WithContext(ctx).Save(env).Error
}

func (r *dashboardEnvRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).Model(&model.DashboardEnvironment{}).Where("id = ?", id).Update("active", active).Error
}

func (r *dashboardEnvRepo) SetVisibility(ctx context.Context, id uuid.UUID, visibility string) error {
	return r.db.WithContext(ctx).Model(&model.DashboardEnvironment{}).Where("id = ?", id).Update("visibility", visibility).Error
}

func (r *dashboardEnvRepo) ListByBuilding(ctx context.Context, building bool) ([]model.DashboardEnvironment, error) {
	var list []model.DashboardEnvironment
	err := r.db.WithContext(ctx).Where("building = ?", building).Order("created_at asc").Find(&list).Error
	return list, err
}

func (r *dashboardEnvRepo) SetBuilding(ctx context.Context, id uuid.UUID, building bool) error {
	return r.db.WithContext(ctx).Model(&model.DashboardEnvironment{}).Where("id = ?", id).Update("building", building).Error
}

func (r *dashboardEnvRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.DashboardEnvironment{}).Error
}

type canvasNodeRepo struct{ db *gorm.DB }

// NewCanvasNodeRepository returns a GORM-backed CanvasNodeRepository.
func NewCanvasNodeRepository(db *gorm.DB) CanvasNodeRepository { return &canvasNodeRepo{db: db} }

func (r *canvasNodeRepo) Replace(ctx context.Context, envID uuid.UUID, nodes []model.CanvasNode) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("environment_id = ?", envID).Delete(&model.CanvasNode{}).Error; err != nil {
			return err
		}
		if len(nodes) == 0 {
			return nil
		}
		return tx.Create(nodes).Error
	})
}

func (r *canvasNodeRepo) ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]model.CanvasNode, error) {
	var list []model.CanvasNode
	err := r.db.WithContext(ctx).Where("environment_id = ?", envID).Order("sort_order asc").Find(&list).Error
	return list, err
}

func (r *canvasNodeRepo) DeleteByEnvironment(ctx context.Context, envID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("environment_id = ?", envID).Delete(&model.CanvasNode{}).Error
}

type apiKeyRepo struct{ db *gorm.DB }

// NewAPIKeyRepository returns a GORM-backed APIKeyRepository.
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository { return &apiKeyRepo{db: db} }

func (r *apiKeyRepo) Create(ctx context.Context, k *model.UserAPIKey) error {
	return r.db.WithContext(ctx).Create(k).Error
}

func (r *apiKeyRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserAPIKey, error) {
	var list []model.UserAPIKey
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *apiKeyRepo) GetByKey(ctx context.Context, key string) (*model.UserAPIKey, error) {
	var k model.UserAPIKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&k).Error
	return firstResult(&k, err)
}

func (r *apiKeyRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.UserAPIKey{}).Error
}

// firstResult translates GORM's RecordNotFound into ErrNotFound.
func firstResult[T any](v *T, err error) (*T, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}
