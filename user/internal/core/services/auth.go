package services

import (
	"errors"
	"os"
	"strings"
	"time"
	"user/pkg/database"
	"user/pkg/models"
	"user/pkg/requests"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 дней
)

// ErrUserAlreadyExists is returned by RegisterUser when the email or username is
// already taken. It lets the HTTP layer respond with 409 Conflict instead of a
// generic 500, regardless of the (localized) message shown to the user.
var ErrUserAlreadyExists = errors.New("Пользователь с таким email или именем уже существует")

// isDuplicateKeyError reports whether a database error was caused by a unique
// constraint violation (duplicate email/username), across Postgres and SQLite
// driver wording.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

// NormalizeEmail trims surrounding whitespace and lowercases an email so that
// registration, login and OAuth all resolve to the same account regardless of
// the casing the user typed.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// GenerateTokens generates access and refresh JWT tokens
func GenerateTokens(user models.UserRegister) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("JWT_SECRET environment variable is not set")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		return "", "", errors.New("JWT_REFRESH_SECRET environment variable is not set")
	}

	// Access Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(AccessTokenExpiry).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(RefreshTokenExpiry).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateAccessToken validates and parses access token
func ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// ValidateRefreshToken validates and parses refresh token
func ValidateRefreshToken(tokenString string) (*jwt.Token, error) {
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		return nil, errors.New("JWT_REFRESH_SECRET environment variable is not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(refreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// LoginUser authenticates user and returns tokens
func LoginUser(req requests.UserLoginRequest) (map[string]interface{}, error) {
	// Find user by email
	var user models.UserRegister
	err := database.Db.Where("email = ?", NormalizeEmail(req.Email)).First(&user).Error
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// OAuth-only accounts have no password set — reject password login explicitly
	if user.Password == "" {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	err = CheckPasswordHash(req.Password, user.Password)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate tokens
	accessToken, refreshToken, err := GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          userPublicInfo(user),
	}, nil
}

// userPublicInfo returns the subset of user fields safe to expose to clients.
func userPublicInfo(user models.UserRegister) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.ID.String(),
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	}
}

// getOrCreateOAuthUser finds an existing user by email or creates a new one,
// generating a unique username when the desired one is already taken. It is the
// shared implementation behind every OAuth/identity provider.
func getOrCreateOAuthUser(email, name string) (*models.UserRegister, error) {
	email = NormalizeEmail(email)

	var user models.UserRegister
	if err := database.Db.Where("email = ?", email).First(&user).Error; err == nil {
		return &user, nil
	}

	baseUsername := strings.TrimSpace(name)
	if baseUsername == "" {
		if len(email) > 10 {
			baseUsername = email[:len(email)-10]
		} else {
			baseUsername = "user_" + uuid.New().String()[:8]
		}
	}

	username := baseUsername
	var count int64
	database.Db.Model(&models.UserRegister{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		username = baseUsername + "_" + uuid.New().String()[:6]
	}

	user = models.UserRegister{
		ID:       uuid.New(),
		Username: username,
		Email:    email,
	}

	if err := database.Db.Create(&user).Error; err != nil {
		if isDuplicateKeyError(err) {
			// Race or username collision — retry once with a random suffix.
			user.Username = baseUsername + "_" + uuid.New().String()[:6]
			if err := database.Db.Create(&user).Error; err != nil {
				return nil, errors.New("Ошибка при создании аккаунта. Попробуйте ещё раз")
			}
		} else {
			return nil, errors.New("Ошибка при создании аккаунта. Попробуйте ещё раз")
		}
	}

	return &user, nil
}

// GetOrCreateUserFromGoogle creates or finds a user from Google OAuth
func GetOrCreateUserFromGoogle(email, name string) (*models.UserRegister, error) {
	return getOrCreateOAuthUser(email, name)
}

// GetOrCreateUserFromGithub creates or finds a user from GitHub OAuth
func GetOrCreateUserFromGithub(email, name string) (*models.UserRegister, error) {
	return getOrCreateOAuthUser(email, name)
}

// GetOrCreateUserFromLeFine creates or finds a user from LeFine/Kefine OAuth-like integration
func GetOrCreateUserFromLeFine(email, name, lefineUserID string) (*models.UserRegister, error) {
	return getOrCreateOAuthUser(email, name)
}

// RefreshTokens generates new tokens from refresh token
func RefreshTokens(req requests.RefreshTokenRequest) (map[string]interface{}, error) {
	// Validate refresh token
	token, err := ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Find user by UUID
	var user models.UserRegister
	userIDStr, _ := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id in token")
	}
	err = database.Db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new tokens
	accessToken, refreshToken, err := GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

// GetMe returns current user info from access token
func GetMe(tokenString string) (map[string]interface{}, error) {
	token, err := ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	var user models.UserRegister
	err = database.Db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, errors.New("user not found")
	}

	hasSubscription := user.SubscriptionEnd != nil && time.Unix(*user.SubscriptionEnd, 0).After(time.Now())
	var subscriptionEnd *time.Time
	if user.SubscriptionEnd != nil {
		t := time.Unix(*user.SubscriptionEnd, 0)
		subscriptionEnd = &t
	}

	return map[string]interface{}{
		"user_id":          user.ID.String(),
		"username":         user.Username,
		"email":            user.Email,
		"created_at":       user.CreatedAt.Format(time.RFC3339),
		"has_subscription": hasSubscription,
		"subscription_end": subscriptionEnd,
	}, nil
}

// RegisterUser registers user and returns tokens
func RegisterUser(req requests.UserRegisterRequest) (map[string]interface{}, error) {
	hashPs, err := HashedPassword(req.Password)
	if err != nil {
		return nil, errors.New("Ошибка при создании аккаунта. Попробуйте ещё раз")
	}

	user := models.UserRegister{
		ID:       uuid.New(),
		Username: strings.TrimSpace(req.Username),
		Email:    NormalizeEmail(req.Email),
		Password: hashPs,
	}

	err = database.Db.Create(&user).Error
	if err != nil {
		println("Register error: " + err.Error())
		if isDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, errors.New("Ошибка при создании аккаунта. Попробуйте ещё раз")
	}

	accessToken, refreshToken, err := GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          userPublicInfo(user),
	}, nil
}
