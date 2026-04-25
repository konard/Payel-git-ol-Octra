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

// GenerateTokens generates access and refresh JWT tokens
func GenerateTokens(user models.UserRegister) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = "default-refresh-secret-change-in-production"
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
		secret = "default-secret-change-in-production"
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
		refreshSecret = "default-refresh-secret-change-in-production"
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
	err := database.Db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
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
		"user": map[string]interface{}{
			"id":       user.ID.String(),
			"username": user.Username,
			"email":    user.Email,
		},
	}, nil
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
		Username: req.Username,
		Email:    req.Email,
		Password: hashPs,
	}

	err = database.Db.Create(&user).Error
	if err != nil {
		println("Register error: " + err.Error())
		// Check for duplicate email/username
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("Пользователь с таким email или именем уже существует")
		}
		return nil, errors.New("Ошибка при создании аккаунта. Попробуйте ещё раз")
	}

	// Generate tokens
	accessToken, refreshToken, err := GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"id":       user.ID.String(),
			"username": user.Username,
			"email":    user.Email,
		},
	}, nil
}
