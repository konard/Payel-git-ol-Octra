// Package oauth implements OAuth 2.0 login flows for Google, GitHub and LeFine.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"backend/internal/config"
	"backend/internal/service"

	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Handler bundles OAuth handlers.
type Handler struct {
	auth *service.AuthService
	cfg  config.Config
}

// New creates an OAuth handler.
func New(auth *service.AuthService, cfg config.Config) *Handler {
	return &Handler{auth: auth, cfg: cfg}
}

func frontendAppRedirectURL(frontendURL, accessToken, refreshToken string) string {
	appURL := strings.TrimRight(frontendURL, "/") + "/app"
	values := url.Values{}
	values.Set("token", accessToken)
	if refreshToken != "" {
		values.Set("refresh_token", refreshToken)
	}
	return appURL + "?" + values.Encode()
}

// --- Google ------------------------------------------------------------------

func (h *Handler) googleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.cfg.GoogleClientID,
		ClientSecret: h.cfg.GoogleClientSecret,
		RedirectURL:  h.cfg.GoogleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func (h *Handler) HandleGoogleLogin(ctx *fasthttp.RequestCtx) {
	config := h.googleConfig()
	url := config.AuthCodeURL("random-state")
	ctx.Redirect(url, fasthttp.StatusTemporaryRedirect)
}

func (h *Handler) HandleGoogleCallback(ctx *fasthttp.RequestCtx) {
	code := string(ctx.QueryArgs().Peek("code"))
	if code == "" {
		writeError(ctx, fasthttp.StatusBadRequest, "code not found")
		return
	}

	config := h.googleConfig()
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "failed to exchange token: "+err.Error())
		return
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "failed to get user info")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &googleUser); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "failed to parse user info")
		return
	}

	if !googleUser.VerifiedEmail {
		writeError(ctx, fasthttp.StatusBadRequest, "google account email not verified")
		return
	}

	user, err := h.auth.GetOrCreateUserFromGoogle(context.Background(), googleUser.Email, googleUser.Name)
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

// --- GitHub ------------------------------------------------------------------

func (h *Handler) githubConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.cfg.GitHubClientID,
		ClientSecret: h.cfg.GitHubClientSecret,
		RedirectURL:  h.cfg.GitHubRedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

type githubUserProfile struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func fetchGitHubUserProfile(client *http.Client) (*githubUserProfile, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub user API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var profile githubUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}
	return &profile, nil
}

func fetchGitHubEmails(client *http.Client) ([]githubEmail, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, fmt.Errorf("failed to get user emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub email API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("failed to parse user emails: %w", err)
	}
	return emails, nil
}

func selectGitHubEmail(profile githubUserProfile, emails []githubEmail) (string, error) {
	for _, email := range emails {
		if email.Primary && email.Verified && strings.TrimSpace(email.Email) != "" {
			return email.Email, nil
		}
	}

	if strings.TrimSpace(profile.Email) != "" {
		return profile.Email, nil
	}

	for _, email := range emails {
		if email.Verified && strings.TrimSpace(email.Email) != "" {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("GitHub account does not expose a verified email")
}

func githubDisplayName(profile githubUserProfile) string {
	if strings.TrimSpace(profile.Login) != "" {
		return profile.Login
	}
	if strings.TrimSpace(profile.Name) != "" {
		return profile.Name
	}
	if profile.ID != 0 {
		return fmt.Sprintf("github_%d", profile.ID)
	}
	return "github_user"
}

func (h *Handler) HandleGitHubLogin(ctx *fasthttp.RequestCtx) {
	config := h.githubConfig()
	url := config.AuthCodeURL("random-state")
	ctx.Redirect(url, fasthttp.StatusTemporaryRedirect)
}

func (h *Handler) HandleGitHubCallback(ctx *fasthttp.RequestCtx) {
	code := string(ctx.QueryArgs().Peek("code"))
	if code == "" {
		writeError(ctx, fasthttp.StatusBadRequest, "code not found")
		return
	}

	config := h.githubConfig()
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "failed to exchange token: "+err.Error())
		return
	}

	client := config.Client(context.Background(), token)
	githubUser, err := fetchGitHubUserProfile(client)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	emails, emailErr := fetchGitHubEmails(client)
	if emailErr != nil && githubUser.Email == "" {
		writeError(ctx, fasthttp.StatusInternalServerError, emailErr.Error())
		return
	}
	email, err := selectGitHubEmail(*githubUser, emails)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	user, err := h.auth.GetOrCreateUserFromGitHub(context.Background(), email, githubDisplayName(*githubUser))
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

// --- helpers ----------------------------------------------------------------

func writeError(ctx *fasthttp.RequestCtx, status int, msg string) {
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(status)
	ctx.SetBodyString(`{"status":"error","error":"` + strings.ReplaceAll(msg, `"`, `\"`) + `"}`)
}
