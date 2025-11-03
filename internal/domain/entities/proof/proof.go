package proof

import (

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type Proof struct {
	ID                uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID            uuid.UUID
	ID_MP             string `gorm:"unique"`
	Date_Approved_MP  utils.FormattedTime
	Operation_Type_MP string
	Status_MP         string
	Amount_MP         float64
	ProofDate         utils.FormattedTime
}
