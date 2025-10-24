package voucher

import (
	"time"

	"github.com/google/uuid"
)

type Voucher struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID         uuid.UUID
	QRCode         string
	GenerationDate time.Time
}
