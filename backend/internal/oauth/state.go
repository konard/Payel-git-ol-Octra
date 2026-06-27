package oauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"time"

	"github.com/valyala/fasthttp"
)

// oauthStateCookie is the cookie used to carry the anti-CSRF state across the
// OAuth redirect round-trip.
const oauthStateCookie = "oauth_state"

// oauthStateTTL is how long the state cookie remains valid — plenty of time to
// complete the provider round-trip, short enough to limit replay.
const oauthStateTTL = 10 * time.Minute

// generateState returns a cryptographically random, URL-safe state string used
// to defend the OAuth round-trip against CSRF (RFC 6749 §10.12).
func generateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// rand.Read only fails in catastrophic situations; fall back to a
		// constant so the flow degrades rather than panics.
		return "octra-oauth-state"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// setStateCookie issues a short-lived, http-only cookie holding the state value.
func setStateCookie(ctx *fasthttp.RequestCtx, state string) {
	c := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(c)
	c.SetKey(oauthStateCookie)
	c.SetValue(state)
	c.SetPath("/")
	c.SetHTTPOnly(true)
	c.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	c.SetExpire(time.Now().Add(oauthStateTTL))
	ctx.Response.Header.SetCookie(c)
}

// clearStateCookie expires the one-time state cookie.
func clearStateCookie(ctx *fasthttp.RequestCtx) {
	c := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(c)
	c.SetKey(oauthStateCookie)
	c.SetValue("")
	c.SetPath("/")
	c.SetHTTPOnly(true)
	c.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	c.SetExpire(fasthttp.CookieExpireDelete)
	ctx.Response.Header.SetCookie(c)
}

// verifyState compares the state returned by the provider with the value stored
// in the cookie and always clears the cookie afterwards. It returns false when
// either value is missing or they do not match (constant-time comparison).
func verifyState(ctx *fasthttp.RequestCtx, queryState string) bool {
	cookieState := string(ctx.Request.Header.Cookie(oauthStateCookie))
	// Always clear the one-time cookie, success or failure.
	clearStateCookie(ctx)

	if cookieState == "" || queryState == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieState), []byte(queryState)) == 1
}
