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
	"time"

	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

var (
	ErrEmailTaken     = errors.New("email already registered")
	ErrUsernameTaken  = errors.New("username already taken")
	ErrInvalidToken   = errors.New("invalid api token")
	ErrInvalidJWT     = errors.New("invalid or expired token")
	ErrInvalidRefresh = errors.New("invalid refresh token")
	ErrUserNotFound   = errors.New("user not found")
)

// AuthService handles registration, JWT, API-token authentication and OAuth.
type AuthService struct {
	users repository.UserRepository
	cfg   config.Config
}

// NewAuthService builds an AuthService.
func NewAuthService(users repository.UserRepository, cfg config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

// --- API key auth (for MCP clients) -----------------------------------------

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

// --- Registration & password login ------------------------------------------

// Register creates a new account and returns it with a freshly minted API key.
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email and password are required")
	}

	if existing, err := s.users.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrEmailTaken
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if existing, err := s.users.GetByUsername(ctx, username); err == nil && existing != nil {
		return nil, ErrUsernameTaken
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
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		APIKey:       apiKey,
		Subscription: "free",
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// LoginUser authenticates by email+password and returns JWT tokens.
func (s *AuthService) LoginUser(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errors.New("invalid email or password")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return s.generateTokens(user)
}

// --- JWT helpers ------------------------------------------------------------

type LoginResult struct {
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	User         *LoginUserInfo      `json:"user"`
}

type LoginUserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

func (s *AuthService) generateTokens(user *model.User) (*LoginResult, error) {
	now := time.Now()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"exp":      now.Add(AccessTokenExpiry).Unix(),
		"iat":      now.Unix(),
		"type":     "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"exp":      now.Add(RefreshTokenExpiry).Unix(),
		"iat":      now.Unix(),
		"type":     "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWTRefreshSecret))
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		User: &LoginUserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// ValidateAccessToken parses and validates an access JWT.
func (s *AuthService) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidJWT
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidJWT
	}
	return claims, nil
}

// ValidateRefreshToken parses and validates a refresh JWT.
func (s *AuthService) ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTRefreshSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidRefresh
	}
	return claims, nil
}

// GetMe returns current user info from an access token.
func (s *AuthService) GetMe(ctx context.Context, tokenString string) (*UserInfo, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidJWT
	}

	user, err := s.users.GetByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return userToInfo(user), nil
}

// RefreshTokens generates new token pair from a valid refresh token.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshTokenString string) (*LoginResult, error) {
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	user, err := s.users.GetByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

// UserInfo is the public profile returned by /me.
type UserInfo struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	Email           string  `json:"email"`
	HasSubscription bool    `json:"has_subscription"`
	SubscriptionEnd *int64  `json:"subscription_end"`
}

func userToInfo(user *model.User) *UserInfo {
	now := time.Now()
	hasSubscription := user.SubscriptionEnd != nil && time.Unix(*user.SubscriptionEnd, 0).After(now)
	return &UserInfo{
		UserID:          user.ID.String(),
		Username:        user.Username,
		Email:           user.Email,
		HasSubscription: hasSubscription,
		SubscriptionEnd: user.SubscriptionEnd,
	}
}

// --- OAuth user creation ----------------------------------------------------

func (s *AuthService) GetOrCreateUserFromGoogle(ctx context.Context, email, name string) (*model.User, error) {
	return s.getOrCreateUser(ctx, email, name)
}

func (s *AuthService) GetOrCreateUserFromGitHub(ctx context.Context, email, name string) (*model.User, error) {
	return s.getOrCreateUser(ctx, email, name)
}

func (s *AuthService) GetOrCreateUserFromLeFine(ctx context.Context, email, name string) (*model.User, error) {
	return s.getOrCreateUser(ctx, email, name)
}

func (s *AuthService) getOrCreateUser(ctx context.Context, email, name string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	username := name
	if username == "" {
		if len(email) > 10 {
			username = email[:len(email)-10]
		} else {
			username = "user_" + uuid.New().String()[:8]
		}
	}

	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	user = &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: "",
		APIKey:       apiKey,
		Subscription: "free",
	}

	if err := s.users.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			user.Username = username + "_" + uuid.New().String()[:6]
			if err := s.users.Create(ctx, user); err != nil {
				return nil, errors.New("failed to create account")
			}
		} else {
			return nil, errors.New("failed to create account")
		}
	}

	return user, nil
}

// LoginResultFromUser builds a LoginResult without needing a password check.
func (s *AuthService) LoginResultFromUser(ctx context.Context, user *model.User) (*LoginResult, error) {
	return s.generateTokens(user)
}

// --- Helpers ----------------------------------------------------------------

func GenerateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "octra_" + hex.EncodeToString(b), nil
}
