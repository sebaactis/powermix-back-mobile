package proof

import (

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type ProofRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	ID_MP  string    `json:"proof_mp_id" validate:"required,max=255"`
}

type ProofResponse struct {
	UserID            uuid.UUID           `json:"user_id"`
	ID_MP             string              `json:"proof_mp_id"`
	ProofDate         utils.FormattedTime `json:"proof_date"`
	Date_Approved_MP  utils.FormattedTime `json:"date_approved_mp"`
	Operation_Type_MP string              `json:"operation_type_mp"`
	Status_MP         string              `json:"status_mp"`
	Amount_MP         float64             `json:"amount_mp"`
}

type ProofOthersRequest struct {
    UserID uuid.UUID `json:"user_id" validate:"required"`
    Date   string    `json:"date" validate:"required"`   // "18/11/2025"
    Time   string    `json:"time" validate:"required"`   // "12:11" o "12.11"
    Amount float64   `json:"amount" validate:"required,gt=0"`

    Last4 *string `json:"last4,omitempty"` // últimos 4 de la tarjeta
    DNI   *string `json:"dni,omitempty"`   // dni del titular
}
