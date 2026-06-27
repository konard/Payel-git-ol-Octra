package oauth

import (
	"strings"
	"testing"
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

func TestSelectGitHubEmailFallbackToProfileEmail(t *testing.T) {
	profile := githubUserProfile{Email: "public@example.com", Login: "user"}
	var emails []githubEmail

	email, err := selectGitHubEmail(profile, emails)
	if err != nil {
		t.Fatalf("selectGitHubEmail returned error: %v", err)
	}
	if email != "public@example.com" {
		t.Fatalf("expected profile email fallback, got %q", email)
	}
}

func TestSelectGitHubEmailFallbackToAnyVerified(t *testing.T) {
	profile := githubUserProfile{Login: "user"}
	emails := []githubEmail{
		{Email: "any-verified@example.com", Primary: false, Verified: true},
	}

	email, err := selectGitHubEmail(profile, emails)
	if err != nil {
		t.Fatalf("selectGitHubEmail returned error: %v", err)
	}
	if email != "any-verified@example.com" {
		t.Fatalf("expected any verified email, got %q", email)
	}
}

func TestSelectGitHubEmailErrorWhenNoEmails(t *testing.T) {
	profile := githubUserProfile{Login: "user"}
	var emails []githubEmail

	_, err := selectGitHubEmail(profile, emails)
	if err == nil {
		t.Fatal("expected error when no emails and no profile email")
	}
}

func TestGithubDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		profile githubUserProfile
		want    string
	}{
		{"uses login", githubUserProfile{Login: "octra-user", Name: "Real Name", ID: 123}, "octra-user"},
		{"uses name when no login", githubUserProfile{Login: "", Name: "Real Name", ID: 123}, "Real Name"},
		{"uses id when no login or name", githubUserProfile{Login: "", Name: "", ID: 456}, "github_456"},
		{"default when empty", githubUserProfile{}, "github_user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := githubDisplayName(tt.profile)
			if got != tt.want {
				t.Fatalf("githubDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFrontendAppRedirectURL(t *testing.T) {
	tests := []struct {
		name         string
		frontendURL  string
		accessToken  string
		refreshToken string
		wantContains string
	}{
		{
			name:         "with refresh token",
			frontendURL:  "http://localhost:5173",
			accessToken:  "abc123",
			refreshToken: "ref456",
			wantContains: "refresh_token=ref456&token=abc123",
		},
		{
			name:         "without refresh token",
			frontendURL:  "http://localhost:5173",
			accessToken:  "abc123",
			refreshToken: "",
			wantContains: "token=abc123",
		},
		{
			name:         "trailing slash stripped",
			frontendURL:  "http://localhost:5173/",
			accessToken:  "tok",
			refreshToken: "",
			wantContains: "http://localhost:5173/app?token=tok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontendAppRedirectURL(tt.frontendURL, tt.accessToken, tt.refreshToken)
			if got != tt.wantContains && !contains(got, tt.wantContains) {
				t.Fatalf("frontendAppRedirectURL() = %q, should contain %q", got, tt.wantContains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
