package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseIssueReferenceRequiresConcreteIssueURL(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantOwner string
		wantRepo  string
		wantNum   int
		wantPR    bool
		wantURL   string
		wantFound bool
	}{
		{
			name:      "issue URL",
			text:      "Please fix https://github.com/Payel-git-ol/Octra/issues/7.",
			wantOwner: "Payel-git-ol",
			wantRepo:  "Octra",
			wantNum:   7,
			wantURL:   "https://github.com/Payel-git-ol/Octra/issues/7",
			wantFound: true,
		},
		{
			name:      "www issue URL with query",
			text:      "See https://www.github.com/octra-labs/app/issues/42?comment=1",
			wantOwner: "octra-labs",
			wantRepo:  "app",
			wantNum:   42,
			wantURL:   "https://github.com/octra-labs/app/issues/42",
			wantFound: true,
		},
		{
			name:      "ordinary repository URL is reference only",
			text:      "Use https://github.com/gin-gonic/gin as a library reference.",
			wantFound: false,
		},
		{
			// Pasting a pull request link must now start the workflow too
			// (issue #44): previously PR URLs were ignored and nothing happened.
			name:      "pull request URL is a target",
			text:      "Review https://github.com/Payel-git-ol/Octra/pull/8",
			wantOwner: "Payel-git-ol",
			wantRepo:  "Octra",
			wantNum:   8,
			wantPR:    true,
			wantURL:   "https://github.com/Payel-git-ol/Octra/pull/8",
			wantFound: true,
		},
		{
			name:      "plural pulls URL is a target",
			text:      "Continue https://github.com/octra-labs/app/pulls/15 please",
			wantOwner: "octra-labs",
			wantRepo:  "app",
			wantNum:   15,
			wantPR:    true,
			wantURL:   "https://github.com/octra-labs/app/pull/15",
			wantFound: true,
		},
		{
			name:      "non github URL",
			text:      "Docs: https://example.com/owner/repo/issues/7",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := ParseIssueReference(tt.text)
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}
			if !tt.wantFound {
				return
			}
			if got.Owner != tt.wantOwner || got.Repo != tt.wantRepo || got.Number != tt.wantNum {
				t.Fatalf("got %#v, want owner=%q repo=%q number=%d", got, tt.wantOwner, tt.wantRepo, tt.wantNum)
			}
			if got.IsPullRequest != tt.wantPR {
				t.Fatalf("IsPullRequest = %v, want %v", got.IsPullRequest, tt.wantPR)
			}
			if got.URL != tt.wantURL {
				t.Fatalf("URL = %q, want %q", got.URL, tt.wantURL)
			}
		})
	}
}

func TestNewIssueBranchNameUsesTaskID(t *testing.T) {
	taskID := "550e8400-e29b-41d4-a716-446655440000"
	got := NewIssueBranchName(taskID)
	want := "Issue-550e8400-e29b-41d4-a716-446655440000"
	if got != want {
		t.Fatalf("NewIssueBranchName() = %q, want %q", got, want)
	}
}

func TestCreatePullRequestSendsExpectedPayload(t *testing.T) {
	var gotAuth string
	var gotReq PullRequestRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/repos/octra-labs/demo/pulls" {
			t.Fatalf("path = %s, want /repos/octra-labs/demo/pulls", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"html_url":"https://github.com/octra-labs/demo/pull/12","number":12}`))
	}))
	defer srv.Close()

	client := NewClient("ghs_test_token", "Octra Bot", "bot@example.com")
	client.apiBaseURL = srv.URL

	pr, err := client.CreatePullRequest(context.Background(), PullRequestRequest{
		Owner: "octra-labs",
		Repo:  "demo",
		Title: "Fix #7: Create pull request",
		Head:  "Issue-550e8400-e29b-41d4-a716-446655440000",
		Base:  "main",
		Body:  "Fixes octra-labs/demo#7",
	})
	if err != nil {
		t.Fatalf("CreatePullRequest returned error: %v", err)
	}
	if pr.HTMLURL != "https://github.com/octra-labs/demo/pull/12" || pr.Number != 12 {
		t.Fatalf("unexpected PR response: %#v", pr)
	}
	if gotAuth != "token ghs_test_token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotReq.Owner != "" || gotReq.Repo != "" {
		t.Fatalf("owner/repo must stay path-only, got request body %#v", gotReq)
	}
	if gotReq.Title != "Fix #7: Create pull request" ||
		gotReq.Head != "Issue-550e8400-e29b-41d4-a716-446655440000" ||
		gotReq.Base != "main" ||
		!strings.Contains(gotReq.Body, "Fixes octra-labs/demo#7") {
		t.Fatalf("unexpected request body: %#v", gotReq)
	}
}

