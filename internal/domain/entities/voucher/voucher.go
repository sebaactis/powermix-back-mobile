package voucher

import (

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type Voucher struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID         uuid.UUID
	QRCode         string
	GenerationDate utils.FormattedTime
}
