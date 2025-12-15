package voucher

import (
	"time"

	"github.com/google/uuid"
)

type Voucher struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID
	QRCode       string
	StoragePath  string
	IsAssigned   bool
	AssignedDate time.Time
}
