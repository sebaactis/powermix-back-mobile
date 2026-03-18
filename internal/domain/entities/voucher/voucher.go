package voucher

import (
	"time"

	"github.com/google/uuid"
)

type VoucherStatus string

const (
	VoucherStatusActive VoucherStatus = "ACTIVE"
	VoucherStatusUsed   VoucherStatus = "USED"
)

type Voucher struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"column:user_id" json:"user_id"`
	QRCode       string    `gorm:"column:qr_code" json:"qr_code"`
	StoragePath  string    `gorm:"column:storage_path" json:"storage_path"`
	IsAssigned   bool      `gorm:"column:is_assigned" json:"is_assigned"`
	AssignedDate time.Time `gorm:"column:assigned_date" json:"assigned_date"`

	Status VoucherStatus `gorm:"type:varchar(20);not null;default:'ACTIVE';column:status" json:"status"`

	UsedAt        *time.Time `gorm:"column:used_at;default:null" json:"used_at"`
	LastCheckedAt *time.Time `gorm:"column:last_checked_at;default:null" json:"last_checked_at"`
}
