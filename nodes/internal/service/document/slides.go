package document

import (
	"fmt"
	"strings"
)

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
			alt, imgURL := parseMarkdownImage(line)
			if imgURL != "" {
				// Реальная ссылка на картинку — её сможет скачать и встроить воркер.
				if cur.Image.URL == "" {
					cur.Image.URL = imgURL
					cur.Image.Alt = alt
				}
				if cur.Visual == "" && alt != "" {
					cur.Visual = alt
				}
			} else if alt != "" {
				cur.Visual = alt
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
	case "visual", "visual direction":
		if value != "" {
			slide.Visual = value
		}
	case "image", "image idea", "illustration":
		// Если значение — настоящий URL, это картинка для встраивания; иначе —
		// текстовое описание визуала.
		switch {
		case value == "":
		case isHTTPURL(value):
			if slide.Image.URL == "" {
				slide.Image.URL = value
			}
		default:
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

// parseMarkdownImage разбирает строку вида ![alt](url) и возвращает alt и url
// по отдельности (любой из них может быть пустым).
func parseMarkdownImage(line string) (alt, url string) {
	altStart := strings.Index(line, "![")
	altEnd := strings.Index(line, "](")
	urlEnd := strings.LastIndex(line, ")")
	if altStart < 0 || altEnd < altStart+2 || urlEnd <= altEnd+2 {
		return "", ""
	}
	alt = strings.TrimSpace(line[altStart+2 : altEnd])
	url = strings.TrimSpace(line[altEnd+2 : urlEnd])
	return alt, url
}

// isHTTPURL сообщает, что строка похожа на http(s)-ссылку.
func isHTTPURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// RenderDeckMarkdown сериализует Deck обратно в Markdown-формат, который понимает
// ParseSlideMarkdown. Воркер использует это, чтобы файл .md и встроенные в .pptx
// картинки оставались согласованными.
func RenderDeckMarkdown(deck Deck) string {
	var b strings.Builder
	if t := strings.TrimSpace(deck.Title); t != "" {
		b.WriteString("# " + t + "\n\n")
	}
	for i, s := range deck.Slides {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = fmt.Sprintf("Slide %d", i+1)
		}
		b.WriteString("## " + title + "\n")
		for _, bullet := range s.Bullets {
			if bullet = strings.TrimSpace(bullet); bullet != "" {
				b.WriteString("- " + bullet + "\n")
			}
		}
		if v := strings.TrimSpace(s.Visual); v != "" {
			b.WriteString("Visual: " + v + "\n")
		}
		if u := strings.TrimSpace(s.Image.URL); u != "" {
			alt := strings.TrimSpace(s.Image.Alt)
			if alt == "" {
				alt = "Illustration"
			}
			b.WriteString(fmt.Sprintf("![%s](%s)\n", alt, u))
		}
		for _, src := range s.Sources {
			if src = strings.TrimSpace(src); src != "" {
				b.WriteString("Source: " + src + "\n")
			}
		}
		if notes := strings.TrimSpace(s.Notes); notes != "" {
			b.WriteString("> " + notes + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
