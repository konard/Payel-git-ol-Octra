package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Image — одна найденная картинка для иллюстрации презентации.
// URL указывает на сам файл изображения (его можно встроить в .pptx),
// а SourcePage/License позволяют корректно сослаться на источник.
type Image struct {
	Title      string // подпись/заголовок изображения
	URL        string // прямой URL файла изображения (jpeg/png)
	Thumbnail  string // уменьшенная превью-версия (если есть)
	SourcePage string // страница-источник (для атрибуции)
	Source     string // провайдер/домен (например, "flickr.com")
	License    string // лицензия (например, "cc-by")
}

// ImageProvider — источник результатов поиска изображений.
// Реализуется OpenverseProvider, а в тестах — заглушкой.
type ImageProvider interface {
	// SearchImages возвращает до limit изображений по запросу query.
	SearchImages(ctx context.Context, query string, limit int) ([]Image, error)
}

// ImageData — скачанные байты изображения вместе с MIME-типом,
// готовые к встраиванию в документ (PPTX).
type ImageData struct {
	URL         string
	Alt         string
	Data        []byte
	ContentType string // "image/jpeg" или "image/png"
}

// maxImageBytes ограничивает размер скачиваемой картинки, чтобы презентация
// не раздувалась и чтобы недобросовестный URL не выкачивал гигабайты.
const maxImageBytes = 4 << 20 // 4 MiB

// OpenverseProvider — провайдер поиска изображений поверх открытого API
// Openverse (https://api.openverse.org). API не требует ключа и отдаёт только
// изображения со свободными лицензиями вместе с прямыми ссылками на файлы,
// поэтому подходит для self-hosted развёртываний Octra и для встраивания.
type OpenverseProvider struct {
	client   *http.Client
	endpoint string
}

// NewOpenverseProvider создаёт провайдер с разумным таймаутом.
func NewOpenverseProvider() *OpenverseProvider {
	return &OpenverseProvider{
		client:   &http.Client{Timeout: 12 * time.Second},
		endpoint: "https://api.openverse.org/v1/images/",
	}
}

type openverseResponse struct {
	Results []openverseResult `json:"results"`
}

type openverseResult struct {
	Title          string `json:"title"`
	URL            string `json:"url"`
	Thumbnail      string `json:"thumbnail"`
	ForeignLanding string `json:"foreign_landing_url"`
	Source         string `json:"source"`
	Provider       string `json:"provider"`
	License        string `json:"license"`
	FileType       string `json:"filetype"`
}

// SearchImages выполняет запрос к Openverse и возвращает до limit изображений.
// Для встраивания в PPTX мы оставляем только jpeg/png — форматы, которые
// PowerPoint гарантированно открывает.
func (p *OpenverseProvider) SearchImages(ctx context.Context, query string, limit int) ([]Image, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("empty query")
	}
	if limit <= 0 {
		limit = 5
	}

	q := url.Values{}
	q.Set("q", query)
	q.Set("page_size", fmt.Sprintf("%d", limit*3)) // запас под фильтрацию по формату
	q.Set("license_type", "all")
	// Просим только растровые форматы, которые умеет встраивать PowerPoint.
	q.Set("extension", "jpg,png")

	reqURL := p.endpoint + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "OctraResearchBot/1.0 (+https://github.com/Payel-git-ol/Octra)")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image search returned status %d", resp.StatusCode)
	}

	var body openverseResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return nil, err
	}

	images := make([]Image, 0, len(body.Results))
	for _, r := range body.Results {
		if !looksLikeEmbeddableImage(r.URL, r.FileType) {
			continue
		}
		source := r.Source
		if source == "" {
			source = r.Provider
		}
		images = append(images, Image{
			Title:      strings.TrimSpace(r.Title),
			URL:        strings.TrimSpace(r.URL),
			Thumbnail:  strings.TrimSpace(r.Thumbnail),
			SourcePage: strings.TrimSpace(r.ForeignLanding),
			Source:     strings.TrimSpace(source),
			License:    strings.TrimSpace(r.License),
		})
		if len(images) >= limit {
			break
		}
	}
	return images, nil
}

// looksLikeEmbeddableImage отсеивает форматы, которые PowerPoint может не открыть
// (gif/svg/webp и т.п.), опираясь на расширение URL и поле filetype.
func looksLikeEmbeddableImage(rawURL, fileType string) bool {
	if strings.TrimSpace(rawURL) == "" {
		return false
	}
	ft := strings.ToLower(strings.TrimSpace(fileType))
	if ft == "jpg" || ft == "jpeg" || ft == "png" {
		return true
	}
	lower := strings.ToLower(rawURL)
	if i := strings.IndexByte(lower, '?'); i >= 0 {
		lower = lower[:i]
	}
	for _, ext := range []string{".jpg", ".jpeg", ".png"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	// filetype пуст и расширение неочевидно — всё равно пробуем: настоящий
	// MIME-тип проверяется при скачивании в FetchImage.
	return ft == ""
}

// SearchImages выполняет поиск изображений по нескольким запросам, убирает дубли
// по URL и возвращает до limit картинок. Поиск устойчив к сбоям: при ошибке или
// выключенном поиске возвращается пустой срез без ошибки, и воркер просто
// продолжает работу без иллюстраций.
func (c *Client) SearchImages(ctx context.Context, queries []string, limit int) []Image {
	if c == nil || c.imageProvider == nil {
		return nil
	}
	if limit <= 0 {
		limit = 6
	}

	seen := make(map[string]struct{})
	var out []Image
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		res, err := c.imageProvider.SearchImages(ctx, q, limit)
		if err != nil {
			continue
		}
		for _, img := range res {
			key := normalizeURL(img.URL)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, img)
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}

// FetchImage скачивает байты изображения и проверяет, что это действительно
// jpeg или png (по магическим байтам, а не по заголовку сервера). Возвращает
// ошибку, если формат не поддерживается или файл слишком большой.
func (c *Client) FetchImage(ctx context.Context, img Image) (*ImageData, error) {
	if c == nil || c.httpClient == nil {
		return nil, fmt.Errorf("image fetch not initialized")
	}
	if strings.TrimSpace(img.URL) == "" {
		return nil, fmt.Errorf("empty image URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, img.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "OctraResearchBot/1.0 (+https://github.com/Payel-git-ol/Octra)")
	req.Header.Set("Accept", "image/jpeg,image/png")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxImageBytes {
		return nil, fmt.Errorf("image exceeds %d bytes", maxImageBytes)
	}

	ct := detectImageContentType(data)
	if ct == "" {
		return nil, fmt.Errorf("unsupported image format (only jpeg/png are embeddable)")
	}

	alt := strings.TrimSpace(img.Title)
	if alt == "" {
		alt = "Illustration"
	}
	return &ImageData{URL: img.URL, Alt: alt, Data: data, ContentType: ct}, nil
}

// detectImageContentType определяет тип по магическим байтам файла.
// Возвращает "" для неподдерживаемых форматов.
func detectImageContentType(data []byte) string {
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "image/png"
	}
	return ""
}

