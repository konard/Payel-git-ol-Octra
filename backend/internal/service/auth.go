// Package service holds the business logic of the monolith: account
// registration, environment provisioning and request routing.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"backend/internal/model"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// ErrEmailTaken is returned when registering with an already-used email.
var ErrEmailTaken = errors.New("email already registered")

// ErrInvalidToken is returned when an API token matches no user.
var ErrInvalidToken = errors.New("invalid api token")

// AuthService handles registration and API-token authentication.
type AuthService struct {
	users        repository.UserRepository
	transactions repository.TransactionRepository
}

// NewAuthService builds an AuthService.
func NewAuthService(users repository.UserRepository, transactions ...repository.TransactionRepository) *AuthService {
	var txs repository.TransactionRepository
	if len(transactions) > 0 {
		txs = transactions[0]
	}
	return &AuthService{users: users, transactions: txs}
}

// Register creates a new account and returns it with a freshly minted API key.
func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	if existing, err := s.users.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrEmailTaken
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:           email,
		PasswordHash:    string(hash),
		APIKey:          apiKey,
		Subscription:    "free",
		Balance:         model.DefaultRegistrationCredits,
		MarginMode:      model.MarginUnlimited,
		AutoPayInterval: model.AutoPayMonthly,
		AutoPayDay:      1,
	}
	user.ApplyBillingDefaults()
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	if s.transactions != nil {
		if err := s.transactions.Create(ctx, &model.Transaction{
			UserID:       user.ID,
			Type:         model.TransactionCredit,
			Amount:       user.Balance,
			Reason:       model.TransactionReasonRegistration,
			BalanceAfter: user.Balance,
		}); err != nil {
			return nil, err
		}
	}
	return user, nil
}

// Authenticate resolves an API token to its owning user.
func (s *AuthService) Authenticate(ctx context.Context, apiKey string) (*model.User, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, ErrInvalidToken
	}
	user, err := s.users.GetByAPIKey(ctx, apiKey)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GenerateAPIKey returns a random, URL-safe API key prefixed for readability.
func GenerateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "octra_" + hex.EncodeToString(b), nil
}
