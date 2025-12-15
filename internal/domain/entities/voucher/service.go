package voucher

import (
	"context"
	"fmt"
	"os"

	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
)

type Service struct {
	repo *Repository
	userRepository *user.Repository
	mailer mailer.Mailer
}

func NewService(repo *Repository, userRepository *user.Repository, mailer mailer.Mailer) *Service {
	return &Service{
		repo: repo,
		userRepository: userRepository,
		mailer: mailer,
	}
}

func (s *Service) AssignNextVoucher(ctx context.Context, voucherRequest *VoucherRequest) (*VoucherResponse, error) {

	voucherEntity, err := s.repo.AssignNextVoucher(ctx, voucherRequest)

	if err != nil {
		return nil, err
	}

	baseURL := os.Getenv("VOUCHER_BUCKER_URL")
	imageURL := fmt.Sprintf("%s/%s", baseURL, voucherEntity.StoragePath)

	voucherResponse := &VoucherResponse{
		UserID: voucherEntity.UserID,
		QRCode: voucherEntity.QRCode,
		ImageURL: imageURL,
	}

	user, err := s.userRepository.FindByID(ctx, voucherEntity.UserID)

	if err != nil {
		return nil, err
	}

	s.mailer.SendVoucherEmail(ctx, user.Email, imageURL)

	return voucherResponse, nil
}
