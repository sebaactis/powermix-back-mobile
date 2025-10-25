package token

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TokenType    string    `json:"token_type" gorm:"size:30;not null"`
	Token        string    `json:"token" gorm:"size:1000;not null"`
	Revoked_Date time.Time `json:"revoked_date" gorm:"default:null"`
	Is_Revoked   bool      `json:"is_revoked" gorm:"default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
