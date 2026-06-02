package worker

import (
	"context"
	"strings"
	"testing"

	"nodes/internal/service/document"
)

func TestShortID(t *testing.T) {
	cases := map[string]string{
		"1a2b3c4d-5e6f-7a8b-9c0d-112233445566": "1a2b3c4d",
		"":                                     "",
		"abcdefghuvwxyz":                       "abcdefgh",
		"short":                                "short",
	}
	for in, want := range cases {
		if got := shortID(in); got != want {
			t.Errorf("shortID(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestRebuildPresentationArtifacts — после правки slide-Markdown во время ревью
// .pptx должен пересобираться из обновлённого Markdown, а не оставаться старым.
func TestRebuildPresentationArtifacts(t *testing.T) {
	originalMD := "# Old deck\n\n## Slide 1\n- old point\n"
	deck := document.ParseSlideMarkdown(originalMD)
	oldPPTX, err := document.BuildPPTX(deck)
	if err != nil {
		t.Fatalf("build old pptx: %v", err)
	}

	original := map[string]string{
		"solution/designer-abc.md":   originalMD,
		"solution/designer-abc.pptx": string(oldPPTX),
	}
	revisedMD := "# New deck\n\n## Slide 1\n- brand new point\n\n## Slide 2\n- another\n"
	fixed := map[string]string{
		"solution/designer-abc.md": revisedMD,
	}

	// Без searchClient повторный поиск картинок пропускается, поведение прежнее.
	(&Service{}).rebuildPresentationArtifacts(context.Background(), original, fixed)

	got, ok := fixed["solution/designer-abc.pptx"]
	if !ok {
		t.Fatal("expected rebuilt .pptx in fixed files")
	}
	if got == string(oldPPTX) {
		t.Error("rebuilt .pptx should differ from the old one")
	}
	if !strings.HasPrefix(got, "PK") {
		t.Errorf("rebuilt .pptx should be a zip (PK header), got prefix %q", safePrefix(got))
	}
}

func safePrefix(s string) string {
	if len(s) > 4 {
		return s[:4]
	}
	return s
}
