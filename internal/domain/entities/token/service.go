package token

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type Service struct {
	repository *Repository
	validator  validations.StructValidator
}

func NewService(repository *Repository, v validations.StructValidator) *Service {
	return &Service{repository: repository, validator: v}
}

func (s *Service) Create(ctx context.Context, tokenRequest *TokenRequest) (*Token, error) {
	if fields, ok := s.validator.ValidateStruct(tokenRequest); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	tokenCreate := &Token{
		TokenType: tokenRequest.TokenType,
		Token:     tokenRequest.Token,
		UserID:    tokenRequest.UserId,
		ExpiresAt: tokenRequest.ExpiresAt,
	}

	token, err := s.repository.Create(ctx, tokenCreate)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) CreateResetPasswordToken(ctx context.Context, userID uuid.UUID, rawToken string, expiresAt time.Time) (*Token, error) {
	tokenCreate := &Token{
		UserID:    userID,
		TokenType: string(jwtx.TokenTypeResetPassword),
		Token:     rawToken,
		ExpiresAt: expiresAt,
	}

	token, err := s.repository.Create(ctx, tokenCreate)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) ValidateAndRevokeResetPasswordToken(ctx context.Context, rawToken string) (*Token, error) {
	now := time.Now()

	t, err := s.repository.GetValidResetPasswordToken(ctx, rawToken, now)

	if err != nil {
		return nil, err
	}

	if err := s.repository.RevokeToken(ctx, rawToken); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) GetAll(ctx context.Context) ([]*Token, error) {
	tokens, err := s.repository.GetAll(ctx)

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *Service) RevokeToken(ctx context.Context, token string) error {
	tokenCheck, err := s.repository.GetByToken(ctx, token)
	if err != nil {
		return err
	}

	now := time.Now()

	if tokenCheck.Is_Revoked {
		return errors.New("el token no es válido o ya fue utilizado")
	}

	if !tokenCheck.ExpiresAt.IsZero() && tokenCheck.ExpiresAt.Before(now) {
		return errors.New("el token ha expirado")
	}

	return s.repository.RevokeToken(ctx, token)
}
