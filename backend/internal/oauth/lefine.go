package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

// LeFineClaims is the payload we receive in the signed "code" from Kefine.
type LeFineClaims struct {
	UserID      string `json:"sub"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Aud         string `json:"aud"` // should be "octra"
	Exp         int64  `json:"exp"`
	Iat         int64  `json:"iat"`
}

// HandleLeFineLogin redirects the user to LeFine's OAuth authorization page.
func (h *Handler) HandleLeFineLogin(ctx *fasthttp.RequestCtx) {
	kefineBase := h.cfg.LeFineBaseURL

	redirectURI := h.cfg.LeFineRedirectURL
	if redirectURI == "" {
		scheme := "https"
		if !ctx.IsTLS() || h.cfg.HTTPAddr == ":8080" {
			scheme = "http"
		}
		redirectURI = fmt.Sprintf("%s://%s/auth/lefine/callback", scheme, string(ctx.Host()))
	}

	state := generateState()
	setStateCookie(ctx, state)

	authURL := fmt.Sprintf("%s/oauth/authorize?client_id=octra&redirect_uri=%s&state=%s&response_type=code",
		kefineBase,
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	ctx.Redirect(authURL, fasthttp.StatusTemporaryRedirect)
}

// HandleLeFineCallback processes the callback from LeFine.
func (h *Handler) HandleLeFineCallback(ctx *fasthttp.RequestCtx) {
	if !verifyState(ctx, string(ctx.QueryArgs().Peek("state"))) {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid oauth state")
		return
	}

	code := string(ctx.QueryArgs().Peek("code"))
	if code == "" {
		writeError(ctx, fasthttp.StatusBadRequest, "no authorization code from LeFine")
		return
	}

	claims, err := h.verifyLeFineCode(code)
	if err != nil {
		writeError(ctx, fasthttp.StatusUnauthorized, "failed to verify LeFine code: "+err.Error())
		return
	}

	name := claims.DisplayName
	if name == "" {
		name = claims.Username
	}
	if name == "" {
		name = claims.Email
	}

	user, err := h.auth.GetOrCreateUserFromLeFine(context.Background(), claims.Email, name)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	result, err := h.auth.LoginResultFromUser(context.Background(), user)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "failed to generate tokens")
		return
	}

	redirectURL := frontendAppRedirectURL(h.cfg.FrontendURL, result.AccessToken, result.RefreshToken)
	ctx.Redirect(redirectURL, fasthttp.StatusTemporaryRedirect)
}

// signLeFineCode creates a compact signed code (payload.sig) — used by LeFine/Kefine.
func (h *Handler) signLeFineCode(claims LeFineClaims) (string, error) {
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha256.New, []byte(h.cfg.LeFineIntegrationSecret))
	mac.Write([]byte(payloadB64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payloadB64 + "." + sig, nil
}

// verifyLeFineCode parses and validates a code from Kefine.
func (h *Handler) verifyLeFineCode(code string) (*LeFineClaims, error) {
	parts := splitTwo(code, '.')
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid code format")
	}

	payloadB64 := parts[0]
	sigB64 := parts[1]

	mac := hmac.New(sha256.New, []byte(h.cfg.LeFineIntegrationSecret))
	mac.Write([]byte(payloadB64))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSig), []byte(sigB64)) {
		return nil, fmt.Errorf("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("bad payload encoding")
	}

	var claims LeFineClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("bad payload json")
	}

	if claims.Aud != "octra" {
		return nil, fmt.Errorf("wrong audience")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("code expired")
	}

	return &claims, nil
}

func splitTwo(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

var _ = (&Handler{}).signLeFineCode // ensure sign is reachable
