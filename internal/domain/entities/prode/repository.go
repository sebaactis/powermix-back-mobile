package prode

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository brinda persistencia vía GORM para las entidades PRODE.
type Repository struct {
	db *gorm.DB
}

// NewRepository crea un Repository con la conexión a DB indicada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTx devuelve un Repository que usa la transacción recibida.
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// DB expone la conexión subyacente para manejo de transacciones.
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// GetMatchByID obtiene un partido por su ID.
func (r *Repository) GetMatchByID(ctx context.Context, id uuid.UUID) (*ProdeMatch, error) {
	var match ProdeMatch
	err := r.db.WithContext(ctx).First(&match, "id = ?", id).Error
	if err != nil {
		return nil, mapProdeMatchErr(ctx, "get match by id", err)
	}
	return &match, nil
}

// GetVisibleMatches obtiene todos los partidos marcados como visibles.
func (r *Repository) GetVisibleMatches(ctx context.Context) ([]ProdeMatch, error) {
	var matches []ProdeMatch
	err := r.db.WithContext(ctx).
		Where("is_visible = ?", true).
		Order("kickoff_at ASC").
		Find(&matches).Error
	if err != nil {
		return nil, mapProdeRepoErr(ctx, "get visible matches", err)
	}
	return matches, nil
}

// GetPredictionsByMatchID obtiene todas las predicciones de un partido.
func (r *Repository) GetPredictionsByMatchID(ctx context.Context, matchID uuid.UUID) ([]ProdePrediction, error) {
	var predictions []ProdePrediction
	err := r.db.WithContext(ctx).
		Where("match_id = ?", matchID).
		Find(&predictions).Error
	if err != nil {
		return nil, mapProdeRepoErr(ctx, "get predictions by match id", err)
	}
	return predictions, nil
}

// GetRewardByPredictionID obtiene el premio asociado a una predicción.
func (r *Repository) GetRewardByPredictionID(ctx context.Context, predictionID uuid.UUID) (*ProdeReward, error) {
	var reward ProdeReward
	err := r.db.WithContext(ctx).First(&reward, "prediction_id = ?", predictionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, mapProdeRepoErr(ctx, "get reward by prediction id", err)
	}
	return &reward, nil
}

// GetUserPrediction obtiene la predicción de un usuario para un partido específico.
func (r *Repository) GetUserPrediction(ctx context.Context, userID uuid.UUID, matchID uuid.UUID) (*ProdePrediction, error) {
	var pred ProdePrediction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND match_id = ?", userID, matchID).
		First(&pred).Error
	if err != nil {
		return nil, mapProdeUserPredictionLookupErr(ctx, err)
	}
	return &pred, nil
}

// UpsertPrediction crea o actualiza la predicción de un usuario para un partido.
// Usa el unique index (user_id, match_id) para determinar si es inserción o actualización.
func (r *Repository) UpsertPrediction(ctx context.Context, pred *ProdePrediction) (*ProdePrediction, error) {
	if pred.ID == uuid.Nil {
		pred.ID = uuid.New()
	}

	// Buscar si ya existe una predicción para este usuario y partido
	var existing ProdePrediction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND match_id = ?", pred.UserID, pred.MatchID).
		First(&existing).Error

	if err == nil {
		// Ya existe → actualizar
		existing.ArgentinaGoals = pred.ArgentinaGoals
		existing.OpponentGoals = pred.OpponentGoals
		existing.Status = PredStatusPending
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, mapProdeRepoErr(ctx, "upsert prediction save", err)
		}
		return &existing, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No existe → crear
		if err := r.db.WithContext(ctx).Create(pred).Error; err != nil {
			return nil, mapProdeRepoErr(ctx, "upsert prediction create", err)
		}
		return pred, nil
	}

	return nil, mapProdeRepoErr(ctx, "upsert prediction lookup", err)
}

// GetPredictionsByUserID obtiene todas las predicciones de un usuario.
func (r *Repository) GetPredictionsByUserID(ctx context.Context, userID uuid.UUID) ([]ProdePrediction, error) {
	var predictions []ProdePrediction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&predictions).Error
	if err != nil {
		return nil, mapProdeRepoErr(ctx, "get predictions by user id", err)
	}
	return predictions, nil
}

// CreateMatch inserta un nuevo partido.
func (r *Repository) CreateMatch(ctx context.Context, match *ProdeMatch) error {
	if err := r.db.WithContext(ctx).Create(match).Error; err != nil {
		return mapProdeRepoErr(ctx, "create match", err)
	}
	return nil
}

// UpdateMatch guarda los cambios de un partido existente.
func (r *Repository) UpdateMatch(ctx context.Context, match *ProdeMatch) error {
	if err := r.db.WithContext(ctx).Save(match).Error; err != nil {
		return mapProdeRepoErr(ctx, "update match", err)
	}
	return nil
}

// CreateReward inserta un nuevo premio en el ledger.
func (r *Repository) CreateReward(ctx context.Context, reward *ProdeReward) error {
	if err := r.db.WithContext(ctx).Create(reward).Error; err != nil {
		return mapProdeRepoErr(ctx, "create reward", err)
	}
	return nil
}

// UpdateReward guarda los cambios de un premio existente.
func (r *Repository) UpdateReward(ctx context.Context, reward *ProdeReward) error {
	if err := r.db.WithContext(ctx).Save(reward).Error; err != nil {
		return mapProdeRepoErr(ctx, "update reward", err)
	}
	return nil
}

// UpdatePrediction actualiza una predicción existente.
func (r *Repository) UpdatePrediction(ctx context.Context, pred *ProdePrediction) error {
	if err := r.db.WithContext(ctx).Save(pred).Error; err != nil {
		return mapProdeRepoErr(ctx, "update prediction", err)
	}
	return nil
}

// GetPredictionByID obtiene una predicción por su ID.
func (r *Repository) GetPredictionByID(ctx context.Context, id uuid.UUID) (*ProdePrediction, error) {
	var pred ProdePrediction
	err := r.db.WithContext(ctx).First(&pred, "id = ?", id).Error
	if err != nil {
		return nil, mapProdePredictionErr(ctx, "get prediction by id", err)
	}
	return &pred, nil
}

// GetPendingInventoryRewards obtiene todos los premios pendientes por falta de inventario.
func (r *Repository) GetPendingInventoryRewards(ctx context.Context) ([]ProdeReward, error) {
	var rewards []ProdeReward
	err := r.db.WithContext(ctx).
		Where("status = ?", RewardStatusPendingInventory).
		Order("created_at ASC").
		Find(&rewards).Error
	if err != nil {
		return nil, mapProdeRepoErr(ctx, "get pending inventory rewards", err)
	}
	return rewards, nil
}

// CountPendingInventoryRewards cuenta los premios pendientes por inventario.
func (r *Repository) CountPendingInventoryRewards(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ProdeReward{}).
		Where("status = ?", RewardStatusPendingInventory).
		Count(&count).Error
	if err != nil {
		return 0, mapProdeRepoErr(ctx, "count pending inventory rewards", err)
	}
	return int(count), nil
}

func mapProdeMatchErr(ctx context.Context, action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("prode: %s: %w", action, ErrMatchNotFound)
	}
	slog.ErrorContext(ctx, "prode repository", "action", action, "error", err)
	return fmt.Errorf("prode: %s: %w", action, ErrInternal)
}

func mapProdePredictionErr(ctx context.Context, action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("prode: %s: %w", action, ErrPredictionNotFound)
	}
	slog.ErrorContext(ctx, "prode repository", "action", action, "error", err)
	return fmt.Errorf("prode: %s: %w", action, ErrInternal)
}

func mapProdeRepoErr(ctx context.Context, action string, err error) error {
	if err == nil {
		return nil
	}
	slog.ErrorContext(ctx, "prode repository", "action", action, "error", err)
	return fmt.Errorf("prode: %s: %w", action, ErrInternal)
}

func mapProdeUserPredictionLookupErr(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	slog.ErrorContext(ctx, "prode repository", "action", "get user prediction", "error", err)
	return fmt.Errorf("prode: get user prediction: %w", ErrInternal)
}
