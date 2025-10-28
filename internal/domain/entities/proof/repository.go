package proof

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, proof *ProofRequest) (*Proof, error) {
	proofCreate := &Proof{
		UserID:    proof.UserID,
		ProofMPID: proof.ProofMPID,
		ProofDate: time.Now(),
	}

	err := r.db.WithContext(ctx).Create(proofCreate).Error

	if err != nil {
		return nil, err
	}

	return proofCreate, nil
}

func (r *Repository) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]*Proof, error) {
	var proofs []*Proof
	result := r.db.WithContext(ctx).Where("user_id = ?", userId).Find(&proofs)

	if result.Error != nil {
		return nil, result.Error
	}

	return proofs, nil
}

func (r *Repository) GetById(ctx context.Context, id string) (*Proof, error) {
	var proof Proof

	
	if err := r.db.WithContext(ctx).First(&proof, "proof_mp_id = ?", id).Error; err != nil {
		return nil, err
	}

	return &proof, nil
}
