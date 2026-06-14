package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnalyzeIssueBuildsFullPassport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/repos/octra-labs/demo/issues/7":
			_, _ = w.Write([]byte(`{
				"html_url":"https://github.com/octra-labs/demo/issues/7",
				"number":7,"title":"Fix login","body":"Body text",
				"state":"open","labels":[{"name":"bug"},{"name":"backend"}]
			}`))
		case r.URL.Path == "/repos/octra-labs/demo":
			_, _ = w.Write([]byte(`{"html_url":"https://github.com/octra-labs/demo","default_branch":"main"}`))
		case r.URL.Path == "/repos/octra-labs/demo/issues/7/comments":
			_, _ = w.Write([]byte(`[{"body":"first","created_at":"2024-01-01","user":{"login":"alice"}}]`))
		case r.URL.Path == "/repos/octra-labs/demo/pulls":
			// PR #11 references #7, PR #12 references unrelated #99.
			_, _ = w.Write([]byte(`[
				{"number":11,"state":"open","title":"WIP fix","html_url":"https://github.com/octra-labs/demo/pull/11","body":"Fixes #7","user":{"login":"bob"}},
				{"number":12,"state":"open","title":"Other","html_url":"https://github.com/octra-labs/demo/pull/12","body":"Closes #99","user":{"login":"carol"}}
			]`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClient("tok", "Octra Bot", "bot@example.com")
	client.apiBaseURL = srv.URL

	ref := IssueReference{Owner: "octra-labs", Repo: "demo", Number: 7, URL: "https://github.com/octra-labs/demo/issues/7"}
	in, err := client.AnalyzeIssue(context.Background(), ref)
	if err != nil {
		t.Fatalf("AnalyzeIssue error: %v", err)
	}
	if in.Title != "Fix login" || in.Body != "Body text" || in.State != "open" {
		t.Fatalf("issue fields wrong: %#v", in)
	}
	if len(in.Labels) != 2 || !in.HasLabel("bug") || !in.HasLabel("BACKEND") {
		t.Fatalf("labels wrong: %#v", in.Labels)
	}
	if in.DefaultBranch != "main" {
		t.Fatalf("default branch = %q", in.DefaultBranch)
	}
	if len(in.Comments) != 1 || in.Comments[0].Author != "alice" || in.Comments[0].Body != "first" {
		t.Fatalf("comments wrong: %#v", in.Comments)
	}
	if len(in.OpenPRs) != 1 || in.OpenPRs[0].Number != 11 {
		t.Fatalf("open PRs should only include the one referencing #7, got: %#v", in.OpenPRs)
	}
}

func TestAnalyzeIssueRequiresIssueFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	client := NewClient("tok", "", "")
	client.apiBaseURL = srv.URL
	_, err := client.AnalyzeIssue(context.Background(), IssueReference{Owner: "o", Repo: "r", Number: 1})
	if err == nil {
		t.Fatal("expected error when issue fetch fails")
	}
}

func TestPullRequestReferencesIssue(t *testing.T) {
	ref := IssueReference{Owner: "octra-labs", Repo: "demo", Number: 7}
	cases := []struct {
		pr   PullRequestListItem
		want bool
	}{
		{PullRequestListItem{Body: "Fixes #7"}, true},
		{PullRequestListItem{Title: "octra-labs/demo#7 fix"}, true},
		{PullRequestListItem{Body: "relates to #70"}, false}, // #7 must not match #70
		{PullRequestListItem{Body: "#7, also touches code"}, true},
		{PullRequestListItem{Body: "nothing here"}, false},
	}
	for i, c := range cases {
		if got := pullRequestReferencesIssue(c.pr, ref); got != c.want {
			t.Errorf("case %d: got %v want %v (%#v)", i, got, c.want, c.pr)
		}
	}
}

func TestIsTrivialIssue(t *testing.T) {
	cases := []struct {
		name string
		in   *IssueInstruction
		want bool
	}{
		{"nil", nil, false},
		{"typo title", &IssueInstruction{Title: "Fix typo in header", Body: "small"}, true},
		{"rename title", &IssueInstruction{Title: "Rename variable", Body: "x"}, true},
		{"long body not trivial", &IssueInstruction{Title: "Fix typo", Body: strings.Repeat("x", 600)}, false},
		{"enhancement label blocks", &IssueInstruction{Title: "Fix typo", Body: "x", Labels: []string{"enhancement"}}, false},
		{"docs only", &IssueInstruction{Title: "Update README", Body: "improve docs in README.md"}, true},
		{"docs but mentions code", &IssueInstruction{Title: "Update README", Body: "also change main.go in README.md"}, false},
		{"regular code issue", &IssueInstruction{Title: "Server crashes", Body: "the api returns 500"}, false},
	}
	for _, c := range cases {
		if got := IsTrivialIssue(c.in); got != c.want {
			t.Errorf("%s: IsTrivialIssue = %v, want %v", c.name, got, c.want)
		}
	}
}
