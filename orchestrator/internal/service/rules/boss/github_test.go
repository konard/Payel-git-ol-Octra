package boss

import (
	"context"
	"strings"
	"testing"

	"orchestrator/pkg/models"
)

// TestPushToGitHubReportsMissingToken — регрессия на issue #75 п.3: раньше при
// отсутствии GITHUB_TOKEN (или незаданном githubClient) pushToGitHub молча
// возвращал "", из-за чего нода GitHub во фронтенде висела в статусе «building»
// без объяснения причины. Теперь функция обязана вернуть человекочитаемую
// причину, которую фронтенд показывает на красной ноде.
func TestPushToGitHubReportsMissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	// githubClient == nil — guard срабатывает до любого обращения к клиенту,
	// поэтому пустого Service достаточно.
	service := &Service{}
	task := &models.Task{}

	repoURL, reason := service.pushToGitHub(context.Background(), task, "/tmp/does-not-matter", nil, nil)

	if repoURL != "" {
		t.Errorf("expected empty repoURL when GitHub is not configured, got %q", repoURL)
	}
	if reason == "" {
		t.Fatal("expected a human-readable reason when GitHub is not configured, got empty string")
	}
	if !strings.Contains(reason, "GITHUB_TOKEN") {
		t.Errorf("reason should mention the missing GITHUB_TOKEN, got %q", reason)
	}
}
