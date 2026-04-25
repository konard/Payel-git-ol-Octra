package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		*s = []string{}
		return nil
	}
	return json.Unmarshal(bytes, s)
}

type PromoCode struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id;primaryKey"`
	Code      string     `json:"code" gorm:"uniqueIndex;not null"`
	Duration  int        `json:"duration" gorm:"not null"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	UsedBy    StringList `json:"used_by" gorm:"type:text"`
	CreatedAt time.Time  `json:"created_at"`
}
