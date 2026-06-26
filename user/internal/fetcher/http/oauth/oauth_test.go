package oauth

import "testing"

func TestSelectGithubEmailPrefersVerifiedPrimaryEmail(t *testing.T) {
	profile := githubUserProfile{Email: "public@example.com", Login: "octra-user"}
	emails := []githubEmail{
		{Email: "secondary@example.com", Primary: false, Verified: true},
		{Email: "primary@example.com", Primary: true, Verified: true},
	}

	email, err := selectGithubEmail(profile, emails)
	if err != nil {
		t.Fatalf("selectGithubEmail returned error: %v", err)
	}
	if email != "primary@example.com" {
		t.Fatalf("expected verified primary email, got %q", email)
	}
}

func TestSelectGithubEmailUsesVerifiedEmailWhenProfileEmailIsHidden(t *testing.T) {
	profile := githubUserProfile{Login: "private-email-user"}
	emails := []githubEmail{
		{Email: "hidden-primary@example.com", Primary: true, Verified: true},
	}

	email, err := selectGithubEmail(profile, emails)
	if err != nil {
		t.Fatalf("selectGithubEmail returned error: %v", err)
	}
	if email != "hidden-primary@example.com" {
		t.Fatalf("expected hidden verified primary email, got %q", email)
	}
}

func TestSelectGithubEmailRejectsUnverifiedSyntheticFallback(t *testing.T) {
	profile := githubUserProfile{Login: "private-email-user"}
	emails := []githubEmail{
		{Email: "unverified@example.com", Primary: true, Verified: false},
	}

	email, err := selectGithubEmail(profile, emails)
	if err == nil {
		t.Fatalf("expected missing verified email error, got email %q", email)
	}
}
