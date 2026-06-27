package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/skillapi"
	ts "backend/internal/typesense"
)

type mockTSClient struct {
	indexed []ts.SkillDocument
}

func (m *mockTSClient) EnsureCollection(ctx context.Context) error { return nil }
func (m *mockTSClient) IndexSkills(ctx context.Context, docs []ts.SkillDocument) error {
	m.indexed = append(m.indexed, docs...)
	return nil
}
func (m *mockTSClient) SearchSkills(ctx context.Context, query string, limit int) (*interface{}, error) {
	return nil, nil
}

func skillapiHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	resp := skillapi.SearchResponse{
		Query:      q,
		SearchType: "fuzzy",
		Count:      1,
		DurationMs: 10,
	}
	switch q {
	case "claude":
		resp.Skills = []skillapi.Skill{
			{ID: "anthropics/skills/claude-api", SkillID: "claude-api", Name: "claude-api", Source: "anthropics/skills"},
		}
	case "python":
		resp.Skills = []skillapi.Skill{
			{ID: "wshobson/agents/python-patterns", SkillID: "python-patterns", Name: "python-patterns", Source: "wshobson/agents"},
		}
	case "go":
		resp.Skills = []skillapi.Skill{
			{ID: "google-labs-code/stitch-skills/design-md", SkillID: "design-md", Name: "design-md", Source: "google-labs-code/stitch-skills"},
		}
	default:
		resp.Skills = nil
		resp.Count = 0
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func TestSkillSyncServiceSync(t *testing.T) {
	db := newTestDB(t)
	skillsRepo := repository.NewSkillRepository(db)
	mockTS := &mockTSClient{}

	srv := httptest.NewServer(http.HandlerFunc(skillapiHandler))
	defer srv.Close()

	apiClient := skillapi.NewWithClient(srv.Client())
	apiClient.BaseURL = srv.URL

	svc := &SkillSyncService{
		skillsRepo: skillsRepo,
		apiClient:  apiClient,
		tsClient:   mockTS,
		interval:   1 * time.Hour,
	}

	queries := []string{"claude", "python", "go", "unknown"}
	svc.SyncWithQueries(context.Background(), queries)

	skills, err := skillsRepo.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills in DB, got %d", len(skills))
	}

	skillMap := make(map[string]*model.Skill)
	for i := range skills {
		skillMap[skills[i].SkillID] = &skills[i]
	}

	if s, ok := skillMap["claude-api"]; !ok {
		t.Fatal("claude-api not found in DB")
	} else if s.Source != "anthropics/skills" {
		t.Fatalf("claude-api source = %q", s.Source)
	}

	if s, ok := skillMap["python-patterns"]; !ok {
		t.Fatal("python-patterns not found in DB")
	} else if s.Name != "python-patterns" {
		t.Fatalf("python-patterns name = %q", s.Name)
	}

	if len(mockTS.indexed) != 3 {
		t.Fatalf("expected 3 skills indexed in typesense, got %d", len(mockTS.indexed))
	}
}

func TestSkillSyncServiceUpdateExisting(t *testing.T) {
	db := newTestDB(t)
	skillsRepo := repository.NewSkillRepository(db)

	existing := &model.Skill{
		Name:       "old-name",
		Type:       model.SkillNixpkgs,
		InstallCmd: "old-cmd",
		SkillID:    "claude-api",
		Source:     "old/source",
	}
	if err := skillsRepo.Upsert(context.Background(), existing); err != nil {
		t.Fatalf("seed: %v", err)
	}

	mockTS := &mockTSClient{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(skillapi.SearchResponse{
			Query: "claude",
			Skills: []skillapi.Skill{
				{ID: "anthropics/skills/claude-api", SkillID: "claude-api", Name: "new-name", Source: "new/source"},
			},
			Count: 1,
		})
	}))
	defer srv.Close()

	apiClient := skillapi.NewWithClient(srv.Client())
	apiClient.BaseURL = srv.URL

	svc := &SkillSyncService{
		skillsRepo: skillsRepo,
		apiClient:  apiClient,
		tsClient:   mockTS,
		interval:   1 * time.Hour,
	}

	svc.SyncWithQueries(context.Background(), []string{"claude"})

	updated, err := skillsRepo.GetBySkillID(context.Background(), "claude-api")
	if err != nil {
		t.Fatalf("GetBySkillID: %v", err)
	}
	if updated.Name != "new-name" {
		t.Fatalf("expected name updated to 'new-name', got %q", updated.Name)
	}
	if updated.Source != "new/source" {
		t.Fatalf("expected source updated to 'new/source', got %q", updated.Source)
	}
}

func TestSkillSyncServiceDeduplicates(t *testing.T) {
	db := newTestDB(t)
	skillsRepo := repository.NewSkillRepository(db)
	mockTS := &mockTSClient{}

	called := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		json.NewEncoder(w).Encode(skillapi.SearchResponse{
			Query: r.URL.Query().Get("q"),
			Skills: []skillapi.Skill{
				{ID: "anthropics/skills/claude-api", SkillID: "claude-api", Name: "claude-api", Source: "anthropics/skills"},
			},
			Count: 1,
		})
	}))
	defer srv.Close()

	apiClient := skillapi.NewWithClient(srv.Client())
	apiClient.BaseURL = srv.URL

	svc := &SkillSyncService{
		skillsRepo: skillsRepo,
		apiClient:  apiClient,
		tsClient:   mockTS,
		interval:   1 * time.Hour,
	}

	svc.SyncWithQueries(context.Background(), []string{"claude", "claude", "claude"})

	skills, _ := skillsRepo.List(context.Background())
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill (deduplicated), got %d", len(skills))
	}
}
