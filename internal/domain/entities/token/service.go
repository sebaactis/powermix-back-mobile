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
	pepper     []byte
}

func NewService(repository *Repository, v validations.StructValidator, tokenHashSecret string) *Service {
	return &Service{
		repository: repository,
		validator:  v,
		pepper:     []byte(tokenHashSecret),
	}
}

func (s *Service) Create(ctx context.Context, tokenRequest *TokenRequest) (*Token, error) {
	if fields, ok := s.validator.ValidateStruct(tokenRequest); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	tokenHash := HashToken(s.pepper, tokenRequest.Token)

	tokenCreate := &Token{
		TokenType: tokenRequest.TokenType,
		TokenHash: tokenHash,
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
	tokenHash := HashToken(s.pepper, rawToken)

	tokenCreate := &Token{
		UserID:    userID,
		TokenType: string(jwtx.TokenTypeResetPassword),
		TokenHash: tokenHash,
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

	hashToken := HashToken(s.pepper, rawToken)

	if err := s.repository.RevokeToken(ctx, hashToken); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) ValidateRefreshToken(ctx context.Context, rawToken string) (*Token, error) {
	return s.repository.GetValidRefreshToken(ctx, rawToken, time.Now())
}

func (s *Service) GetAll(ctx context.Context) ([]*Token, error) {
	tokens, err := s.repository.GetAll(ctx)

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *Service) RevokeToken(ctx context.Context, token string) error {
	hashToken := HashToken(s.pepper, token)
	tokenCheck, err := s.repository.GetByToken(ctx, hashToken)
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

	return s.repository.RevokeToken(ctx, hashToken)
}

func (s *Service) RevokeRefreshIfValid(ctx context.Context, tokenIn string, now time.Time) (bool, error) {
	hashToken := HashToken(s.pepper, tokenIn)
	return s.repository.RevokeRefreshIfValid(ctx, hashToken, now)
}

func (s *Service) WithRepo(repo *Repository) *Service {
	return &Service{
		repository: repo,
		validator:  s.validator,
		pepper:     s.pepper,
	}
}

func (s *Service) Transaction(ctx context.Context, fn func(sTx *Service) error) error {
	return s.repository.Transaction(ctx, func(rTx *Repository) error {
		return fn(s.WithRepo(rTx))
	})
}
