// Package skills implements a file-based "system skills" library for Octra
// agents. Instead of relying only on hardcoded inline prompts, the agents READ
// skill definitions from Markdown files at runtime and inject the matching
// expert guidance into their LLM prompts.
//
// Each skill lives in library/<slug>.md as a Markdown document with a small
// YAML-style frontmatter block:
//
//	---
//	name: Backend Development
//	slug: backend
//	area: Backend
//	keywords: backend, server, api, ...
//	tech: go, golang, python, ...
//	---
//	## Backend Development skill
//	...expert guidance...
//
// The library is embedded into the binary so skills travel with the service,
// and new skills can be added simply by dropping another .md file in library/.
// Content is curated/inspired by community skill and prompt collections such as
// FlowiseAI, LangChain Hub, PromptBase and awesome-chatgpt-prompts.
package skills

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed library/*.md
var libraryFS embed.FS

// Skill is a single specialized capability loaded from a Markdown file.
type Skill struct {
	Name     string   // human-readable name, e.g. "Backend Development"
	Slug     string   // stable identifier, e.g. "backend"
	Area     string   // coverage area from the issue, e.g. "Backend"
	Keywords []string // lowercase keywords used to match a task/role
	Tech     []string // lowercase tech-stack tokens that imply this skill
	Body     string   // Markdown guidance injected into prompts
}

// registry holds every skill parsed from the embedded library, in stable order.
var registry []Skill

func init() {
	registry = loadAll(libraryFS, "library")
}

// loadAll reads and parses every *.md file under dir in the given filesystem.
func loadAll(fsys fs.FS, dir string) []Skill {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil
	}
	var out []Skill
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, dir+"/"+e.Name())
		if err != nil {
			continue
		}
		if s, ok := parseSkill(string(data)); ok {
			if s.Slug == "" {
				s.Slug = strings.TrimSuffix(e.Name(), ".md")
			}
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

// parseSkill splits a skill file into frontmatter and body, returning the skill.
func parseSkill(content string) (Skill, bool) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return Skill{}, false
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Skill{}, false
	}
	header := rest[:end]
	body := strings.TrimLeft(rest[end+len("\n---"):], "\n")

	s := Skill{Body: strings.TrimSpace(body)}
	for _, line := range strings.Split(header, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		switch key {
		case "name":
			s.Name = val
		case "slug":
			s.Slug = val
		case "area":
			s.Area = val
		case "keywords":
			s.Keywords = splitList(val)
		case "tech":
			s.Tech = splitList(val)
		}
	}
	if s.Body == "" {
		return Skill{}, false
	}
	return s, true
}

// splitList parses a comma-separated, lowercased token list.
func splitList(val string) []string {
	var out []string
	for _, part := range strings.Split(val, ",") {
		p := strings.ToLower(strings.TrimSpace(part))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// All returns every loaded skill (stable order). The slice is a copy so callers
// cannot mutate the registry.
func All() []Skill {
	out := make([]Skill, len(registry))
	copy(out, registry)
	return out
}

// score rates how well a skill matches the given role/techStack/task text.
// Role matches weigh most (the manager hires by role), tech-stack matches next,
// and free-text keyword matches least.
func (s Skill) score(role, techStack, task string) int {
	role = strings.ToLower(role)
	tech := strings.ToLower(techStack)
	text := strings.ToLower(task)

	score := 0
	for _, kw := range s.Keywords {
		if strings.Contains(role, kw) {
			score += 5
		}
		if strings.Contains(text, kw) {
			score += 2
		}
	}
	for _, t := range s.Tech {
		if strings.Contains(tech, t) {
			score += 3
		}
	}
	// The slug itself is a strong signal when it appears in the role.
	if s.Slug != "" && strings.Contains(role, s.Slug) {
		score += 4
	}
	return score
}

// Match returns the skills that best fit a worker's role, tech stack and task,
// ordered from most to least relevant. At most max skills are returned; pass a
// non-positive max to use the default of 2. Returns nil when nothing matches.
func Match(role, techStack, task string, max int) []Skill {
	if max <= 0 {
		max = 2
	}
	type scored struct {
		skill Skill
		score int
	}
	var ranked []scored
	for _, s := range registry {
		if sc := s.score(role, techStack, task); sc > 0 {
			ranked = append(ranked, scored{s, sc})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if len(ranked) > max {
		ranked = ranked[:max]
	}
	out := make([]Skill, len(ranked))
	for i, r := range ranked {
		out[i] = r.skill
	}
	return out
}

// Guidance returns a ready-to-inject Markdown block with the expert guidance of
// the best-matching skill(s) for the given role/tech/task, or "" when no skill
// applies. It is safe to append directly to any worker prompt.
func Guidance(role, techStack, task string) string {
	matched := Match(role, techStack, task, 2)
	if len(matched) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n=== APPLY THIS EXPERT SKILL ===\n")
	for i, s := range matched {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(s.Body)
		b.WriteString("\n")
	}
	b.WriteString("=== END SKILL ===")
	return b.String()
}

// Catalog returns a short one-line-per-skill summary of the available skill
// areas, suitable for telling the boss which specialities the system covers.
func Catalog() string {
	if len(registry) == 0 {
		return ""
	}
	var b strings.Builder
	for _, s := range registry {
		area := s.Area
		if area == "" {
			area = s.Name
		}
		b.WriteString("- ")
		b.WriteString(area)
		b.WriteString(" (")
		b.WriteString(s.Slug)
		b.WriteString(")\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

