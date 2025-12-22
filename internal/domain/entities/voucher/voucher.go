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
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID
	QRCode       string
	StoragePath  string
	IsAssigned   bool
	AssignedDate time.Time

	Status VoucherStatus `gorm:"type:varchar(20);not null;default:'ACTIVE'"`

	UsedAt        *time.Time `gorm:"default:null"`
	LastCheckedAt *time.Time `gorm:"default:null"`
}
