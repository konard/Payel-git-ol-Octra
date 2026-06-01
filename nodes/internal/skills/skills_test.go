package skills

import (
	"strings"
	"testing"
)

// TestLibraryLoads — все скилл-файлы из issue должны успешно прочитаться из
// встроенной библиотеки (полноценное «чтение скиллов», а не инлайновые промпты).
func TestLibraryLoads(t *testing.T) {
	want := []string{"research", "presentation", "frontend", "backend", "devops", "proxy", "vpn"}
	all := All()
	if len(all) < len(want) {
		t.Fatalf("expected at least %d skills, got %d", len(want), len(all))
	}
	have := map[string]Skill{}
	for _, s := range all {
		have[s.Slug] = s
	}
	for _, slug := range want {
		s, ok := have[slug]
		if !ok {
			t.Errorf("missing skill %q", slug)
			continue
		}
		if s.Name == "" || s.Body == "" || len(s.Keywords) == 0 {
			t.Errorf("skill %q not fully parsed: name=%q keywords=%v bodyLen=%d", slug, s.Name, s.Keywords, len(s.Body))
		}
	}
}

// TestMatchByRole — роль воркера должна выбирать соответствующий скилл.
func TestMatchByRole(t *testing.T) {
	cases := []struct {
		role, tech, task, wantSlug string
	}{
		{"frontend developer", "react", "build a dashboard UI", "frontend"},
		{"backend engineer", "go", "build a REST API", "backend"},
		{"devops engineer", "docker", "set up CI/CD pipeline", "devops"},
		{"proxy developer", "go", "write a reverse proxy", "proxy"},
		{"network engineer", "wireguard", "configure a VPN tunnel", "vpn"},
		{"research analyst", "markdown", "investigate the market", "research"},
		{"presentation designer", "pptx", "make a slide deck", "presentation"},
	}
	for _, c := range cases {
		got := Match(c.role, c.tech, c.task, 1)
		if len(got) == 0 {
			t.Errorf("role %q: no skill matched", c.role)
			continue
		}
		if got[0].Slug != c.wantSlug {
			t.Errorf("role %q: want top skill %q, got %q", c.role, c.wantSlug, got[0].Slug)
		}
	}
}

// TestGuidanceInjectable — Guidance возвращает непустой блок для известной роли
// и пустую строку, когда ничего не подходит.
func TestGuidanceInjectable(t *testing.T) {
	g := Guidance("backend engineer", "go", "build an API")
	if !strings.Contains(g, "APPLY THIS EXPERT SKILL") {
		t.Errorf("expected skill marker in guidance:\n%s", g)
	}
	if !strings.Contains(g, "Backend") {
		t.Errorf("expected backend guidance, got:\n%s", g)
	}
	if empty := Guidance("", "", "zzzz no match here qqq"); empty != "" {
		t.Errorf("expected empty guidance for no match, got:\n%s", empty)
	}
}

// TestCatalogListsAreas — каталог для босса перечисляет области покрытия.
func TestCatalogListsAreas(t *testing.T) {
	cat := Catalog()
	for _, want := range []string{"Backend", "Frontend", "Proxy", "Vpn"} {
		if !strings.Contains(cat, want) {
			t.Errorf("catalog missing %q:\n%s", want, cat)
		}
	}
}

// TestMatchLimit — Match не возвращает больше скиллов, чем запрошено.
func TestMatchLimit(t *testing.T) {
	got := Match("backend frontend devops engineer", "go react docker", "api ui pipeline", 2)
	if len(got) > 2 {
		t.Errorf("expected at most 2 skills, got %d", len(got))
	}
}
