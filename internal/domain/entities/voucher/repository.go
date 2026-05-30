package voucher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNoAvailableVouchers = errors.New("no hay vouchers disponibles en este momento")
var ErrVoucherNotFound = errors.New("voucher no encontrado")
var ErrVoucherNotBelongsToUser = errors.New("el voucher no pertenece al usuario")
var ErrVoucherNotUsed = errors.New("solo se pueden eliminar vouchers usados")
var ErrInternal = errors.New("voucher: error interno de persistencia")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// WithTx devuelve un nuevo Repository que usa la transacción que le pasamos
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// DB expone la conexión subyacente para manejo de transacciones
func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) AssignNextVoucher(ctx context.Context, voucherRequest *VoucherRequest) (*Voucher, error) {
	var result *Voucher

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var v Voucher

		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("is_assigned = ?", false).
			Order("id").
			First(&v).Error; err != nil {

			return mapVoucherAssignErr(ctx, "assign next", err)
		}

		now := time.Now()
		v.IsAssigned = true
		v.UserID = voucherRequest.UserID
		v.AssignedDate = now

		if err := tx.Save(&v).Error; err != nil {
			return mapVoucherAssignErr(ctx, "assign next save", err)
		}

		result = &v
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Voucher, error) {
	var result []*Voucher

	tx := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&result)

	if tx.Error != nil {
		return nil, mapVoucherRepoErr(ctx, "get all by user id", tx.Error)
	}

	return result, nil
}

func (r *Repository) ListAssignedActive(ctx context.Context, limit int) ([]*Voucher, error) {
	var v []*Voucher
	err := r.db.WithContext(ctx).
		Where("is_assigned = ?", true).
		Where("status = ?", VoucherStatusActive).
		Limit(limit).
		Find(&v).Error
	if err != nil {
		return nil, mapVoucherRepoErr(ctx, "list assigned active", err)
	}
	return v, nil
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID, now time.Time) error {
	updates := map[string]any{
		"status":  VoucherStatusUsed,
		"used_at": &now,
	}
	if err := r.db.WithContext(ctx).
		Model(&Voucher{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return mapVoucherRepoErr(ctx, "mark used", err)
	}
	return nil
}

func (r *Repository) TouchChecked(ctx context.Context, id uuid.UUID, now time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&Voucher{}).
		Where("id = ?", id).
		Update("last_checked_at", &now).Error; err != nil {
		return mapVoucherRepoErr(ctx, "touch checked", err)
	}
	return nil
}

func (r *Repository) DeleteUsedVoucher(ctx context.Context, voucherID uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", voucherID).
		Where("user_id = ?", userID).
		Where("status = ?", VoucherStatusUsed).
		Delete(&Voucher{})

	if result.Error != nil {
		return mapVoucherRepoErr(ctx, "delete used voucher", result.Error)
	}

	if result.RowsAffected == 0 {
		var v Voucher
		err := r.db.WithContext(ctx).Where("id = ?", voucherID).First(&v).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("voucher: delete used: %w", ErrVoucherNotFound)
		}
		if err != nil {
			return mapVoucherRepoErr(ctx, "delete used voucher lookup", err)
		}
		if v.UserID != userID {
			return fmt.Errorf("voucher: delete used: %w", ErrVoucherNotBelongsToUser)
		}
		return fmt.Errorf("voucher: delete used: %w", ErrVoucherNotUsed)
	}

	return nil
}

func (r *Repository) CountAvailable(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Voucher{}).
		Where("is_assigned = ?", false).
		Count(&count).Error
	if err != nil {
		return 0, mapVoucherRepoErr(ctx, "count available", err)
	}
	return count, nil
}

func mapVoucherAssignErr(ctx context.Context, action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("voucher: %s: %w", action, ErrNoAvailableVouchers)
	}
	slog.ErrorContext(ctx, "voucher repository", "action", action, "error", err)
	return fmt.Errorf("voucher: %s: %w", action, ErrInternal)
}

func mapVoucherRepoErr(ctx context.Context, action string, err error) error {
	if err == nil {
		return nil
	}
	slog.ErrorContext(ctx, "voucher repository", "action", action, "error", err)
	return fmt.Errorf("voucher: %s: %w", action, ErrInternal)
}
