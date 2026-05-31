package document

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func openZip(t *testing.T, data []byte) *zip.Reader {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("output is not a valid zip: %v", err)
	}
	return r
}

func readPart(t *testing.T, r *zip.Reader, name string) string {
	t.Helper()
	for _, f := range r.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open %s: %v", name, err)
			}
			defer rc.Close()
			var b bytes.Buffer
			if _, err := b.ReadFrom(rc); err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			return b.String()
		}
	}
	t.Fatalf("part %q missing from pptx", name)
	return ""
}

func TestBuildPPTXProducesValidPackage(t *testing.T) {
	deck := Deck{
		Title: "Quarterly Report",
		Slides: []Slide{
			{Title: "Overview", Bullets: []string{"Revenue up 12%", "New markets"}},
			{Title: "Risks", Bullets: []string{"Supply chain"}},
		},
	}
	data, err := BuildPPTX(deck)
	if err != nil {
		t.Fatalf("BuildPPTX error: %v", err)
	}
	r := openZip(t, data)

	required := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/theme/theme1.xml",
		"ppt/slideMasters/slideMaster1.xml",
		"ppt/slideLayouts/slideLayout1.xml",
		"ppt/slides/slide1.xml",
		"ppt/slides/slide2.xml",
		"ppt/slides/_rels/slide1.xml.rels",
	}
	for _, name := range required {
		readPart(t, r, name) // fails if missing
	}

	// Content types must declare every slide override.
	ct := readPart(t, r, "[Content_Types].xml")
	if !strings.Contains(ct, "/ppt/slides/slide2.xml") {
		t.Errorf("content types missing slide2 override")
	}

	// Slide text must be present and XML-escaped.
	s1 := readPart(t, r, "ppt/slides/slide1.xml")
	if !strings.Contains(s1, "Overview") || !strings.Contains(s1, "Revenue up 12%") {
		t.Errorf("slide1 missing expected content: %s", s1)
	}
}

func TestBuildPPTXEscapesSpecialChars(t *testing.T) {
	deck := Deck{Slides: []Slide{{Title: "A & B <C>", Bullets: []string{"x > y && a < b"}}}}
	data, err := BuildPPTX(deck)
	if err != nil {
		t.Fatalf("BuildPPTX error: %v", err)
	}
	r := openZip(t, data)
	s1 := readPart(t, r, "ppt/slides/slide1.xml")
	if strings.Contains(s1, "A & B <C>") {
		t.Errorf("special characters were not escaped: %s", s1)
	}
	if !strings.Contains(s1, "A &amp; B") {
		t.Errorf("expected escaped ampersand in %s", s1)
	}
}

func TestBuildPPTXEmptyDeckHasOneSlide(t *testing.T) {
	data, err := BuildPPTX(Deck{Title: "Empty"})
	if err != nil {
		t.Fatalf("BuildPPTX error: %v", err)
	}
	r := openZip(t, data)
	readPart(t, r, "ppt/slides/slide1.xml")
	for _, f := range r.File {
		if f.Name == "ppt/slides/slide2.xml" {
			t.Errorf("empty deck should produce exactly one slide")
		}
	}
}

func TestParseSlideMarkdown(t *testing.T) {
	md := `# My Deck

## Intro
- first point
- second point
> remember to smile

## Conclusion
* wrap up
`
	deck := ParseSlideMarkdown(md)
	if deck.Title != "My Deck" {
		t.Errorf("title = %q, want My Deck", deck.Title)
	}
	if len(deck.Slides) != 2 {
		t.Fatalf("slides = %d, want 2", len(deck.Slides))
	}
	if deck.Slides[0].Title != "Intro" || len(deck.Slides[0].Bullets) != 2 {
		t.Errorf("slide0 = %+v", deck.Slides[0])
	}
	if deck.Slides[0].Notes != "remember to smile" {
		t.Errorf("notes = %q", deck.Slides[0].Notes)
	}
	if deck.Slides[1].Title != "Conclusion" || deck.Slides[1].Bullets[0] != "wrap up" {
		t.Errorf("slide1 = %+v", deck.Slides[1])
	}
}

func TestParseSlideMarkdownRoundTripsToValidPPTX(t *testing.T) {
	deck := ParseSlideMarkdown("# T\n## S1\n- a\n## S2\n- b\n")
	data, err := BuildPPTX(deck)
	if err != nil {
		t.Fatalf("BuildPPTX error: %v", err)
	}
	openZip(t, data) // must be a valid zip
}
