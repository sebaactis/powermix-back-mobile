package token

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

var ErrRefreshReuseDetected = errors.New("refresh_reuse_detected")
var ErrRefreshInvalid = errors.New("refresh_invalid")

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

func (s *Service) CreateInitialRefreshToken(ctx context.Context, userID uuid.UUID, rawToken string, expiresAt time.Time) (*Token, error) {
	familyID := uuid.New()

	t := &Token{
		UserID:     userID,
		TokenType:  string(jwtx.TokenTypeRefresh),
		TokenHash:  HashToken(s.pepper, rawToken),
		FamilyID:   familyID,
		ParentID:   nil,
		ExpiresAt:  expiresAt,
		Is_Revoked: false,
	}

	return s.repository.Create(ctx, t)
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


func (s *Service) RotateRefresh(ctx context.Context, rawRefresh string, now time.Time, signAccess func() (string, time.Time, error), signRefresh func() (string, time.Time, error),
) (newAccess string, accessExp time.Time, newRefresh string, refreshExp time.Time, err error) {

	err = s.Transaction(ctx, func(sTx *Service) error {
		hash := HashToken(sTx.pepper, rawRefresh)

		current, err := sTx.repository.GetRefreshByHashForUpdate(ctx, hash)
		if err != nil {
			return ErrRefreshInvalid
		}

		if !current.ExpiresAt.IsZero() && current.ExpiresAt.Before(now) {
			return ErrRefreshInvalid
		}

		if current.Is_Revoked {
			_ = sTx.repository.RevokeFamily(ctx, current.FamilyID, now, "reuse_detected")
			return ErrRefreshReuseDetected
		}

		newAccess, accessExp, err = signAccess()
		if err != nil {
			return err
		}

		newRefresh, refreshExp, err = signRefresh()
		if err != nil {
			return err
		}

		next := &Token{
			UserID:     current.UserID,
			TokenType:  string(jwtx.TokenTypeRefresh),
			TokenHash:  HashToken(sTx.pepper, newRefresh),
			FamilyID:   current.FamilyID,
			ParentID:   &current.ID,
			ExpiresAt:  refreshExp,
			Is_Revoked: false,
		}

		created, err := sTx.repository.Create(ctx, next)
		if err != nil {
			return err
		}

		if err := sTx.repository.MarkRotated(ctx, current.ID, created.ID, now); err != nil {
			return err
		}

		return nil
	})

	return
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
