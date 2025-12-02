package proof

import (
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type ProofRequest struct {
	UserID uuid.UUID `json:"user_id"`
	ID_MP  string    `json:"proof_mp_id" validate:"required,max=255"`
}

type ProofOthersRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Date   string    `json:"date" validate:"required"` // "18/11/2025"
	Time   string    `json:"time" validate:"required"` // "12:11" o "12.11"
	Amount float64   `json:"amount" validate:"required,gt=0"`

	Last4       *string `json:"last4,omitempty"`
	DNI         *string `json:"dni,omitempty"`
}

type ProofResponse struct {
	UserID            uuid.UUID           `json:"user_id"`
	ID_MP             string              `json:"proof_mp_id"`
	ProofDate         utils.FormattedTime `json:"proof_date"`
	Date_Approved_MP  utils.FormattedTime `json:"date_approved_mp"`
	Operation_Type_MP string              `json:"operation_type_mp"`
	Status_MP         string              `json:"status_mp"`
	Amount_MP         float64             `json:"amount_mp"`
	Dni               *string             `json:"dni,omitempty"`
	CardType          *string             `json:"card_type,omitempty"`
	Last4Card         *string             `json:"last4_card,omitempty"`
	ExternalID        *string             `json:"external_id,omitempty"`
	ProductName       *string             `json:"product_name,omitempty"`
}

type PaginatedProofResponse struct {
	Items    []*ProofResponse `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Total    int64            `json:"total"`
	HasMore  bool             `json:"hasMore"`
}

type ProofFilters struct {
	ID_MP         string
	FromProofDate *time.Time
	ToProofDate   *time.Time
	MinAmount     *float64
	MaxAmount     *float64
}
