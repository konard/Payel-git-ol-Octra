package context

import (
	"context"
	"fmt"
	"time"

	"orchestrator/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(db *gorm.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Save(ctx context.Context, entry *models.ContextEntry) error {
	return s.db.WithContext(ctx).Create(entry).Error
}

func (s *PostgresStore) GetGlobal(ctx context.Context, projectID uuid.UUID) ([]models.ContextEntry, error) {
	var entries []models.ContextEntry
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND scope = 'global' AND forgotten = false", projectID).
		Order("timestamp DESC").
		Find(&entries).Error
	return entries, err
}

func (s *PostgresStore) GetTeam(ctx context.Context, projectID uuid.UUID, managerID string) ([]models.ContextEntry, error) {
	var entries []models.ContextEntry
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND scope = 'team' AND target_id = ? AND forgotten = false", projectID, managerID).
		Order("timestamp DESC").
		Find(&entries).Error
	return entries, err
}

func (s *PostgresStore) GetIndividual(ctx context.Context, projectID uuid.UUID, agentID string) ([]models.ContextEntry, error) {
	var entries []models.ContextEntry
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND scope = 'individual' AND target_id = ? AND forgotten = false", projectID, agentID).
		Order("timestamp DESC").
		Find(&entries).Error
	return entries, err
}

func (s *PostgresStore) GetByProject(ctx context.Context, projectID uuid.UUID) ([]models.ContextEntry, error) {
	var entries []models.ContextEntry
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND forgotten = false", projectID).
		Order("timestamp DESC").
		Find(&entries).Error
	return entries, err
}

func (s *PostgresStore) ForgetByContent(ctx context.Context, projectID uuid.UUID, content string) (int64, error) {
	res := s.db.WithContext(ctx).
		Model(&models.ContextEntry{}).
		Where("project_id = ? AND content ILIKE ?", projectID, "%"+content+"%").
		Update("forgotten", true)
	return res.RowsAffected, res.Error
}

func (s *PostgresStore) ForgetByID(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&models.ContextEntry{}).
		Where("id = ?", id).
		Update("forgotten", true).Error
}

func (s *PostgresStore) CleanupMessages(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	s.db.WithContext(ctx).
		Model(&models.ContextEntry{}).
		Where("project_id = ? AND context_type = 'message' AND forgotten = false", projectID).
		Count(&count)

	if count <= MessageLimit {
		return 0, nil
	}

	excess := count - MessageLimit
	var toDelete []string
	s.db.WithContext(ctx).
		Model(&models.ContextEntry{}).
		Where("project_id = ? AND context_type = 'message' AND forgotten = false", projectID).
		Order("timestamp ASC").
		Limit(int(excess)).
		Pluck("id", &toDelete)

	if len(toDelete) == 0 {
		return 0, nil
	}

	ids := make([]uuid.UUID, len(toDelete))
	for i, id := range toDelete {
		uid, err := uuid.Parse(id)
		if err != nil {
			return 0, fmt.Errorf("parse uuid %s: %w", id, err)
		}
		ids[i] = uid
	}

	res := s.db.WithContext(ctx).
		Model(&models.ContextEntry{}).
		Where("id IN ?", ids).
		Update("forgotten", true)
	return res.RowsAffected, res.Error
}

func (s *PostgresStore) SaveGlobalRule(ctx context.Context, projectID uuid.UUID, sourceAgentID, content string) error {
	entry := &models.ContextEntry{
		ID:            uuid.New(),
		ProjectID:     projectID,
		Scope:         "global",
		SourceAgentID: sourceAgentID,
		ContextType:   "global_rule",
		Content:       content,
		Timestamp:     time.Now().UTC(),
	}
	return s.db.WithContext(ctx).Create(entry).Error
}
