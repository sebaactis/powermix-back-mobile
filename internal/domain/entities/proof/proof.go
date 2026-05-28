package proof

import (
	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type Proof struct {
	ID              uuid.UUID           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID          uuid.UUID           `gorm:"not null"`
	IDMP            string              `json:"id_mp" gorm:"column:id_mp;unique;not null"`
	DateApprovedMP  utils.FormattedTime `json:"date_approved_mp" gorm:"column:date_approved_mp;not null"`
	OperationTypeMP string              `json:"operation_type_mp" gorm:"column:operation_type_mp;not null"`
	StatusMP        string              `json:"status_mp" gorm:"column:status_mp;not null"`
	AmountMP        float64             `json:"amount_mp" gorm:"column:amount_mp;not null"`
	ProofDate       utils.FormattedTime `gorm:"not null"`
	Dni             *string
	CardID          *string
	CardType        *string
	Last4Card       *string
	ExternalID      *string
	ProductName     *string
}
