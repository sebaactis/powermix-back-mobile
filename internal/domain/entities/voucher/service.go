package voucher

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/coffeeji"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"gorm.io/gorm"
)

type Service struct {
	repo           *Repository
	userRepository *user.Repository
	mailer         mailer.Mailer
	coffejiClient  *coffeeji.Client
}

func NewService(repo *Repository, userRepository *user.Repository, mailer mailer.Mailer, coffejiClient *coffeeji.Client) *Service {
	return &Service{
		repo:           repo,
		userRepository: userRepository,
		mailer:         mailer,
		coffejiClient:  coffejiClient,
	}
}

// WithTx returns a new Service that uses the given transaction
func (s *Service) WithTx(tx *gorm.DB) *Service {
	return &Service{
		repo:           s.repo.WithTx(tx),
		userRepository: s.userRepository.WithTx(tx),
		mailer:         s.mailer,
		coffejiClient:  s.coffejiClient,
	}
}

func (s *Service) AssignNextVoucher(ctx context.Context, voucherRequest *VoucherRequest) (*VoucherResponse, error) {

	voucherEntity, err := s.repo.AssignNextVoucher(ctx, voucherRequest)

	if err != nil {
		return nil, err
	}

	imageURL := s.GetVoucherImageUrl(voucherEntity.StoragePath)

	voucherResponse := &VoucherResponse{
		UserID:   voucherEntity.UserID,
		QRCode:   voucherEntity.QRCode,
		ImageURL: imageURL,
	}

	user, err := s.userRepository.FindByID(ctx, voucherEntity.UserID)

	if err != nil {
		return nil, err
	}

	s.mailer.SendVoucherEmail(ctx, user.Email, imageURL)

	return voucherResponse, nil
}

func (s *Service) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]*VoucherResponse, error) {
	var voucherResponse []*VoucherResponse

	vouchers, err := s.repo.GetAllByUserId(ctx, userId)

	if err != nil {
		return nil, err
	}

	for i := range vouchers {
		voucherResponse = append(voucherResponse, &VoucherResponse{
			VoucherID:     vouchers[i].ID,
			UserID:        vouchers[i].UserID,
			QRCode:        vouchers[i].QRCode,
			ImageURL:      s.GetVoucherImageUrl(vouchers[i].StoragePath),
			Status:        vouchers[i].Status,
			UsedAt:        vouchers[i].UsedAt,
			LastCheckedAt: vouchers[i].LastCheckedAt,
		})
	}

	return voucherResponse, nil
}

func (s *Service) CheckUsedVouchers(ctx context.Context, batch int) error {
	now := time.Now()

	vouchers, err := s.repo.ListAssignedActive(ctx, batch)
	if err != nil {
		return err
	}

	for _, v := range vouchers {
		vCtx, cancel := context.WithTimeout(ctx, 8*time.Second)

		used, err := s.coffejiClient.ValidateVoucherCode(vCtx, v.QRCode)

		cancel()

		_ = s.repo.TouchChecked(ctx, v.ID, now)

		if err != nil {
			log.Printf("[cron] ValidateVoucherCode error voucherID=%s code=%s err=%v",
				v.ID.String(), v.QRCode, err)
			continue
		}

		if used {
			if err := s.repo.MarkUsed(ctx, v.ID, now); err != nil {
				log.Printf("[cron] MarkUsed failed voucherID=%s err=%v", v.ID.String(), err)
			}
		}
	}

	return nil
}

// -------- PRIVADO -------- //

func (s *Service) GetVoucherImageUrl(storagePath string) string {

	baseURL := os.Getenv("VOUCHER_BUCKET_URL")
	imageURL := fmt.Sprintf("%s/%s", baseURL, storagePath)

	return imageURL

}

func (s *Service) DeleteVoucher(ctx context.Context, voucherID uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteUsedVoucher(ctx, voucherID, userID)
}
