package proof

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type Service struct {
	repo      *Repository
	validator validations.StructValidator
}

func NewService(repo *Repository, validator validations.StructValidator) *Service {
	return &Service{repo: repo, validator: validator}
}

func (s *Service) Create(ctx context.Context, proof *ProofRequest) (*ProofResponse, error) {
	if fields, ok := s.validator.ValidateStruct(proof); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	proofCreated, err := s.repo.Create(ctx, proof)

	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		UserID:    proofCreated.UserID,
		ProofMPID: proofCreated.ProofMPID,
		ProofDate: proofCreated.ProofDate.String(),
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
			UserID:    proofs[i].UserID,
			ProofMPID: proofs[i].ProofMPID,
			ProofDate: proofs[i].ProofDate.String(),
		})
	}

	return proofsResponse, nil

}

func (s *Service) GetById(ctx context.Context, id string) (*ProofResponse, error) {

	if id == "" {
		return nil, errors.New("id is required")
	}

	proof, err := s.repo.GetById(ctx, id)

	if err != nil {
		return nil, err
	}

	return &ProofResponse{
		UserID:    proof.UserID,
		ProofMPID: proof.ProofMPID,
		ProofDate: proof.ProofDate.String(),
	}, nil

}
