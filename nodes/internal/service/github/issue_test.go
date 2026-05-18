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
		wantFound bool
	}{
		{
			name:      "issue URL",
			text:      "Please fix https://github.com/Payel-git-ol/Octra/issues/7.",
			wantOwner: "Payel-git-ol",
			wantRepo:  "Octra",
			wantNum:   7,
			wantFound: true,
		},
		{
			name:      "www issue URL with query",
			text:      "See https://www.github.com/octra-labs/app/issues/42?comment=1",
			wantOwner: "octra-labs",
			wantRepo:  "app",
			wantNum:   42,
			wantFound: true,
		},
		{
			name:      "ordinary repository URL is reference only",
			text:      "Use https://github.com/gin-gonic/gin as a library reference.",
			wantFound: false,
		},
		{
			name:      "pull request URL is not an issue target",
			text:      "Review https://github.com/Payel-git-ol/Octra/pull/8",
			wantFound: false,
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
		})
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
		Head:  "octra/issue-7-abc12345",
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
		gotReq.Head != "octra/issue-7-abc12345" ||
		gotReq.Base != "main" ||
		!strings.Contains(gotReq.Body, "Fixes octra-labs/demo#7") {
		t.Fatalf("unexpected request body: %#v", gotReq)
	}
}
