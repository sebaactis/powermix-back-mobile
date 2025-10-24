package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name          string    `gorm:"not null"`
	Email         string    `gorm:"not null;unique"`
	Password      string    `gorm:"not null"`
	StampsCounter int       `gorm:"default:0"`
	LoginAttempt  int       `json:"login_attempt" gorm:"default:0"`
	Locked_until  time.Time `json:"locked_until" gorm:"default:null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
