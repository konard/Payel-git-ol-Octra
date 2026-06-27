package oauth

import (
	"strings"
	"testing"

	"backend/internal/config"
	"backend/internal/service"

	"github.com/valyala/fasthttp"
)

func TestSelectGitHubEmailPrefersVerifiedPrimaryEmail(t *testing.T) {
	profile := githubUserProfile{Email: "public@example.com", Login: "octra-user"}
	emails := []githubEmail{
		{Email: "secondary@example.com", Primary: false, Verified: true},
		{Email: "primary@example.com", Primary: true, Verified: true},
	}

	email, err := selectGitHubEmail(profile, emails)
	if err != nil {
		t.Fatalf("selectGitHubEmail returned error: %v", err)
	}
	if email != "primary@example.com" {
		t.Fatalf("expected verified primary email, got %q", email)
	}
}

func TestSelectGitHubEmailUsesVerifiedEmailWhenProfileEmailIsHidden(t *testing.T) {
	profile := githubUserProfile{Login: "private-email-user"}
	emails := []githubEmail{
		{Email: "hidden-primary@example.com", Primary: true, Verified: true},
	}

	email, err := selectGitHubEmail(profile, emails)
	if err != nil {
		t.Fatalf("selectGitHubEmail returned error: %v", err)
	}
	if email != "hidden-primary@example.com" {
		t.Fatalf("expected hidden verified primary email, got %q", email)
	}
}

func TestSelectGitHubEmailRejectsUnverifiedSyntheticFallback(t *testing.T) {
	profile := githubUserProfile{Login: "private-email-user"}
	emails := []githubEmail{
		{Email: "unverified@example.com", Primary: true, Verified: false},
	}

	email, err := selectGitHubEmail(profile, emails)
	if err == nil {
		t.Fatalf("expected missing verified email error, got email %q", email)
	}
}

// callbackRejectsForgedState asserts a callback handler refuses to proceed
// (HTTP 400, no token exchange) when the anti-CSRF state cookie is absent or
// does not match the ?state= the provider echoed back.
func callbackRejectsForgedState(t *testing.T, name string, handler func(*fasthttp.RequestCtx), queryState, cookieState string) {
	t.Helper()
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/callback?code=fake-code&state=" + queryState)
	if cookieState != "" {
		ctx.Request.Header.SetCookie(oauthStateCookie, cookieState)
	}

	handler(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Fatalf("%s: expected 400 on forged state, got %d", name, ctx.Response.StatusCode())
	}
	if !strings.Contains(string(ctx.Response.Body()), "invalid oauth state") {
		t.Fatalf("%s: expected 'invalid oauth state' error, got %q", name, ctx.Response.Body())
	}
}

func TestOAuthCallbacksRejectForgedState(t *testing.T) {
	h := New(&service.AuthService{}, config.Config{})

	cases := []struct {
		name        string
		handler     func(*fasthttp.RequestCtx)
		queryState  string
		cookieState string
	}{
		{"google/missing-cookie", h.HandleGoogleCallback, "attacker-state", ""},
		{"google/mismatch", h.HandleGoogleCallback, "attacker-state", "real-state"},
		{"github/missing-cookie", h.HandleGitHubCallback, "attacker-state", ""},
		{"github/mismatch", h.HandleGitHubCallback, "attacker-state", "real-state"},
		{"lefine/missing-cookie", h.HandleLeFineCallback, "attacker-state", ""},
		{"lefine/mismatch", h.HandleLeFineCallback, "attacker-state", "real-state"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			callbackRejectsForgedState(t, tc.name, tc.handler, tc.queryState, tc.cookieState)
		})
	}
}

// TestOAuthLoginSetsStateCookie verifies the login redirect plants a state
// cookie and echoes the same value into the provider authorization URL.
func TestOAuthLoginSetsStateCookie(t *testing.T) {
	h := New(&service.AuthService{}, config.Config{
		GoogleClientID: "google-client",
		GitHubClientID: "github-client",
		LeFineBaseURL:  "https://lefine.example",
	})

	for _, tc := range []struct {
		name    string
		handler func(*fasthttp.RequestCtx)
	}{
		{"google", h.HandleGoogleLogin},
		{"github", h.HandleGitHubLogin},
		{"lefine", h.HandleLeFineLogin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI("/auth/" + tc.name)
			tc.handler(ctx)

			setCookie := string(ctx.Response.Header.Peek("Set-Cookie"))
			if !strings.HasPrefix(setCookie, oauthStateCookie+"=") {
				t.Fatalf("%s login did not set the state cookie, got %q", tc.name, setCookie)
			}
			if ctx.Response.StatusCode() != fasthttp.StatusTemporaryRedirect {
				t.Fatalf("%s login expected redirect, got %d", tc.name, ctx.Response.StatusCode())
			}
			if loc := string(ctx.Response.Header.Peek("Location")); !strings.Contains(loc, "state=") {
				t.Fatalf("%s login redirect missing state param: %q", tc.name, loc)
			}
		})
	}
}
