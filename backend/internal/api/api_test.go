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

	authSvc := service.NewAuthService(users)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, fakeProvisioner{})
	chatSvc := service.NewChatService(agents, fakeCLIRouter{}, fakeLLM{}, fakeEnvPaths{})

	handler := New(authSvc, envSvc, chatSvc).Router().Handler

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

	// chat before environment -> 400
	if code, _ := do(t, client, "POST", "/api/chat", reg.APIKey, `{"prompt":"hi"}`); code != 400 {
		t.Fatalf("expected 400 before env, got %d", code)
	}

	// create environment (proxy mode: no cli)
	code, body = do(t, client, "POST", "/environment", reg.APIKey,
		`{"llm":{"api_key":"sk","base_url":"https://api.anthropic.com"},"agent":{"cli":""},"skills":[]}`)
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
