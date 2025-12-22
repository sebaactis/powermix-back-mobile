package token

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
    ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
    TokenType string    `json:"token_type" gorm:"size:30;not null;index"`

    TokenHash string  `json:"token_hash" gorm:"size:64;not null;uniqueIndex"`

    FamilyID      uuid.UUID  `json:"family_id" gorm:"type:uuid;not null;index"`
    ParentID      *uuid.UUID `json:"parent_id" gorm:"type:uuid;default:null;index"`
    ReplacedByID  *uuid.UUID `json:"replaced_by_id" gorm:"type:uuid;default:null;index"`
    UsedAt        *time.Time `json:"used_at" gorm:"default:null"`
    RevokedReason *string    `json:"revoked_reason" gorm:"size:30;default:null"`

    Revoked_Date time.Time `json:"revoked_date" gorm:"default:null"`
    Is_Revoked   bool      `json:"is_revoked" gorm:"default:false;index"`
    ExpiresAt    time.Time `json:"expires_at" gorm:"index"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
