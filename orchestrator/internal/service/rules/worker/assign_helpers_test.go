package worker

import "testing"

func TestParseWorkerMetadataSearchConfig(t *testing.T) {
	meta := parseWorkerMetadata(map[string]string{
		"search": `{"provider":"apodex","model":"apodex-1-0-deepresearch-mini","base-url":"https://api.apodex.ai/v1/responses","api-key":"sk-test","striming":true}`,
	})

	if meta.searchConfig == nil {
		t.Fatal("expected search config to be parsed")
	}
	if meta.searchConfig.Provider != "apodex" {
		t.Fatalf("provider = %q, want apodex", meta.searchConfig.Provider)
	}
	if meta.searchConfig.Model != "apodex-1-0-deepresearch-mini" {
		t.Fatalf("model = %q", meta.searchConfig.Model)
	}
	if meta.searchConfig.BaseURL != "https://api.apodex.ai/v1/responses" {
		t.Fatalf("base URL = %q", meta.searchConfig.BaseURL)
	}
	if meta.searchConfig.APIKey != "sk-test" {
		t.Fatalf("API key = %q", meta.searchConfig.APIKey)
	}
	if !meta.searchConfig.Streaming {
		t.Fatal("expected striming=true to enable streaming")
	}
}

func TestParseWorkerMetadataIgnoresIncompleteSearchConfig(t *testing.T) {
	meta := parseWorkerMetadata(map[string]string{
		"search": `{"provider":"apodex","model":"apodex-1-0-deepresearch-mini","base-url":"https://api.apodex.ai/v1/responses","striming":true}`,
	})

	if meta.searchConfig != nil {
		t.Fatalf("expected incomplete search config to be ignored, got %#v", meta.searchConfig)
	}
}
