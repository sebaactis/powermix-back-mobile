package voucher

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNoAvailableVouchers = errors.New("no hay vouchers disponibles en este momento")
var ErrVoucherNotFound = errors.New("voucher no encontrado")
var ErrVoucherNotBelongsToUser = errors.New("el voucher no pertenece al usuario")
var ErrVoucherNotUsed = errors.New("solo se pueden eliminar vouchers usados")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// WithTx returns a new Repository that uses the given transaction
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// DB exposes the underlying db connection for transaction management
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

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoAvailableVouchers
			}
			return err
		}

		now := time.Now()
		v.IsAssigned = true
		v.UserID = voucherRequest.UserID
		v.AssignedDate = now

		if err := tx.Save(&v).Error; err != nil {
			return err
		}

		result = &v
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]*Voucher, error) {
	var result []*Voucher

	tx := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Find(&result)

	if tx.Error != nil {
		return nil, tx.Error
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
	return v, err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID, now time.Time) error {
	updates := map[string]any{
		"status":  VoucherStatusUsed,
		"used_at": &now,
	}
	return r.db.WithContext(ctx).
		Model(&Voucher{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) TouchChecked(ctx context.Context, id uuid.UUID, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&Voucher{}).
		Where("id = ?", id).
		Update("last_checked_at", &now).Error
}

func (r *Repository) DeleteUsedVoucher(ctx context.Context, voucherID uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", voucherID).
		Where("user_id = ?", userID).
		Where("status = ?", VoucherStatusUsed).
		Delete(&Voucher{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		var v Voucher
		err := r.db.WithContext(ctx).Where("id = ?", voucherID).First(&v).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVoucherNotFound
		}
		if v.UserID != userID {
			return ErrVoucherNotBelongsToUser
		}
		return ErrVoucherNotUsed
	}

	return nil
}
