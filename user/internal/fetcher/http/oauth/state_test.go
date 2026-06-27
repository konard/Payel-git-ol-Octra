package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenerateStateIsRandom(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		s := generateState()
		if s == "" {
			t.Fatal("generateState returned empty string")
		}
		if seen[s] {
			t.Fatalf("generateState produced a duplicate value: %q", s)
		}
		seen[s] = true
	}
}

func newCtx(state, cookieValue string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/callback?state="+state, nil)
	if cookieValue != "" {
		c.Request.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: cookieValue})
	}
	return c
}

func TestVerifyState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	if !verifyState(newCtx("abc123", "abc123"), "abc123") {
		t.Error("matching state and cookie should verify")
	}
	if verifyState(newCtx("abc123", "different"), "abc123") {
		t.Error("mismatched cookie should fail")
	}
	if verifyState(newCtx("abc123", ""), "abc123") {
		t.Error("missing cookie should fail")
	}
	if verifyState(newCtx("", "abc123"), "") {
		t.Error("empty query state should fail")
	}
}
