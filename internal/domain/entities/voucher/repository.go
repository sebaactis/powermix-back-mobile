package voucher

import (
	"context"

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

func (r *Repository) Create(ctx context.Context, voucherRequest *Voucher) error {
	return r.db.WithContext(ctx).Create(voucherRequest).Error
}
