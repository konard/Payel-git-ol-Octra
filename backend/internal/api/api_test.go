package api

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"backend/internal/cli"
	"backend/internal/llm"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- fakes ------------------------------------------------------------------

type fakeProvisioner struct{}

func (fakeProvisioner) CreateEnvironment(context.Context, string, model.CLIType) error { return nil }
func (fakeProvisioner) InstallSkill(context.Context, string, model.Skill) error        { return nil }

type fakeCLIRouter struct{}

func (fakeCLIRouter) Send(_ context.Context, _ cli.LaunchSpec, prompt string) (string, error) {
	return "cli:" + prompt, nil
}

type fakeLLM struct{}

func (fakeLLM) Complete(_ context.Context, req llm.Request) (string, error) {
	return "llm:" + req.Prompt, nil
}

type fakeEnvPaths struct{}

func (fakeEnvPaths) EnvPath(id string) string { return "/envs/" + id }

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

	billingSvc := service.NewBillingService(users, agents, transactions, usageMetrics)
	authSvc := service.NewAuthService(users, transactions)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, fakeProvisioner{}, billingSvc)
	chatSvc := service.NewChatService(agents, fakeCLIRouter{}, fakeLLM{}, fakeEnvPaths{})

	handler := New(authSvc, envSvc, chatSvc, billingSvc).Router().Handler

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

	// health
	if code, _ := do(t, client, "GET", "/health", "", ""); code != 200 {
		t.Fatalf("health code %d", code)
	}

	// register
	code, body := do(t, client, "POST", "/register", "", `{"email":"a@b.com","password":"pw"}`)
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

	// chat before environment -> 400
	if code, _ := do(t, client, "POST", "/api/chat", reg.APIKey, `{"prompt":"hi"}`); code != 400 {
		t.Fatalf("expected 400 before env, got %d", code)
	}

	// create environment (proxy mode: no cli)
	code, body = do(t, client, "POST", "/environment", reg.APIKey,
		`{"llm":{"api_key":"sk","base_url":"https://api.anthropic.com"},"agent":{"cli":"","priority":3},"skills":[]}`)
	if code != 200 {
		t.Fatalf("environment code %d: %s", code, body)
	}

	// chat -> proxy
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

func TestBillingEndpoints(t *testing.T) {
	client, cleanup := newTestServer(t)
	defer cleanup()

	_, body := do(t, client, "POST", "/register", "", `{"email":"billing@example.com","password":"pw"}`)
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

	_, body := do(t, client, "POST", "/register", "", `{"email":"debt@example.com","password":"pw"}`)
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

	_, body := do(t, client, "POST", "/register", "", `{"email":"c@d.com","password":"pw"}`)
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
