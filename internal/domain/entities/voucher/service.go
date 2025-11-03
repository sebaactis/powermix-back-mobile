package voucher

import (
	"context"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, voucherRequest *VoucherRequest) (*VoucherResponse, error) {

	newVoucher := &Voucher{
		UserID:         voucherRequest.UserID,
		QRCode:         "QRCODE_TO_DO",
		GenerationDate: utils.FormattedTime{Time: time.Now().Truncate(time.Second)},
	}

	err := s.repo.Create(ctx, newVoucher)
	if err != nil {
		return nil, err
	}

	voucherResponse := &VoucherResponse{
		UserID:         newVoucher.UserID,
		QRCode:         newVoucher.QRCode,
		GenerationDate: newVoucher.GenerationDate.Time,
	}

	return voucherResponse, nil
}
