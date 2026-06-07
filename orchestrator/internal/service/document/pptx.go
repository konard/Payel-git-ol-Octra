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

// Image — иллюстрация слайда. URL/Alt заполняются парсером из Markdown,
// а Data/ContentType — воркером после скачивания файла, и только при наличии
// валидных байтов картинка действительно встраивается в .pptx.
type Image struct {
	Alt         string
	URL         string
	Source      string // подпись/атрибуция (домен или название источника)
	Data        []byte // байты файла для встраивания (jpeg/png)
	ContentType string // "image/jpeg" | "image/png"
}

// embeddable сообщает, можно ли встроить картинку как бинарную часть .pptx.
func (im Image) embeddable() bool {
	return len(im.Data) > 0 && (im.ContentType == "image/jpeg" || im.ContentType == "image/png")
}

// Embeddable — экспортированная версия embeddable для других пакетов (воркера).
func (im Image) Embeddable() bool { return im.embeddable() }

func (im Image) ext() string {
	if im.ContentType == "image/png" {
		return "png"
	}
	return "jpeg"
}

// Slide — одна страница презентации.
type Slide struct {
	Title   string
	Bullets []string
	Notes   string
	Visual  string
	Image   Image
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

	pal := pickPalette(deck.Title)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	hasJPEG, hasPNG := false, false

	parts := map[string]string{
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
		imgRID := ""
		if s.Image.embeddable() {
			imgRID = "rId2"
			parts[fmt.Sprintf("ppt/media/image%d.%s", n, s.Image.ext())] = string(s.Image.Data)
			if s.Image.ContentType == "image/png" {
				hasPNG = true
			} else {
				hasJPEG = true
			}
		}
		parts[fmt.Sprintf("ppt/slides/slide%d.xml", n)] = slideXML(s, deck.Title, n, len(deck.Slides), pal, imgRID)
		parts[fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", n)] = slideRels(s, n)
	}
	parts["[Content_Types].xml"] = contentTypes(len(deck.Slides), hasJPEG, hasPNG)

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

func contentTypes(slides int, hasJPEG, hasPNG bool) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	if hasJPEG {
		b.WriteString(`<Default Extension="jpeg" ContentType="image/jpeg"/>`)
	}
	if hasPNG {
		b.WriteString(`<Default Extension="png" ContentType="image/png"/>`)
	}
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

// slideRels — связи слайда: всегда макет (rId1), и при наличии встроенной
// картинки — relationship на медиа-часть (rId2). Имя медиа-файла совпадает с
// тем, что пишет BuildPPTX: image<n>.<ext>.
func slideRels(s Slide, n int) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	b.WriteString(`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>`)
	if s.Image.embeddable() {
		target := fmt.Sprintf("../media/image%d.%s", n, s.Image.ext())
		b.WriteString(`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="` + esc(target) + `"/>`)
	}
	b.WriteString(`</Relationships>`)
	return b.String()
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

	// colorInk остаётся дефолтом текста для textBox, когда палитра не задана явно.
	colorInk = "0B1F28"

	// onAccentText/onAccentMuted — близкие к белому цвета текста поверх тёмных
	// акцентных панелей; читаются на любой палитре.
	onAccentText  = "FFFFFF"
	onAccentMuted = "E9F2F0"
)

// palette — связный набор цветов одной презентации. Каждая колода детерминированно
// получает свою палитру (по заголовку), поэтому соседние презентации выглядят
// по-разному, а внутри одной колоды цвета согласованы.
type palette struct {
	Background string
	Ink        string
	Muted      string
	Card       string
	Pale       string
	Line       string
	Accents    []string // первый — основной (primary), остальные чередуются по слайдам
}

func (p palette) primary() string { return p.Accents[0] }

func (p palette) accent(n int) string {
	return p.Accents[(n-1+len(p.Accents))%len(p.Accents)]
}

// palettes — несколько светлых, но разнотональных тем. Все фоны светлые, а текст
// тёмный, чтобы гарантировать контраст; различаются оттенок фона и набор акцентов.
var palettes = []palette{
	{ // Harbor — исходная тёплая бежево-бирюзовая тема.
		Background: "F6F4EF", Ink: "0B1F28", Muted: "5D6970", Card: "FFFFFF", Pale: "E9F2F0", Line: "D9DFDA",
		Accents: []string{"0A4B5A", "D45D3A", "E6A23C", "1E6F91", "3D6B5B"},
	},
	{ // Indigo — холодная сине-фиолетовая тема.
		Background: "F4F5FB", Ink: "141B2E", Muted: "5B617A", Card: "FFFFFF", Pale: "ECEEFB", Line: "DBDFEE",
		Accents: []string{"3B3B8F", "E0529C", "F2A93B", "2F7CC4", "4C9A8E"},
	},
	{ // Sunrise — тёплая кораллово-оранжевая тема.
		Background: "FBF6F0", Ink: "2A1A12", Muted: "7A6253", Card: "FFFFFF", Pale: "FBEEE2", Line: "EBDDCE",
		Accents: []string{"E0552B", "C2185B", "F2A93B", "7A4FB5", "2E8B7F"},
	},
	{ // Forest — спокойная зелёная тема.
		Background: "F1F6F1", Ink: "12241A", Muted: "55695C", Card: "FFFFFF", Pale: "E5F1E7", Line: "D6E2D7",
		Accents: []string{"2E6B4F", "1E6F91", "C9772E", "8E5BB5", "B23A48"},
	},
	{ // Slate — нейтральная сине-серая тема.
		Background: "F4F6F8", Ink: "14222E", Muted: "5A6A78", Card: "FFFFFF", Pale: "E8EEF3", Line: "D8E0E7",
		Accents: []string{"22566B", "D45D3A", "E6A23C", "4763A8", "3D7A6B"},
	},
}

// pickPalette детерминированно выбирает палитру по заголовку колоды: одинаковые
// заголовки дают одинаковую тему, разные — чаще всего разные.
func pickPalette(deckTitle string) palette {
	var h uint32 = 2166136261 // FNV-1a
	for _, r := range strings.ToLower(strings.TrimSpace(deckTitle)) {
		h ^= uint32(r)
		h *= 16777619
	}
	return palettes[int(h)%len(palettes)]
}

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

func slideXML(s Slide, deckTitle string, n, total int, pal palette, imgRID string) string {
	if strings.TrimSpace(s.Title) == "" {
		s.Title = fallbackSlideTitle(deckTitle, n)
	}

	id := 2
	accent := pal.accent(n)
	// Чередуем расположение колонок на чётных слайдах, чтобы структура не была
	// однотипной от слайда к слайду.
	mirror := n%2 == 0
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">`)
	b.WriteString(`<p:cSld><p:spTree>`)
	b.WriteString(`<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>`)
	b.WriteString(`<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>`)
	b.WriteString(shape(&id, "Canvas", "rect", 0, 0, slideW, slideH, pal.Background, ""))
	b.WriteString(shape(&id, "Top Accent", "rect", 0, 0, slideW, 152400, pal.primary(), ""))
	b.WriteString(shape(&id, "Side Accent", "rect", slideW-228600, 0, 228600, slideH, accent, ""))

	switch {
	case n == 1:
		b.WriteString(titleSlide(&id, s, deckTitle, total, accent, pal, imgRID))
	case isClosingSlide(s):
		b.WriteString(closingSlide(&id, s, n, total, accent, pal))
	default:
		b.WriteString(contentSlide(&id, s, n, total, accent, pal, mirror, imgRID))
	}

	b.WriteString(`</p:spTree></p:cSld>`)
	b.WriteString(`<p:clrMapOvr><a:overrideClrMapping bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/></p:clrMapOvr>`)
	b.WriteString(`</p:sld>`)
	return b.String()
}

func titleSlide(id *int, s Slide, deckTitle string, total int, accent string, pal palette, imgRID string) string {
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
	b.WriteString(shape(id, "Title Panel", "rect", 0, 152400, 8200000, slideH-152400, pal.Card, ""))
	b.WriteString(shape(id, "Title Accent Line", "rect", 685800, 3950000, 3657600, 76200, accent, ""))
	b.WriteString(textBox(id, "Deck Title", 650000, 1420000, 7200000, 1780000, []string{title}, textOptions{Size: 4300, Color: pal.Ink, Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Deck Subtitle", 700000, 3280000, 6200000, 700000, []string{subtitle}, textOptions{Size: 1800, Color: pal.Muted, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Deck Meta", 700000, 5180000, 5600000, 500000, []string{fmt.Sprintf("%d-slide presentation", total)}, textOptions{Size: 1200, Color: pal.Muted, LIns: 0, TIns: 0}))

	b.WriteString(shape(id, "Visual Panel", "roundRect", 8400000, 870000, 3000000, 4950000, pal.primary(), ""))
	if imgRID != "" {
		// Реальная картинка как hero-иллюстрация титульного слайда.
		b.WriteString(picture(id, "Hero Image", imgRID, 8560000, 1080000, 2680000, 2680000))
		caption := strings.TrimSpace(s.Image.Alt)
		if caption == "" {
			caption = "Search-backed illustration"
		}
		b.WriteString(textBox(id, "Hero Caption", 8560000, 3900000, 2680000, 850000, []string{truncateText(caption, 120)}, textOptions{Size: 1100, Color: onAccentText, Bold: true, LIns: 0, TIns: 0}))
		if src := strings.TrimSpace(s.Image.Source); src != "" {
			b.WriteString(textBox(id, "Hero Source", 8560000, 4900000, 2680000, 500000, []string{truncateText(src, 90)}, textOptions{Size: 900, Color: onAccentMuted, LIns: 0, TIns: 0}))
		}
	} else {
		visual := s.Visual
		if visual == "" {
			visual = "Use a hero image or clean diagram that introduces the deck topic."
		}
		b.WriteString(textBox(id, "Visual Label", 8720000, 1280000, 2300000, 350000, []string{"Visual direction"}, textOptions{Size: 1050, Color: onAccentMuted, Bold: true, LIns: 0, TIns: 0}))
		b.WriteString(textBox(id, "Visual Text", 8720000, 1740000, 2300000, 2500000, []string{truncateText(visual, 170)}, textOptions{Size: 1900, Color: onAccentText, Bold: true, LIns: 0, TIns: 0}))
		b.WriteString(shape(id, "Visual Rule", "rect", 8720000, 4520000, 1500000, 50800, pal.accent(3), ""))
		b.WriteString(textBox(id, "Visual Hint", 8720000, 4760000, 2300000, 680000, []string{"Search-backed image, chart, or screenshot ideas are rendered on content slides."}, textOptions{Size: 1000, Color: onAccentMuted, LIns: 0, TIns: 0}))
	}
	return b.String()
}

func contentSlide(id *int, s Slide, n, total int, accent string, pal palette, mirror bool, imgRID string) string {
	var b strings.Builder

	// Колонка с пунктами и колонка-визуал меняются местами на чётных слайдах.
	bulletX := 650000
	visualX := 8350000
	if mirror {
		bulletX = 3842000
		visualX = 650000
	}

	b.WriteString(textBox(id, "Slide Number", 650000, 430000, 700000, 350000, []string{fmt.Sprintf("%02d", n)}, textOptions{Size: 1300, Color: accent, Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Slide Title", 1350000, 350000, 7500000, 760000, []string{truncateText(s.Title, 72)}, textOptions{Size: 2750, Color: pal.Ink, Bold: true, LIns: 0, TIns: 0}))
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
	// Форма карточек тоже чередуется, усиливая структурное разнообразие.
	cardShape := "roundRect"
	if mirror {
		cardShape = "rect"
	}
	for i, bullet := range bullets {
		y := cardY + i*(cardH+gap)
		b.WriteString(shape(id, fmt.Sprintf("Point Card %d", i+1), cardShape, bulletX, y, 7150000, cardH, pal.Card, pal.Line))
		b.WriteString(shape(id, fmt.Sprintf("Point Accent %d", i+1), "rect", bulletX, y, 76200, cardH, accent, ""))
		b.WriteString(textBox(id, fmt.Sprintf("Point Text %d", i+1), bulletX+250000, y+90000, 6500000, cardH-120000, []string{bullet}, textOptions{Size: 1550, Color: pal.Ink, LIns: 0, TIns: 0}))
	}

	b.WriteString(visualPanel(id, s, accent, pal, visualX, imgRID))
	b.WriteString(footer(id, n, total, accent, pal))
	return b.String()
}

func closingSlide(id *int, s Slide, n, total int, accent string, pal palette) string {
	title := s.Title
	if title == "" {
		title = "Closing"
	}
	bullets := limitStrings(s.Bullets, 3, 92)
	if len(bullets) == 0 {
		bullets = []string{"Align on next steps", "Use the deck as a working artifact"}
	}

	var b strings.Builder
	b.WriteString(shape(id, "Closing Panel", "roundRect", 1050000, 1050000, 9850000, 4750000, pal.primary(), ""))
	b.WriteString(shape(id, "Closing Accent", "rect", 1050000, 1050000, 9850000, 152400, accent, ""))
	b.WriteString(textBox(id, "Closing Title", 1550000, 1750000, 8200000, 900000, []string{truncateText(title, 72)}, textOptions{Size: 3400, Color: onAccentText, Bold: true, Align: "ctr", LIns: 0, TIns: 0}))
	y := 3050000
	for i, bullet := range bullets {
		b.WriteString(textBox(id, fmt.Sprintf("Closing Point %d", i+1), 2150000, y+i*620000, 7000000, 420000, []string{bullet}, textOptions{Size: 1650, Color: onAccentMuted, Align: "ctr", LIns: 0, TIns: 0}))
	}
	b.WriteString(footer(id, n, total, accent, pal))
	return b.String()
}

// visualPanel рисует правую (или левую — при mirror) колонку слайда: либо
// встроенную реальную картинку с подписью/атрибуцией (imgRID != ""), либо
// текстовую «визуальную подсказку» с источником.
func visualPanel(id *int, s Slide, accent string, pal palette, x int, imgRID string) string {
	const (
		panelY = 1500000
		panelW = 3000000
		panelH = 3750000
	)
	var b strings.Builder
	b.WriteString(shape(id, "Visual Card", "roundRect", x, panelY, panelW, panelH, pal.Pale, pal.Line))
	b.WriteString(shape(id, "Visual Card Accent", "rect", x, panelY, panelW, 101600, accent, ""))

	if imgRID != "" {
		b.WriteString(picture(id, "Slide Image", imgRID, x+150000, panelY+260000, panelW-300000, 2600000))
		caption := strings.TrimSpace(s.Image.Alt)
		if caption == "" {
			caption = strings.TrimSpace(s.Visual)
		}
		if caption != "" {
			b.WriteString(textBox(id, "Image Caption", x+150000, panelY+2980000, panelW-300000, 420000, []string{truncateText(caption, 90)}, textOptions{Size: 1000, Color: pal.Ink, Bold: true, LIns: 0, TIns: 0}))
		}
		attribution := strings.TrimSpace(s.Image.Source)
		if attribution == "" && len(s.Sources) > 0 {
			attribution = s.Sources[0]
		}
		if attribution != "" {
			b.WriteString(textBox(id, "Image Source", x+150000, panelY+3400000, panelW-300000, 300000, []string{truncateText(attribution, 110)}, textOptions{Size: 850, Color: pal.Muted, LIns: 0, TIns: 0}))
		}
		return b.String()
	}

	visual := strings.TrimSpace(s.Visual)
	if visual == "" {
		visual = `Use a clean supporting image, chart, or diagram for "` + s.Title + `".`
	}
	b.WriteString(textBox(id, "Visual Label", x+300000, panelY+340000, panelW-650000, 360000, []string{"Visual direction"}, textOptions{Size: 1050, Color: pal.primary(), Bold: true, LIns: 0, TIns: 0}))
	b.WriteString(textBox(id, "Visual Body", x+300000, panelY+800000, panelW-650000, 1250000, []string{truncateText(visual, 145)}, textOptions{Size: 1350, Color: pal.Ink, Bold: true, LIns: 0, TIns: 0}))
	if len(s.Sources) > 0 {
		b.WriteString(shape(id, "Source Divider", "rect", x+300000, panelY+2280000, panelW-950000, 38100, pal.Line, ""))
		b.WriteString(textBox(id, "Source Label", x+300000, panelY+2460000, panelW-650000, 300000, []string{"Source"}, textOptions{Size: 900, Color: pal.Muted, Bold: true, LIns: 0, TIns: 0}))
		b.WriteString(textBox(id, "Source Text", x+300000, panelY+2760000, panelW-650000, 650000, []string{truncateText(s.Sources[0], 118)}, textOptions{Size: 900, Color: pal.Muted, LIns: 0, TIns: 0}))
	}
	return b.String()
}

func footer(id *int, n, total int, accent string, pal palette) string {
	return shape(id, "Footer Rule", "rect", 650000, 6350000, 4200000, 25400, accent, "") +
		textBox(id, "Footer Brand", 650000, 6420000, 3000000, 260000, []string{"Octra presentation"}, textOptions{Size: 850, Color: pal.Muted, LIns: 0, TIns: 0}) +
		textBox(id, "Footer Page", 9750000, 6420000, 1400000, 260000, []string{fmt.Sprintf("Slide %d of %d", n, total)}, textOptions{Size: 850, Color: pal.Muted, Align: "r", LIns: 0, TIns: 0})
}

// picture встраивает реальное изображение (медиа-часть, на которую ссылается rID)
// как <p:pic> в указанную область слайда.
func picture(id *int, name, rID string, x, y, cx, cy int) string {
	cur := *id
	*id = *id + 1
	return fmt.Sprintf(
		`<p:pic><p:nvPicPr><p:cNvPr id="%d" name="%s"/><p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr><p:nvPr/></p:nvPicPr>`+
			`<p:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></p:blipFill>`+
			`<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr></p:pic>`,
		cur, esc(name), rID, x, y, cx, cy,
	)
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

