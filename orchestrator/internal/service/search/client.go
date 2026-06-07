package search

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client — высокоуровневый поисковый клиент, которым пользуются воркеры.
// Он оборачивает Provider, выполняет несколько запросов (по диапазону, заданному
// менеджером), убирает дубли по URL и переранжирует объединённый набор алгоритмом
// BM25 относительно темы ресёрча. Для презентаций он дополнительно умеет искать
// изображения (imageProvider) и скачивать их байты (httpClient) для встраивания.
type Client struct {
	provider      Provider
	imageProvider ImageProvider
	httpClient    *http.Client
}

// NewClient создаёт клиент с дефолтными провайдерами: DuckDuckGo для текстового
// поиска и Openverse для поиска изображений.
func NewClient() *Client {
	return &Client{
		provider:      NewDuckDuckGoProvider(),
		imageProvider: NewOpenverseProvider(),
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

// NewClientWithProvider создаёт клиент с произвольным текстовым провайдером (для тестов).
func NewClientWithProvider(p Provider) *Client {
	return &Client{provider: p}
}

// NewClientWithImageProvider создаёт клиент с произвольным провайдером изображений
// и (опционально) http-клиентом для скачивания картинок (для тестов).
func NewClientWithImageProvider(p ImageProvider, httpClient *http.Client) *Client {
	return &Client{imageProvider: p, httpClient: httpClient}
}

// Enabled сообщает, разрешён ли веб-поиск. Его можно полностью выключить
// переменной окружения WEB_SEARCH_DISABLED=1 (например, в офлайн-окружении).
func Enabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("WEB_SEARCH_DISABLED")))
	return v != "1" && v != "true" && v != "yes"
}

// Research выполняет несколько поисковых запросов и возвращает топ-результаты,
// уже переранжированные BM25 относительно темы topic. perQuery ограничивает число
// результатов на один запрос, limit — итоговое число лучших источников.
func (c *Client) Research(ctx context.Context, topic string, queries []string, perQuery, limit int) ([]Result, error) {
	if c == nil || c.provider == nil {
		return nil, fmt.Errorf("search client not initialized")
	}
	if perQuery <= 0 {
		perQuery = 5
	}

	seen := make(map[string]struct{})
	var merged []Result
	var firstErr error

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		res, err := c.provider.Search(ctx, q, perQuery)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, r := range res {
			key := normalizeURL(r.URL)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, r)
		}
	}

	if len(merged) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, nil
	}

	ranked := RankBM25(topic, merged)
	if limit > 0 && len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked, nil
}

// FormatForPrompt превращает результаты в компактный блок для LLM-промпта.
func FormatForPrompt(results []Result) string {
	if len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, r := range results {
		fmt.Fprintf(&sb, "[%d] %s\n", i+1, r.Title)
		fmt.Fprintf(&sb, "    URL: %s\n", r.URL)
		if r.Snippet != "" {
			fmt.Fprintf(&sb, "    %s\n", collapseSpaces(r.Snippet))
		}
	}
	return sb.String()
}

// FormatSourcesMarkdown собирает Markdown-список найденных источников для файла
// решения (solution/sources-*.md), чтобы у ответа были проверяемые ссылки.
func FormatSourcesMarkdown(results []Result) string {
	if len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Sources\n\n")
	for _, r := range results {
		title := r.Title
		if title == "" {
			title = r.URL
		}
		fmt.Fprintf(&sb, "- [%s](%s)", title, r.URL)
		if r.Source != "" {
			fmt.Fprintf(&sb, " — %s", r.Source)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimRight(u, "/")
	return strings.ToLower(u)
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

