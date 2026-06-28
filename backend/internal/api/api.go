package api

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/oauth"
	"backend/internal/repository"
	"backend/internal/service"
	ts "backend/internal/typesense"

	"github.com/fasthttp/router"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const authHeader = "octra-api-token"
const authHeaderBearer = "Authorization"
const ctxUserKey = "octra_user"

type API struct {
	auth       *service.AuthService
	env        *service.EnvironmentService
	chat       *service.ChatService
	billing    *service.BillingService
	oauthH     *oauth.Handler
	dashboardEnvRepo repository.DashboardEnvironmentRepository
	cliRepo          repository.CLIRepository
	typesense        *ts.Client
}

func New(auth *service.AuthService, env *service.EnvironmentService, chat *service.ChatService, billing *service.BillingService, oauthH *oauth.Handler, dashboardEnvRepo repository.DashboardEnvironmentRepository, cliRepo repository.CLIRepository, typesense *ts.Client) *API {
	return &API{auth: auth, env: env, chat: chat, billing: billing, oauthH: oauthH, dashboardEnvRepo: dashboardEnvRepo, cliRepo: cliRepo, typesense: typesense}
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
	r.DELETE("/api/keys/:id", a.withAuth(a.handleDeleteAPIKey))

	// Dashboard environments
	r.POST("/api/environments", a.withAuth(a.handleCreateDashboardEnvironment))
	r.GET("/api/environments", a.withAuth(a.handleListDashboardEnvironments))
	r.POST("/api/environments/patch", a.withAuth(a.handlePatchDashboardEnvironment))
	r.DELETE("/api/environments/:id", a.withAuth(a.handleDeleteDashboardEnvironment))

	// Skills search
	r.GET("/skills/search", a.handleSkillSearch)

	// CLI catalogue
	r.GET("/api/cli", a.handleListCLIs)
	r.GET("/api/cli/search", a.handleCLISearch)
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
		APIKey  string `json:"api_key"`
		BaseURL string `json:"base_url"`
		Model   string `json:"model"`
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
		LLMAPIKey:  req.LLM.APIKey,
		LLMBaseURL: req.LLM.BaseURL,
		LLMModel:   req.LLM.Model,
		CLI:        model.CLIType(req.Agent.CLI),
		Priority:   req.Agent.Priority,
		Skills:     req.Skills,
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

type skillSearchResponse struct {
	Query  string              `json:"query"`
	Skills []skillSearchHit    `json:"skills"`
	Count  int                 `json:"count"`
}

type skillSearchHit struct {
	ID         string `json:"id"`
	SkillID    string `json:"skill_id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	InstallCmd string `json:"install_cmd"`
}

func (a *API) handleSkillSearch(ctx *fasthttp.RequestCtx) {
	if a.typesense == nil {
		writeError(ctx, fasthttp.StatusServiceUnavailable, "search not available")
		return
	}
	q := string(ctx.QueryArgs().Peek("q"))
	if q == "" {
		writeError(ctx, fasthttp.StatusBadRequest, "query parameter q is required")
		return
	}
	limit := queryInt(ctx, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := a.typesense.SearchSkills(ctx, q, limit)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "search failed")
		return
	}

	hits := make([]skillSearchHit, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		doc := *hit.Document
		hits = append(hits, skillSearchHit{
			ID:         getStrField(doc, "id"),
			SkillID:    getStrField(doc, "skill_id"),
			Name:       getStrField(doc, "name"),
			Source:     getStrField(doc, "source"),
			InstallCmd: getStrField(doc, "install_cmd"),
		})
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
	if a.typesense == nil {
		writeError(ctx, fasthttp.StatusServiceUnavailable, "search not available")
		return
	}
	q := string(ctx.QueryArgs().Peek("q"))
	if q == "" {
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
		writeJSON(ctx, fasthttp.StatusOK, cliSearchResult{Query: q, CLIs: out, Count: len(out)})
		return
	}
	limit := queryInt(ctx, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}
	result, err := a.typesense.SearchCLIs(ctx, q, limit)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, "search failed")
		return
	}
	hits := make([]cliResponse, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		doc := *hit.Document
		hits = append(hits, cliResponse{
			ID:         getStrField(doc, "id"),
			Name:       getStrField(doc, "name"),
			NixAttr:    getStrField(doc, "nix_attr"),
			InstallCmd: getStrField(doc, "install_cmd"),
		})
	}
	writeJSON(ctx, fasthttp.StatusOK, cliSearchResult{Query: q, CLIs: hits, Count: len(hits)})
}

func getStrField(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

var _ context.Context = (*fasthttp.RequestCtx)(nil)
