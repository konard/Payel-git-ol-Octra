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
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

// AgentRepository persists agents (user environments).
type AgentRepository interface {
	Upsert(ctx context.Context, a *model.Agent) error
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

// --- GORM implementations ---------------------------------------------------

type userRepo struct{ db *gorm.DB }

// NewUserRepository returns a GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
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
