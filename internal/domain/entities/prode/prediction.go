package prode

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Constantes de estado de predicción
const (
	PredStatusPending  = "PENDING"
	PredStatusCorrect  = "CORRECT"
	PredStatusIncorrect = "INCORRECT"
)

type ProdePrediction struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_prode_pred_user_match" json:"user_id"`
	MatchID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_prode_pred_user_match" json:"match_id"`
	ArgentinaGoals int       `gorm:"type:smallint;not null" json:"argentina_goals"`
	OpponentGoals  int       `gorm:"type:smallint;not null" json:"opponent_goals"`
	Status         string    `gorm:"type:varchar(30);not null;default:PENDING" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ProdePrediction) TableName() string { return "prode_predictions" }

func (p *ProdePrediction) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ValidateScore verifica que los resultados estén dentro de rangos válidos.
func (p *ProdePrediction) ValidateScore() error {
	if p.ArgentinaGoals < 0 || p.OpponentGoals < 0 {
		return ErrInvalidScore
	}
	if p.ArgentinaGoals > 50 || p.OpponentGoals > 50 {
		return ErrScoreOutOfRange
	}
	return nil
}
