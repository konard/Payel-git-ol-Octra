package skillapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstallCmd(t *testing.T) {
	tests := []struct {
		name string
		skill Skill
		want string
	}{
		{
			name: "standard skill",
			skill: Skill{ID: "anthropics/skills/claude-api", SkillID: "claude-api", Name: "claude-api", Source: "anthropics/skills"},
			want: "npx skills add https://github.com/anthropics/skills --skill claude-api",
		},
		{
			name: "nested source",
			skill: Skill{ID: "vercel-labs/agent-skills/vercel-react-best-practices", SkillID: "vercel-react-best-practices", Name: "vercel-react-best-practices", Source: "vercel-labs/agent-skills"},
			want: "npx skills add https://github.com/vercel-labs/agent-skills --skill vercel-react-best-practices",
		},
		{
			name:   "empty skill",
			skill:  Skill{},
			want:   "npx skills add https://github.com/ --skill ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InstallCmd(tt.skill)
			if got != tt.want {
				t.Fatalf("InstallCmd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultQueriesNotEmpty(t *testing.T) {
	if len(DefaultQueries) == 0 {
		t.Fatal("DefaultQueries should not be empty")
	}
}

func TestClientSearch(t *testing.T) {
	expected := SearchResponse{
		Query:      "claude",
		SearchType: "fuzzy",
		Skills: []Skill{
			{ID: "anthropics/skills/claude-api", SkillID: "claude-api", Name: "claude-api", Source: "anthropics/skills"},
		},
		Count:      1,
		DurationMs: 100,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "claude" {
			t.Fatalf("unexpected query: %s", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("limit") != "50" {
			t.Fatalf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	client := NewWithClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Search("claude", 50)
	if err != nil {
		t.Fatalf("Search() returned error: %v", err)
	}
	if resp.Query != "claude" {
		t.Fatalf("resp.Query = %q, want %q", resp.Query, "claude")
	}
	if len(resp.Skills) != 1 {
		t.Fatalf("len(resp.Skills) = %d, want 1", len(resp.Skills))
	}
	if resp.Skills[0].SkillID != "claude-api" {
		t.Fatalf("skill ID = %q, want %q", resp.Skills[0].SkillID, "claude-api")
	}
}

func TestClientSearchServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewWithClient(srv.Client())
	client.BaseURL = srv.URL

	_, err := client.Search("test", 10)
	if err == nil {
		t.Fatal("expected error on server error, got nil")
	}
}

func TestClientSearchInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer srv.Close()

	client := NewWithClient(srv.Client())
	client.BaseURL = srv.URL

	_, err := client.Search("test", 10)
	if err == nil {
		t.Fatal("expected error on invalid JSON, got nil")
	}
}
