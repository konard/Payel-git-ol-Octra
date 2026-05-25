package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"user/internal/core/services"
)

// LeFineClaims is the payload we receive in the signed "code" from Kefine
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

// getLeFineSecret returns the shared secret for signing/verifying LeFine integration codes
func getLeFineSecret() []byte {
	secret := os.Getenv("LEFINE_INTEGRATION_SECRET")
	if secret == "" {
		secret = "dev-only-lefine-octra-shared-secret-change-me"
	}
	return []byte(secret)
}

// signLeFineCode creates a compact signed code (payload.sig)
func signLeFineCode(claims LeFineClaims) (string, error) {
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha256.New, getLeFineSecret())
	mac.Write([]byte(payloadB64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payloadB64 + "." + sig, nil
}

// verifyLeFineCode parses and validates a code from Kefine
func verifyLeFineCode(code string) (*LeFineClaims, error) {
	parts := splitTwo(code, '.')
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid code format")
	}

	payloadB64 := parts[0]
	sigB64 := parts[1]

	// Verify signature
	mac := hmac.New(sha256.New, getLeFineSecret())
	mac.Write([]byte(payloadB64))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSig), []byte(sigB64)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("bad payload encoding")
	}

	var claims LeFineClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("bad payload json")
	}

	// Check audience
	if claims.Aud != "octra" {
		return nil, fmt.Errorf("wrong audience")
	}

	// Check expiry
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

// RegisterLeFineRoutes registers the custom "Login with LeFine" flow (OAuth-like but simplified for internal use)
func RegisterLeFineRoutes(r *gin.Engine) {
	r.GET("/auth/lefine", func(c *gin.Context) {
		kefineBase := os.Getenv("LEFINE_BASE_URL")
		if kefineBase == "" {
			kefineBase = "https://lefine.pro" // production default (use docker-compose override for local dev)
		}

		redirectURI := os.Getenv("LEFINE_REDIRECT_URL")
		if redirectURI == "" {
			// Default to current host + path (works for most deploys)
			scheme := "https"
			if c.Request.TLS == nil && os.Getenv("ENV") != "prod" {
				scheme = "http"
			}
			redirectURI = fmt.Sprintf("%s://%s/auth/lefine/callback", scheme, c.Request.Host)
		}

		state := fmt.Sprintf("%d", time.Now().UnixNano()) // simple stateless state for now

		authURL := fmt.Sprintf("%s/oauth/authorize?client_id=octra&redirect_uri=%s&state=%s&response_type=code",
			kefineBase,
			url.QueryEscape(redirectURI),
			state,
		)

		c.Redirect(http.StatusTemporaryRedirect, authURL)
	})

	r.GET("/auth/lefine/callback", func(c *gin.Context) {
		code := c.Query("code")
		// state := c.Query("state") // we can verify state if we persist it later

		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "No authorization code from LeFine"})
			return
		}

		claims, err := verifyLeFineCode(code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "Failed to verify LeFine code: " + err.Error()})
			return
		}

		// Create or get local Octra user from LeFine identity
		name := claims.DisplayName
		if name == "" {
			name = claims.Username
		}
		if name == "" {
			name = claims.Email
		}

		user, err := services.GetOrCreateUserFromLeFine(claims.Email, name, claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}

		accessToken, refreshToken, err := services.GenerateTokens(*user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to generate tokens"})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", refreshToken, 604800, "/", "", false, true)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://octra.env.pm"
		}
		redirectURL := fmt.Sprintf("%s/app?token=%s", frontendURL, accessToken)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	})
}
