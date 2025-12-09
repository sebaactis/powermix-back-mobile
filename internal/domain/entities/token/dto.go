package token

import (
	"time"

	"github.com/google/uuid"
)

type TokenRequest struct {
	TokenType string    `json:"token_type" validate:"required,max=30"`
	Token     string    `json:"token" validate:"required,max=1000"`
	UserId    uuid.UUID `json:"user_id" validate:"required,uuid"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TokenResponse struct {
	TokenType   string `json:"token_type"`
	Token       string `json:"token"`
	RevokedDate string `json:"revoked_date"`
	IsRevoked   bool   `json:"is_revoked"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ToResponse(t *Token) *TokenResponse {
	return &TokenResponse{
		TokenType:   t.TokenType,
		Token:       t.Token,
		RevokedDate: t.Revoked_Date.Format("2006-01-02 15:04:05"),
		IsRevoked:   t.Is_Revoked,
	}
}

func ToResponseMany(tokens []*Token) []*TokenResponse {

	response := make([]*TokenResponse, len(tokens))

	for i, token := range tokens {
		response[i] = ToResponse(token)
	}

	return response

}
