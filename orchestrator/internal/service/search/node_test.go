package search

import "testing"

// TestNeedsWebSearch проверяет эвристику решения о необходимости веб-поиска
// (issue #97): явные запросы и факты — поиск; код и чистая математика — нет.
func TestNeedsWebSearch(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		// Явные просьбы найти информацию.
		{"explicit english", "search for the best pizza in town", true},
		{"explicit look up", "look up the capital of France", true},
		{"explicit russian", "найди информацию про новый закон", true},
		{"explicit poishi", "поищи курс доллара", true},

		// Свежие/актуальные факты, которые модель знать не может.
		{"recency latest", "what is the latest version of Go", true},
		{"recency today", "what is the weather today", true},
		{"recency russian", "какие последние новости", true},

		// Факт-запросы (lookup).
		{"lookup who is", "who is the CEO of Acme", true},
		{"lookup price", "what is the price of bitcoin", true},
		{"lookup russian", "какая столица Австралии", true},

		// Год как сигнал конкретного периода.
		{"year reference", "Olympic Games 2026 schedule", true},

		// Самодостаточные задачи — поиск НЕ нужен.
		{"code request", "напиши hello world на python", false},
		{"code english", "write a script that sorts a list", false},
		{"math expression", "2 + 2 * 2 =", false},
		{"plain question", "explain how recursion works", false},
		{"empty", "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, reason := NeedsWebSearch(c.text)
			if got != c.want {
				t.Errorf("NeedsWebSearch(%q) = %v (%s), want %v", c.text, got, reason, c.want)
			}
		})
	}
}

// TestPlanNode_AutoCreate — поиск нужен, ноды на канве нет → создаём её.
func TestPlanNode_AutoCreate(t *testing.T) {
	n := PlanNode("find the latest news about Go", false, "", "", "gpt-4o", "openai", "universal")
	if !n.Enabled {
		t.Fatalf("expected Enabled, got disabled (reason=%q)", n.Reason)
	}
	if !n.Created {
		t.Errorf("expected Created=true for auto-created node")
	}
	if n.Explicit {
		t.Errorf("expected Explicit=false")
	}
	if n.Model != "gpt-4o" || n.Provider != "openai" {
		t.Errorf("expected fallback to task model/provider, got model=%q provider=%q", n.Model, n.Provider)
	}
	if n.AttachTo != "universal" {
		t.Errorf("expected AttachTo=universal, got %q", n.AttachTo)
	}
}

// TestPlanNode_ExplicitButUnused — нода размещена, но поиск не нужен → не активна.
func TestPlanNode_ExplicitButUnused(t *testing.T) {
	n := PlanNode("напиши hello world на python", true, "claude-3", "anthropic", "gpt-4o", "openai", "worker")
	if n.Enabled {
		t.Errorf("expected disabled for a code task even with an explicit node (reason=%q)", n.Reason)
	}
	if !n.Explicit {
		t.Errorf("expected Explicit=true")
	}
	if n.Created {
		t.Errorf("expected Created=false for explicit node")
	}
}

// TestPlanNode_ExplicitModel — модель ноды имеет приоритет над моделью задачи.
func TestPlanNode_ExplicitModel(t *testing.T) {
	n := PlanNode("search for current exchange rates", true, "claude-3", "anthropic", "gpt-4o", "openai", "manager")
	if !n.Enabled {
		t.Fatalf("expected Enabled")
	}
	if n.Model != "claude-3" || n.Provider != "anthropic" {
		t.Errorf("expected node model/provider to win, got model=%q provider=%q", n.Model, n.Provider)
	}
	if n.Created {
		t.Errorf("expected Created=false because the node is explicit")
	}
}
