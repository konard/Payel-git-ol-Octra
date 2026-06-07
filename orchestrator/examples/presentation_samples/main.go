package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/service/document"
)

type sampleDeck struct {
	FileName string
	Deck     document.Deck
}

func main() {
	repoRoot := filepath.Clean("..")
	outDir := filepath.Join(repoRoot, "docs", "examples", "presentations", "issue-23")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	samples := buildSamples()
	for _, sample := range samples {
		data, err := document.BuildPPTX(sample.Deck)
		if err != nil {
			panic(fmt.Errorf("build %s: %w", sample.FileName, err))
		}
		if err := os.WriteFile(filepath.Join(outDir, sample.FileName+".pptx"), data, 0644); err != nil {
			panic(err)
		}
	}

	page, err := renderPreview(samples)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "preview.html"), page, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d sample decks to %s\n", len(samples), outDir)
}

func buildSamples() []sampleDeck {
	return []sampleDeck{
		{
			FileName: "market-expansion-brief",
			Deck: document.Deck{
				Title: "Market Expansion Brief",
				Slides: []document.Slide{
					{Title: "Market Expansion Brief", Bullets: []string{"Compare demand signals", "Prioritize near-term launch markets"}, Visual: "Hero map with highlighted target regions"},
					{Title: "Demand Signals", Bullets: []string{"Search volume clusters around three use cases", "Customer interviews show urgency in operations", "Budget owners prefer measurable rollout milestones"}, Visual: "Stacked demand bars by region", Sources: []string{"SBA market research guide - https://www.sba.gov/business-guide/plan-your-business/market-research-competitive-analysis"}},
					{Title: "Audience Segments", Bullets: []string{"Operations leaders need faster reporting", "Finance teams want lower manual review load", "Founders value simple implementation paths"}, Visual: "Three persona cards with workflow snapshots", Sources: []string{"U.S. Census retail trade data - https://www.census.gov/retail/"}},
					{Title: "Go-To-Market Motion", Bullets: []string{"Anchor messaging in time saved", "Pair webinars with customer proof", "Use partner demos for first meetings"}, Visual: "Funnel diagram from awareness to pilot"},
					{Title: "Risks To Manage", Bullets: []string{"Regulatory language differs by region", "Support coverage must match launch hours", "Procurement cycles can delay conversion"}, Visual: "Risk matrix with impact and mitigation"},
					{Title: "Next Steps", Bullets: []string{"Validate the top two regions", "Prepare localized sales assets", "Launch a four-week pilot campaign"}, Visual: "Checklist over a launch calendar"},
				},
			},
		},
		{
			FileName: "bakery-product-launch",
			Deck: document.Deck{
				Title: "Neighborhood Bakery Launch",
				Slides: []document.Slide{
					{Title: "Neighborhood Bakery Launch", Bullets: []string{"Introduce a seasonal pastry line", "Build repeat morning traffic"}, Visual: "Warm product photo with storefront detail"},
					{Title: "Customer Moment", Bullets: []string{"Weekday commuters want fast premium options", "Families shop visually on weekends", "Office orders need predictable packaging"}, Visual: "Customer journey strip from sidewalk to counter"},
					{Title: "Product Story", Bullets: []string{"Lead with one signature pastry", "Use local ingredients in naming", "Offer three price points for trial"}, Visual: "Pastry hero shot with ingredient callouts", Sources: []string{"SBA market research guide - https://www.sba.gov/business-guide/plan-your-business/market-research-competitive-analysis"}},
					{Title: "Launch Calendar", Bullets: []string{"Preview tastings one week before launch", "Post daily behind-the-scenes reels", "Invite nearby offices to preorder"}, Visual: "Four-week calendar with tasting markers"},
					{Title: "In-Store Experience", Bullets: []string{"Place the hero item near checkout", "Use clear allergen cards", "Bundle coffee upgrades at the register"}, Visual: "Counter layout diagram with product placement"},
					{Title: "Wrap Up", Bullets: []string{"Pilot for two weekends", "Track sell-through by hour", "Keep the winning flavor in rotation"}, Visual: "Sales dashboard beside bakery display"},
				},
			},
		},
		{
			FileName: "campus-sustainability-plan",
			Deck: document.Deck{
				Title: "Campus Sustainability Plan",
				Slides: []document.Slide{
					{Title: "Campus Sustainability Plan", Bullets: []string{"Reduce waste across daily routines", "Make progress visible to students"}, Visual: "Aerial campus photo with green route overlay"},
					{Title: "Current Baseline", Bullets: []string{"Dining produces the largest visible waste stream", "Dorm participation varies by building", "Events create short spikes in landfill volume"}, Visual: "Waste stream bar chart by source", Sources: []string{"EPA Reduce, Reuse, Recycle - https://www.epa.gov/recycle"}},
					{Title: "Behavior Design", Bullets: []string{"Put sorting cues at decision points", "Make reuse stations easy to find", "Reward teams for monthly improvement"}, Visual: "Before-after bin station comparison"},
					{Title: "Operations Changes", Bullets: []string{"Standardize bin colors across buildings", "Add event kits for high-traffic venues", "Publish pickup schedules for transparency"}, Visual: "Operations workflow map from bins to reporting"},
					{Title: "Measurement Plan", Bullets: []string{"Track landfill diversion weekly", "Share building-level leaderboards", "Review contamination with facilities staff"}, Visual: "Dashboard mockup with trend lines", Sources: []string{"EPA reducing waste guide - https://www.epa.gov/recycle/reducing-waste-what-you-can-do"}},
					{Title: "Conclusion", Bullets: []string{"Start with dining and residence halls", "Report progress in plain language", "Use visible wins to expand participation"}, Visual: "Student team reviewing dashboard"},
				},
			},
		},
	}
}

func renderPreview(samples []sampleDeck) ([]byte, error) {
	const page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="icon" href="data:,">
<title>Issue 23 Presentation Samples</title>
<style>
body { margin: 0; background: #ece8df; color: #0b1f28; font-family: Inter, Arial, sans-serif; }
main { max-width: 1440px; margin: 0 auto; padding: 42px; }
h1 { margin: 0 0 10px; font-size: 32px; }
.intro { margin: 0 0 28px; color: #5d6970; }
.deck { margin: 0 0 36px; }
.deck h2 { font-size: 22px; margin: 0 0 14px; }
.slides { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 18px; }
.slide { aspect-ratio: 16 / 9; background: #f6f4ef; border: 1px solid #d9dfda; box-shadow: 0 18px 45px rgba(11,31,40,.14); position: relative; overflow: hidden; }
.slide::before { content: ""; position: absolute; left: 0; top: 0; right: 0; height: 8px; background: #0a4b5a; }
.slide::after { content: ""; position: absolute; top: 0; right: 0; bottom: 0; width: 10px; background: var(--accent); }
.title-slide { display: grid; grid-template-columns: 1fr 36%; }
.title-copy { padding: 58px 34px 26px; }
.title-copy h3 { font-size: 28px; line-height: 1.05; margin: 0 0 14px; }
.title-copy p { color: #5d6970; margin: 0; font-size: 12px; }
.visual-dark { margin: 38px 28px 34px 0; border-radius: 8px; background: #0a4b5a; color: white; padding: 24px; display: flex; flex-direction: column; justify-content: center; }
.visual-dark span, .visual span { color: #bfe3dc; font-size: 10px; font-weight: 700; text-transform: uppercase; }
.visual-dark strong { font-size: 17px; line-height: 1.25; margin-top: 8px; }
.content { padding: 34px 28px 24px; display: grid; grid-template-columns: 1fr 30%; gap: 20px; }
.content h3 { margin: 0 0 10px; font-size: 20px; line-height: 1.12; }
.rule { width: 88px; height: 4px; background: var(--accent); margin-bottom: 14px; }
.point { background: white; border: 1px solid #d9dfda; border-left: 5px solid var(--accent); border-radius: 8px; margin-bottom: 8px; padding: 9px 11px; font-size: 11px; line-height: 1.3; }
.visual { background: #e9f2f0; border: 1px solid #b8cdc8; border-top: 5px solid var(--accent); border-radius: 8px; padding: 14px; font-size: 11px; line-height: 1.32; }
.source { border-top: 1px solid #b8cdc8; color: #5d6970; font-size: 9px; margin-top: 12px; padding-top: 8px; }
@media (max-width: 980px) { main { padding: 24px; } .slides { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<main>
<h1>Issue 23 Presentation Samples</h1>
<p class="intro">Generated from the same slide model that writes the PPTX files.</p>
{{range .}}
{{$deck := .Deck}}
<section class="deck">
<h2>{{$deck.Title}}</h2>
<div class="slides">
{{range $i, $slide := $deck.Slides}}
<article class="slide {{if eq $i 0}}title-slide{{else}}content{{end}}" style="--accent: {{accent $i}}">
{{if eq $i 0}}
<div class="title-copy">
<h3>{{$deck.Title}}</h3>
<p>{{join $slide.Bullets}}</p>
</div>
<div class="visual-dark"><span>Visual direction</span><strong>{{$slide.Visual}}</strong></div>
{{else}}
<div>
<h3>{{$slide.Title}}</h3>
<div class="rule"></div>
{{range $slide.Bullets}}<div class="point">{{.}}</div>{{end}}
</div>
<aside class="visual"><span>Visual direction</span><p>{{$slide.Visual}}</p>{{with first $slide.Sources}}<div class="source">{{.}}</div>{{end}}</aside>
{{end}}
</article>
{{end}}
</div>
</section>
{{end}}
</main>
</body>
</html>`
	tmpl, err := template.New("preview").Funcs(template.FuncMap{
		"accent": func(i int) string {
			colors := []string{"#0a4b5a", "#d45d3a", "#e6a23c", "#1e6f91", "#3d6b5b"}
			return colors[i%len(colors)]
		},
		"join": func(in []string) string {
			return strings.Join(in, " / ")
		},
		"first": func(in []string) string {
			if len(in) == 0 {
				return ""
			}
			return in[0]
		},
	}).Parse(page)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, samples); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

