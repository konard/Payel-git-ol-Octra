// Package repository provides persistence for the Octra models. Interfaces are
// declared here so the service layer can be unit-tested with in-memory fakes;
// GORM-backed implementations live alongside in the same package.
package repository

import (
	"context"
	"errors"

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
	List(ctx context.Context) ([]model.Skill, error)
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
	Update(ctx context.Context, env *model.DashboardEnvironment) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// UsageMetricsRepository persists resource usage snapshots.
type UsageMetricsRepository interface {
	Create(ctx context.Context, metric *model.UsageMetric) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.UsageMetric, error)
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

func (r *dashboardEnvRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.DashboardEnvironment{}).Error
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
