// Package document turns Markdown deliverables into office formats.
// pptx.go is a small, dependency-free PowerPoint (.pptx) writer: it produces a valid
// OOXML package (a ZIP of XML parts) from a simple slide model, so Octra can deliver
// presentations without any external service or CGo dependency.
package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// Slide — одна страница презентации.
type Slide struct {
	Title   string
	Bullets []string
	Notes   string
	Visual  string
	Sources []string
}

// Deck — презентация целиком.
type Deck struct {
	Title  string
	Slides []Slide
}

// BuildPPTX — сериализует Deck в валидный .pptx (ZIP с OOXML-частями).
func BuildPPTX(deck Deck) ([]byte, error) {
	if len(deck.Slides) == 0 {
		// Гарантируем хотя бы один слайд, иначе PowerPoint считает файл повреждённым.
		title := deck.Title
		if title == "" {
			title = "Presentation"
		}
		deck.Slides = []Slide{{Title: title}}
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	parts := map[string]string{
		"[Content_Types].xml":                          contentTypes(len(deck.Slides)),
		"_rels/.rels":                                  rootRels(),
		"ppt/presentation.xml":                         presentationXML(len(deck.Slides)),
		"ppt/_rels/presentation.xml.rels":              presentationRels(len(deck.Slides)),
		"ppt/presProps.xml":                            presProps(),
		"ppt/theme/theme1.xml":                         themeXML(),
		"ppt/slideMasters/slideMaster1.xml":            slideMasterXML(),
		"ppt/slideMasters/_rels/slideMaster1.xml.rels": slideMasterRels(),
		"ppt/slideLayouts/slideLayout1.xml":            slideLayoutXML(),
		"ppt/slideLayouts/_rels/slideLayout1.xml.rels": slideLayoutRels(),
	}
	for i, s := range deck.Slides {
		n := i + 1
		parts[fmt.Sprintf("ppt/slides/slide%d.xml", n)] = slideXML(s, deck.Title, n, len(deck.Slides))
		parts[fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", n)] = slideRels()
	}

	names := make([]string, 0, len(parts))
	for name := range parts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(parts[name])); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func esc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

func contentTypes(slides int) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>`)
	b.WriteString(`<Override PartName="/ppt/presProps.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presProps+xml"/>`)
	b.WriteString(`<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>`)
	b.WriteString(`<Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>`)
	b.WriteString(`<Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>`)
	for i := 1; i <= slides; i++ {
		b.WriteString(fmt.Sprintf(`<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i))
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func rootRels() string {
	return xmlHeader +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>` +
		`</Relationships>`
}

func presentationXML(slides int) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">`)
	b.WriteString(`<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>`)
	b.WriteString(`<p:sldIdLst>`)
	for i := 0; i < slides; i++ {
		// Relationship ids for slides start at rId2 (rId1 is the master).
		b.WriteString(fmt.Sprintf(`<p:sldId id="%d" r:id="rId%d"/>`, 256+i, i+2))
	}
	b.WriteString(`</p:sldIdLst>`)
	b.WriteString(`<p:sldSz cx="12192000" cy="6858000" type="screen16x9"/>`)
	b.WriteString(`<p:notesSz cx="6858000" cy="9144000"/>`)
	b.WriteString(`</p:presentation>`)
	return b.String()
}

func presentationRels(slides int) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	b.WriteString(`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>`)
	for i := 0; i < slides; i++ {
		b.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, i+2, i+1))
	}
	// Theme and presProps relationships use ids above the slide range.
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>`, slides+2))
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/presProps" Target="presProps.xml"/>`, slides+3))
	b.WriteString(`</Relationships>`)
	return b.String()
}

func presProps() string {
	return xmlHeader +
		`<p:presentationPr xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`
}

func slideMasterRels() string {
	return xmlHeader +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>` +
		`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>` +
		`</Relationships>`
}

func slideLayoutRels() string {
	return xmlHeader +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>` +
		`</Relationships>`
}

func slideRels() string {
	return xmlHeader +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>` +
		`</Relationships>`
}

func slideMasterXML() string {
	return xmlHeader +
		`<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">` +
		`<p:cSld><p:bg><p:bgRef idx="1001"><a:schemeClr val="bg1"/></p:bgRef></p:bg>` +
		`<p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>` +
		`<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>` +
		`</p:spTree></p:cSld>` +
		`<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>` +
		`<p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst>` +
		`<p:txStyles>` +
		`<p:titleStyle><a:lvl1pPr><a:defRPr sz="4400"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill></a:defRPr></a:lvl1pPr></p:titleStyle>` +
		`<p:bodyStyle><a:lvl1pPr marL="342900" indent="-342900"><a:buChar char="&#8226;"/><a:defRPr sz="2800"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill></a:defRPr></a:lvl1pPr></p:bodyStyle>` +
		`<p:otherStyle/>` +
		`</p:txStyles>` +
		`</p:sldMaster>`
}

func slideLayoutXML() string {
	return xmlHeader +
		`<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="obj" preserve="1">` +
		`<p:cSld name="Title and Content"><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>` +
		`<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>` +
		`</p:spTree></p:cSld>` +
		`<p:clrMapOvr><a:overrideClrMapping bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/></p:clrMapOvr>` +
		`</p:sldLayout>`
}

const (
	slideW = 12192000
	slideH = 6858000

	colorBackground = "F6F4EF"
	colorInk        = "0B1F28"
	colorMuted      = "5D6970"
	colorTeal       = "0A4B5A"
	colorBlue       = "1E6F91"
	colorCoral      = "D45D3A"
	colorGold       = "E6A23C"
	colorGreen      = "3D6B5B"
	colorCard       = "FFFFFF"
	colorPale       = "E9F2F0"
	colorLine       = "D9DFDA"
)

var accentColors = []string{colorTeal, colorCoral, colorGold, colorBlue, colorGreen}

type textOptions struct {
	Size   int
	Color  string
	Bold   bool
	Italic bool
	Align  string
	Fill   string
	Line   string
	Shape  string
	LIns   int
	TIns   int
	RIns   int
	BIns   int
}

func slideXML(s Slide, deckTitle string, n, total int) string {
	if strings.TrimSpace(s.Title) == "" {
		s.Title = fallbackSlideTitle(deckTitle, n)
	}

	id := 2
	accent := accentColors[(n-1)%len(accentColors)]
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">`)
	b.WriteString(`<p:cSld><p:spTree>`)
	b.WriteString(`<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>`)
	b.WriteString(`<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>`)
	b.WriteString(shape(&id, "Canvas", "rect", 0, 0, slideW, slideH, colorBackground, ""))
	b.WriteString(shape(&id, "Top Accent", "rect", 0, 0, slideW, 152400, colorTeal, ""))
	b.WriteString(shape(&id, "Side Accent", "rect", slideW-228600, 0, 228600, slideH, accent, ""))

	switch {
	case n == 1:
		b.WriteString(titleSlide(&id, s, deckTitle, total, accent))
	case isClosingSlide(s):
		b.WriteString(closingSlide(&id, s, n, total, accent))
	default:
		b.WriteString(contentSlide(&id, s, n, total, accent))
	}

	b.WriteString(`</p:spTree></p:cSld>`)
	b.WriteString(`<p:clrMapOvr><a:overrideClrMapping bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/></p:clrMapOvr>`)
	b.WriteString(`</p:sld>`)
	return b.String()
}

func titleSlide(id *int, s Slide, deckTitle string, total int, accent string) string {
	title := strings.TrimSpace(deckTitle)
	if title == "" {
		title = s.Title
	}
	subtitle := "A structured deck with sourced visual direction"
	if len(s.Bullets) > 0 {
		subtitle = strings.Join(limitStrings(s.Bullets, 2, 88), " / ")
	}
	if slideTitle := strings.TrimSpace(s.Title); slideTitle != "" && !strings.EqualFold(slideTitle, title) {
		subtitle = slideTitle + " / " + subtitle
	}

	var b strings.Builder
	b.WriteString(shape(id, "Title Panel", "rect", 0, 152400, 8200000, slideH-152400, "FDFCF8", ""))
	b.WriteString(shape(id, "Title Accent Line", "rect", 685800, 3950000, 3657600, 76200, accent, ""))
	b.WriteString(textBox(id, "Deck Title", 650000, 1420000, 7200000, 1780000, []string{title}, textOptions{Size: 4300, Color: colorInk, Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Deck Subtitle", 700000, 3280000, 6200000, 700000, []string{subtitle}, textOptions{Size: 1800, Color: colorMuted, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Deck Meta", 700000, 5180000, 5600000, 500000, []string{fmt.Sprintf("%d-slide presentation", total)}, textOptions{Size: 1200, Color: colorMuted, LIns: 0, TIns: 0}))

	visual := s.Visual
	if visual == "" {
		visual = "Use a hero image or clean diagram that introduces the deck topic."
	}
	b.WriteString(shape(id, "Visual Panel", "roundRect", 8400000, 870000, 3000000, 4950000, colorTeal, ""))
	b.WriteString(textBox(id, "Visual Label", 8720000, 1280000, 2300000, 350000, []string{"Visual direction"}, textOptions{Size: 1050, Color: "BFE3DC", Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Visual Text", 8720000, 1740000, 2300000, 2500000, []string{truncateText(visual, 170)}, textOptions{Size: 1900, Color: "FFFFFF", Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(shape(id, "Visual Rule", "rect", 8720000, 4520000, 1500000, 50800, colorGold, ""))
	b.WriteString(textBox(id, "Visual Hint", 8720000, 4760000, 2300000, 680000, []string{"Search-backed image, chart, or screenshot ideas are rendered on content slides."}, textOptions{Size: 1000, Color: "E9F2F0", LIns: 0, TIns: 0}))
	return b.String()
}

func contentSlide(id *int, s Slide, n, total int, accent string) string {
	var b strings.Builder
	b.WriteString(textBox(id, "Slide Number", 650000, 430000, 700000, 350000, []string{fmt.Sprintf("%02d", n)}, textOptions{Size: 1300, Color: accent, Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Slide Title", 1350000, 350000, 7500000, 760000, []string{truncateText(s.Title, 72)}, textOptions{Size: 2750, Color: colorInk, Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(shape(id, "Title Rule", "rect", 1350000, 1120000, 2200000, 50800, accent, ""))

	bullets := limitStrings(s.Bullets, 5, 115)
	if len(bullets) == 0 {
		bullets = []string{"Use this slide to explain the key message clearly."}
	}
	cardY := 1520000
	cardH := 760000
	gap := 130000
	if len(bullets) >= 5 {
		cardH = 650000
		gap = 90000
	}
	for i, bullet := range bullets {
		y := cardY + i*(cardH+gap)
		b.WriteString(shape(id, fmt.Sprintf("Point Card %d", i+1), "roundRect", 650000, y, 7150000, cardH, colorCard, colorLine))
		b.WriteString(shape(id, fmt.Sprintf("Point Accent %d", i+1), "rect", 650000, y, 76200, cardH, accent, ""))
		b.WriteString(textBox(id, fmt.Sprintf("Point Text %d", i+1), 900000, y+90000, 6500000, cardH-120000, []string{bullet}, textOptions{Size: 1550, Color: colorInk, LIns: 0, TIns: 0}))
	}

	b.WriteString(visualPanel(id, s, accent))
	b.WriteString(footer(id, n, total, accent))
	return b.String()
}

func closingSlide(id *int, s Slide, n, total int, accent string) string {
	title := s.Title
	if title == "" {
		title = "Closing"
	}
	bullets := limitStrings(s.Bullets, 3, 92)
	if len(bullets) == 0 {
		bullets = []string{"Align on next steps", "Use the deck as a working artifact"}
	}

	var b strings.Builder
	b.WriteString(shape(id, "Closing Panel", "roundRect", 1050000, 1050000, 9850000, 4750000, colorTeal, ""))
	b.WriteString(shape(id, "Closing Accent", "rect", 1050000, 1050000, 9850000, 152400, accent, ""))
	b.WriteString(textBox(id, "Closing Title", 1550000, 1750000, 8200000, 900000, []string{truncateText(title, 72)}, textOptions{Size: 3400, Color: "FFFFFF", Bold: true, Align: "ctr", LIns: 0, TIns: 0}))
	y := 3050000
	for i, bullet := range bullets {
		b.WriteString(textBox(id, fmt.Sprintf("Closing Point %d", i+1), 2150000, y+i*620000, 7000000, 420000, []string{bullet}, textOptions{Size: 1650, Color: "E9F2F0", Align: "ctr", LIns: 0, TIns: 0}))
	}
	b.WriteString(footer(id, n, total, accent))
	return b.String()
}

func visualPanel(id *int, s Slide, accent string) string {
	visual := strings.TrimSpace(s.Visual)
	if visual == "" {
		visual = `Use a clean supporting image, chart, or diagram for "` + s.Title + `".`
	}

	var b strings.Builder
	b.WriteString(shape(id, "Visual Card", "roundRect", 8350000, 1500000, 3000000, 3750000, colorPale, "B8CDC8"))
	b.WriteString(shape(id, "Visual Card Accent", "rect", 8350000, 1500000, 3000000, 101600, accent, ""))
	b.WriteString(textBox(id, "Visual Label", 8650000, 1840000, 2350000, 360000, []string{"Visual direction"}, textOptions{Size: 1050, Color: colorTeal, Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Visual Body", 8650000, 2300000, 2350000, 1250000, []string{truncateText(visual, 145)}, textOptions{Size: 1350, Color: colorInk, Bold: true, LIns: 0, TIns: 0}))
	if len(s.Sources) > 0 {
		b.WriteString(shape(id, "Source Divider", "rect", 8650000, 3780000, 2050000, 38100, "B8CDC8", ""))
		b.WriteString(textBox(id, "Source Label", 8650000, 3960000, 2350000, 300000, []string{"Source"}, textOptions{Size: 900, Color: colorMuted, Bold: true, LIns: 0, TIns: 0}))
		b.WriteString(textBox(id, "Source Text", 8650000, 4260000, 2350000, 650000, []string{truncateText(s.Sources[0], 118)}, textOptions{Size: 900, Color: colorMuted, LIns: 0, TIns: 0}))
	}
	return b.String()
}

func footer(id *int, n, total int, accent string) string {
	return shape(id, "Footer Rule", "rect", 650000, 6350000, 4200000, 25400, accent, "") +
		textBox(id, "Footer Brand", 650000, 6420000, 3000000, 260000, []string{"Octra presentation"}, textOptions{Size: 850, Color: colorMuted, LIns: 0, TIns: 0}) +
		textBox(id, "Footer Page", 9750000, 6420000, 1400000, 260000, []string{fmt.Sprintf("Slide %d of %d", n, total)}, textOptions{Size: 850, Color: colorMuted, Align: "r", LIns: 0, TIns: 0})
}

func shape(id *int, name, prst string, x, y, cx, cy int, fill, line string) string {
	cur := *id
	*id = *id + 1
	if prst == "" {
		prst = "rect"
	}
	return fmt.Sprintf(
		`<p:sp><p:nvSpPr><p:cNvPr id="%d" name="%s"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="%s"><a:avLst/></a:prstGeom>%s%s</p:spPr><p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		cur, esc(name), x, y, cx, cy, prst, fillXML(fill), lineXML(line),
	)
}

func textBox(id *int, name string, x, y, cx, cy int, paragraphs []string, opt textOptions) string {
	cur := *id
	*id = *id + 1
	if opt.Shape == "" {
		opt.Shape = "rect"
	}
	if opt.Color == "" {
		opt.Color = colorInk
	}
	if opt.Size == 0 {
		opt.Size = 1400
	}
	if opt.LIns == 0 {
		opt.LIns = 91440
	}
	if opt.RIns == 0 {
		opt.RIns = 91440
	}
	if opt.TIns == 0 {
		opt.TIns = 45720
	}
	if opt.BIns == 0 {
		opt.BIns = 45720
	}
	var body strings.Builder
	if len(paragraphs) == 0 {
		paragraphs = []string{""}
	}
	for _, p := range paragraphs {
		body.WriteString(paragraphXML(p, opt))
	}
	return fmt.Sprintf(
		`<p:sp><p:nvSpPr><p:cNvPr id="%d" name="%s"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="%s"><a:avLst/></a:prstGeom>%s%s</p:spPr><p:txBody><a:bodyPr wrap="square" lIns="%d" tIns="%d" rIns="%d" bIns="%d"/><a:lstStyle/>%s</p:txBody></p:sp>`,
		cur, esc(name), x, y, cx, cy, opt.Shape, fillXML(opt.Fill), lineXML(opt.Line), opt.LIns, opt.TIns, opt.RIns, opt.BIns, body.String(),
	)
}

func paragraphXML(text string, opt textOptions) string {
	align := ""
	if opt.Align != "" {
		align = fmt.Sprintf(` algn="%s"`, opt.Align)
	}
	bold := ""
	if opt.Bold {
		bold = ` b="1"`
	}
	italic := ""
	if opt.Italic {
		italic = ` i="1"`
	}
	return fmt.Sprintf(
		`<a:p><a:pPr%s/><a:r><a:rPr lang="en-US" sz="%d"%s%s dirty="0"><a:solidFill><a:srgbClr val="%s"/></a:solidFill><a:latin typeface="Aptos"/></a:rPr><a:t>%s</a:t></a:r></a:p>`,
		align, opt.Size, bold, italic, opt.Color, esc(text),
	)
}

func fillXML(color string) string {
	if color == "" {
		return `<a:noFill/>`
	}
	return `<a:solidFill><a:srgbClr val="` + esc(color) + `"/></a:solidFill>`
}

func lineXML(color string) string {
	if color == "" {
		return `<a:ln><a:noFill/></a:ln>`
	}
	return `<a:ln w="12700"><a:solidFill><a:srgbClr val="` + esc(color) + `"/></a:solidFill></a:ln>`
}

func fallbackSlideTitle(deckTitle string, n int) string {
	if n == 1 && strings.TrimSpace(deckTitle) != "" {
		return deckTitle
	}
	return fmt.Sprintf("Slide %d", n)
}

func isClosingSlide(s Slide) bool {
	t := strings.ToLower(s.Title)
	for _, token := range []string{"closing", "conclusion", "summary", "next steps", "thank", "wrap up", "final"} {
		if strings.Contains(t, token) {
			return true
		}
	}
	return false
}

func limitStrings(in []string, max, maxLen int) []string {
	out := make([]string, 0, max)
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, truncateText(s, maxLen))
		if len(out) >= max {
			break
		}
	}
	return out
}

func truncateText(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if max <= 0 || len([]rune(s)) <= max {
		return s
	}
	runes := []rune(s)
	cut := max - 1
	for cut > max-24 && cut > 0 && runes[cut] != ' ' {
		cut--
	}
	if cut <= 0 {
		cut = max - 1
	}
	return strings.TrimSpace(string(runes[:cut])) + "..."
}

func themeXML() string {
	return xmlHeader +
		`<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">` +
		`<a:themeElements>` +
		`<a:clrScheme name="Office">` +
		`<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>` +
		`<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>` +
		`<a:dk2><a:srgbClr val="44546A"/></a:dk2>` +
		`<a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>` +
		`<a:accent1><a:srgbClr val="4472C4"/></a:accent1>` +
		`<a:accent2><a:srgbClr val="ED7D31"/></a:accent2>` +
		`<a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>` +
		`<a:accent4><a:srgbClr val="FFC000"/></a:accent4>` +
		`<a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>` +
		`<a:accent6><a:srgbClr val="70AD47"/></a:accent6>` +
		`<a:hlink><a:srgbClr val="0563C1"/></a:hlink>` +
		`<a:folHlink><a:srgbClr val="954F72"/></a:folHlink>` +
		`</a:clrScheme>` +
		`<a:fontScheme name="Office">` +
		`<a:majorFont><a:latin typeface="Calibri Light"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>` +
		`<a:minorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>` +
		`</a:fontScheme>` +
		`<a:fmtScheme name="Office">` +
		`<a:fillStyleLst>` +
		`<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>` +
		`<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>` +
		`<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>` +
		`</a:fillStyleLst>` +
		`<a:lnStyleLst>` +
		`<a:ln w="6350" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>` +
		`<a:ln w="12700" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>` +
		`<a:ln w="19050" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>` +
		`</a:lnStyleLst>` +
		`<a:effectStyleLst>` +
		`<a:effectStyle><a:effectLst/></a:effectStyle>` +
		`<a:effectStyle><a:effectLst/></a:effectStyle>` +
		`<a:effectStyle><a:effectLst/></a:effectStyle>` +
		`</a:effectStyleLst>` +
		`<a:bgFillStyleLst>` +
		`<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>` +
		`<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>` +
		`<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>` +
		`</a:bgFillStyleLst>` +
		`</a:fmtScheme>` +
		`</a:themeElements>` +
		`</a:theme>`
}
