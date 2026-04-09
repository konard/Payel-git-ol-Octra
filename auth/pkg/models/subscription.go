package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan string

const (
	PlanMonthly    SubscriptionPlan = "monthly"
	PlanQuarterly  SubscriptionPlan = "quarterly"
	PlanSemiYearly SubscriptionPlan = "semi_annually"
	PlanAnnually   SubscriptionPlan = "annually"
)

type Subscription struct {
	Id        uuid.UUID        `json:"id" gorm:"column:id;primaryKey"`
	UserId    uuid.UUID        `json:"user_id" gorm:"index;not null"`
	Plan      SubscriptionPlan `json:"plan" gorm:"not null"`
	StartDate time.Time        `json:"start_date" gorm:"not null"`
	EndDate   time.Time        `json:"end_date" gorm:"not null"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
