package voucher

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNoAvailableVouchers = errors.New("no hay vouchers disponibles en este momento")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
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
