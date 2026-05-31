package document

import "strings"

// ParseSlideMarkdown — разбирает Markdown слайд-формат, который выдаёт воркер-презентатор:
//
//	# Deck title
//	## Slide title
//	- bullet
//	- bullet
//	> speaker notes
//
// Первый уровень "# " задаёт заголовок презентации (и титульный слайд),
// каждый "## " начинает новый слайд, "- "/"* " — пункты, "> " — заметки докладчика.
func ParseSlideMarkdown(md string) Deck {
	var deck Deck
	var cur *Slide

	flush := func() {
		if cur != nil {
			deck.Slides = append(deck.Slides, *cur)
			cur = nil
		}
	}

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "## "):
			flush()
			title := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			cur = &Slide{Title: title}
		case strings.HasPrefix(line, "# "):
			deck.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			if cur == nil {
				cur = &Slide{Title: deck.Title}
			}
			cur.Bullets = append(cur.Bullets, strings.TrimSpace(line[2:]))
		case strings.HasPrefix(line, "> "):
			if cur != nil {
				note := strings.TrimSpace(strings.TrimPrefix(line, "> "))
				if cur.Notes == "" {
					cur.Notes = note
				} else {
					cur.Notes += " " + note
				}
			}
		}
	}
	flush()

	// Если автор дал только заголовок и пункты без "## ", соберём один слайд.
	if len(deck.Slides) == 0 && deck.Title != "" {
		deck.Slides = []Slide{{Title: deck.Title}}
	}
	return deck
}
