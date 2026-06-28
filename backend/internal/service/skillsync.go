package service

import (
	"context"
	"log"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/skillapi"
	ts "backend/internal/typesense"
)

type SkillSyncService struct {
	skillsRepo repository.SkillRepository
	apiClient  *skillapi.Client
	tsClient   ts.ClientInterface
	interval   time.Duration
}

func NewSkillSyncService(skillsRepo repository.SkillRepository, tsClient ts.ClientInterface, interval time.Duration, baseURL string) *SkillSyncService {
	return &SkillSyncService{
		skillsRepo: skillsRepo,
		apiClient:  skillapi.NewWithBaseURL(baseURL),
		tsClient:   tsClient,
		interval:   interval,
	}
}

func (s *SkillSyncService) Start(ctx context.Context) {
	log.Printf("skillsync: starting sync every %v", s.interval)

	go func() {
		s.Sync(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.Sync(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *SkillSyncService) Sync(ctx context.Context) {
	s.SyncWithQueries(ctx, skillapi.DefaultQueries)
}

func (s *SkillSyncService) SyncWithQueries(ctx context.Context, queries []string) {
	log.Println("skillsync: starting sync cycle...")
	seen := make(map[string]bool)
	var allDocs []ts.SkillDocument

	for _, q := range queries {
		resp, err := s.apiClient.Search(q, 100)
		if err != nil {
			log.Printf("skillsync: search %q failed: %v", q, err)
			continue
		}
		if resp == nil {
			continue
		}
		for _, apiSkill := range resp.Skills {
			if seen[apiSkill.SkillID] {
				continue
			}
			seen[apiSkill.SkillID] = true

			installCmd := skillapi.InstallCmd(apiSkill)

			existing, err := s.skillsRepo.GetBySkillID(ctx, apiSkill.SkillID)
			if err != nil && err != repository.ErrNotFound {
				log.Printf("skillsync: lookup %s: %v", apiSkill.SkillID, err)
				continue
			}

			skill := &model.Skill{
				Name:       apiSkill.Name,
				Type:       model.SkillNixpkgs,
				InstallCmd: installCmd,
				SkillID:    apiSkill.SkillID,
				Source:     apiSkill.Source,
			}

			if existing == nil {
				if err := s.skillsRepo.Upsert(ctx, skill); err != nil {
					log.Printf("skillsync: create %s: %v", apiSkill.SkillID, err)
				}
			} else if existing.Name != skill.Name || existing.Source != skill.Source || existing.InstallCmd != skill.InstallCmd {
				skill.ID = existing.ID
				skill.CreatedAt = existing.CreatedAt
				if err := s.skillsRepo.Upsert(ctx, skill); err != nil {
					log.Printf("skillsync: update %s: %v", apiSkill.SkillID, err)
				}
			}

			allDocs = append(allDocs, ts.SkillDocument{
				ID:         apiSkill.ID,
				SkillID:    apiSkill.SkillID,
				Name:       apiSkill.Name,
				Source:     apiSkill.Source,
				InstallCmd: installCmd,
			})
		}
		log.Printf("skillsync: %q -> %d skills (total unique: %d)", q, len(resp.Skills), len(seen))
		time.Sleep(200 * time.Millisecond)
	}

	if s.tsClient != nil && len(allDocs) > 0 {
		if err := s.tsClient.IndexSkills(ctx, allDocs); err != nil {
			log.Printf("skillsync: typesense index: %v", err)
		} else {
			log.Printf("skillsync: indexed %d skills in typesense", len(allDocs))
		}
	}

	log.Printf("skillsync: sync complete: %d total unique skills", len(seen))
}
