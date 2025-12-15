package voucher

import (
	"github.com/google/uuid"
)

type VoucherRequest struct {
	UserID uuid.UUID
}

type VoucherResponse struct {
	UserID   uuid.UUID
	QRCode   string
	ImageURL string
}
