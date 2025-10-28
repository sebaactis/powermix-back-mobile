package proof

import "github.com/google/uuid"

type ProofRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	ProofMPID string    `json:"proof_mp_id" validate:"required,max=255"`
}

type ProofResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	ProofMPID string    `json:"proof_mp_id"`
	ProofDate string    `json:"proof_date"`
}
