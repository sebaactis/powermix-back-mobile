package proof

import (
	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type Proof struct {
	ID                uuid.UUID           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID            uuid.UUID           `gorm:"not null"`
	ID_MP             string              `gorm:"unique, not null"`
	Date_Approved_MP  utils.FormattedTime `gorm:"not null"`
	Operation_Type_MP string              `gorm:"not null"`
	Status_MP         string              `gorm:"not null"`
	Amount_MP         float64             `gorm:"not null"`
	ProofDate         utils.FormattedTime `gorm:"not null"`
	Dni               *string
	CardId            *string
	CardType          *string
	Last4Card         *string
}
