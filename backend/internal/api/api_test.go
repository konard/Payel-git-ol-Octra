package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"backend/internal/cli"
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/oauth"
	"backend/internal/provider"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type fakeProvisioner struct{}

func (fakeProvisioner) CreateEnvironment(context.Context, string) error         { return nil }
func (fakeProvisioner) InstallSkill(context.Context, string, model.Skill) error { return nil }

type fakeEnvPaths struct{}

func (fakeEnvPaths) EnvPath(id string) string { return "/envs/" + id }

type fakeOcaweProvider struct {
	port int
}

func (f *fakeOcaweProvider) EnsureOcawe(_ context.Context, _ cli.LaunchSpec) (int, error) {
	if f.port > 0 {
		return f.port, nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		model, _ := body["model"].(string)

		var prefix string
		if len(model) > 4 && model[:4] == "cli/" {
			prefix = "cli:"
		} else {
			prefix = "llm:"
		}

		messages := body["messages"].([]any)
		msg := messages[0].(map[string]any)
		prompt, _ := msg["content"].(string)

		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": prefix + prompt,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	f.port = listener.Addr().(*net.TCPAddr).Port

	go http.Serve(listener, mux)

	time.Sleep(50 * time.Millisecond)
	return f.port, nil
}

func testCfg() config.Config {
	return config.Config{
		JWTSecret:               "test-jwt-secret",
		JWTRefreshSecret:        "test-jwt-refresh-secret",
		FrontendURL:             "http://localhost:5173",
		LeFineIntegrationSecret: "test-lefine-secret",
	}
}

func newTestServer(t *testing.T) (*fasthttp.Client, func()) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		t.Fatal(err)
	}

	users := repository.NewUserRepository(db)
	agents := repository.NewAgentRepository(db)
	skills := repository.NewSkillRepository(db)
	userSkills := repository.NewUserSkillRepository(db)
	transactions := repository.NewTransactionRepository(db)
	usageMetrics := repository.NewUsageMetricsRepository(db)
	requestMetrics := repository.NewRequestMetricsRepository(db)
	dashboardEnvs := repository.NewDashboardEnvironmentRepository(db)
	canvasNodes := repository.NewCanvasNodeRepository(db)
	clis := repository.NewCLIRepository(db)
	providers := repository.NewProviderRepository(db)

	for _, p := range cli.BuiltinCLIs() {
		if err := clis.Upsert(context.Background(), &model.CLI{Name: p.Name, NixAttr: p.NixAttr, InstallCmd: p.InstallCmd}); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range provider.BuiltinProviders() {
		if err := providers.Upsert(context.Background(), &model.Provider{
			Key:          p.Key,
			Name:         p.Name,
			BaseURL:      p.BaseURL,
			AuthEnv:      p.AuthEnv,
			DefaultModel: p.DefaultModel,
			Description:  p.Description,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := skills.Upsert(context.Background(), &model.Skill{
		Name:       "github-tools",
		Type:       model.SkillNixpkgs,
		InstallCmd: "gh",
		SkillID:    "github-tools",
		Source:     "octra/tests",
	}); err != nil {
		t.Fatal(err)
	}

	cfg := testCfg()
	billingSvc := service.NewBillingService(users, agents, transactions, usageMetrics)
	metricsSvc := service.NewMetricsService(requestMetrics, dashboardEnvs)
	authSvc := service.NewAuthService(users, cfg, transactions)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, fakeProvisioner{}, billingSvc)
	chatSvc := service.NewChatService(agents, &fakeOcaweProvider{}, fakeEnvPaths{})
	oauthH := oauth.New(authSvc, cfg)

	handler := New(authSvc, envSvc, chatSvc, billingSvc, metricsSvc, oauthH, dashboardEnvs, canvasNodes, skills, clis, providers, nil).Router().Handler

	ln := fasthttputil.NewInmemoryListener()
	server := &fasthttp.Server{Handler: handler}
	go func() { _ = server.Serve(ln) }()

	client := &fasthttp.Client{
		Dial: func(string) (net.Conn, error) { return ln.Dial() },
	}
	cleanup := func() { _ = ln.Close() }
	return client, cleanup
}

func do(t *testing.T, client *fasthttp.Client, method, path, token, body string) (int, []byte) {
	t.Helper()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://localhost" + path)
	req.Header.SetMethod(method)
	if token != "" {
		req.Header.Set(authHeader, token)
	}
	if body != "" {
		req.Header.SetContentType("application/json")
		req.SetBodyString(body)
	}
	if err := client.DoTimeout(req, resp, 5*time.Second); err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	out := make([]byte, len(resp.Body()))
	copy(out, resp.Body())
	return resp.StatusCode(), out
}

func TestEndToEndProxyFlow(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	if code, _ := do(t, client, "GET", "/health", "", ""); code != 200 {
		t.Fatalf("health code %d", code)
	}

	code, body := do(t, client, "POST", "/register", "", `{"username":"testuser","email":"a@b.com","password":"pw"}`)
	if code != 201 {
		t.Fatalf("register code %d: %s", code, body)
	}
	var reg registerResponse
	if err := json.Unmarshal(body, &reg); err != nil {
		t.Fatal(err)
	}
	if reg.APIKey == "" {
		t.Fatal("no api key returned")
	}
	if reg.Balance != model.DefaultRegistrationCredits {
		t.Fatalf("register balance = %d", reg.Balance)
	}

	if code, _ := do(t, client, "POST", "/api/chat", reg.APIKey, `{"prompt":"hi"}`); code != 400 {
		t.Fatalf("expected 400 before env, got %d", code)
	}

	code, body = do(t, client, "POST", "/environment", reg.APIKey,
		`{"llm":{"api_key":"sk","base_url":"https://api.anthropic.com"},"agent":{"cli":"","priority":3},"skills":[]}`)
	if code != 200 {
		t.Fatalf("environment code %d: %s", code, body)
	}

	code, body = do(t, client, "POST", "/api/chat", reg.APIKey, `{"prompt":"hello"}`)
	if code != 200 {
		t.Fatalf("chat code %d: %s", code, body)
	}
	var chat chatResponse
	if err := json.Unmarshal(body, &chat); err != nil {
		t.Fatal(err)
	}
	if chat.Response != "llm:hello" {
		t.Fatalf("unexpected chat response %q", chat.Response)
	}
}

func TestRequestMetricsEndpoint(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	// Auth is required.
	if code, _ := do(t, client, "GET", "/api/metrics/requests", "", ""); code != 401 {
		t.Fatalf("expected 401 without token, got %d", code)
	}

	code, body := do(t, client, "POST", "/register", "", `{"username":"m","email":"m@b.com","password":"pw"}`)
	if code != 201 {
		t.Fatalf("register code %d: %s", code, body)
	}
	var reg registerResponse
	if err := json.Unmarshal(body, &reg); err != nil {
		t.Fatal(err)
	}

	// No traffic yet: a well-formed empty series is returned.
	code, body = do(t, client, "GET", "/api/metrics/requests", reg.APIKey, "")
	if code != 200 {
		t.Fatalf("metrics code %d: %s", code, body)
	}
	var empty service.RequestMetricsResult
	if err := json.Unmarshal(body, &empty); err != nil {
		t.Fatalf("unmarshal metrics: %v (%s)", err, body)
	}
	if empty.Total != 0 || len(empty.Series) != 7 {
		t.Fatalf("expected empty 7d series, got total=%d series=%d", empty.Total, len(empty.Series))
	}

	// Provision an environment and send a chat so a metric row is recorded.
	code, body = do(t, client, "POST", "/environment", reg.APIKey,
		`{"llm":{"api_key":"sk","base_url":"https://api.anthropic.com"},"agent":{"cli":"","priority":3},"skills":[]}`)
	if code != 200 {
		t.Fatalf("environment code %d: %s", code, body)
	}
	if code, body = do(t, client, "POST", "/api/chat", reg.APIKey, `{"prompt":"hello"}`); code != 200 {
		t.Fatalf("chat code %d: %s", code, body)
	}

	// The metric is now reflected in the aggregated series.
	code, body = do(t, client, "GET", "/api/metrics/requests?range=24h", reg.APIKey, "")
	if code != 200 {
		t.Fatalf("metrics code %d: %s", code, body)
	}
	var res service.RequestMetricsResult
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal metrics: %v (%s)", err, body)
	}
	if res.Range != "24h" || res.Bucket != "hour" {
		t.Fatalf("unexpected range/bucket: %s/%s", res.Range, res.Bucket)
	}
	if res.Total != 1 || res.Success != 1 {
		t.Fatalf("expected 1 successful request, got total=%d success=%d", res.Total, res.Success)
	}

	// A bad env filter is rejected.
	if code, _ = do(t, client, "GET", "/api/metrics/requests?env=not-a-uuid", reg.APIKey, ""); code != 400 {
		t.Fatalf("expected 400 for bad env id, got %d", code)
	}
}

func TestPublicProfileAndLeaderboard(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	code, body := do(t, client, "POST", "/register", "", `{"username":"alpha","email":"alpha@example.com","password":"pw"}`)
	if code != 201 {
		t.Fatalf("register alpha code %d: %s", code, body)
	}
	var alpha registerResponse
	if err := json.Unmarshal(body, &alpha); err != nil {
		t.Fatal(err)
	}

	code, body = do(t, client, "POST", "/register", "", `{"username":"beta","email":"beta@example.com","password":"pw"}`)
	if code != 201 {
		t.Fatalf("register beta code %d: %s", code, body)
	}
	var beta registerResponse
	if err := json.Unmarshal(body, &beta); err != nil {
		t.Fatal(err)
	}

	code, body = do(t, client, "POST", "/api/environments", alpha.APIKey, `{"name":"Shared Pipeline","visibility":"public"}`)
	if code != 201 {
		t.Fatalf("create public env code %d: %s", code, body)
	}
	code, body = do(t, client, "POST", "/api/environments", alpha.APIKey, `{"name":"Private Lab","visibility":"private"}`)
	if code != 201 {
		t.Fatalf("create private env code %d: %s", code, body)
	}
	code, body = do(t, client, "POST", "/api/environments", beta.APIKey, `{"name":"Quiet Public Env","visibility":"public"}`)
	if code != 201 {
		t.Fatalf("create beta env code %d: %s", code, body)
	}

	code, body = do(t, client, "POST", "/environment", alpha.APIKey, `{"llm":{"api_key":"sk"},"agent":{"cli":""}}`)
	if code != 200 {
		t.Fatalf("agent environment code %d: %s", code, body)
	}
	code, body = do(t, client, "POST", "/api/chat", alpha.APIKey, `{"prompt":"hello"}`)
	if code != 200 {
		t.Fatalf("chat code %d: %s", code, body)
	}

	code, body = do(t, client, "GET", "/api/users/"+alpha.UserID+"/profile?range=24h", "", "")
	if code != 200 {
		t.Fatalf("profile code %d: %s", code, body)
	}
	if strings.Contains(string(body), "alpha@example.com") || strings.Contains(string(body), alpha.APIKey) {
		t.Fatalf("public profile leaked private user fields: %s", body)
	}

	var profile publicProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		t.Fatalf("unmarshal profile: %v (%s)", err, body)
	}
	if profile.User.ID != alpha.UserID || profile.User.Username != "alpha" {
		t.Fatalf("unexpected profile user: %+v", profile.User)
	}
	if profile.User.PublicEnvCount != 1 || profile.User.ActivePublicEnvCount != 1 {
		t.Fatalf("unexpected public env counts: %+v", profile.User)
	}
	if len(profile.PublicEnvironments) != 1 || profile.PublicEnvironments[0].Name != "Shared Pipeline" {
		t.Fatalf("profile should expose only public envs, got %+v", profile.PublicEnvironments)
	}
	if profile.Workload.Total != 1 || len(profile.Workload.Candles) != 24 {
		t.Fatalf("expected one request and hourly candles, got total=%d candles=%d", profile.Workload.Total, len(profile.Workload.Candles))
	}

	code, body = do(t, client, "GET", "/api/users/leaderboard?range=7d", "", "")
	if code != 200 {
		t.Fatalf("leaderboard code %d: %s", code, body)
	}
	if strings.Contains(string(body), alpha.APIKey) || strings.Contains(string(body), beta.APIKey) {
		t.Fatalf("leaderboard leaked api keys: %s", body)
	}

	var leaderboard publicLeaderboardResponse
	if err := json.Unmarshal(body, &leaderboard); err != nil {
		t.Fatalf("unmarshal leaderboard: %v (%s)", err, body)
	}
	if len(leaderboard.Users) != 2 {
		t.Fatalf("expected two leaderboard users, got %+v", leaderboard.Users)
	}
	if leaderboard.Users[0].Rank != 1 || leaderboard.Users[0].User.ID != alpha.UserID {
		t.Fatalf("expected alpha first by workload, got %+v", leaderboard.Users)
	}
	if len(leaderboard.Users[0].Trend) != 7 {
		t.Fatalf("expected weekly trend points, got %d", len(leaderboard.Users[0].Trend))
	}
}

func TestCLISearchFallsBackWithoutTypesense(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	code, body := do(t, client, "GET", "/api/cli/search?q=cod&limit=5", "", "")
	if code != 200 {
		t.Fatalf("cli search code %d: %s", code, body)
	}
	var result cliSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if result.Count == 0 {
		t.Fatalf("expected CLI results: %s", body)
	}
	found := false
	for _, item := range result.CLIs {
		if item.Name == "codex" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected codex in CLI results: %+v", result.CLIs)
	}
}

func TestSkillSearchFallsBackWithoutTypesense(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	code, body := do(t, client, "GET", "/skills/search?q=github&limit=5", "", "")
	if code != 200 {
		t.Fatalf("skill search code %d: %s", code, body)
	}
	var result skillSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 || result.Skills[0].Name != "github-tools" {
		t.Fatalf("unexpected skill search result: %+v", result)
	}
}

func TestProviderSearchFallsBackWithoutTypesense(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	code, body := do(t, client, "GET", "/api/providers/search?q=open&limit=5", "", "")
	if code != 200 {
		t.Fatalf("provider search code %d: %s", code, body)
	}
	var result providerSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	foundOpenAI := false
	for _, item := range result.Providers {
		if item.Key == "openai" {
			foundOpenAI = true
		}
	}
	if !foundOpenAI {
		t.Fatalf("expected openai provider in results: %+v", result.Providers)
	}
}

func TestCatalogSearchAggregatesProvidersCLIsSkillsAndCustom(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	code, body := do(t, client, "GET", "/api/catalog/search?q=&category=all&limit=50", "", "")
	if code != 200 {
		t.Fatalf("catalog search code %d: %s", code, body)
	}
	var result catalogSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, item := range result.Items {
		seen[item.Type] = true
	}
	for _, typ := range []string{"provider", "cli", "skill", "custom_provider"} {
		if !seen[typ] {
			t.Fatalf("expected catalog type %s in %+v", typ, result.Items)
		}
	}
}

func TestBillingEndpoints(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	_, body := do(t, client, "POST", "/register", "", `{"username":"billinguser","email":"billing@example.com","password":"pw"}`)
	var reg registerResponse
	if err := json.Unmarshal(body, &reg); err != nil {
		t.Fatal(err)
	}

	code, body := do(t, client, "GET", "/billing/balance", reg.APIKey, "")
	if code != 200 {
		t.Fatalf("balance code %d: %s", code, body)
	}
	var balance billingBalanceResponse
	if err := json.Unmarshal(body, &balance); err != nil {
		t.Fatal(err)
	}
	if balance.Balance != model.DefaultRegistrationCredits || balance.MarginMode != model.MarginUnlimited {
		t.Fatalf("unexpected initial balance: %+v", balance)
	}

	code, body = do(t, client, "POST", "/billing/topup", reg.APIKey, `{"amount":25}`)
	if code != 200 {
		t.Fatalf("topup code %d: %s", code, body)
	}
	var topup transactionResponse
	if err := json.Unmarshal(body, &topup); err != nil {
		t.Fatal(err)
	}
	if topup.Amount != 25 || topup.BalanceAfter != 125 || topup.Reason != model.TransactionReasonTopUp {
		t.Fatalf("unexpected top-up response: %+v", topup)
	}

	code, body = do(t, client, "POST", "/billing/lefine-reward", reg.APIKey, `{"amount":10}`)
	if code != 200 {
		t.Fatalf("lefine reward code %d: %s", code, body)
	}
	var reward transactionResponse
	if err := json.Unmarshal(body, &reward); err != nil {
		t.Fatal(err)
	}
	if reward.BalanceAfter != 135 || reward.Reason != model.TransactionReasonLefineReward {
		t.Fatalf("unexpected reward response: %+v", reward)
	}

	code, body = do(t, client, "PATCH", "/billing/settings", reg.APIKey,
		`{"margin_mode":"safe","safe_margin_limit":5,"auto_pay_interval":"week","auto_pay_day":2}`)
	if code != 200 {
		t.Fatalf("settings code %d: %s", code, body)
	}
	if err := json.Unmarshal(body, &balance); err != nil {
		t.Fatal(err)
	}
	if balance.MarginMode != model.MarginSafe || balance.SafeMarginLimit != 5 || balance.AutoPayInterval != model.AutoPayWeekly || balance.AutoPayDay != 2 {
		t.Fatalf("settings were not applied: %+v", balance)
	}

	code, body = do(t, client, "POST", "/billing/usage", reg.APIKey,
		`{"cpu_seconds":12,"memory_mb_hours":24,"disk_mb":48,"load_percent":150,"standard_payment":30}`)
	if code != 200 {
		t.Fatalf("usage code %d: %s", code, body)
	}
	var usage usageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		t.Fatal(err)
	}
	if usage.Transaction == nil || usage.Transaction.Amount != 45 || usage.Transaction.BalanceAfter != 90 {
		t.Fatalf("unexpected usage response: %+v", usage)
	}

	code, body = do(t, client, "GET", "/billing/transactions", reg.APIKey, "")
	if code != 200 {
		t.Fatalf("transactions code %d: %s", code, body)
	}
	var transactions []transactionResponse
	if err := json.Unmarshal(body, &transactions); err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 4 {
		t.Fatalf("expected registration, top-up, reward, usage transactions; got %d: %+v", len(transactions), transactions)
	}
}

func TestEnvironmentRequiresNonNegativeBalanceForNewAgent(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	_, body := do(t, client, "POST", "/register", "", `{"username":"debtuser","email":"debt@example.com","password":"pw"}`)
	var reg registerResponse
	_ = json.Unmarshal(body, &reg)

	code, body := do(t, client, "POST", "/billing/usage", reg.APIKey,
		`{"cpu_seconds":1,"memory_mb_hours":1,"disk_mb":1,"load_percent":100,"standard_payment":150}`)
	if code != 200 {
		t.Fatalf("usage code %d: %s", code, body)
	}

	code, body = do(t, client, "POST", "/environment", reg.APIKey, `{"agent":{"cli":"claude-code"}}`)
	if code != fasthttp.StatusPaymentRequired {
		t.Fatalf("expected 402 for new environment with debt, got %d: %s", code, body)
	}
}

func TestChatRequiresAuth(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()
	if code, _ := do(t, client, "POST", "/api/chat", "", `{"prompt":"hi"}`); code != 401 {
		t.Fatalf("expected 401 without token, got %d", code)
	}
	if code, _ := do(t, client, "POST", "/api/chat", "bad-token", `{"prompt":"hi"}`); code != 401 {
		t.Fatalf("expected 401 with bad token, got %d", code)
	}
}

func TestCLIModeChat(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	_, body := do(t, client, "POST", "/register", "", `{"username":"cliclient","email":"c@d.com","password":"pw"}`)
	var reg registerResponse
	_ = json.Unmarshal(body, &reg)

	do(t, client, "POST", "/environment", reg.APIKey, `{"llm":{"api_key":"sk"},"agent":{"cli":"claude-code"}}`)

	code, body := do(t, client, "POST", "/api/chat", reg.APIKey, `{"prompt":"go"}`)
	if code != 200 {
		t.Fatalf("chat code %d: %s", code, body)
	}
	var chat chatResponse
	_ = json.Unmarshal(body, &chat)
	if chat.Response != "cli:go" {
		t.Fatalf("expected cli route, got %q", chat.Response)
	}
}
