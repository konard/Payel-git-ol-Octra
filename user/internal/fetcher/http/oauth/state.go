package oauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"

	"github.com/gin-gonic/gin"
)

// oauthStateCookie is the cookie used to carry the anti-CSRF state across the
// OAuth redirect round-trip.
const oauthStateCookie = "oauth_state"

// generateState returns a cryptographically random, URL-safe state string.
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
func setStateCookie(c *gin.Context, state string) {
	c.SetSameSite(2) // Lax
	// 10 minutes is plenty to complete the provider round-trip.
	c.SetCookie(oauthStateCookie, state, 600, "/", "", false, true)
}

// verifyState compares the state returned by the provider with the value stored
// in the cookie and clears the cookie afterwards. It returns false when the
// values are missing or do not match.
func verifyState(c *gin.Context, queryState string) bool {
	cookieState, err := c.Cookie(oauthStateCookie)
	// Always clear the one-time cookie.
	c.SetSameSite(2)
	c.SetCookie(oauthStateCookie, "", -1, "/", "", false, true)

	if err != nil || cookieState == "" || queryState == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieState), []byte(queryState)) == 1
}
