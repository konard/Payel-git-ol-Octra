package oauth

import (
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
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

// ctxWithStateCookie builds a request carrying the given cookie value (when
// non-empty) and the given ?state= query argument.
func ctxWithStateCookie(queryState, cookieValue string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("state", queryState)
	ctx.Request.SetRequestURI("/callback?state=" + queryState)
	if cookieValue != "" {
		ctx.Request.Header.SetCookie(oauthStateCookie, cookieValue)
	}
	return ctx
}

func TestVerifyState(t *testing.T) {
	if !verifyState(ctxWithStateCookie("abc123", "abc123"), "abc123") {
		t.Error("matching state and cookie should verify")
	}
	if verifyState(ctxWithStateCookie("abc123", "different"), "abc123") {
		t.Error("mismatched cookie should fail")
	}
	if verifyState(ctxWithStateCookie("abc123", ""), "abc123") {
		t.Error("missing cookie should fail")
	}
	if verifyState(ctxWithStateCookie("", "abc123"), "") {
		t.Error("empty query state should fail")
	}
}

// TestVerifyStateClearsCookie ensures the one-time state cookie is always
// expired after a verification attempt, regardless of outcome.
func TestVerifyStateClearsCookie(t *testing.T) {
	ctx := ctxWithStateCookie("abc123", "abc123")
	verifyState(ctx, "abc123")

	setCookie := string(ctx.Response.Header.Peek("Set-Cookie"))
	// The one-time cookie is reset to an empty value...
	if !strings.HasPrefix(setCookie, oauthStateCookie+"=;") {
		t.Fatalf("expected the state cookie to be reset to empty, got %q", setCookie)
	}
	// ...and expired (fasthttp's delete marker uses a fixed past expiry).
	if !strings.Contains(strings.ToLower(setCookie), "expires=") {
		t.Errorf("expected the state cookie to carry an expiry, got %q", setCookie)
	}
}

func TestSetStateCookieIsHTTPOnly(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	setStateCookie(ctx, "some-state")
	setCookie := strings.ToLower(string(ctx.Response.Header.Peek("Set-Cookie")))
	if !strings.Contains(setCookie, "httponly") {
		t.Errorf("expected state cookie to be HttpOnly, got %q", setCookie)
	}
	if !strings.Contains(setCookie, "samesite=lax") {
		t.Errorf("expected state cookie to be SameSite=Lax, got %q", setCookie)
	}
}
