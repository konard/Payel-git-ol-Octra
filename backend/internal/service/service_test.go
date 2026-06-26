package service

import (
	"context"
	"testing"

	"backend/internal/cli"
	"backend/internal/config"
	"backend/internal/llm"
	"backend/internal/model"
	"backend/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testCfg() config.Config {
	return config.Config{
		JWTSecret:              "test-jwt-secret",
		JWTRefreshSecret:       "test-jwt-refresh-secret",
		LeFineIntegrationSecret: "test-lefine-secret",
	}
}

// newTestDB returns a migrated in-memory SQLite database.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// --- auth -------------------------------------------------------------------

func TestRegisterAndAuthenticate(t *testing.T) {
	db := newTestDB(t)
	users := repository.NewUserRepository(db)
	svc := NewAuthService(users, testCfg())
	ctx := context.Background()

	user, err := svc.Register(ctx, "testuser", "Test@Example.com", "secret")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("email not normalised: %q", user.Email)
	}
	if user.APIKey == "" {
		t.Fatal("expected an api key")
	}

	got, err := svc.Authenticate(ctx, user.APIKey)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if got.ID != user.ID {
		t.Fatal("authenticated wrong user")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	db := newTestDB(t)
	svc := NewAuthService(repository.NewUserRepository(db), testCfg())
	ctx := context.Background()

	if _, err := svc.Register(ctx, "user1", "a@b.com", "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(ctx, "user2", "a@b.com", "y"); err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	db := newTestDB(t)
	svc := NewAuthService(repository.NewUserRepository(db), testCfg())
	if _, err := svc.Authenticate(context.Background(), "nope"); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

// --- environment ------------------------------------------------------------

// fakeProvisioner records nix interactions.
type fakeProvisioner struct {
	created   []model.CLIType
	skills    []string
	failSkill string
}

func (f *fakeProvisioner) CreateEnvironment(_ context.Context, _ string, c model.CLIType) error {
	f.created = append(f.created, c)
	return nil
}
func (f *fakeProvisioner) InstallSkill(_ context.Context, _ string, s model.Skill) error {
	f.skills = append(f.skills, s.Name)
	if s.Name == f.failSkill {
		return context.DeadlineExceeded
	}
	return nil
}

func seedUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()
	u, err := NewAuthService(repository.NewUserRepository(db), testCfg()).Register(context.Background(), "seeduser", "u@e.com", "pw")
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func seedSkill(t *testing.T, db *gorm.DB, name string, typ model.SkillType) {
	t.Helper()
	if err := db.Create(&model.Skill{Name: name, Type: typ}).Error; err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentCreate(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	seedSkill(t, db, "filesystem", model.SkillBuiltin)
	seedSkill(t, db, "github", model.SkillNixpkgs)

	prov := &fakeProvisioner{}
	svc := NewEnvironmentService(
		repository.NewAgentRepository(db),
		repository.NewSkillRepository(db),
		repository.NewUserSkillRepository(db),
		prov,
	)

	agent, err := svc.Create(context.Background(), user, EnvironmentInput{
		LLMAPIKey:  "sk-1",
		LLMBaseURL: "https://api.anthropic.com",
		CLI:        "claude-code",
		Skills:     []string{"filesystem", "github", "unknown-skill"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if agent.CLI != "claude-code" {
		t.Fatalf("agent cli = %q", agent.CLI)
	}
	if len(prov.created) != 1 || prov.created[0] != "claude-code" {
		t.Fatalf("env not created: %v", prov.created)
	}
	if len(prov.skills) != 2 { // unknown-skill is skipped
		t.Fatalf("expected 2 installed skills, got %v", prov.skills)
	}

	links, _ := repository.NewUserSkillRepository(db).ListByAgent(context.Background(), agent.ID)
	if len(links) != 2 {
		t.Fatalf("expected 2 user_skill rows, got %d", len(links))
	}
}

func TestEnvironmentCreateIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	svc := NewEnvironmentService(
		repository.NewAgentRepository(db),
		repository.NewSkillRepository(db),
		repository.NewUserSkillRepository(db),
		&fakeProvisioner{},
	)
	ctx := context.Background()

	a1, err := svc.Create(ctx, user, EnvironmentInput{CLI: "claude-code"})
	if err != nil {
		t.Fatal(err)
	}
	a2, err := svc.Create(ctx, user, EnvironmentInput{CLI: "opencode"})
	if err != nil {
		t.Fatal(err)
	}
	if a1.ID != a2.ID {
		t.Fatal("expected agent to be reused (upsert), got new id")
	}
	if a2.CLI != "opencode" {
		t.Fatalf("expected cli updated to opencode, got %q", a2.CLI)
	}
}

func TestEnvironmentSkillFailureRecordsStatus(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	seedSkill(t, db, "broken", model.SkillCustom)
	prov := &fakeProvisioner{failSkill: "broken"}
	svc := NewEnvironmentService(
		repository.NewAgentRepository(db),
		repository.NewSkillRepository(db),
		repository.NewUserSkillRepository(db),
		prov,
	)
	agent, err := svc.Create(context.Background(), user, EnvironmentInput{Skills: []string{"broken"}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	links, _ := repository.NewUserSkillRepository(db).ListByAgent(context.Background(), agent.ID)
	if len(links) != 1 || links[0].Status != "failed" {
		t.Fatalf("expected failed status, got %+v", links)
	}
}

// --- chat -------------------------------------------------------------------

type fakeCLIRouter struct {
	called bool
	spec   cli.LaunchSpec
}

func (f *fakeCLIRouter) Send(_ context.Context, spec cli.LaunchSpec, prompt string) (string, error) {
	f.called = true
	f.spec = spec
	return "cli:" + prompt, nil
}

type fakeLLM struct {
	called bool
	req    llm.Request
}

func (f *fakeLLM) Complete(_ context.Context, req llm.Request) (string, error) {
	f.called = true
	f.req = req
	return "llm:" + req.Prompt, nil
}

type fakeEnvPaths struct{}

func (fakeEnvPaths) EnvPath(userID string) string { return "/envs/" + userID }

func TestChatRoutesToCLI(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	agents := repository.NewAgentRepository(db)
	if err := agents.Upsert(context.Background(), &model.Agent{UserID: user.ID, CLI: "claude-code", LLMAPIKey: "sk", Active: true}); err != nil {
		t.Fatal(err)
	}

	cliR := &fakeCLIRouter{}
	llmC := &fakeLLM{}
	svc := NewChatService(agents, cliR, llmC, fakeEnvPaths{})

	out, err := svc.Chat(context.Background(), user, "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "cli:hello" {
		t.Fatalf("unexpected output %q", out)
	}
	if !cliR.called || llmC.called {
		t.Fatal("expected CLI router to be used, not LLM")
	}
	if cliR.spec.EnvPath != "/envs/"+user.ID.String() {
		t.Fatalf("env path not wired: %q", cliR.spec.EnvPath)
	}
}

func TestChatRoutesToLLMWhenNoCLI(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	agents := repository.NewAgentRepository(db)
	if err := agents.Upsert(context.Background(), &model.Agent{UserID: user.ID, CLI: "", LLMAPIKey: "sk", LLMBaseURL: "https://x", Active: true}); err != nil {
		t.Fatal(err)
	}

	cliR := &fakeCLIRouter{}
	llmC := &fakeLLM{}
	svc := NewChatService(agents, cliR, llmC, fakeEnvPaths{})

	out, err := svc.Chat(context.Background(), user, "hi", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "llm:hi" {
		t.Fatalf("unexpected output %q", out)
	}
	if cliR.called || !llmC.called {
		t.Fatal("expected LLM to be used, not CLI router")
	}
}

func TestChatWithoutEnvironment(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	svc := NewChatService(repository.NewAgentRepository(db), &fakeCLIRouter{}, &fakeLLM{}, fakeEnvPaths{})
	if _, err := svc.Chat(context.Background(), user, "hi", nil); err != ErrNoEnvironment {
		t.Fatalf("expected ErrNoEnvironment, got %v", err)
	}
}
