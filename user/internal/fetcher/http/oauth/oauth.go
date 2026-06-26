package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"user/internal/core/services"
)

func frontendAppRedirectURL(frontendURL, accessToken, refreshToken string) string {
	appURL := strings.TrimRight(frontendURL, "/") + "/app"
	values := url.Values{}
	values.Set("token", accessToken)
	if refreshToken != "" {
		values.Set("refresh_token", refreshToken)
	}
	return appURL + "?" + values.Encode()
}

func GetGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func GetGithubOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
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

func fetchGithubUserProfile(client *http.Client) (*githubUserProfile, error) {
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

func fetchGithubEmails(client *http.Client) ([]githubEmail, error) {
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

func selectGithubEmail(profile githubUserProfile, emails []githubEmail) (string, error) {
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

func RegisterGoogleRoutes(r *gin.Engine) {
	r.GET("/auth/google", func(c *gin.Context) {
		config := GetGoogleOauthConfig()
		url := config.AuthCodeURL("random-state")
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	r.GET("/auth/google/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(400, gin.H{"status": "error", "error": "Code not found"})
			return
		}

		config := GetGoogleOauthConfig()
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to exchange token: " + err.Error()})
			return
		}

		client := config.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to get user info"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var googleUser struct {
			ID            string `json:"id"`
			Email         string `json:"email"`
			Name          string `json:"name"`
			Picture       string `json:"picture"`
			VerifiedEmail bool   `json:"verified_email"`
		}
		if err := json.Unmarshal(body, &googleUser); err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to parse user info"})
			return
		}

		if !googleUser.VerifiedEmail {
			c.JSON(400, gin.H{"status": "error", "error": "Google account email not verified"})
			return
		}

		user, err := services.GetOrCreateUserFromGoogle(googleUser.Email, googleUser.Name)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}

		accessToken, refreshToken, err := services.GenerateTokens(*user)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to generate tokens"})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", refreshToken, 604800, "/", "", false, true)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://octra.env.pm"
		}
		c.Redirect(http.StatusTemporaryRedirect, frontendAppRedirectURL(frontendURL, accessToken, refreshToken))
	})
}

func RegisterGithubRoutes(r *gin.Engine) {
	r.GET("/auth/github", func(c *gin.Context) {
		config := GetGithubOauthConfig()
		url := config.AuthCodeURL("random-state")
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	r.GET("/auth/github/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(400, gin.H{"status": "error", "error": "Code not found"})
			return
		}

		config := GetGithubOauthConfig()
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to exchange token: " + err.Error()})
			return
		}

		client := config.Client(context.Background(), token)
		githubUser, err := fetchGithubUserProfile(client)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}

		emails, emailErr := fetchGithubEmails(client)
		if emailErr != nil && githubUser.Email == "" {
			c.JSON(500, gin.H{"status": "error", "error": emailErr.Error()})
			return
		}
		email, err := selectGithubEmail(*githubUser, emails)
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": err.Error()})
			return
		}

		user, err := services.GetOrCreateUserFromGithub(email, githubDisplayName(*githubUser))
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}

		accessToken, refreshToken, err := services.GenerateTokens(*user)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to generate tokens"})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", refreshToken, 604800, "/", "", false, true)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://octra.env.pm"
		}
		c.Redirect(http.StatusTemporaryRedirect, frontendAppRedirectURL(frontendURL, accessToken, refreshToken))
	})
}
