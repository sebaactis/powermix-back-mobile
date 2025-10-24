package token

import (
	"context"
	"errors"
	"time"

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
			return nil, errors.New("token not found")
		}
		return nil, errors.New("unexpected error")
	}

	return &token, nil
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
		return errors.New("Token not found")
	}

	return nil
}

func (r *Repository) RevokeToken(ctx context.Context, token string) error {
	return r.Update(ctx, token, map[string]interface{}{
		"revoked_date": time.Now(),
		"is_revoked":   true,
	})
}
