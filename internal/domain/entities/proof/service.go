package proof

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mercadopago"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type Service struct {
	repo           *Repository
	userService    *user.Service
	voucherService *voucher.Service
	validator      validations.StructValidator
	mpClient       *mercadopago.Client
}

func NewService(repo *Repository, userService *user.Service, voucherService *voucher.Service, validator validations.StructValidator, mpClient *mercadopago.Client) *Service {
	return &Service{repo: repo, userService: userService, voucherService: voucherService, validator: validator, mpClient: mpClient}
}

func (s *Service) Create(ctx context.Context, proof *ProofRequest) (*ProofResponse, error) {

	if fields, ok := s.validator.ValidateStruct(proof); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	proofExistsValidate, err := s.GetById(ctx, proof.ID_MP)

	if err != nil {
		return nil, err
	}

	if proofExistsValidate != nil {
		return nil, fmt.Errorf("ya guardaste un comprobante con este ID: %s", proof.ID_MP)
	}

	payment, err := s.mpClient.ValidatePaymentExists(ctx, proof.ID_MP)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("el comprobante %s no existe en Mercado Pago", proof.ID_MP)
	}

	newProof := &Proof{
		UserID:            proof.UserID,
		ID_MP:             proof.ID_MP,
		Date_Approved_MP:  utils.FormattedTime{Time: payment.DateApproved.Truncate(time.Second)},
		Operation_Type_MP: payment.OperationType,
		Status_MP:         payment.Status,
		Amount_MP:         payment.TotalPaidAmount,
		ProofDate:         utils.NowFormatted(),
	}

	proofResult, err := s.repo.Create(ctx, newProof)

	if err != nil {
		return nil, err
	}

	quantityStamps, err := s.userService.IncrementStampsCounter(ctx, proofResult.UserID)

	log.Printf("✅ proofResult: %+v", proofResult)
	log.Printf("✅ quantityStamps: %+v", quantityStamps)

	if err != nil {
		return nil, err
	}

	if quantityStamps == 10 {

		_, err = s.voucherService.Create(ctx, &voucher.VoucherRequest{UserID: proofResult.UserID})

		if err != nil {
			return nil, err
		}

		_, err = s.userService.ResetStampsCounter(ctx, proofResult.UserID)

		if err != nil {
			return nil, err
		}
	}

	return &ProofResponse{
		UserID:            proofResult.UserID,
		ID_MP:             proofResult.ID_MP,
		ProofDate:         proofResult.ProofDate,
		Status_MP:         proofResult.Status_MP,
		Date_Approved_MP:  proofResult.Date_Approved_MP,
		Operation_Type_MP: proofResult.Operation_Type_MP,
		Amount_MP:         proofResult.Amount_MP,
	}, nil

}

func (s *Service) CreateFromOthers(ctx context.Context, req *ProofOthersRequest) (*ProofResponse, error) {

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	mpReq := mercadopago.ReconcileOthersRequest{
		Date:   req.Date,
		Time:   req.Time,
		Amount: req.Amount,
		Last4:  req.Last4,
		DNI:    req.DNI,
	}

	result, err := s.mpClient.ReconcileOthers(ctx, mpReq)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("no se encontró un pago en Mercado Pago que coincida con el comprobante")
	}

	idMP := fmt.Sprintf("%d", result.PaymentID)

	proofExistsValidate, err := s.GetById(ctx, idMP)
	if err != nil {
		return nil, err
	}
	if proofExistsValidate != nil {
		return nil, fmt.Errorf("ya guardaste un comprobante con este pago de Mercado Pago (ID: %s)", idMP)
	}

	newProof := &Proof{
		UserID:            req.UserID,
		ID_MP:             idMP,
		Date_Approved_MP:  utils.FormattedTime{Time: result.DateApproved.Truncate(time.Second)},
		Operation_Type_MP: "Others",
		Status_MP:         result.Status,
		Amount_MP:         result.TotalPaidAmount,
		ProofDate:         utils.NowFormatted(),
	}

	proofResult, err := s.repo.Create(ctx, newProof)
	if err != nil {
		return nil, err
	}

	quantityStamps, err := s.userService.IncrementStampsCounter(ctx, proofResult.UserID)
	log.Printf("✅ proofResult (others): %+v", proofResult)
	log.Printf("✅ quantityStamps (others): %+v", quantityStamps)
	if err != nil {
		return nil, err
	}

	if quantityStamps == 10 {
		_, err = s.voucherService.Create(ctx, &voucher.VoucherRequest{UserID: proofResult.UserID})
		if err != nil {
			return nil, err
		}
		_, err = s.userService.ResetStampsCounter(ctx, proofResult.UserID)
		if err != nil {
			return nil, err
		}
	}

	return &ProofResponse{
		UserID:            proofResult.UserID,
		ID_MP:             proofResult.ID_MP,
		ProofDate:         proofResult.ProofDate,
		Status_MP:         proofResult.Status_MP,
		Date_Approved_MP:  proofResult.Date_Approved_MP,
		Operation_Type_MP: proofResult.Operation_Type_MP,
		Amount_MP:         proofResult.Amount_MP,
	}, nil
}

func (s *Service) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]*ProofResponse, error) {

	var proofsResponse []*ProofResponse

	proofs, err := s.repo.GetAllByUserId(ctx, userId)

	if err != nil {
		return nil, err
	}

	for i := range proofs {
		proofsResponse = append(proofsResponse, &ProofResponse{
			UserID:            proofs[i].UserID,
			ID_MP:             proofs[i].ID_MP,
			ProofDate:         proofs[i].ProofDate,
			Status_MP:         proofs[i].Status_MP,
			Date_Approved_MP:  proofs[i].Date_Approved_MP,
			Operation_Type_MP: proofs[i].Operation_Type_MP,
			Amount_MP:         proofs[i].Amount_MP,
		})
	}

	return proofsResponse, nil

}

func (s *Service) GetById(ctx context.Context, id string) (*ProofResponse, error) {
	if id == "" {
		return nil, errors.New("id es requerido")
	}

	proof, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if proof == nil {
		return nil, nil
	}

	return &ProofResponse{
		UserID:            proof.UserID,
		ID_MP:             proof.ID_MP,
		ProofDate:         proof.ProofDate,
		Status_MP:         proof.Status_MP,
		Date_Approved_MP:  proof.Date_Approved_MP,
		Operation_Type_MP: proof.Operation_Type_MP,
		Amount_MP:         proof.Amount_MP,
	}, nil
}
