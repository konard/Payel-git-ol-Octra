package service

import (
	"context"
	"errors"
	"math"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrBalanceNegative       = errors.New("balance is negative; top up before starting a new agent")
	ErrInvalidBillingAmount  = errors.New("amount must be positive")
	ErrSafeMarginLimit       = errors.New("safe margin limit reached; agent suspended until balance is topped up")
	ErrInvalidBillingSetting = errors.New("invalid billing setting")
)

// UsageInput is the resource snapshot accepted by the billing service.
type UsageInput struct {
	Date           time.Time
	CPUSeconds     int64
	MemoryMBHours  int64
	DiskMB         int64
	LoadPercent    int
	StandardCharge int
	AgentID        *uuid.UUID
}

// BillingSettingsInput is a partial update for user billing preferences.
type BillingSettingsInput struct {
	MarginMode      *model.MarginMode
	SafeMarginLimit *int
	AutoPayInterval *model.AutoPayInterval
	AutoPayDay      *int
}

// BillingService owns credit balance, usage and margin rules.
type BillingService struct {
	users        repository.UserRepository
	agents       repository.AgentRepository
	transactions repository.TransactionRepository
	usage        repository.UsageMetricsRepository
}

// NewBillingService builds a BillingService.
func NewBillingService(
	users repository.UserRepository,
	agents repository.AgentRepository,
	transactions repository.TransactionRepository,
	usage repository.UsageMetricsRepository,
) *BillingService {
	return &BillingService{users: users, agents: agents, transactions: transactions, usage: usage}
}

// GetBalance returns the latest persisted user row with billing fields.
func (s *BillingService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.users.GetByID(ctx, userID)
}

// CanCreateEnvironment enforces the issue rule: debt blocks starting new
// agents, while already-running agents are handled by chat/environment update
// paths.
func (s *BillingService) CanCreateEnvironment(ctx context.Context, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Balance < 0 {
		return ErrBalanceNegative
	}
	return nil
}

// Credit adds credits for top-ups, LeFine rewards, or registration bonuses.
func (s *BillingService) Credit(ctx context.Context, userID uuid.UUID, amount int, reason model.TransactionReason, agentID *uuid.UUID) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidBillingAmount
	}
	if reason == "" {
		reason = model.TransactionReasonTopUp
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Balance += amount
	user.ApplyBillingDefaults()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	tx := &model.Transaction{
		UserID:       userID,
		Type:         model.TransactionCredit,
		Amount:       amount,
		Reason:       reason,
		AgentID:      agentID,
		BalanceAfter: user.Balance,
	}
	if err := s.transactions.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// ApplyHostingCharge debits resource usage. Unlimited mode may create debt.
// Safe mode preserves SafeMarginLimit; when the charge cannot be paid, the
// user's current agent is suspended and no debit is recorded.
func (s *BillingService) ApplyHostingCharge(ctx context.Context, userID uuid.UUID, agentID *uuid.UUID, amount int) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidBillingAmount
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.ApplyBillingDefaults()

	nextBalance := user.Balance - amount
	if user.MarginMode == model.MarginSafe && nextBalance < user.SafeMarginLimit {
		if err := s.suspendCurrentAgent(ctx, userID); err != nil {
			return nil, err
		}
		return nil, ErrSafeMarginLimit
	}

	user.Balance = nextBalance
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	tx := &model.Transaction{
		UserID:       userID,
		Type:         model.TransactionDebit,
		Amount:       amount,
		Reason:       model.TransactionReasonHosting,
		AgentID:      agentID,
		BalanceAfter: user.Balance,
	}
	if err := s.transactions.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// RecordUsage stores resource usage and optionally applies the matching hosting
// charge when StandardCharge is provided.
func (s *BillingService) RecordUsage(ctx context.Context, userID uuid.UUID, in UsageInput) (*model.UsageMetric, *model.Transaction, error) {
	date := in.Date
	if date.IsZero() {
		date = time.Now().UTC()
	}
	date = time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)

	metric := &model.UsageMetric{
		UserID:        userID,
		Date:          date,
		CPUSeconds:    in.CPUSeconds,
		MemoryMBHours: in.MemoryMBHours,
		DiskMB:        in.DiskMB,
		LoadPercent:   in.LoadPercent,
	}
	if err := s.usage.Create(ctx, metric); err != nil {
		return nil, nil, err
	}

	charge := CalculateHostingChargeFromPercent(in.StandardCharge, in.LoadPercent)
	if charge <= 0 {
		return metric, nil, nil
	}
	tx, err := s.ApplyHostingCharge(ctx, userID, in.AgentID, charge)
	if err != nil {
		return metric, nil, err
	}
	return metric, tx, nil
}

// ListTransactions returns the user's ledger in newest-first order.
func (s *BillingService) ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	return s.transactions.ListByUserID(ctx, userID, limit, offset)
}

// UpdateSettings validates and persists margin / auto-pay preferences.
func (s *BillingService) UpdateSettings(ctx context.Context, userID uuid.UUID, in BillingSettingsInput) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.ApplyBillingDefaults()

	if in.MarginMode != nil {
		if !validMarginMode(*in.MarginMode) {
			return nil, ErrInvalidBillingSetting
		}
		user.MarginMode = *in.MarginMode
	}
	if in.SafeMarginLimit != nil {
		user.SafeMarginLimit = *in.SafeMarginLimit
	}
	if in.AutoPayInterval != nil {
		if !validAutoPayInterval(*in.AutoPayInterval) {
			return nil, ErrInvalidBillingSetting
		}
		user.AutoPayInterval = *in.AutoPayInterval
	}
	if in.AutoPayDay != nil {
		if *in.AutoPayDay < 1 || *in.AutoPayDay > 31 {
			return nil, ErrInvalidBillingSetting
		}
		user.AutoPayDay = *in.AutoPayDay
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *BillingService) suspendCurrentAgent(ctx context.Context, userID uuid.UUID) error {
	agent, err := s.agents.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	agent.Active = false
	return s.agents.Update(ctx, agent)
}

func validMarginMode(mode model.MarginMode) bool {
	return mode == model.MarginUnlimited || mode == model.MarginSafe
}

func validAutoPayInterval(interval model.AutoPayInterval) bool {
	switch interval {
	case model.AutoPayDaily, model.AutoPayWeekly, model.AutoPayMonthly, model.AutoPayHalfYear, model.AutoPayYearly:
		return true
	default:
		return false
	}
}

// CalculateHostingCharge scales a standard charge by user load relative to the
// platform average. Average load pays the standard amount; lower load is
// cheaper, higher load is more expensive.
func CalculateHostingCharge(standardPayment int, userLoad, averageLoad float64) int {
	if standardPayment <= 0 {
		return 0
	}
	if averageLoad <= 0 {
		return standardPayment
	}
	if userLoad <= 0 {
		return 0
	}
	return int(math.Round(float64(standardPayment) * userLoad / averageLoad))
}

// CalculateHostingChargeFromPercent applies a precomputed load_percent value.
func CalculateHostingChargeFromPercent(standardPayment, loadPercent int) int {
	if standardPayment <= 0 || loadPercent <= 0 {
		return 0
	}
	return int(math.Round(float64(standardPayment) * float64(loadPercent) / 100))
}
