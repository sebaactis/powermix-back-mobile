package voucher

import (
	"time"

	"github.com/google/uuid"
)

type VoucherRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type VoucherResponse struct {
	VoucherID uuid.UUID `json:"voucher_id"`
	UserID    uuid.UUID `json:"user_id"`
	QRCode    string    `json:"qr_code"`
	ImageURL  string    `json:"image_url"`

	Status VoucherStatus `json:"status"`

	UsedAt        *time.Time `json:"used_at"`
	LastCheckedAt *time.Time `json:"last_checked_at"`
}
