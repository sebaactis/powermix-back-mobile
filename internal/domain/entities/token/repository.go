package token

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{
		db: tx,
	}
}

func (r *Repository) Transaction(ctx context.Context, fn func(rTx *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithTx(tx))
	})
}

func (r *Repository) Create(ctx context.Context, token *Token) (*Token, error) {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, mapTokenRepoErr("create", err)
	}
	return token, nil
}


func (r *Repository) GetByToken(ctx context.Context, tokenIn string) (*Token, error) {
	var token Token

	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenIn).First(&token).Error
	if err != nil {
		return nil, mapTokenRepoErr("get by hash", err)
	}

	return &token, nil
}

func (r *Repository) GetValidResetPasswordToken(ctx context.Context, tokenIn string, now time.Time) (*Token, error) {
	var t Token

	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenIn).
		Where("token_type = ?", string(jwtx.TokenTypeResetPassword)).
		Where("is_revoked = ?", false).
		Where("expires_at > ?", now).
		First(&t).Error

	if err != nil {
		return nil, mapResetTokenRepoErr("get valid reset token", err)
	}

	return &t, nil
}


func (r *Repository) Update(ctx context.Context, token string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&Token{}).
		Where("token_hash = ?", token).
		Updates(updates)

	if result.Error != nil {
		return mapTokenRepoErr("update", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("token: update: %w", ErrTokenNotFound)
	}

	return nil
}

func (r *Repository) RevokeToken(ctx context.Context, token string) error {
	return r.Update(ctx, token, map[string]interface{}{
		"revoked_date": time.Now(),
		"is_revoked":   true,
	})
}


func (r *Repository) GetRefreshByHashForUpdate(ctx context.Context, tokenHash string) (*Token, error) {
	var t Token

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", tokenHash).
		Where("token_type = ?", string(jwtx.TokenTypeRefresh)).
		First(&t).Error

	if err != nil {
		return nil, mapTokenRepoErr("get refresh by hash for update", err)
	}

	return &t, nil
}

func (r *Repository) RevokeFamily(ctx context.Context, familyID uuid.UUID, now time.Time, reason string) error {
	result := r.db.WithContext(ctx).
		Model(&Token{}).
		Where("family_id = ?", familyID).
		Where("token_type = ?", string(jwtx.TokenTypeRefresh)).
		Where("is_revoked = ?", false).
		Updates(map[string]any{
			"revoked_date":   now,
			"is_revoked":     true,
			"revoked_reason": reason,
		})

	if result.Error != nil {
		return mapTokenRepoErr("revoke family", result.Error)
	}
	return nil
}

func (r *Repository) MarkRotated(ctx context.Context, tokenID uuid.UUID, replacedBy uuid.UUID, now time.Time) error {
	reason := "rotated"
	result := r.db.WithContext(ctx).
		Model(&Token{}).
		Where("id = ?", tokenID).
		Updates(map[string]any{
			"is_revoked":     true,
			"revoked_date":   now,
			"replaced_by_id": replacedBy,
			"used_at":        now,
			"revoked_reason": reason,
		})

	if result.Error != nil {
		return mapTokenRepoErr("mark rotated", result.Error)
	}
	return nil
}

func mapTokenRepoErr(action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("token: %s: %w", action, ErrTokenNotFound)
	}
	slog.Error("token repository", "action", action, "error", err)
	return fmt.Errorf("token: %s: %w", action, ErrInternal)
}

func mapResetTokenRepoErr(action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("token: %s: %w", action, ErrTokenInvalid)
	}
	slog.Error("token repository", "action", action, "error", err)
	return fmt.Errorf("token: %s: %w", action, ErrInternal)
}
