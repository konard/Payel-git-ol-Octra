package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// DuckDuckGoProvider — провайдер веб-поиска поверх HTML-эндпоинта DuckDuckGo
// (https://html.duckduckgo.com/html/). Он не требует API-ключа, что удобно для
// self-hosted развёртываний Octra. Результаты приходят как обычная HTML-страница,
// которую мы разбираем парсером golang.org/x/net/html.
type DuckDuckGoProvider struct {
	client   *http.Client
	endpoint string
}

// NewDuckDuckGoProvider создаёт провайдер с разумным таймаутом.
func NewDuckDuckGoProvider() *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		client:   &http.Client{Timeout: 12 * time.Second},
		endpoint: "https://html.duckduckgo.com/html/",
	}
}

// Search выполняет запрос и возвращает до limit результатов.
func (p *DuckDuckGoProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("empty query")
	}

	reqURL := p.endpoint + "?q=" + url.QueryEscape(query) + "&kl=wt-wt"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	// DuckDuckGo отдаёт HTML только клиентам с правдоподобным User-Agent.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OctraResearchBot/1.0)")
	req.Header.Set("Accept", "text/html")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search provider returned status %d", resp.StatusCode)
	}

	results := parseDuckDuckGoHTML(resp.Body)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// parseDuckDuckGoHTML извлекает результаты из HTML-страницы DuckDuckGo.
// Вынесено в отдельную функцию, чтобы покрыть парсинг unit-тестами без сети.
//
// В разметке html.duckduckgo.com каждый результат — это блок с ссылкой
// <a class="result__a" href="…">Title</a> и сниппетом
// <a class="result__snippet">…</a>. href часто является редиректом вида
// //duckduckgo.com/l/?uddg=<urlencoded>, поэтому реальный URL берём из uddg.
func parseDuckDuckGoHTML(r io.Reader) []Result {
	doc, err := html.Parse(r)
	if err != nil {
		return nil
	}

	var results []Result
	var current *Result

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			class := attr(n, "class")
			switch {
			case strings.Contains(class, "result__a"):
				// Начало нового результата.
				href := decodeDDGHref(attr(n, "href"))
				title := strings.TrimSpace(textContent(n))
				if href != "" && title != "" {
					results = append(results, Result{
						Title:  title,
						URL:    href,
						Source: domainOf(href),
					})
					current = &results[len(results)-1]
				}
			case strings.Contains(class, "result__snippet"):
				if current != nil && current.Snippet == "" {
					current.Snippet = strings.TrimSpace(textContent(n))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return results
}

// decodeDDGHref разворачивает редирект DuckDuckGo (uddg=…) в реальный URL.
func decodeDDGHref(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	// Относительный протокол → https.
	parseTarget := href
	if strings.HasPrefix(parseTarget, "//") {
		parseTarget = "https:" + parseTarget
	}
	u, err := url.Parse(parseTarget)
	if err != nil {
		return href
	}
	if uddg := u.Query().Get("uddg"); uddg != "" {
		return uddg
	}
	if strings.HasPrefix(u.Scheme, "http") {
		return u.String()
	}
	return href
}

// domainOf возвращает домен URL без схемы и www.
func domainOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return strings.TrimPrefix(u.Host, "www.")
}

// attr возвращает значение атрибута элемента (или "").
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// textContent собирает весь текст внутри узла.
func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
