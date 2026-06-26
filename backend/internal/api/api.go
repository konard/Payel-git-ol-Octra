// Package api exposes the public HTTP API of the monolith over fasthttp.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"backend/internal/model"
	"backend/internal/oauth"
	"backend/internal/service"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// authHeader is the header carrying the per-user API token.
const authHeader = "octra-api-token"

const authHeaderBearer = "Authorization"

// ctxUserKey is the fasthttp UserValue key for the authenticated user.
const ctxUserKey = "octra_user"

// API bundles the services exposed over HTTP.
type API struct {
	auth   *service.AuthService
	env    *service.EnvironmentService
	chat   *service.ChatService
	oauthH *oauth.Handler
}

// New builds an API.
func New(auth *service.AuthService, env *service.EnvironmentService, chat *service.ChatService, oauthH *oauth.Handler) *API {
	return &API{auth: auth, env: env, chat: chat, oauthH: oauthH}
}

// Router wires routes to handlers.
func (a *API) Router() *router.Router {
	r := router.New()
	r.GET("/health", a.handleHealth)

	// Auth
	r.POST("/register", a.handleRegister)
	r.POST("/login", a.handleLogin)
	r.POST("/logout", a.handleLogout)
	r.GET("/me", a.handleMe)
	r.POST("/refresh", a.handleRefresh)

	// OAuth
	r.GET("/auth/google", a.oauthH.HandleGoogleLogin)
	r.GET("/auth/google/callback", a.oauthH.HandleGoogleCallback)
	r.GET("/auth/github", a.oauthH.HandleGitHubLogin)
	r.GET("/auth/github/callback", a.oauthH.HandleGitHubCallback)
	r.GET("/auth/lefine", a.oauthH.HandleLeFineLogin)
	r.GET("/auth/lefine/callback", a.oauthH.HandleLeFineCallback)

	// API key auth (MCP clients)
	r.POST("/environment", a.withAuth(a.handleEnvironment))
	r.POST("/api/chat", a.withAuth(a.handleChat))
	return r
}

// --- middleware -------------------------------------------------------------

// withAuth authenticates the request via the octra-api-token header and stores
// the user on the request context.
func (a *API) withAuth(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		token := string(ctx.Request.Header.Peek(authHeader))
		user, err := a.auth.Authenticate(ctx, token)
		if err != nil {
			writeError(ctx, fasthttp.StatusUnauthorized, "invalid or missing api token")
			return
		}
		ctx.SetUserValue(ctxUserKey, user)
		next(ctx)
	}
}

func userFrom(ctx *fasthttp.RequestCtx) *model.User {
	u, _ := ctx.UserValue(ctxUserKey).(*model.User)
	return u
}

// extractBearerToken extracts a JWT from the Authorization header.
func extractBearerToken(ctx *fasthttp.RequestCtx) string {
	token := string(ctx.Request.Header.Peek(authHeaderBearer))
	if token == "" {
		return ""
	}
	if len(token) > 7 && strings.ToUpper(token[:7]) == "BEARER " {
		return token[7:]
	}
	return token
}

// --- handlers ---------------------------------------------------------------

func (a *API) handleHealth(ctx *fasthttp.RequestCtx) {
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "ok"})
}

// --- Register ---------------------------------------------------------------

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID string `json:"user_id"`
	APIKey string `json:"api_key"`
}

func (a *API) handleRegister(ctx *fasthttp.RequestCtx) {
	var req registerRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	user, err := a.auth.Register(ctx, req.Username, req.Email, req.Password)
	if errors.Is(err, service.ErrEmailTaken) || errors.Is(err, service.ErrUsernameTaken) {
		writeError(ctx, fasthttp.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusCreated, registerResponse{
		UserID: user.ID.String(),
		APIKey: user.APIKey,
	})
}

// --- Login ------------------------------------------------------------------

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) handleLogin(ctx *fasthttp.RequestCtx) {
	var req loginRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	result, err := a.auth.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		writeError(ctx, fasthttp.StatusUnauthorized, "invalid email or password")
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   result,
	})
}

// --- Logout -----------------------------------------------------------------

func (a *API) handleLogout(ctx *fasthttp.RequestCtx) {
	writeJSON(ctx, fasthttp.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "logged out successfully",
	})
}

// --- Me ---------------------------------------------------------------------

func (a *API) handleMe(ctx *fasthttp.RequestCtx) {
	token := extractBearerToken(ctx)
	if token == "" {
		writeError(ctx, fasthttp.StatusUnauthorized, "authorization token is required")
		return
	}

	user, err := a.auth.GetMe(ctx, token)
	if err != nil {
		writeError(ctx, fasthttp.StatusUnauthorized, "invalid or expired token")
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   user,
	})
}

// --- Refresh Tokens ---------------------------------------------------------

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *API) handleRefresh(ctx *fasthttp.RequestCtx) {
	var req refreshRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	result, err := a.auth.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		writeError(ctx, fasthttp.StatusUnauthorized, "invalid refresh token")
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   result,
	})
}

// --- Environment (unchanged) ------------------------------------------------

type environmentRequest struct {
	LLM struct {
		APIKey  string `json:"api_key"`
		BaseURL string `json:"base_url"`
		Model   string `json:"model"`
	} `json:"llm"`
	Agent struct {
		CLI string `json:"cli"`
	} `json:"agent"`
	Skills []string `json:"skills"`
}

type environmentResponse struct {
	UserID  string `json:"user_id"`
	AgentID string `json:"agent_id"`
	APIKey  string `json:"api_key"`
}

func (a *API) handleEnvironment(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	var req environmentRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	agent, err := a.env.Create(ctx, user, service.EnvironmentInput{
		LLMAPIKey:  req.LLM.APIKey,
		LLMBaseURL: req.LLM.BaseURL,
		LLMModel:   req.LLM.Model,
		CLI:        model.CLIType(req.Agent.CLI),
		Skills:     req.Skills,
	})
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, environmentResponse{
		UserID:  user.ID.String(),
		AgentID: agent.ID.String(),
		APIKey:  user.APIKey,
	})
}

// --- Chat (unchanged) -------------------------------------------------------

type chatRequest struct {
	Prompt string   `json:"prompt"`
	Skills []string `json:"skills"`
}

type chatResponse struct {
	Response string `json:"response"`
}

func (a *API) handleChat(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	var req chatRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}
	if req.Prompt == "" {
		writeError(ctx, fasthttp.StatusBadRequest, "prompt is required")
		return
	}

	resp, err := a.chat.Chat(ctx, user, req.Prompt, req.Skills)
	if errors.Is(err, service.ErrNoEnvironment) || errors.Is(err, service.ErrEnvironmentInactive) {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(ctx, fasthttp.StatusBadGateway, err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, chatResponse{Response: resp})
}

// --- helpers ----------------------------------------------------------------

func writeJSON(ctx *fasthttp.RequestCtx, status int, body any) {
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(status)
	if err := json.NewEncoder(ctx).Encode(body); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}

func writeError(ctx *fasthttp.RequestCtx, status int, msg string) {
	writeJSON(ctx, status, map[string]string{"error": msg})
}

// ensure fasthttp ctx satisfies context.Context for service calls.
var _ context.Context = (*fasthttp.RequestCtx)(nil)
