package skills

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed fragments/*.md
var fragmentsFS embed.FS

// Fragment is a single atomic skill fragment — a small, focused piece of
// expert guidance that can be independently selected and injected.
// Compared to the monolithic Skill, fragments are 3-10 lines each, making
// them cheap to send only when needed.
type Fragment struct {
	Name     string   // human-readable, e.g. "Backend API Design"
	Slug     string   // stable identifier, e.g. "backend-api"
	Area     string   // domain area, e.g. "Backend"
	Keywords []string // lowercase keywords for matching
	Tech     []string // lowercase tech-stack tokens
	Weight   int      // importance (1-5), for priority selection
	Body     string   // markdown guidance (3-10 lines)
}

// Warehouse is the registry of all skill fragments. It provides catalog,
// search, and selection operations so that a Manager can pick exactly the
// fragments each Worker needs — no more sending entire skill files.
type Warehouse struct {
	fragments  []*Fragment
	bySlug     map[string]*Fragment
	byArea     map[string][]*Fragment
	byTech     map[string][]*Fragment
}

// NewWarehouse loads all fragments from the embedded library. It panics on
// parse errors so startup fails fast if a fragment is malformed.
func NewWarehouse() *Warehouse {
	w := &Warehouse{
		bySlug: make(map[string]*Fragment),
		byArea: make(map[string][]*Fragment),
		byTech: make(map[string][]*Fragment),
	}
	for _, f := range loadFragments(fragmentsFS, "fragments") {
		w.fragments = append(w.fragments, f)
		w.bySlug[f.Slug] = f
		w.byArea[f.Area] = append(w.byArea[f.Area], f)
		for _, t := range f.Tech {
			w.byTech[t] = append(w.byTech[t], f)
		}
	}
	return w
}

func loadFragments(fsys fs.FS, dir string) []*Fragment {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil
	}
	var out []*Fragment
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, dir+"/"+e.Name())
		if err != nil {
			continue
		}
		if f, ok := parseFragment(string(data)); ok {
			if f.Slug == "" {
				f.Slug = strings.TrimSuffix(e.Name(), ".md")
			}
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

func parseFragment(content string) (*Fragment, bool) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return nil, false
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, false
	}
	header := rest[:end]
	body := strings.TrimLeft(rest[end+len("\n---"):], "\n")

	f := &Fragment{Body: strings.TrimSpace(body), Weight: 1}
	for _, line := range strings.Split(header, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		switch key {
		case "name":
			f.Name = val
		case "slug":
			f.Slug = val
		case "area":
			f.Area = val
		case "keywords":
			f.Keywords = splitList(val)
		case "tech":
			f.Tech = splitList(val)
		case "weight":
			if w, err := parseInt(val); err == nil && w > 0 {
				f.Weight = w
			}
		}
	}
	if f.Body == "" {
		return nil, false
	}
	return f, true
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// AllFragments returns all loaded fragments (copy to prevent mutation).
func (w *Warehouse) AllFragments() []*Fragment {
	out := make([]*Fragment, len(w.fragments))
	copy(out, w.fragments)
	return out
}

// Get returns a fragment by slug, or nil if not found.
func (w *Warehouse) Get(slug string) *Fragment {
	return w.bySlug[slug]
}

// Catalog returns a one-line-per-fragment summary, suitable for telling the
// Boss / Manager which specialities are available on the "warehouse".
func (w *Warehouse) Catalog() string {
	if len(w.fragments) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("=== SKILL WAREHOUSE CATALOG ===\n")
	for _, f := range w.fragments {
		b.WriteString("- ")
		b.WriteString(f.Name)
		b.WriteString(" (slug: ")
		b.WriteString(f.Slug)
		b.WriteString(", area: ")
		b.WriteString(f.Area)
		b.WriteString(", weight: ")
		writeInt(&b, f.Weight)
		b.WriteString(")\n")
	}
	b.WriteString("=== END CATALOG ===")
	return b.String()
}

func writeInt(b *strings.Builder, n int) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	b.Write(buf[i:])
}

// Search returns the top-N fragments that best match the given role, tech
// stack, and task description. At most max fragments are returned (default 3
// if max <= 0). Returns nil when nothing matches.
func (w *Warehouse) Search(role, techStack, task string, max int) []*Fragment {
	if max <= 0 {
		max = 3
	}
	type scored struct {
		f     *Fragment
		score int
	}
	var ranked []scored
	for _, f := range w.fragments {
		if sc := scoreFragment(f, role, techStack, task); sc > 0 {
			ranked = append(ranked, scored{f, sc})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if len(ranked) > max {
		ranked = ranked[:max]
	}
	out := make([]*Fragment, len(ranked))
	for i, r := range ranked {
		out[i] = r.f
	}
	return out
}

// scoreFragment rates how well a fragment matches the given role/tech/task.
// Scoring follows the same logic as Skill.score but includes weight as a
// multiplier.
func scoreFragment(f *Fragment, role, techStack, task string) int {
	role = strings.ToLower(role)
	tech := strings.ToLower(techStack)
	text := strings.ToLower(task)

	score := 0
	for _, kw := range f.Keywords {
		if strings.Contains(role, kw) {
			score += 5
		}
		if strings.Contains(text, kw) {
			score += 2
		}
	}
	for _, t := range f.Tech {
		if strings.Contains(tech, t) {
			score += 3
		}
	}
	if f.Slug != "" {
		slug := strings.ToLower(f.Slug)
		if strings.Contains(role, slug) {
			score += 4
		}
		// Also boost when the role mentions the area
		if strings.Contains(role, strings.ToLower(f.Area)) {
			score += 3
		}
	}
	// Apply weight multiplier
	score *= f.Weight
	return score
}

// Select returns a formatted Markdown block with the bodies of the requested
// fragments (by slug), or "" when none of the slugs are found. This is the
// method that Workers call — they only get the specific fragments the Manager
// picked for them.
func (w *Warehouse) Select(slugs ...string) string {
	if len(slugs) == 0 {
		return ""
	}
	var b strings.Builder
	count := 0
	for _, slug := range slugs {
		f := w.bySlug[slug]
		if f == nil {
			continue
		}
		if count == 0 {
			b.WriteString("\n\n=== APPLY THESE SKILL FRAGMENTS ===\n")
		}
		b.WriteString(f.Body)
		b.WriteString("\n")
		count++
	}
	if count == 0 {
		return ""
	}
	b.WriteString("=== END SKILL FRAGMENTS ===")
	return b.String()
}
