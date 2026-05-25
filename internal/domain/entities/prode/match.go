package prode

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Constantes de estado de partido
const (
	MatchStatusDraft          = "DRAFT"
	MatchStatusScheduled      = "SCHEDULED"
	MatchStatusOpen           = "OPEN"
	MatchStatusClosed         = "CLOSED"
	MatchStatusResultRecorded = "RESULT_RECORDED"
	MatchStatusEvaluated      = "EVALUATED"
	MatchStatusCancelled      = "CANCELLED"
)

type ProdeMatch struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Stage          string     `gorm:"type:varchar(50);not null" json:"stage"`
	Opponent       string     `gorm:"type:varchar(100);not null" json:"opponent"`
	KickoffAt      time.Time  `gorm:"type:timestamptz;not null" json:"kickoff_at"`
	Status         string     `gorm:"type:varchar(30);not null;default:DRAFT" json:"status"`
	ArgentinaGoals *int       `gorm:"type:smallint" json:"argentina_goals,omitempty"`
	OpponentGoals  *int       `gorm:"type:smallint" json:"opponent_goals,omitempty"`
	IsVisible      bool       `gorm:"not null;default:false" json:"is_visible"`
	ExternalID     string     `gorm:"type:varchar(100)" json:"external_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ProdeMatch) TableName() string { return "prode_matches" }

func (m *ProdeMatch) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}


// CutoffAt devuelve la hora límite para aceptar predicciones.
// El corte es 1 hora antes del inicio en huso horario argentino.
func (m *ProdeMatch) CutoffAt() time.Time {
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	return m.KickoffAt.In(loc).Add(-1 * time.Hour)
}

// IsOpenForPrediction devuelve true si el partido sigue aceptando
// predicciones en el momento indicado.
func (m *ProdeMatch) IsOpenForPrediction(now time.Time) bool {
	if m.Status != MatchStatusScheduled && m.Status != MatchStatusOpen {
		return false
	}
	return now.Before(m.CutoffAt())
}

// BeforeCutoff es un helper independiente que verifica si una hora
// está antes del corte (kickoff - 1 hora en zona Argentina).
func BeforeCutoff(kickoff time.Time, now time.Time) bool {
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	cutoff := kickoff.In(loc).Add(-1 * time.Hour)
	return now.Before(cutoff)
}
