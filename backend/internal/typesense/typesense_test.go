package typesense

import (
	"testing"
)

func TestSkillDocumentStruct(t *testing.T) {
	doc := SkillDocument{
		ID:         "anthropics/skills/claude-api",
		SkillID:    "claude-api",
		Name:       "claude-api",
		Source:     "anthropics/skills",
		InstallCmd: "npx skills add https://github.com/anthropics/skills --skill claude-api",
	}
	if doc.ID != "anthropics/skills/claude-api" {
		t.Fatalf("doc.ID = %q", doc.ID)
	}
	if doc.SkillID != "claude-api" {
		t.Fatalf("doc.SkillID = %q", doc.SkillID)
	}
	if doc.Name != "claude-api" {
		t.Fatalf("doc.Name = %q", doc.Name)
	}
	if doc.InstallCmd == "" {
		t.Fatal("InstallCmd should not be empty")
	}
}

func TestProviderDocumentStruct(t *testing.T) {
	doc := ProviderDocument{
		ID:           "provider-openai",
		Key:          "openai",
		Name:         "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		AuthEnv:      "OPENAI_API_KEY",
		DefaultModel: "gpt-4.1",
		Description:  "OpenAI-compatible chat completions provider.",
	}
	if doc.Key != "openai" {
		t.Fatalf("doc.Key = %q", doc.Key)
	}
	if doc.BaseURL == "" || doc.AuthEnv == "" {
		t.Fatalf("provider document is missing connection metadata: %+v", doc)
	}
}

func TestStrPtr(t *testing.T) {
	s := "hello"
	ptr := strPtr(s)
	if ptr == nil {
		t.Fatal("strPtr returned nil")
	}
	if *ptr != s {
		t.Fatalf("strPtr() = %q, want %q", *ptr, s)
	}
}

func TestNew(t *testing.T) {
	c := New("localhost:8108", "test-key")
	if c == nil {
		t.Fatal("New() returned nil")
	}
}
