package services

import (
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"user/pkg/models"
	"user/pkg/requests"
)

func TestNormalizeEmail(t *testing.T) {
	cases := map[string]string{
		"  User@Example.COM ": "user@example.com",
		"plain@mail.io":       "plain@mail.io",
		"\tMixed@Case.Org\n":  "mixed@case.org",
		"":                    "",
	}
	for in, want := range cases {
		if got := NormalizeEmail(in); got != want {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"postgres unique": {errors.New("duplicate key value violates unique constraint"), true},
		"sqlite unique":   {errors.New("UNIQUE constraint failed: users.email"), true},
		"mixed case":      {errors.New("Duplicate entry"), true},
		"unrelated":       {errors.New("connection refused"), false},
		"nil":             {nil, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := isDuplicateKeyError(tc.err); got != tc.want {
				t.Errorf("isDuplicateKeyError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func newTestUser() models.UserRegister {
	return models.UserRegister{
		ID:       uuid.New(),
		Username: "tester",
		Email:    "tester@example.com",
	}
}

func TestGenerateAndValidateTokens(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	user := newTestUser()
	access, refresh, err := GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens returned error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected non-empty tokens")
	}

	accessTok, err := ValidateAccessToken(access)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}
	claims := accessTok.Claims.(jwt.MapClaims)
	if claims["type"] != "access" {
		t.Errorf("expected access token type, got %v", claims["type"])
	}
	if claims["user_id"] != user.ID.String() {
		t.Errorf("expected user_id %q, got %v", user.ID.String(), claims["user_id"])
	}

	refreshTok, err := ValidateRefreshToken(refresh)
	if err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}
	if refreshTok.Claims.(jwt.MapClaims)["type"] != "refresh" {
		t.Errorf("expected refresh token type, got %v", refreshTok.Claims.(jwt.MapClaims)["type"])
	}
}

// Access and refresh tokens are signed with different secrets and must not be
// interchangeable.
func TestTokensAreNotCrossValid(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	access, refresh, err := GenerateTokens(newTestUser())
	if err != nil {
		t.Fatalf("GenerateTokens returned error: %v", err)
	}

	if _, err := ValidateAccessToken(refresh); err == nil {
		t.Error("refresh token must not validate as an access token")
	}
	if _, err := ValidateRefreshToken(access); err == nil {
		t.Error("access token must not validate as a refresh token")
	}
}

func TestGenerateTokensRequiresSecrets(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")
	if _, _, err := GenerateTokens(newTestUser()); err == nil {
		t.Error("expected error when JWT secrets are unset")
	}
}

func TestRefreshTokensRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	if _, err := RefreshTokens(requests.RefreshTokenRequest{RefreshToken: "not-a-token"}); err == nil {
		t.Error("expected error for invalid refresh token")
	}
}
