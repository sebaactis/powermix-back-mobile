package proof

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTx returns a new Repository that uses the given transaction
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// DB exposes the underlying db connection for transaction management
func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) Create(ctx context.Context, proof *Proof) (*Proof, error) {

	err := r.db.WithContext(ctx).Create(proof).Error

	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (r *Repository) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]*Proof, error) {
	var proofs []*Proof
	result := r.db.WithContext(ctx).Where("user_id = ?", userId).Find(&proofs)

	if result.Error != nil {
		return nil, result.Error
	}

	return proofs, nil
}

func (r *Repository) GetAllByUserIdPaginated(ctx context.Context, userId uuid.UUID, page int, pageSize int, filters ProofFilters) ([]*Proof, int64, error) {
	if page < 1 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	var total int64

	baseQuery := r.db.WithContext(ctx).Model(&Proof{}).Where("user_id = ?", userId)

	// FILTROS
	if filters.ID_MP != "" {
		baseQuery = baseQuery.Where("id_mp ILIKE ?", "%"+filters.ID_MP+"%")
	}

	if filters.FromProofDate != nil {
		baseQuery = baseQuery.Where("date_approved_mp::timestamp >= ?", *filters.FromProofDate)
	}
	if filters.ToProofDate != nil {
		baseQuery = baseQuery.Where("date_approved_mp::timestamp <= ?", *filters.ToProofDate)
	}

	if filters.MinAmount != nil {
		baseQuery = baseQuery.Where("amount_mp >= ?", *filters.MinAmount)
	}
	if filters.MaxAmount != nil {
		baseQuery = baseQuery.Where("amount_mp <= ?", *filters.MaxAmount)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var proofs []*Proof
	offset := (page - 1) * pageSize

	if err := baseQuery.
		Order("proof_date DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&proofs).Error; err != nil {
		return nil, 0, err
	}

	return proofs, total, nil
}

func (r *Repository) GetLastThreeByUserId(ctx context.Context, userId uuid.UUID) ([]*Proof, error) {
	var proofs []*Proof

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("proof_date DESC").
		Limit(3).
		Find(&proofs)

	if result.Error != nil {
		return nil, result.Error
	}

	return proofs, nil
}

func (r *Repository) GetById(ctx context.Context, id string) (*Proof, error) {
	var proof Proof

	err := r.db.WithContext(ctx).First(&proof, "ID_MP = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &proof, nil
}
