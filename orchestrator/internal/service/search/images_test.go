package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeImageProvider — детерминированная заглушка провайдера изображений.
type fakeImageProvider struct {
	byQuery map[string][]Image
	err     error
	calls   []string
}

func (f *fakeImageProvider) SearchImages(_ context.Context, query string, limit int) ([]Image, error) {
	f.calls = append(f.calls, query)
	if f.err != nil {
		return nil, f.err
	}
	res := f.byQuery[query]
	if limit > 0 && len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}

func TestClientSearchImages_DedupesByURL(t *testing.T) {
	fp := &fakeImageProvider{byQuery: map[string][]Image{
		"bakery": {
			{Title: "Croissant", URL: "https://img.example/a.jpg"},
			{Title: "Dup", URL: "https://img.example/dup.jpg"},
		},
		"pastry": {
			{Title: "Dup again", URL: "https://img.example/dup.jpg/"},
			{Title: "Tart", URL: "https://img.example/b.png"},
		},
	}}
	c := NewClientWithImageProvider(fp, nil)

	imgs := c.SearchImages(context.Background(), []string{"bakery", "pastry"}, 10)
	if len(imgs) != 3 {
		t.Fatalf("expected 3 deduped images, got %d: %#v", len(imgs), imgs)
	}
	if len(fp.calls) != 2 {
		t.Fatalf("expected 2 provider calls, got %d", len(fp.calls))
	}
}

func TestClientSearchImages_RespectsLimit(t *testing.T) {
	fp := &fakeImageProvider{byQuery: map[string][]Image{
		"q": {
			{URL: "https://img.example/1.jpg"},
			{URL: "https://img.example/2.jpg"},
			{URL: "https://img.example/3.jpg"},
		},
	}}
	c := NewClientWithImageProvider(fp, nil)
	imgs := c.SearchImages(context.Background(), []string{"q"}, 2)
	if len(imgs) != 2 {
		t.Fatalf("expected limit of 2, got %d", len(imgs))
	}
}

func TestClientSearchImages_NilProvider(t *testing.T) {
	c := NewClientWithProvider(&fakeProvider{})
	if imgs := c.SearchImages(context.Background(), []string{"q"}, 5); imgs != nil {
		t.Fatalf("expected nil when no image provider, got %#v", imgs)
	}
}

var jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}

func TestFetchImage_AcceptsJPEGandPNG(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/photo.jpg":
			_, _ = w.Write(jpegMagic)
		case "/photo.png":
			_, _ = w.Write(pngMagic)
		case "/doc.svg":
			_, _ = w.Write([]byte("<svg></svg>"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := NewClientWithImageProvider(nil, srv.Client())

	jpg, err := c.FetchImage(context.Background(), Image{URL: srv.URL + "/photo.jpg", Title: "Hero"})
	if err != nil {
		t.Fatalf("jpeg fetch failed: %v", err)
	}
	if jpg.ContentType != "image/jpeg" || jpg.Alt != "Hero" {
		t.Fatalf("unexpected jpeg result: %+v", jpg)
	}

	png, err := c.FetchImage(context.Background(), Image{URL: srv.URL + "/photo.png"})
	if err != nil {
		t.Fatalf("png fetch failed: %v", err)
	}
	if png.ContentType != "image/png" || png.Alt != "Illustration" {
		t.Fatalf("unexpected png result: %+v", png)
	}

	if _, err := c.FetchImage(context.Background(), Image{URL: srv.URL + "/doc.svg"}); err == nil {
		t.Fatal("expected error for non-image content")
	}
}

func TestOpenverseSearchImages_ParsesAndFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[
			{"title":"A photo","url":"https://img.example/a.jpg","thumbnail":"https://img.example/a-thumb.jpg","foreign_landing_url":"https://page.example/a","source":"flickr","license":"cc-by","filetype":"jpg"},
			{"title":"Vector","url":"https://img.example/v.svg","filetype":"svg"},
			{"title":"PNG","url":"https://img.example/b.png","filetype":"png"}
		]}`))
	}))
	defer srv.Close()

	p := &OpenverseProvider{client: srv.Client(), endpoint: srv.URL + "/"}
	imgs, err := p.SearchImages(context.Background(), "logos", 5)
	if err != nil {
		t.Fatalf("SearchImages error: %v", err)
	}
	if len(imgs) != 2 {
		t.Fatalf("expected svg to be filtered out, got %d: %#v", len(imgs), imgs)
	}
	if imgs[0].Source != "flickr" || imgs[0].SourcePage != "https://page.example/a" {
		t.Fatalf("attribution not parsed: %+v", imgs[0])
	}
}

