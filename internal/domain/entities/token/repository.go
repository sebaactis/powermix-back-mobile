package token

import (
	"context"
	"errors"
	"time"

	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, token *Token) (*Token, error) {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*Token, error) {
	var tokens []*Token
	if err := r.db.WithContext(ctx).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *Repository) GetByToken(ctx context.Context, tokenIn string) (*Token, error) {
	var token Token

	err := r.db.WithContext(ctx).Where("token = ?", tokenIn).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no se encontró el token proporcionado")
		}
		return nil, errors.New("error inesperado")
	}

	return &token, nil
}

func (r *Repository) GetValidResetPasswordToken(ctx context.Context, tokenIn string, now time.Time) (*Token, error) {
	var t Token

	err := r.db.WithContext(ctx).
		Where("token = ?", tokenIn).
		Where("token_type = ?", string(jwtx.TokenTypeResetPassword)).
		Where("is_revoked = ?", false).
		Where("expires_at > ?", now).
		First(&t).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token inválido o expirado")
		}

		return nil, errors.New("error inesperado al buscar el token de recuperación")
	}

	return &t, nil
}

func (r *Repository) Update(ctx context.Context, token string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&Token{}).
		Where("token = ?", token).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no se encontró el token proporcionado")
	}

	return nil
}

func (r *Repository) RevokeToken(ctx context.Context, token string) error {
	return r.Update(ctx, token, map[string]interface{}{
		"revoked_date": time.Now(),
		"is_revoked":   true,
	})
}
