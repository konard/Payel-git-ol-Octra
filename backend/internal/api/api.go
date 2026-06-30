package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/oauth"
	"backend/internal/repository"
	"backend/internal/service"
	ts "backend/internal/typesense"

	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const authHeader = "octra-api-token"
const authHeaderBearer = "Authorization"
const ctxUserKey = "octra_user"

type WorkflowProxy interface {
	OcawePort(ctx context.Context, userID string) (int, error)
}

type API struct {
	auth    *service.AuthService
	env     *service.EnvironmentService
	chat    *service.ChatService
	billing *service.BillingService
	oauthH  *oauth.Handler

	dashboardEnvRepo repository.DashboardEnvironmentRepository
	canvasNodeRepo   repository.CanvasNodeRepository
	skillsRepo       repository.SkillRepository
	cliRepo          repository.CLIRepository
	providerRepo     repository.ProviderRepository
	typesense        *ts.Client
	workflowProxy    WorkflowProxy
}

func New(
	auth *service.AuthService,
	env *service.EnvironmentService,
	chat *service.ChatService,
	billing *service.BillingService,
	oauthH *oauth.Handler,
	dashboardEnvRepo repository.DashboardEnvironmentRepository,
	canvasNodeRepo repository.CanvasNodeRepository,
	skillsRepo repository.SkillRepository,
	cliRepo repository.CLIRepository,
	providerRepo repository.ProviderRepository,
	typesense *ts.Client,
	workflowProxy ...WorkflowProxy,
) *API {
	api := &API{
		auth:             auth,
		env:              env,
		chat:             chat,
		billing:          billing,
		oauthH:           oauthH,
		dashboardEnvRepo: dashboardEnvRepo,
		canvasNodeRepo:   canvasNodeRepo,
		skillsRepo:       skillsRepo,
		cliRepo:          cliRepo,
		providerRepo:     providerRepo,
		typesense:        typesense,
	}
	if len(workflowProxy) > 0 {
		api.workflowProxy = workflowProxy[0]
	}
	return api
}

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

	// Billing
	r.GET("/billing/balance", a.withAuth(a.handleBillingBalance))
	r.PATCH("/billing/settings", a.withAuth(a.handleBillingSettings))
	r.GET("/billing/transactions", a.withAuth(a.handleBillingTransactions))
	r.POST("/billing/topup", a.withAuth(a.handleBillingTopUp))
	r.POST("/billing/lefine-reward", a.withAuth(a.handleBillingLefineReward))
	r.POST("/billing/usage", a.withAuth(a.handleBillingUsage))

	// User API keys
	r.POST("/api/keys", a.withAuth(a.handleCreateAPIKey))
	r.GET("/api/keys", a.withAuth(a.handleListAPIKeys))
	r.DELETE("/api/keys/{id}", a.withAuth(a.handleDeleteAPIKey))

	// Dashboard environments
	r.POST("/api/environments", a.withAuth(a.handleCreateDashboardEnvironment))
	r.GET("/api/environments", a.withAuth(a.handleListDashboardEnvironments))
	r.POST("/api/environments/patch", a.withAuth(a.handlePatchDashboardEnvironment))
	r.DELETE("/api/environments/{id}", a.withAuth(a.handleDeleteDashboardEnvironment))

	// Workflow canvas per environment
	r.GET("/api/environments/{id}/canvas", a.withAuth(a.handleGetCanvas))
	r.PUT("/api/environments/{id}/canvas", a.withAuth(a.handlePutCanvas))

	// WebSocket for real-time canvas sync (auth via query param)
	r.GET("/ws/canvas/{id}", a.handleWSCanvas)

	// Skills search
	r.GET("/skills/search", a.handleSkillSearch)

	// CLI catalogue
	r.GET("/api/cli", a.handleListCLIs)
	r.GET("/api/cli/search", a.handleCLISearch)

	// Provider and combined catalogue search
	r.GET("/api/providers", a.handleListProviders)
	r.GET("/api/providers/search", a.handleProviderSearch)
	r.GET("/api/catalog/search", a.handleCatalogSearch)

	// Workflow proxy to Ocawe: forwards /v1/workflows/* to the user's Ocawe
	r.ANY("/v1/workflows/{rest:*}", a.withAuth(a.handleWorkflowProxy))
	return r
}

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

func (a *API) handleHealth(ctx *fasthttp.RequestCtx) {
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "ok"})
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID  string `json:"user_id"`
	APIKey  string `json:"api_key"`
	Balance int    `json:"balance"`
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
		UserID:  user.ID.String(),
		APIKey:  user.APIKey,
		Balance: user.Balance,
	})
}

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

func (a *API) handleLogout(ctx *fasthttp.RequestCtx) {
	writeJSON(ctx, fasthttp.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "logged out successfully",
	})
}

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

type environmentRequest struct {
	LLM struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		BaseURL  string `json:"base_url"`
		Model    string `json:"model"`
	} `json:"llm"`
	Agent struct {
		CLI      string `json:"cli"`
		Priority int    `json:"priority"`
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
		LLMProvider: req.LLM.Provider,
		LLMAPIKey:   req.LLM.APIKey,
		LLMBaseURL:  req.LLM.BaseURL,
		LLMModel:    req.LLM.Model,
		CLI:         model.CLIType(req.Agent.CLI),
		Priority:    req.Agent.Priority,
		Skills:      req.Skills,
	})
	if errors.Is(err, service.ErrBalanceNegative) {
		writeError(ctx, fasthttp.StatusPaymentRequired, err.Error())
		return
	}
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

func (a *API) handleWorkflowProxy(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	if a.workflowProxy == nil {
		writeError(ctx, fasthttp.StatusNotFound, "workflow proxy not configured")
		return
	}

	port, err := a.workflowProxy.OcawePort(ctx, user.ID.String())
	if err != nil {
		writeError(ctx, fasthttp.StatusBadGateway, fmt.Sprintf("ocawe unavailable: %v", err))
		return
	}

	path := string(ctx.Path())
	suffix := strings.TrimPrefix(path, "/v1/workflows")
	target := fmt.Sprintf("http://127.0.0.1:%d/v1/workflows%s", port, suffix)
	a.reverseProxy(ctx, target)
}

func (a *API) reverseProxy(ctx *fasthttp.RequestCtx, target string) {
	body := ctx.PostBody()
	req, err := http.NewRequest(string(ctx.Method()), target, bytes.NewReader(body))
	if err != nil {
		writeError(ctx, fasthttp.StatusBadGateway, err.Error())
		return
	}
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
	req.Header.Set("X-Forwarded-For", ctx.RemoteIP().String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadGateway, err.Error())
		return
	}

	ctx.SetStatusCode(resp.StatusCode)
	ctx.SetContentType(resp.Header.Get("Content-Type"))
	ctx.Write(respBody)
}

type billingBalanceResponse struct {
	UserID          string                `json:"user_id"`
	Balance         int                   `json:"balance"`
	MarginMode      model.MarginMode      `json:"margin_mode"`
	SafeMarginLimit int                   `json:"safe_margin_limit"`
	AutoPayInterval model.AutoPayInterval `json:"auto_pay_interval"`
	AutoPayDay      int                   `json:"auto_pay_day"`
}

func (a *API) handleBillingBalance(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	fresh, err := a.billing.GetBalance(ctx, user.ID)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, balanceResponseFromUser(fresh))
}

type billingSettingsRequest struct {
	MarginMode      *model.MarginMode      `json:"margin_mode"`
	SafeMarginLimit *int                   `json:"safe_margin_limit"`
	AutoPayInterval *model.AutoPayInterval `json:"auto_pay_interval"`
	AutoPayDay      *int                   `json:"auto_pay_day"`
}

func (a *API) handleBillingSettings(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	var req billingSettingsRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	fresh, err := a.billing.UpdateSettings(ctx, user.ID, service.BillingSettingsInput{
		MarginMode:      req.MarginMode,
		SafeMarginLimit: req.SafeMarginLimit,
		AutoPayInterval: req.AutoPayInterval,
		AutoPayDay:      req.AutoPayDay,
	})
	if errors.Is(err, service.ErrInvalidBillingSetting) {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, balanceResponseFromUser(fresh))
}

type billingAmountRequest struct {
	Amount  int    `json:"amount"`
	AgentID string `json:"agent_id"`
}

func (a *API) handleBillingTopUp(ctx *fasthttp.RequestCtx) {
	a.handleCredit(ctx, model.TransactionReasonTopUp)
}

func (a *API) handleBillingLefineReward(ctx *fasthttp.RequestCtx) {
	a.handleCredit(ctx, model.TransactionReasonLefineReward)
}

func (a *API) handleCredit(ctx *fasthttp.RequestCtx, reason model.TransactionReason) {
	user := userFrom(ctx)
	var req billingAmountRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}
	agentID, err := parseOptionalUUID(req.AgentID)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid agent_id")
		return
	}
	tx, err := a.billing.Credit(ctx, user.ID, req.Amount, reason, agentID)
	if errors.Is(err, service.ErrInvalidBillingAmount) {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, transactionResponseFromModel(*tx))
}

func (a *API) handleBillingTransactions(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	limit := queryInt(ctx, "limit", 100)
	offset := queryInt(ctx, "offset", 0)
	txs, err := a.billing.ListTransactions(ctx, user.ID, limit, offset)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	out := make([]transactionResponse, 0, len(txs))
	for _, tx := range txs {
		out = append(out, transactionResponseFromModel(tx))
	}
	writeJSON(ctx, fasthttp.StatusOK, out)
}

type billingUsageRequest struct {
	Date            *time.Time `json:"date"`
	CPUSeconds      int64      `json:"cpu_seconds"`
	MemoryMBHours   int64      `json:"memory_mb_hours"`
	DiskMB          int64      `json:"disk_mb"`
	LoadPercent     int        `json:"load_percent"`
	StandardPayment int        `json:"standard_payment"`
	AgentID         string     `json:"agent_id"`
}

type usageResponse struct {
	Usage       usageMetricResponse  `json:"usage"`
	Transaction *transactionResponse `json:"transaction,omitempty"`
}

type usageMetricResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Date          time.Time `json:"date"`
	CPUSeconds    int64     `json:"cpu_seconds"`
	MemoryMBHours int64     `json:"memory_mb_hours"`
	DiskMB        int64     `json:"disk_mb"`
	LoadPercent   int       `json:"load_percent"`
}

func (a *API) handleBillingUsage(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	var req billingUsageRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}
	agentID, err := parseOptionalUUID(req.AgentID)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid agent_id")
		return
	}
	var date time.Time
	if req.Date != nil {
		date = *req.Date
	}
	metric, tx, err := a.billing.RecordUsage(ctx, user.ID, service.UsageInput{
		Date:           date,
		CPUSeconds:     req.CPUSeconds,
		MemoryMBHours:  req.MemoryMBHours,
		DiskMB:         req.DiskMB,
		LoadPercent:    req.LoadPercent,
		StandardCharge: req.StandardPayment,
		AgentID:        agentID,
	})
	if errors.Is(err, service.ErrInvalidBillingAmount) {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrSafeMarginLimit) {
		writeError(ctx, fasthttp.StatusPaymentRequired, err.Error())
		return
	}
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	resp := usageResponse{Usage: usageMetricResponseFromModel(*metric)}
	if tx != nil {
		txResp := transactionResponseFromModel(*tx)
		resp.Transaction = &txResp
	}
	writeJSON(ctx, fasthttp.StatusOK, resp)
}

type transactionResponse struct {
	ID           string                  `json:"id"`
	UserID       string                  `json:"user_id"`
	Type         model.TransactionType   `json:"type"`
	Amount       int                     `json:"amount"`
	Reason       model.TransactionReason `json:"reason"`
	AgentID      string                  `json:"agent_id,omitempty"`
	BalanceAfter int                     `json:"balance_after"`
	CreatedAt    time.Time               `json:"created_at"`
}

type createAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type createAPIKeyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (a *API) handleCreateAPIKey(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	var req createAPIKeyRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	result, err := a.auth.CreateAPIKey(ctx, user.ID, service.CreateAPIKeyInput{
		Name:      req.Name,
		ExpiresAt: req.ExpiresAt,
	})
	if errors.Is(err, service.ErrAPIKeyNameEmpty) {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusCreated, createAPIKeyResponse{
		ID:        result.ID,
		Name:      result.Name,
		Key:       result.Key,
		ExpiresAt: result.ExpiresAt,
		CreatedAt: result.CreatedAt,
	})
}

func (a *API) handleListAPIKeys(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	keys, err := a.auth.ListUserAPIKeys(ctx, user.ID)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, keys)
}

func (a *API) handleDeleteAPIKey(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	idStr := ctx.UserValue("id").(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid key id")
		return
	}
	if err := a.auth.DeleteUserAPIKey(ctx, id, user.ID); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "deleted"})
}

type createDashboardEnvironmentRequest struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

type dashboardEnvironmentResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Visibility string    `json:"visibility"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

func (a *API) handleCreateDashboardEnvironment(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	var req createDashboardEnvironmentRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}
	if req.Name == "" {
		writeError(ctx, fasthttp.StatusBadRequest, "name is required")
		return
	}
	if req.Visibility != "private" && req.Visibility != "public" {
		req.Visibility = "private"
	}

	env := &model.DashboardEnvironment{
		UserID:     user.ID,
		Name:       req.Name,
		Visibility: req.Visibility,
	}
	if err := a.dashboardEnvRepo.Create(ctx, env); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusCreated, dashboardEnvironmentResponse{
		ID:         env.ID.String(),
		Name:       env.Name,
		Visibility: env.Visibility,
		Active:     true,
		CreatedAt:  env.CreatedAt,
	})
}

func (a *API) handleListDashboardEnvironments(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	envs, err := a.dashboardEnvRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	out := make([]dashboardEnvironmentResponse, 0, len(envs))
	for i := range envs {
		out = append(out, dashboardEnvironmentResponse{
			ID:         envs[i].ID.String(),
			Name:       envs[i].Name,
			Visibility: envs[i].Visibility,
			Active:     envs[i].Active,
			CreatedAt:  envs[i].CreatedAt,
		})
	}
	writeJSON(ctx, fasthttp.StatusOK, out)
}

type patchDashboardEnvironmentRequest struct {
	ID         string  `json:"id"`
	Active     *bool   `json:"active"`
	Visibility *string `json:"visibility"`
}

func (a *API) handlePatchDashboardEnvironment(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)

	var req patchDashboardEnvironmentRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}
	if req.Active == nil && req.Visibility == nil {
		writeError(ctx, fasthttp.StatusBadRequest, "nothing to update")
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid environment id")
		return
	}

	target, err := a.dashboardEnvRepo.GetByID(ctx, id)
	if err != nil {
		writeError(ctx, fasthttp.StatusNotFound, "environment not found")
		return
	}
	if target.UserID != user.ID {
		writeError(ctx, fasthttp.StatusForbidden, "not your environment")
		return
	}

	resp := dashboardEnvironmentResponse{
		ID:         target.ID.String(),
		Name:       target.Name,
		Visibility: target.Visibility,
		Active:     target.Active,
		CreatedAt:  target.CreatedAt,
	}

	if req.Active != nil {
		if err := a.dashboardEnvRepo.SetActive(ctx, id, *req.Active); err != nil {
			writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
		resp.Active = *req.Active
	}
	if req.Visibility != nil {
		vis := *req.Visibility
		if vis != "private" && vis != "public" {
			writeError(ctx, fasthttp.StatusBadRequest, "visibility must be private or public")
			return
		}
		if err := a.dashboardEnvRepo.SetVisibility(ctx, id, vis); err != nil {
			writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
		resp.Visibility = vis
	}

	writeJSON(ctx, fasthttp.StatusOK, resp)
}

func (a *API) handleDeleteDashboardEnvironment(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	idStr := ctx.UserValue("id").(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid environment id")
		return
	}
	if err := a.dashboardEnvRepo.Delete(ctx, id, user.ID); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "deleted"})
}

type canvasNodeRequest struct {
	ItemID    string             `json:"item_id"`
	Kind      string             `json:"kind"`
	Name      string             `json:"name"`
	Detail    string             `json:"detail"`
	DetailSet bool               `json:"-"`
	Desc      string             `json:"description"`
	DescSet   bool               `json:"-"`
	Meta      map[string]*string `json:"meta"`
	PositionX float64            `json:"position_x"`
	PositionY float64            `json:"position_y"`
	SortOrder int                `json:"sort_order"`
}

type putCanvasRequest struct {
	Nodes []canvasNodeRequest `json:"nodes"`
}

type canvasNodeResponse struct {
	ID          string             `json:"id"`
	ItemID      string             `json:"item_id"`
	Kind        string             `json:"kind"`
	Name        string             `json:"name"`
	Detail      string             `json:"detail,omitempty"`
	Description string             `json:"description,omitempty"`
	Meta        map[string]*string `json:"meta,omitempty"`
	PositionX   float64            `json:"position_x"`
	PositionY   float64            `json:"position_y"`
	SortOrder   int                `json:"sort_order"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func (a *API) handleGetCanvas(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	envID, err := uuid.Parse(ctx.UserValue("id").(string))
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid environment id")
		return
	}
	env, err := a.dashboardEnvRepo.GetByID(ctx, envID)
	if err != nil {
		writeError(ctx, fasthttp.StatusNotFound, "environment not found")
		return
	}
	if env.UserID != user.ID {
		writeError(ctx, fasthttp.StatusForbidden, "not your environment")
		return
	}

	nodes, err := a.canvasNodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	out := make([]canvasNodeResponse, len(nodes))
	for i, n := range nodes {
		var meta map[string]*string
		if n.Meta != "" {
			json.Unmarshal([]byte(n.Meta), &meta)
		}
		out[i] = canvasNodeResponse{
			ID:          n.ID.String(),
			ItemID:      n.ItemID,
			Kind:        n.Kind,
			Name:        n.Name,
			Detail:      n.Detail,
			Description: n.Description,
			Meta:        meta,
			PositionX:   n.PositionX,
			PositionY:   n.PositionY,
			SortOrder:   n.SortOrder,
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		}
	}
	writeJSON(ctx, fasthttp.StatusOK, out)
}

func (a *API) handlePutCanvas(ctx *fasthttp.RequestCtx) {
	user := userFrom(ctx)
	envID, err := uuid.Parse(ctx.UserValue("id").(string))
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid environment id")
		return
	}
	env, err := a.dashboardEnvRepo.GetByID(ctx, envID)
	if err != nil {
		writeError(ctx, fasthttp.StatusNotFound, "environment not found")
		return
	}
	if env.UserID != user.ID {
		writeError(ctx, fasthttp.StatusForbidden, "not your environment")
		return
	}

	var req putCanvasRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid json body")
		return
	}

	nodes := make([]model.CanvasNode, len(req.Nodes))
	for i, n := range req.Nodes {
		var metaBytes []byte
		if n.Meta != nil {
			metaBytes, _ = json.Marshal(n.Meta)
		}
		nodes[i] = model.CanvasNode{
			EnvironmentID: envID,
			UserID:        user.ID,
			ItemID:        n.ItemID,
			Kind:          n.Kind,
			Name:          n.Name,
			Detail:        n.Detail,
			Description:   n.Desc,
			Meta:          string(metaBytes),
			PositionX:     n.PositionX,
			PositionY:     n.PositionY,
			SortOrder:     n.SortOrder,
		}
	}

	if err := a.canvasNodeRepo.Replace(ctx, envID, nodes); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "saved"})
}

type wsCanvasMessage struct {
	Type  string             `json:"type"`
	Nodes []canvasNodeRequest `json:"nodes,omitempty"`
}

type wsCanvasResponse struct {
	Type  string              `json:"type"`
	Nodes []canvasNodeResponse `json:"nodes,omitempty"`
	Error string              `json:"error,omitempty"`
}

func (a *API) handleWSCanvas(ctx *fasthttp.RequestCtx) {
	envID, err := uuid.Parse(ctx.UserValue("id").(string))
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid environment id")
		return
	}

	upgrader := websocket.FastHTTPUpgrader{
		CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
	}

	err = upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		conn.SetPongHandler(func(_ string) error {
			conn.SetReadDeadline(time.Now().Add(45 * time.Second))
			return nil
		})
		conn.SetReadDeadline(time.Now().Add(45 * time.Second))

		token := string(ctx.QueryArgs().Peek("token"))
		if token == "" {
			token = string(ctx.Request.Header.Peek(authHeader))
		}
		user, err := a.auth.Authenticate(ctx, token)
		if err != nil {
			conn.WriteJSON(&wsCanvasResponse{Type: "error", Error: "invalid or missing api token"})
			return
		}

		env, err := a.dashboardEnvRepo.GetByID(ctx, envID)
		if err != nil {
			conn.WriteJSON(&wsCanvasResponse{Type: "error", Error: "environment not found"})
			return
		}
		if env.UserID != user.ID {
			conn.WriteJSON(&wsCanvasResponse{Type: "error", Error: "not your environment"})
			return
		}

		nodes, err := a.canvasNodeRepo.ListByEnvironment(ctx, envID)
		if err != nil {
			conn.WriteJSON(&wsCanvasResponse{Type: "error", Error: err.Error()})
			return
		}

		out := make([]canvasNodeResponse, len(nodes))
		for i, n := range nodes {
			var meta map[string]*string
			if n.Meta != "" {
				json.Unmarshal([]byte(n.Meta), &meta)
			}
			out[i] = canvasNodeResponse{
				ID:          n.ID.String(),
				ItemID:      n.ItemID,
				Kind:        n.Kind,
				Name:        n.Name,
				Detail:      n.Detail,
				Description: n.Description,
				Meta:        meta,
				PositionX:   n.PositionX,
				PositionY:   n.PositionY,
				SortOrder:   n.SortOrder,
				CreatedAt:   n.CreatedAt,
				UpdatedAt:   n.UpdatedAt,
			}
		}
		conn.WriteJSON(&wsCanvasResponse{Type: "init", Nodes: out})

		done := make(chan struct{})
		defer close(done)

		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					conn.WriteMessage(websocket.PingMessage, nil)
				case <-done:
					return
				}
			}
		}()

		for {
			var msg wsCanvasMessage
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}

			switch msg.Type {
			case "save":
				modelNodes := make([]model.CanvasNode, len(msg.Nodes))
				for i, n := range msg.Nodes {
					var metaBytes []byte
					if n.Meta != nil {
						metaBytes, _ = json.Marshal(n.Meta)
					}
					modelNodes[i] = model.CanvasNode{
						EnvironmentID: envID,
						UserID:        user.ID,
						ItemID:        n.ItemID,
						Kind:          n.Kind,
						Name:          n.Name,
						Detail:        n.Detail,
						Description:   n.Desc,
						Meta:          string(metaBytes),
						PositionX:     n.PositionX,
						PositionY:     n.PositionY,
						SortOrder:     n.SortOrder,
					}
				}
				if err := a.canvasNodeRepo.Replace(ctx, envID, modelNodes); err != nil {
					conn.WriteJSON(&wsCanvasResponse{Type: "error", Error: err.Error()})
				} else {
					conn.WriteJSON(&wsCanvasResponse{Type: "saved"})
				}
			default:
				conn.WriteJSON(&wsCanvasResponse{Type: "error", Error: "unknown message type"})
			}
		}
	})
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, err.Error())
	}
}

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

func balanceResponseFromUser(user *model.User) billingBalanceResponse {
	return billingBalanceResponse{
		UserID:          user.ID.String(),
		Balance:         user.Balance,
		MarginMode:      user.MarginMode,
		SafeMarginLimit: user.SafeMarginLimit,
		AutoPayInterval: user.AutoPayInterval,
		AutoPayDay:      user.AutoPayDay,
	}
}

func transactionResponseFromModel(tx model.Transaction) transactionResponse {
	resp := transactionResponse{
		ID:           tx.ID.String(),
		UserID:       tx.UserID.String(),
		Type:         tx.Type,
		Amount:       tx.Amount,
		Reason:       tx.Reason,
		BalanceAfter: tx.BalanceAfter,
		CreatedAt:    tx.CreatedAt,
	}
	if tx.AgentID != nil {
		resp.AgentID = tx.AgentID.String()
	}
	return resp
}

func usageMetricResponseFromModel(metric model.UsageMetric) usageMetricResponse {
	return usageMetricResponse{
		ID:            metric.ID.String(),
		UserID:        metric.UserID.String(),
		Date:          metric.Date,
		CPUSeconds:    metric.CPUSeconds,
		MemoryMBHours: metric.MemoryMBHours,
		DiskMB:        metric.DiskMB,
		LoadPercent:   metric.LoadPercent,
	}
}

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func queryInt(ctx *fasthttp.RequestCtx, key string, fallback int) int {
	raw := string(ctx.QueryArgs().Peek(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func boundedQueryInt(ctx *fasthttp.RequestCtx, key string, fallback, min, max int) int {
	value := queryInt(ctx, key, fallback)
	if value < min || value > max {
		return fallback
	}
	return value
}

type skillSearchResponse struct {
	Query  string           `json:"query"`
	Skills []skillSearchHit `json:"skills"`
	Count  int              `json:"count"`
}

type skillSearchHit struct {
	ID         string `json:"id"`
	SkillID    string `json:"skill_id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	InstallCmd string `json:"install_cmd"`
}

func (a *API) handleSkillSearch(ctx *fasthttp.RequestCtx) {
	q := strings.TrimSpace(string(ctx.QueryArgs().Peek("q")))
	limit := boundedQueryInt(ctx, "limit", 20, 1, 100)
	hits, err := a.searchSkillHits(ctx, q, limit)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, skillSearchResponse{
		Query:  q,
		Skills: hits,
		Count:  len(hits),
	})
}

type cliResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NixAttr    string `json:"nix_attr"`
	InstallCmd string `json:"install_cmd,omitempty"`
}

func (a *API) handleListCLIs(ctx *fasthttp.RequestCtx) {
	if a.cliRepo == nil {
		writeJSON(ctx, fasthttp.StatusOK, []cliResponse{})
		return
	}
	list, err := a.cliRepo.List(ctx)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	out := make([]cliResponse, 0, len(list))
	for _, c := range list {
		out = append(out, cliResponse{
			ID:         c.ID.String(),
			Name:       c.Name,
			NixAttr:    c.NixAttr,
			InstallCmd: c.InstallCmd,
		})
	}
	writeJSON(ctx, fasthttp.StatusOK, out)
}

type cliSearchResult struct {
	Query string        `json:"query"`
	CLIs  []cliResponse `json:"clis"`
	Count int           `json:"count"`
}

func (a *API) handleCLISearch(ctx *fasthttp.RequestCtx) {
	q := strings.TrimSpace(string(ctx.QueryArgs().Peek("q")))
	limit := boundedQueryInt(ctx, "limit", 20, 1, 100)
	hits, err := a.searchCLIResponses(ctx, q, limit)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, cliSearchResult{Query: q, CLIs: hits, Count: len(hits)})
}

type providerResponse struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Name         string `json:"name"`
	BaseURL      string `json:"base_url"`
	AuthEnv      string `json:"auth_env"`
	DefaultModel string `json:"default_model,omitempty"`
	Description  string `json:"description,omitempty"`
}

func (a *API) handleListProviders(ctx *fasthttp.RequestCtx) {
	if a.providerRepo == nil {
		writeJSON(ctx, fasthttp.StatusOK, []providerResponse{})
		return
	}
	list, err := a.providerRepo.List(ctx)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	out := make([]providerResponse, 0, len(list))
	for _, p := range list {
		out = append(out, providerResponseFromModel(p))
	}
	writeJSON(ctx, fasthttp.StatusOK, out)
}

type providerSearchResult struct {
	Query     string             `json:"query"`
	Providers []providerResponse `json:"providers"`
	Count     int                `json:"count"`
}

func (a *API) handleProviderSearch(ctx *fasthttp.RequestCtx) {
	q := strings.TrimSpace(string(ctx.QueryArgs().Peek("q")))
	limit := boundedQueryInt(ctx, "limit", 20, 1, 100)
	hits, err := a.searchProviderResponses(ctx, q, limit)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, providerSearchResult{Query: q, Providers: hits, Count: len(hits)})
}

type catalogSearchResponse struct {
	Query    string                `json:"query"`
	Category string                `json:"category"`
	Items    []catalogItemResponse `json:"items"`
	Count    int                   `json:"count"`
}

type catalogItemResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Subtitle     string `json:"subtitle,omitempty"`
	Description  string `json:"description,omitempty"`
	Key          string `json:"key,omitempty"`
	BaseURL      string `json:"base_url,omitempty"`
	AuthEnv      string `json:"auth_env,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
	NixAttr      string `json:"nix_attr,omitempty"`
	InstallCmd   string `json:"install_cmd,omitempty"`
	SkillID      string `json:"skill_id,omitempty"`
	Source       string `json:"source,omitempty"`
}

func (a *API) handleCatalogSearch(ctx *fasthttp.RequestCtx) {
	q := strings.TrimSpace(string(ctx.QueryArgs().Peek("q")))
	category := normalizeCatalogCategory(string(ctx.QueryArgs().Peek("category")))
	limit := boundedQueryInt(ctx, "limit", 20, 1, 100)
	items := make([]catalogItemResponse, 0, limit)

	if includesCatalogCategory(category, "providers") {
		providers, err := a.searchProviderResponses(ctx, q, limit-len(items))
		if err != nil {
			writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
		for _, p := range providers {
			items = append(items, catalogItemResponse{
				ID:           p.ID,
				Type:         "provider",
				Name:         p.Name,
				Subtitle:     p.BaseURL,
				Description:  p.Description,
				Key:          p.Key,
				BaseURL:      p.BaseURL,
				AuthEnv:      p.AuthEnv,
				DefaultModel: p.DefaultModel,
			})
		}
	}
	if len(items) < limit && includesCatalogCategory(category, "cli") {
		clis, err := a.searchCLIResponses(ctx, q, limit-len(items))
		if err != nil {
			writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
		for _, c := range clis {
			items = append(items, catalogItemResponse{
				ID:         c.ID,
				Type:       "cli",
				Name:       c.Name,
				Subtitle:   c.NixAttr,
				NixAttr:    c.NixAttr,
				InstallCmd: c.InstallCmd,
			})
		}
	}
	if len(items) < limit && includesCatalogCategory(category, "skills") {
		skills, err := a.searchSkillHits(ctx, q, limit-len(items))
		if err != nil {
			writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
		for _, s := range skills {
			items = append(items, catalogItemResponse{
				ID:         s.ID,
				Type:       "skill",
				Name:       s.Name,
				Subtitle:   s.Source,
				SkillID:    s.SkillID,
				Source:     s.Source,
				InstallCmd: s.InstallCmd,
			})
		}
	}
	if len(items) < limit && includesCustomProvider(category) && customProviderMatches(q) {
		items = append(items, catalogItemResponse{
			ID:          "custom-provider",
			Type:        "custom_provider",
			Name:        "Custom provider",
			Subtitle:    "Base URL and API key",
			Description: "Configure an OpenAI-compatible provider endpoint.",
		})
	}

	writeJSON(ctx, fasthttp.StatusOK, catalogSearchResponse{
		Query:    q,
		Category: category,
		Items:    items,
		Count:    len(items),
	})
}

func (a *API) searchSkillHits(ctx context.Context, q string, limit int) ([]skillSearchHit, error) {
	if limit <= 0 {
		return []skillSearchHit{}, nil
	}
	if a.typesense != nil && q != "" {
		result, err := a.typesense.SearchSkills(ctx, q, limit)
		if err == nil {
			hits := make([]skillSearchHit, 0, limit)
			if result.Hits != nil {
				for _, hit := range *result.Hits {
					if hit.Document == nil {
						continue
					}
					doc := *hit.Document
					hits = append(hits, skillSearchHit{
						ID:         getStrField(doc, "id"),
						SkillID:    getStrField(doc, "skill_id"),
						Name:       getStrField(doc, "name"),
						Source:     getStrField(doc, "source"),
						InstallCmd: getStrField(doc, "install_cmd"),
					})
				}
			}
			return hits, nil
		}
	}
	if a.skillsRepo == nil {
		return []skillSearchHit{}, nil
	}
	list, err := a.skillsRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	hits := make([]skillSearchHit, 0, minInt(limit, len(list)))
	for _, s := range list {
		if !textMatches(q, s.Name, s.SkillID, s.Source, s.InstallCmd, s.Description) {
			continue
		}
		hits = append(hits, skillSearchHit{
			ID:         s.ID.String(),
			SkillID:    s.SkillID,
			Name:       s.Name,
			Source:     s.Source,
			InstallCmd: s.InstallCmd,
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func (a *API) searchCLIResponses(ctx context.Context, q string, limit int) ([]cliResponse, error) {
	if limit <= 0 {
		return []cliResponse{}, nil
	}
	if a.typesense != nil && q != "" {
		result, err := a.typesense.SearchCLIs(ctx, q, limit)
		if err == nil {
			hits := make([]cliResponse, 0, limit)
			if result.Hits != nil {
				for _, hit := range *result.Hits {
					if hit.Document == nil {
						continue
					}
					doc := *hit.Document
					hits = append(hits, cliResponse{
						ID:         getStrField(doc, "id"),
						Name:       getStrField(doc, "name"),
						NixAttr:    getStrField(doc, "nix_attr"),
						InstallCmd: getStrField(doc, "install_cmd"),
					})
				}
			}
			return hits, nil
		}
	}
	if a.cliRepo == nil {
		return []cliResponse{}, nil
	}
	list, err := a.cliRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	hits := make([]cliResponse, 0, minInt(limit, len(list)))
	for _, c := range list {
		if !textMatches(q, c.Name, c.NixAttr, c.InstallCmd) {
			continue
		}
		hits = append(hits, cliResponseFromModel(c))
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func (a *API) searchProviderResponses(ctx context.Context, q string, limit int) ([]providerResponse, error) {
	if limit <= 0 {
		return []providerResponse{}, nil
	}
	if a.typesense != nil && q != "" {
		result, err := a.typesense.SearchProviders(ctx, q, limit)
		if err == nil {
			hits := make([]providerResponse, 0, limit)
			if result.Hits != nil {
				for _, hit := range *result.Hits {
					if hit.Document == nil {
						continue
					}
					doc := *hit.Document
					hits = append(hits, providerResponse{
						ID:           getStrField(doc, "id"),
						Key:          getStrField(doc, "key"),
						Name:         getStrField(doc, "name"),
						BaseURL:      getStrField(doc, "base_url"),
						AuthEnv:      getStrField(doc, "auth_env"),
						DefaultModel: getStrField(doc, "default_model"),
						Description:  getStrField(doc, "description"),
					})
				}
			}
			return hits, nil
		}
	}
	if a.providerRepo == nil {
		return []providerResponse{}, nil
	}
	list, err := a.providerRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	hits := make([]providerResponse, 0, minInt(limit, len(list)))
	for _, p := range list {
		if !textMatches(q, p.Key, p.Name, p.BaseURL, p.AuthEnv, p.DefaultModel, p.Description) {
			continue
		}
		hits = append(hits, providerResponseFromModel(p))
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func cliResponseFromModel(c model.CLI) cliResponse {
	return cliResponse{
		ID:         c.ID.String(),
		Name:       c.Name,
		NixAttr:    c.NixAttr,
		InstallCmd: c.InstallCmd,
	}
}

func providerResponseFromModel(p model.Provider) providerResponse {
	return providerResponse{
		ID:           p.ID.String(),
		Key:          p.Key,
		Name:         p.Name,
		BaseURL:      p.BaseURL,
		AuthEnv:      p.AuthEnv,
		DefaultModel: p.DefaultModel,
		Description:  p.Description,
	}
}

func normalizeCatalogCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "providers", "provider":
		return "providers"
	case "cli", "clis":
		return "cli"
	case "skills", "skill":
		return "skills"
	case "custom", "custom_provider", "custom-provider":
		return "custom"
	default:
		return "all"
	}
}

func includesCatalogCategory(category, target string) bool {
	return category == "all" || category == target
}

func includesCustomProvider(category string) bool {
	return category == "all" || category == "providers" || category == "custom"
}

func customProviderMatches(q string) bool {
	return textMatches(q, "custom provider", "custom", "provider", "base url", "api key", "openai compatible")
}

func textMatches(query string, fields ...string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getStrField(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

var _ context.Context = (*fasthttp.RequestCtx)(nil)
