package oauth

import "testing"

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
