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
			if !applyStructuredSlideLine(cur, strings.TrimSpace(line[2:])) {
				cur.Bullets = append(cur.Bullets, strings.TrimSpace(line[2:]))
			}
		case isStructuredSlideLine(line):
			if cur == nil {
				cur = &Slide{Title: deck.Title}
			}
			applyStructuredSlideLine(cur, line)
		case strings.HasPrefix(line, "!["):
			if cur == nil {
				cur = &Slide{Title: deck.Title}
			}
			if visual := parseMarkdownImage(line); visual != "" {
				cur.Visual = visual
			}
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

func isStructuredSlideLine(line string) bool {
	_, _, ok := splitStructuredLine(line)
	return ok
}

func applyStructuredSlideLine(slide *Slide, line string) bool {
	key, value, ok := splitStructuredLine(line)
	if !ok {
		return false
	}
	switch key {
	case "visual", "visual direction", "image", "image idea", "illustration":
		if value != "" {
			slide.Visual = value
		}
	case "source", "sources":
		if value != "" {
			slide.Sources = append(slide.Sources, value)
		}
	default:
		return false
	}
	return true
}

func splitStructuredLine(line string) (key, value string, ok bool) {
	before, after, found := strings.Cut(line, ":")
	if !found {
		return "", "", false
	}
	key = strings.ToLower(strings.TrimSpace(before))
	switch key {
	case "visual", "visual direction", "image", "image idea", "illustration", "source", "sources":
		return key, strings.TrimSpace(after), true
	default:
		return "", "", false
	}
}

func parseMarkdownImage(line string) string {
	altStart := strings.Index(line, "![")
	altEnd := strings.Index(line, "](")
	urlEnd := strings.LastIndex(line, ")")
	if altStart < 0 || altEnd < altStart+2 || urlEnd <= altEnd+2 {
		return ""
	}
	alt := strings.TrimSpace(line[altStart+2 : altEnd])
	url := strings.TrimSpace(line[altEnd+2 : urlEnd])
	switch {
	case alt != "" && url != "":
		return alt + " — " + url
	case alt != "":
		return alt
	default:
		return url
	}
}
