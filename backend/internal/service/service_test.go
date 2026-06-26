package service

import (
	"context"
	"testing"

	"backend/internal/cli"
	"backend/internal/llm"
	"backend/internal/model"
	"backend/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
	transactions := repository.NewTransactionRepository(db)
	svc := NewAuthService(users, transactions)
	ctx := context.Background()

	user, err := svc.Register(ctx, "Test@Example.com", "secret")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("email not normalised: %q", user.Email)
	}
	if user.APIKey == "" {
		t.Fatal("expected an api key")
	}
	if user.Balance != model.DefaultRegistrationCredits {
		t.Fatalf("expected registration credits %d, got %d", model.DefaultRegistrationCredits, user.Balance)
	}
	if user.MarginMode != model.MarginUnlimited {
		t.Fatalf("expected default margin mode %q, got %q", model.MarginUnlimited, user.MarginMode)
	}
	if user.AutoPayInterval != model.AutoPayMonthly {
		t.Fatalf("expected default auto-pay interval %q, got %q", model.AutoPayMonthly, user.AutoPayInterval)
	}

	got, err := svc.Authenticate(ctx, user.APIKey)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if got.ID != user.ID {
		t.Fatal("authenticated wrong user")
	}

	list, err := transactions.ListByUserID(ctx, user.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected registration transaction, got %d rows", len(list))
	}
	if list[0].Reason != model.TransactionReasonRegistration || list[0].Amount != model.DefaultRegistrationCredits {
		t.Fatalf("unexpected registration transaction: %+v", list[0])
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	db := newTestDB(t)
	svc := NewAuthService(repository.NewUserRepository(db))
	ctx := context.Background()

	if _, err := svc.Register(ctx, "a@b.com", "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(ctx, "a@b.com", "y"); err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	db := newTestDB(t)
	svc := NewAuthService(repository.NewUserRepository(db))
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
	u, err := NewAuthService(repository.NewUserRepository(db), repository.NewTransactionRepository(db)).Register(context.Background(), "u@e.com", "pw")
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
		Priority:   7,
		Skills:     []string{"filesystem", "github", "unknown-skill"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if agent.CLI != "claude-code" {
		t.Fatalf("agent cli = %q", agent.CLI)
	}
	if agent.Priority != 7 {
		t.Fatalf("agent priority = %d", agent.Priority)
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

func TestEnvironmentCreateBlockedForNegativeBalance(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	user.Balance = -1
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	agents := repository.NewAgentRepository(db)
	billing := NewBillingService(
		repository.NewUserRepository(db),
		agents,
		repository.NewTransactionRepository(db),
		repository.NewUsageMetricsRepository(db),
	)
	svc := NewEnvironmentService(
		agents,
		repository.NewSkillRepository(db),
		repository.NewUserSkillRepository(db),
		&fakeProvisioner{},
		billing,
	)

	if _, err := svc.Create(context.Background(), user, EnvironmentInput{CLI: "claude-code"}); err != ErrBalanceNegative {
		t.Fatalf("expected ErrBalanceNegative, got %v", err)
	}
}

func TestEnvironmentUpdateAllowedForExistingAgentWithNegativeBalance(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	agents := repository.NewAgentRepository(db)
	if err := agents.Upsert(context.Background(), &model.Agent{UserID: user.ID, CLI: "claude-code", Active: true}); err != nil {
		t.Fatal(err)
	}
	user.Balance = -5
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	billing := NewBillingService(
		repository.NewUserRepository(db),
		agents,
		repository.NewTransactionRepository(db),
		repository.NewUsageMetricsRepository(db),
	)
	svc := NewEnvironmentService(
		agents,
		repository.NewSkillRepository(db),
		repository.NewUserSkillRepository(db),
		&fakeProvisioner{},
		billing,
	)

	agent, err := svc.Create(context.Background(), user, EnvironmentInput{CLI: "opencode"})
	if err != nil {
		t.Fatalf("existing environment should still be configurable: %v", err)
	}
	if agent.CLI != "opencode" {
		t.Fatalf("expected existing agent update, got %q", agent.CLI)
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

// --- billing ---------------------------------------------------------------

func newBillingService(db *gorm.DB) *BillingService {
	return NewBillingService(
		repository.NewUserRepository(db),
		repository.NewAgentRepository(db),
		repository.NewTransactionRepository(db),
		repository.NewUsageMetricsRepository(db),
	)
}

func TestBillingTopUpAndUnlimitedHostingCharge(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	billing := newBillingService(db)
	ctx := context.Background()

	topup, err := billing.Credit(ctx, user.ID, 50, model.TransactionReasonTopUp, nil)
	if err != nil {
		t.Fatalf("Credit: %v", err)
	}
	if topup.Type != model.TransactionCredit || topup.BalanceAfter != 150 {
		t.Fatalf("unexpected top-up transaction: %+v", topup)
	}

	charge, err := billing.ApplyHostingCharge(ctx, user.ID, nil, 180)
	if err != nil {
		t.Fatalf("ApplyHostingCharge: %v", err)
	}
	if charge.Type != model.TransactionDebit || charge.BalanceAfter != -30 {
		t.Fatalf("unexpected charge transaction: %+v", charge)
	}

	var reloaded model.User
	if err := db.First(&reloaded, "id = ?", user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if reloaded.Balance != -30 {
		t.Fatalf("expected negative balance in unlimited mode, got %d", reloaded.Balance)
	}
}

func TestBillingDoesNotResetZeroBalanceOnTopUp(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	user.Balance = 0
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	tx, err := newBillingService(db).Credit(context.Background(), user.ID, 10, model.TransactionReasonTopUp, nil)
	if err != nil {
		t.Fatalf("Credit: %v", err)
	}
	if tx.BalanceAfter != 10 {
		t.Fatalf("expected top-up from zero to 10, got %+v", tx)
	}
}

func TestSafeMarginChargeSuspendsAgentInsteadOfGoingBelowLimit(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	user.Balance = 20
	user.MarginMode = model.MarginSafe
	user.SafeMarginLimit = 0
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	agents := repository.NewAgentRepository(db)
	agent := &model.Agent{UserID: user.ID, CLI: "claude-code", Active: true, Priority: 1}
	if err := agents.Upsert(context.Background(), agent); err != nil {
		t.Fatal(err)
	}

	billing := newBillingService(db)
	tx, err := billing.ApplyHostingCharge(context.Background(), user.ID, &agent.ID, 25)
	if err != ErrSafeMarginLimit {
		t.Fatalf("expected ErrSafeMarginLimit, got tx=%+v err=%v", tx, err)
	}
	if tx != nil {
		t.Fatalf("safe-margin rejection should not create a debit transaction: %+v", tx)
	}

	var reloadedUser model.User
	if err := db.First(&reloadedUser, "id = ?", user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if reloadedUser.Balance != 20 {
		t.Fatalf("safe margin should preserve balance, got %d", reloadedUser.Balance)
	}
	reloadedAgent, err := agents.GetByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloadedAgent.Active {
		t.Fatal("expected safe-margin failure to suspend the agent")
	}
}

func TestRecordUsageCalculatesAndDebitsHostingCharge(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db)
	billing := newBillingService(db)

	usage, tx, err := billing.RecordUsage(context.Background(), user.ID, UsageInput{
		CPUSeconds:     120,
		MemoryMBHours:  256,
		DiskMB:         512,
		LoadPercent:    150,
		StandardCharge: 40,
	})
	if err != nil {
		t.Fatalf("RecordUsage: %v", err)
	}
	if usage.LoadPercent != 150 {
		t.Fatalf("usage load percent = %d", usage.LoadPercent)
	}
	if tx == nil || tx.Amount != 60 || tx.BalanceAfter != 40 {
		t.Fatalf("expected 60 credit charge, got %+v", tx)
	}
}

func TestCalculateHostingCharge(t *testing.T) {
	tests := []struct {
		name     string
		standard int
		load     float64
		average  float64
		want     int
	}{
		{name: "average load pays standard", standard: 100, load: 100, average: 100, want: 100},
		{name: "half average is cheaper", standard: 100, load: 50, average: 100, want: 50},
		{name: "above average is more expensive", standard: 100, load: 150, average: 100, want: 150},
		{name: "zero average falls back to standard", standard: 100, load: 10, average: 0, want: 100},
		{name: "zero load is free", standard: 100, load: 0, average: 100, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateHostingCharge(tt.standard, tt.load, tt.average); got != tt.want {
				t.Fatalf("CalculateHostingCharge() = %d, want %d", got, tt.want)
			}
		})
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
