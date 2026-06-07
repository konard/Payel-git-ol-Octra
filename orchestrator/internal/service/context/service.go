package context

import (
	"context"
	"fmt"
	"log"
	"strings"

	"orchestrator/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	pg  *PostgresStore
	rdb *RedisCache
}

func NewService(db *gorm.DB, rdb *RedisCache) *Service {
	return &Service{
		pg:  NewPostgresStore(db),
		rdb: rdb,
	}
}

func (s *Service) Save(ctx context.Context, entry *models.ContextEntry) error {
	if err := s.pg.Save(ctx, entry); err != nil {
		return fmt.Errorf("context save: %w", err)
	}
	s.rdb.Save(ctx, entry)
	return nil
}

func (s *Service) SaveFromAI(ctx context.Context, projectID uuid.UUID, sourceAgentID, aiJSON string) *models.ContextEntry {
	entry := ExtractEntryFromJSON(projectID, sourceAgentID, aiJSON)
	if entry == nil {
		return nil
	}
	if IsForgetRequest(entry.Content) {
		n, err := s.pg.ForgetByContent(ctx, projectID, entry.Content)
		if err != nil {
			log.Printf("[Context] Forget error: %v", err)
		} else {
			log.Printf("[Context] Forgotten %d entries by request", n)
		}
		s.rdb.ClearProject(ctx, projectID.String())
		return nil
	}
	if err := s.Save(ctx, entry); err != nil {
		log.Printf("[Context] Save error: %v", err)
		return nil
	}
	s.cleanupAfterSave(ctx, entry.ProjectID)
	return entry
}

func (s *Service) SaveGlobalRule(ctx context.Context, projectID uuid.UUID, sourceAgentID, content string) error {
	return s.pg.SaveGlobalRule(ctx, projectID, sourceAgentID, content)
}

func (s *Service) GetForPrompt(ctx context.Context, projectID uuid.UUID, agentID, managerID string) string {
	pid := projectID.String()

	var global, team, individual []models.ContextEntry

	if cachedGlobal, err := s.rdb.GetGlobal(ctx, pid); err == nil && len(cachedGlobal) > 0 {
		global = cachedGlobal
	} else if entries, err := s.pg.GetGlobal(ctx, projectID); err == nil {
		global = entries
		s.rdb.Populate(ctx, entries)
	}

	if managerID != "" {
		if cachedTeam, err := s.rdb.GetTeam(ctx, pid, managerID); err == nil && len(cachedTeam) > 0 {
			team = cachedTeam
		} else if entries, err := s.pg.GetTeam(ctx, projectID, managerID); err == nil {
			team = entries
			s.rdb.Populate(ctx, entries)
		}
	}

	if agentID != "" {
		if cachedIndiv, err := s.rdb.GetIndividual(ctx, pid, agentID); err == nil && len(cachedIndiv) > 0 {
			individual = cachedIndiv
		} else if entries, err := s.pg.GetIndividual(ctx, projectID, agentID); err == nil {
			individual = entries
			s.rdb.Populate(ctx, entries)
		}
	}

	return formatPromptBlock(global, team, individual)
}

func (s *Service) ForgetContent(ctx context.Context, projectID uuid.UUID, content string) (int64, error) {
	n, err := s.pg.ForgetByContent(ctx, projectID, content)
	if err == nil {
		s.rdb.ClearProject(ctx, projectID.String())
	}
	return n, err
}

func (s *Service) ForgetID(ctx context.Context, id uuid.UUID) error {
	return s.pg.ForgetByID(ctx, id)
}

func (s *Service) Cleanup(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return s.pg.CleanupMessages(ctx, projectID)
}

func (s *Service) cleanupAfterSave(ctx context.Context, projectID uuid.UUID) {
	n, err := s.pg.CleanupMessages(ctx, projectID)
	if err != nil {
		log.Printf("[Context] Cleanup error: %v", err)
	}
	if n > 0 {
		s.rdb.ClearProject(ctx, projectID.String())
		log.Printf("[Context] Cleaned up %d old messages for %s", n, projectID)
	}
}

func formatPromptBlock(global, team, individual []models.ContextEntry) string {
	var parts []string

	if len(global) > 0 {
		var lines []string
		for _, e := range global {
			mark := ""
			if e.Important {
				mark = " (important)"
			}
			lines = append(lines, fmt.Sprintf("  - %s%s", e.Content, mark))
		}
		parts = append(parts, "--- GLOBAL RULES ---\n"+strings.Join(lines, "\n"))
	}

	if len(team) > 0 {
		var lines []string
		for _, e := range team {
			lines = append(lines, fmt.Sprintf("  - %s", e.Content))
		}
		parts = append(parts, "--- TEAM CONTEXT ---\n"+strings.Join(lines, "\n"))
	}

	if len(individual) > 0 {
		var lines []string
		for _, e := range individual {
			lines = append(lines, fmt.Sprintf("  - %s", e.Content))
		}
		parts = append(parts, "--- YOUR CONTEXT ---\n"+strings.Join(lines, "\n"))
	}

	if len(parts) == 0 {
		return ""
	}

	return "\n\n=== CONTEXT ===\n" + strings.Join(parts, "\n\n") + "\n=== END CONTEXT ==="
}
