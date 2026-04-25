package services

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
	"user/pkg/database"
	"user/pkg/models"

	"github.com/google/uuid"
)

// SubscriptionPlanConfig хранит информацию о плане подписки
type SubscriptionPlanConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"` // days
	Price    string `json:"price"`
}

// GetSubscriptionPlans возвращает все планы подписок из env
func GetSubscriptionPlans() []SubscriptionPlanConfig {
	// Формат: id:name:duration:price,id:name:duration:price
	// Пример: monthly:1 месяц:30:990,quarterly:3 месяца:90:2490
	raw := os.Getenv("SUBSCRIPTION_PLANS")
	if raw == "" {
		// Default plans - цены в копейках для YooKassa (10$ = 1000₽, 25$ = 2500₽, 50$ = 5000₽, 100$ = 10000₽)
		return []SubscriptionPlanConfig{
			{ID: "monthly", Name: "1 месяц", Duration: 30, Price: "100000"},          // 10$ = 1000₽
			{ID: "quarterly", Name: "3 месяца", Duration: 90, Price: "250000"},       // 25$ = 2500₽
			{ID: "semi_annually", Name: "6 месяцев", Duration: 180, Price: "500000"}, // 50$ = 5000₽
			{ID: "annually", Name: "1 год", Duration: 365, Price: "1000000"},         // 100$ = 10000₽
		}
	}

	var plans []SubscriptionPlanConfig
	for _, planStr := range strings.Split(raw, ",") {
		parts := strings.Split(planStr, ":")
		if len(parts) != 4 {
			continue
		}
		duration, _ := strconv.Atoi(parts[2])
		plans = append(plans, SubscriptionPlanConfig{
			ID:       parts[0],
			Name:     parts[1],
			Duration: duration,
			Price:    parts[3],
		})
	}

	return plans
}

// GetPlanByID returns a plan config by ID
func GetPlanByID(planID string) (*SubscriptionPlanConfig, error) {
	plans := GetSubscriptionPlans()
	for _, plan := range plans {
		if plan.ID == planID {
			return &plan, nil
		}
	}
	return nil, errors.New("plan not found")
}

// SubscribeUser создает подписку для пользователя
func SubscribeUser(userID uuid.UUID, plan string) error {
	planConfig, err := GetPlanByID(plan)
	if err != nil {
		return err
	}

	// Find user
	var user models.UserRegister
	err = database.Db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return errors.New("user not found")
	}

	// Calculate end date
	now := time.Now()
	var endDate time.Time

	// If user already has active subscription, extend it
	if user.SubscriptionEnd != nil && time.Unix(*user.SubscriptionEnd, 0).After(now) {
		endDate = time.Unix(*user.SubscriptionEnd, 0).AddDate(0, 0, planConfig.Duration)
	} else {
		endDate = now.AddDate(0, 0, planConfig.Duration)
	}

	endTimestamp := endDate.Unix()
	user.SubscriptionEnd = &endTimestamp

	// Save user
	err = database.Db.Save(&user).Error
	if err != nil {
		return err
	}

	// Create subscription record
	subscription := models.Subscription{
		Id:        uuid.New(),
		UserId:    userID,
		Plan:      models.SubscriptionPlan(plan),
		StartDate: now,
		EndDate:   endDate,
	}

	return database.Db.Create(&subscription).Error
}

// ActivatePromoCode активирует промокод для пользователя
func ActivatePromoCode(userID uuid.UUID, code string) error {
	// Find promo code
	var promo models.PromoCode
	err := database.Db.Where("code = ? AND is_active = ?", code, true).First(&promo).Error
	if err != nil {
		return errors.New("invalid promo code")
	}

	// Find user
	var user models.UserRegister
	err = database.Db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return errors.New("user not found")
	}

	// Check if user already used this promo
	for _, usedBy := range promo.UsedBy {
		if usedBy == userID.String() {
			return errors.New("promo code already used")
		}
	}

	// Activate subscription for promo duration
	now := time.Now()
	endDate := now.AddDate(0, 0, promo.Duration)
	endTimestamp := endDate.Unix()
	user.SubscriptionEnd = &endTimestamp

	// Save user
	err = database.Db.Save(&user).Error
	if err != nil {
		return err
	}

	// Mark promo as used
	if promo.UsedBy == nil {
		promo.UsedBy = models.StringList{}
	}
	promo.UsedBy = append(promo.UsedBy, userID.String())
	err = database.Db.Save(&promo).Error
	if err != nil {
		return err
	}

	// Create subscription record
	subscription := models.Subscription{
		Id:        uuid.New(),
		UserId:    userID,
		Plan:      "promo",
		StartDate: now,
		EndDate:   endDate,
	}

	return database.Db.Create(&subscription).Error
}

// CheckSubscription проверяет есть ли у пользователя активная подписка
func CheckSubscription(userID uuid.UUID) (bool, error) {
	var user models.UserRegister
	err := database.Db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return false, err
	}

	if user.SubscriptionEnd == nil {
		return false, nil
	}

	return time.Unix(*user.SubscriptionEnd, 0).After(time.Now()), nil
}

// GetUserSubscriptionInfo возвращает информацию о подписке пользователя
func GetUserSubscriptionInfo(userID uuid.UUID) (map[string]interface{}, error) {
	var user models.UserRegister
	err := database.Db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	if user.SubscriptionEnd == nil {
		return map[string]interface{}{
			"has_subscription": false,
		}, nil
	}

	endDate := time.Unix(*user.SubscriptionEnd, 0)
	isActive := endDate.After(time.Now())

	return map[string]interface{}{
		"has_subscription": isActive,
		"subscription_end": endDate,
	}, nil
}

// InitDefaultPromoCodes создает промокоды по умолчанию
func InitDefaultPromoCodes() {
	// Create shard2026 promo code (1 year)
	var existing models.PromoCode
	err := database.Db.Where("code = ?", "shard2026").First(&existing).Error

	if err != nil {
		promo := models.PromoCode{
			Id:       uuid.New(),
			Code:     "shard2026",
			Duration: 365, // 1 year
			IsActive: true,
			UsedBy:   models.StringList{},
		}
		database.Db.Create(&promo)
	}
}
