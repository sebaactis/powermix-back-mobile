package proof

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/coffeeji"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mercadopago"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
	"gorm.io/gorm"
)

type Service struct {
	repo           *Repository
	userService    *user.Service
	voucherService *voucher.Service
	validator      validations.StructValidator
	mpClient       *mercadopago.Client
	coffejiClient  *coffeeji.Client
}

func NewService(repo *Repository, userService *user.Service, voucherService *voucher.Service, validator validations.StructValidator, mpClient *mercadopago.Client, coffejiClient *coffeeji.Client) *Service {
	return &Service{repo: repo, userService: userService, voucherService: voucherService, validator: validator, mpClient: mpClient, coffejiClient: coffejiClient}
}

func (s *Service) Create(ctx context.Context, proof *ProofRequest) (*ProofResponse, error) {

	if fields, ok := s.validator.ValidateStruct(proof); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	proofExistsValidate, err := s.GetById(ctx, proof.IDMP)

	if err != nil {
		return nil, err
	}

	if proofExistsValidate != nil {
		return nil, fmt.Errorf("ya tenes guardado un comprobante con este ID: %s", proof.IDMP)
	}

	payment, err := s.mpClient.ValidatePaymentExists(ctx, proof.IDMP)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("el comprobante %s no existe, por favor, verifique los datos ingresados", proof.IDMP)
	}

	goodsName, err := s.coffejiClient.GetGoodsNameByOrderNo(ctx, *payment.ExternalID)

	if err != nil {
		return nil, err
	}

	newProof := &Proof{
		UserID:          proof.UserID,
		IDMP:            proof.IDMP,
		DateApprovedMP:  utils.FormattedTime{Time: payment.DateApproved.Truncate(time.Second)},
		OperationTypeMP: payment.OperationType,
		StatusMP:        payment.Status,
		AmountMP:        payment.TotalPaidAmount,
		ProofDate:       utils.NowFormatted(),
		Dni:             payment.PayerDNI,
		CardID:          payment.CardId,
		CardType:        payment.CardType,
		Last4Card:       payment.CardLast4,
		ExternalID:      payment.ExternalID,
		ProductName:     &goodsName,
	}

	// Use a transaction to ensure atomicity
	var proofResult *Proof
	var quantityStamps int

	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create repositories/services that use this transaction
		txProofRepo := s.repo.WithTx(tx)
		txUserService := s.userService.WithTx(tx)
		txVoucherService := s.voucherService.WithTx(tx)

		// 1. Create the proof
		var createErr error
		proofResult, createErr = txProofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		// 2. Increment stamps counter
		quantityStamps, createErr = txUserService.IncrementStampsCounter(ctx, proofResult.UserID)
		if createErr != nil {
			return createErr
		}

		// 3. If stamps == 5, assign voucher and reset counter
		if quantityStamps == 5 {
			_, createErr = txVoucherService.AssignNextVoucher(ctx, &voucher.VoucherRequest{
				UserID: proofResult.UserID,
			})
			if createErr != nil {
				return createErr
			}

			_, createErr = txUserService.ResetStampsCounter(ctx, proofResult.UserID)
			if createErr != nil {
				return createErr
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		UserID:          proofResult.UserID,
		IDMP:            proofResult.IDMP,
		ProofDate:       proofResult.ProofDate,
		StatusMP:        proofResult.StatusMP,
		DateApprovedMP:  proofResult.DateApprovedMP,
		OperationTypeMP: proofResult.OperationTypeMP,
		AmountMP:        proofResult.AmountMP,
		Dni:             proofResult.Dni,
		CardType:        proofResult.CardType,
		Last4Card:       proofResult.Last4Card,
		ExternalID:      proofResult.ExternalID,
		ProductName:     proofResult.ProductName,
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
		return nil, fmt.Errorf("no se encontró un pago que coincida con los datos ingresados")
	}

	idMP := fmt.Sprintf("%d", result.PaymentID)

	proofExistsValidate, err := s.GetById(ctx, idMP)
	if err != nil {
		return nil, err
	}
	if proofExistsValidate != nil {
		return nil, fmt.Errorf("ya guardaste un comprobante con este pago de Mercado Pago (ID: %s)", idMP)
	}

	goodsName, err := s.coffejiClient.GetGoodsNameByOrderNo(ctx, *result.ExternalID)

	if err != nil {
		return nil, err
	}

	newProof := &Proof{
		UserID:          req.UserID,
		IDMP:            idMP,
		DateApprovedMP:  utils.FormattedTime{Time: result.DateApproved.Add(time.Hour).Truncate(time.Second)},
		OperationTypeMP: result.OperationType,
		StatusMP:        result.Status,
		AmountMP:        result.TotalPaidAmount,
		ProofDate:       utils.NowFormatted(),
		Dni:             result.PayerDNI,
		CardID:          result.CardId,
		CardType:        result.CardType,
		Last4Card:       result.CardLast4,
		ExternalID:      result.ExternalID,
		ProductName:     &goodsName,
	}

	// Use a transaction to ensure atomicity
	var proofResult *Proof
	var quantityStamps int

	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create repositories/services that use this transaction
		txProofRepo := s.repo.WithTx(tx)
		txUserService := s.userService.WithTx(tx)
		txVoucherService := s.voucherService.WithTx(tx)

		// 1. Create the proof
		var createErr error
		proofResult, createErr = txProofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		// 2. Increment stamps counter
		quantityStamps, createErr = txUserService.IncrementStampsCounter(ctx, proofResult.UserID)
		if createErr != nil {
			return createErr
		}

		log.Printf("proofResult (others): userID=%s idMP=%s amount=%.2f", proofResult.UserID, proofResult.IDMP, proofResult.AmountMP)
		log.Printf("quantityStamps (others): %d", quantityStamps)

		// 3. If stamps == 5, assign voucher and reset counter
		if quantityStamps == 5 {
			_, createErr = txVoucherService.AssignNextVoucher(ctx, &voucher.VoucherRequest{
				UserID: proofResult.UserID,
			})
			if createErr != nil {
				return createErr
			}

			_, createErr = txUserService.ResetStampsCounter(ctx, proofResult.UserID)
			if createErr != nil {
				return createErr
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		UserID:          proofResult.UserID,
		IDMP:            proofResult.IDMP,
		ProofDate:       proofResult.ProofDate,
		StatusMP:        proofResult.StatusMP,
		DateApprovedMP:  proofResult.DateApprovedMP,
		OperationTypeMP: proofResult.OperationTypeMP,
		AmountMP:        proofResult.AmountMP,
		Dni:             proofResult.Dni,
		CardType:        proofResult.CardType,
		Last4Card:       proofResult.Last4Card,
		ExternalID:      proofResult.ExternalID,
		ProductName:     proofResult.ProductName,
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
			UserID:          proofs[i].UserID,
			IDMP:            proofs[i].IDMP,
			ProofDate:       proofs[i].ProofDate,
			StatusMP:        proofs[i].StatusMP,
			DateApprovedMP:  proofs[i].DateApprovedMP,
			OperationTypeMP: proofs[i].OperationTypeMP,
			AmountMP:        proofs[i].AmountMP,
			Dni:             proofs[i].Dni,
			CardType:        proofs[i].CardType,
			Last4Card:       proofs[i].Last4Card,
			ExternalID:      proofs[i].ExternalID,
			ProductName:     proofs[i].ProductName,
		})
	}

	return proofsResponse, nil

}

func (s *Service) GetAllByUserIdPaginated(ctx context.Context, userId uuid.UUID, page int, pageSize int, filters ProofFilters) (*PaginatedProofResponse, error) {
	proofs, total, err := s.repo.GetAllByUserIdPaginated(ctx, userId, page, pageSize, filters)

	if err != nil {
		return nil, err
	}

	proofsResponse := make([]*ProofResponse, len(proofs))

	for i := range proofs {
		proofsResponse[i] = &ProofResponse{
			UserID:          proofs[i].UserID,
			IDMP:            proofs[i].IDMP,
			ProofDate:       proofs[i].ProofDate,
			StatusMP:        proofs[i].StatusMP,
			DateApprovedMP:  proofs[i].DateApprovedMP,
			OperationTypeMP: proofs[i].OperationTypeMP,
			AmountMP:        proofs[i].AmountMP,
			Dni:             proofs[i].Dni,
			CardType:        proofs[i].CardType,
			Last4Card:       proofs[i].Last4Card,
			ExternalID:      proofs[i].ExternalID,
			ProductName:     proofs[i].ProductName,
		}
	}

	resp := &PaginatedProofResponse{
		Items:    proofsResponse,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}

	return resp, nil
}

func (s *Service) GetLastThreeByUserId(ctx context.Context, userId uuid.UUID) ([]*ProofResponse, error) {

	var proofsResponse []*ProofResponse

	proofs, err := s.repo.GetLastThreeByUserId(ctx, userId)

	if err != nil {
		return nil, err
	}

	for i := range proofs {
		proofsResponse = append(proofsResponse, &ProofResponse{
			UserID:          proofs[i].UserID,
			IDMP:            proofs[i].IDMP,
			ProofDate:       proofs[i].ProofDate,
			StatusMP:        proofs[i].StatusMP,
			DateApprovedMP:  proofs[i].DateApprovedMP,
			OperationTypeMP: proofs[i].OperationTypeMP,
			AmountMP:        proofs[i].AmountMP,
			Dni:             proofs[i].Dni,
			CardType:        proofs[i].CardType,
			Last4Card:       proofs[i].Last4Card,
			ExternalID:      proofs[i].ExternalID,
			ProductName:     proofs[i].ProductName,
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
		UserID:          proof.UserID,
		IDMP:            proof.IDMP,
		ProofDate:       proof.ProofDate,
		StatusMP:        proof.StatusMP,
		DateApprovedMP:  proof.DateApprovedMP,
		OperationTypeMP: proof.OperationTypeMP,
		AmountMP:        proof.AmountMP,
		Dni:             proof.Dni,
		CardType:        proof.CardType,
		Last4Card:       proof.Last4Card,
		ExternalID:      proof.ExternalID,
		ProductName:     proof.ProductName,
	}, nil
}
