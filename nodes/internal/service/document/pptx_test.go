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
			{Title: "Overview", Bullets: []string{"Revenue up 12%", "New markets"}, Visual: "Regional growth map"},
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

func TestBuildPPTXUsesDesignedWidescreenSlides(t *testing.T) {
	deck := Deck{
		Title: "Launch Plan",
		Slides: []Slide{
			{Title: "Launch Plan", Bullets: []string{"Position the release", "Coordinate teams"}},
			{Title: "Market Signal", Bullets: []string{"Analysts expect higher demand"}, Visual: "Show a simple demand curve", Sources: []string{"Example Market Report — https://example.com/report"}},
		},
	}
	data, err := BuildPPTX(deck)
	if err != nil {
		t.Fatalf("BuildPPTX error: %v", err)
	}
	r := openZip(t, data)

	pres := readPart(t, r, "ppt/presentation.xml")
	if !strings.Contains(pres, `type="screen16x9"`) || !strings.Contains(pres, `cx="12192000"`) {
		t.Fatalf("presentation must use a 16:9 widescreen canvas, got: %s", pres)
	}

	slide2 := readPart(t, r, "ppt/slides/slide2.xml")
	for _, want := range []string{
		"Visual direction",
		"Show a simple demand curve",
		"Example Market Report",
		"Slide 2 of 2",
		`prst="roundRect"`,
	} {
		if !strings.Contains(slide2, want) {
			t.Fatalf("designed slide missing %q:\n%s", want, slide2)
		}
	}
	// Цвет «визуальной подсказки» берётся из палитры колоды, которая выбирается
	// детерминированно по заголовку — проверяем именно её основной цвет.
	primary := pickPalette(deck.Title).primary()
	if !strings.Contains(slide2, `val="`+primary+`"`) {
		t.Fatalf("designed slide must use deck palette primary %q:\n%s", primary, slide2)
	}
}

func TestBuildPPTXVariesPaletteByTitle(t *testing.T) {
	// Разные заголовки должны (в большинстве случаев) давать разные палитры,
	// иначе все презентации выглядят одинаково.
	seen := map[string]int{}
	for _, title := range []string{"Launch Plan", "Market Research", "Q3 Strategy", "Onboarding Guide", "Climate Report"} {
		seen[pickPalette(title).primary()]++
	}
	if len(seen) < 2 {
		t.Fatalf("expected varied palettes across titles, got %#v", seen)
	}
}

func TestBuildPPTXEmbedsImageWhenBytesPresent(t *testing.T) {
	// Минимальный валидный JPEG-заголовок достаточно для embeddable().
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	deck := Deck{
		Title: "Visual Deck",
		Slides: []Slide{
			{Title: "Visual Deck"},
			{Title: "Proof", Bullets: []string{"It works"}, Image: Image{
				URL: "https://img.example/a.jpg", Alt: "A photo", Source: "flickr.com",
				Data: jpeg, ContentType: "image/jpeg",
			}},
		},
	}
	data, err := BuildPPTX(deck)
	if err != nil {
		t.Fatalf("BuildPPTX error: %v", err)
	}
	r := openZip(t, data)

	// Медиа-часть с байтами картинки должна присутствовать.
	media := readPart(t, r, "ppt/media/image2.jpeg")
	if media != string(jpeg) {
		t.Fatalf("embedded image bytes mismatch")
	}
	// Слайд ссылается на картинку, а связь указывает на медиа-файл.
	slide2 := readPart(t, r, "ppt/slides/slide2.xml")
	if !strings.Contains(slide2, `<p:pic>`) || !strings.Contains(slide2, `r:embed="rId2"`) {
		t.Fatalf("slide2 must embed the picture:\n%s", slide2)
	}
	rels := readPart(t, r, "ppt/slides/_rels/slide2.xml.rels")
	if !strings.Contains(rels, "../media/image2.jpeg") {
		t.Fatalf("slide2 rels must target the media file:\n%s", rels)
	}
	// Тип содержимого должен объявлять расширение jpeg.
	ct := readPart(t, r, "[Content_Types].xml")
	if !strings.Contains(ct, `Extension="jpeg"`) {
		t.Fatalf("content types must declare jpeg default:\n%s", ct)
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
Visual: Product screenshot over a workflow diagram
Source: Product docs — https://example.com/docs
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
	if deck.Slides[0].Visual != "Product screenshot over a workflow diagram" {
		t.Errorf("visual = %q", deck.Slides[0].Visual)
	}
	if len(deck.Slides[0].Sources) != 1 || !strings.Contains(deck.Slides[0].Sources[0], "https://example.com/docs") {
		t.Errorf("sources = %#v", deck.Slides[0].Sources)
	}
	if deck.Slides[1].Title != "Conclusion" || deck.Slides[1].Bullets[0] != "wrap up" {
		t.Errorf("slide1 = %+v", deck.Slides[1])
	}
}

func TestParseSlideMarkdownCapturesBulletStructuredLines(t *testing.T) {
	deck := ParseSlideMarkdown(`# Visual Deck

## Proof
- Insight one
- Visual: customer photo next to adoption chart
- Source: Case study — https://example.com/case
- Insight two
`)
	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1", len(deck.Slides))
	}
	s := deck.Slides[0]
	if s.Visual != "customer photo next to adoption chart" {
		t.Fatalf("visual = %q", s.Visual)
	}
	if len(s.Sources) != 1 || !strings.Contains(s.Sources[0], "https://example.com/case") {
		t.Fatalf("sources = %#v", s.Sources)
	}
	if len(s.Bullets) != 2 {
		t.Fatalf("structured lines should not become bullets: %#v", s.Bullets)
	}
}

func TestParseSlideMarkdownCapturesImageURL(t *testing.T) {
	deck := ParseSlideMarkdown(`# Deck

## Proof
- point
![A field of flowers](https://img.example/flowers.jpg)
Source: Openverse — flickr.com
`)
	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1", len(deck.Slides))
	}
	s := deck.Slides[0]
	if s.Image.URL != "https://img.example/flowers.jpg" {
		t.Fatalf("image url = %q", s.Image.URL)
	}
	if s.Image.Alt != "A field of flowers" {
		t.Fatalf("image alt = %q", s.Image.Alt)
	}
}

func TestParseSlideMarkdownImageStructuredLine(t *testing.T) {
	deck := ParseSlideMarkdown("# D\n## S\n- a\nImage: https://img.example/x.png\n")
	if deck.Slides[0].Image.URL != "https://img.example/x.png" {
		t.Fatalf("image url = %q", deck.Slides[0].Image.URL)
	}
	// Нессылочное значение Image: остаётся текстовым визуалом.
	deck2 := ParseSlideMarkdown("# D\n## S\n- a\nImage: a clean product photo\n")
	if deck2.Slides[0].Image.URL != "" {
		t.Fatalf("non-url image should not set URL: %q", deck2.Slides[0].Image.URL)
	}
	if deck2.Slides[0].Visual != "a clean product photo" {
		t.Fatalf("non-url image should set visual: %q", deck2.Slides[0].Visual)
	}
}

func TestRenderDeckMarkdownRoundTrip(t *testing.T) {
	deck := Deck{
		Title: "My Deck",
		Slides: []Slide{
			{Title: "Intro", Bullets: []string{"one", "two"}, Visual: "diagram", Notes: "smile",
				Image:   Image{URL: "https://img.example/a.jpg", Alt: "Hero"},
				Sources: []string{"Docs — https://example.com/docs"}},
		},
	}
	md := RenderDeckMarkdown(deck)
	got := ParseSlideMarkdown(md)
	if got.Title != "My Deck" || len(got.Slides) != 1 {
		t.Fatalf("round-trip deck = %+v\nmd:\n%s", got, md)
	}
	s := got.Slides[0]
	if s.Title != "Intro" || len(s.Bullets) != 2 || s.Visual != "diagram" || s.Notes != "smile" {
		t.Fatalf("round-trip slide = %+v", s)
	}
	if s.Image.URL != "https://img.example/a.jpg" || s.Image.Alt != "Hero" {
		t.Fatalf("round-trip image = %+v", s.Image)
	}
	if len(s.Sources) != 1 {
		t.Fatalf("round-trip sources = %#v", s.Sources)
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
