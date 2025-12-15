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

	proofExistsValidate, err := s.GetById(ctx, proof.ID_MP)

	if err != nil {
		return nil, err
	}

	if proofExistsValidate != nil {
		return nil, fmt.Errorf("ya tenes guardado un comprobante con este ID: %s", proof.ID_MP)
	}

	payment, err := s.mpClient.ValidatePaymentExists(ctx, proof.ID_MP)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("el comprobante %s no existe, por favor, verifique los datos ingresados", proof.ID_MP)
	}

	goodsName, err := s.coffejiClient.GetGoodsNameByOrderNo(ctx, *payment.ExternalID)

	if err != nil {
		return nil, err
	}

	newProof := &Proof{
		UserID:            proof.UserID,
		ID_MP:             proof.ID_MP,
		Date_Approved_MP:  utils.FormattedTime{Time: payment.DateApproved.Truncate(time.Second)},
		Operation_Type_MP: payment.OperationType,
		Status_MP:         payment.Status,
		Amount_MP:         payment.TotalPaidAmount,
		ProofDate:         utils.NowFormatted(),
		Dni:               payment.PayerDNI,
		CardId:            payment.CardId,
		CardType:          payment.CardType,
		Last4Card:         payment.CardLast4,
		ExternalID:        payment.ExternalID,
		ProductName:       &goodsName,
	}

	proofResult, err := s.repo.Create(ctx, newProof)

	if err != nil {
		return nil, err
	}

	quantityStamps, err := s.userService.IncrementStampsCounter(ctx, proofResult.UserID)

	if err != nil {
		return nil, err
	}

	if quantityStamps == 10 {

		_, err = s.voucherService.AssignNextVoucher(ctx, &voucher.VoucherRequest{
			UserID: proofResult.UserID})

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
		Dni:               proofResult.Dni,
		CardType:          proofResult.CardType,
		Last4Card:         proofResult.Last4Card,
		ExternalID:        proofResult.ExternalID,
		ProductName:       proofResult.ProductName,
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
		UserID:            req.UserID,
		ID_MP:             idMP,
		Date_Approved_MP:  utils.FormattedTime{Time: result.DateApproved.Add(time.Hour).Truncate(time.Second)},
		Operation_Type_MP: result.OperationType,
		Status_MP:         result.Status,
		Amount_MP:         result.TotalPaidAmount,
		ProofDate:         utils.NowFormatted(),
		Dni:               result.PayerDNI,
		CardId:            result.CardId,
		CardType:          result.CardType,
		Last4Card:         result.CardLast4,
		ExternalID:        result.ExternalID,
		ProductName:       &goodsName,
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
		_, err = s.voucherService.AssignNextVoucher(ctx, &voucher.VoucherRequest{UserID: proofResult.UserID})
		
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
		Dni:               proofResult.Dni,
		CardType:          proofResult.CardType,
		Last4Card:         proofResult.Last4Card,
		ExternalID:        proofResult.ExternalID,
		ProductName:       proofResult.ProductName,
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
			Dni:               proofs[i].Dni,
			CardType:          proofs[i].CardType,
			Last4Card:         proofs[i].Last4Card,
			ExternalID:        proofs[i].ExternalID,
			ProductName:       proofs[i].ProductName,
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
			UserID:            proofs[i].UserID,
			ID_MP:             proofs[i].ID_MP,
			ProofDate:         proofs[i].ProofDate,
			Status_MP:         proofs[i].Status_MP,
			Date_Approved_MP:  proofs[i].Date_Approved_MP,
			Operation_Type_MP: proofs[i].Operation_Type_MP,
			Amount_MP:         proofs[i].Amount_MP,
			Dni:               proofs[i].Dni,
			CardType:          proofs[i].CardType,
			Last4Card:         proofs[i].Last4Card,
			ExternalID:        proofs[i].ExternalID,
			ProductName:       proofs[i].ProductName,
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
			UserID:            proofs[i].UserID,
			ID_MP:             proofs[i].ID_MP,
			ProofDate:         proofs[i].ProofDate,
			Status_MP:         proofs[i].Status_MP,
			Date_Approved_MP:  proofs[i].Date_Approved_MP,
			Operation_Type_MP: proofs[i].Operation_Type_MP,
			Amount_MP:         proofs[i].Amount_MP,
			Dni:               proofs[i].Dni,
			CardType:          proofs[i].CardType,
			Last4Card:         proofs[i].Last4Card,
			ExternalID:        proofs[i].ExternalID,
			ProductName:       proofs[i].ProductName,
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
		Dni:               proof.Dni,
		CardType:          proof.CardType,
		Last4Card:         proof.Last4Card,
		ExternalID:        proof.ExternalID,
		ProductName:       proof.ProductName,
	}, nil
}
