// Package api exposes the public HTTP API of the monolith over fasthttp.
package api

import (
	"context"
	"encoding/json"
	"errors"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// authHeader is the header carrying the per-user API token.
const authHeader = "octra-api-token"

// ctxUserKey is the fasthttp UserValue key for the authenticated user.
const ctxUserKey = "octra_user"

// API bundles the services exposed over HTTP.
type API struct {
	auth *service.AuthService
	env  *service.EnvironmentService
	chat *service.ChatService
}

// New builds an API.
func New(auth *service.AuthService, env *service.EnvironmentService, chat *service.ChatService) *API {
	return &API{auth: auth, env: env, chat: chat}
}

// Router wires routes to handlers.
func (a *API) Router() *router.Router {
	r := router.New()
	r.GET("/health", a.handleHealth)
	r.POST("/register", a.handleRegister)
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

// --- handlers ---------------------------------------------------------------

func (a *API) handleHealth(ctx *fasthttp.RequestCtx) {
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "ok"})
}

type registerRequest struct {
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

	user, err := a.auth.Register(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrEmailTaken) {
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
