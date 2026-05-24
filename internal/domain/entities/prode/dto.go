package prode

import "time"

// MatchResponse devuelve la información de un partido al usuario.
type MatchResponse struct {
	ID               string     `json:"id"`
	Stage            string     `json:"stage"`
	Opponent         string     `json:"opponent"`
	KickoffAt        time.Time  `json:"kickoff_at"`
	CutoffAt         time.Time  `json:"cutoff_at"`
	Status           string     `json:"status"`
	IsOpen           bool       `json:"is_open"`
	ArgentinaGoals   *int       `json:"argentina_goals,omitempty"`
	OpponentGoals    *int       `json:"opponent_goals,omitempty"`
	MyPrediction     *PredictionResponse `json:"my_prediction,omitempty"`
}

// PredictionRequest es el body para crear o editar una predicción.
type PredictionRequest struct {
	ArgentinaGoals int `json:"argentina_goals"`
	OpponentGoals  int `json:"opponent_goals"`
}

// PredictionResponse devuelve los datos de la predicción del usuario.
type PredictionResponse struct {
	ID             string    `json:"id"`
	MatchID        string    `json:"match_id"`
	ArgentinaGoals int       `json:"argentina_goals"`
	OpponentGoals  int       `json:"opponent_goals"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ---- Admin DTOs ----

// CreateMatchRequest es el body para crear un partido.
type CreateMatchRequest struct {
	Stage     string    `json:"stage"`
	Opponent  string    `json:"opponent"`
	KickoffAt time.Time `json:"kickoff_at"`
	IsVisible bool      `json:"is_visible"`
}

// UpdateMatchRequest es el body para actualizar un partido.
type UpdateMatchRequest struct {
	Stage     *string    `json:"stage,omitempty"`
	Opponent  *string    `json:"opponent,omitempty"`
	KickoffAt *time.Time `json:"kickoff_at,omitempty"`
	IsVisible *bool      `json:"is_visible,omitempty"`
	Status    *string    `json:"status,omitempty"`
}

// RecordResultRequest es el body para cargar el resultado de un partido.
type RecordResultRequest struct {
	ArgentinaGoals int `json:"argentina_goals"`
	OpponentGoals  int `json:"opponent_goals"`
}

// AdminMatchResponse devuelve la info completa de un partido para el admin.
type AdminMatchResponse struct {
	ID             string     `json:"id"`
	Stage          string     `json:"stage"`
	Opponent       string     `json:"opponent"`
	KickoffAt      time.Time  `json:"kickoff_at"`
	CutoffAt       time.Time  `json:"cutoff_at"`
	Status         string     `json:"status"`
	IsVisible      bool       `json:"is_visible"`
	ArgentinaGoals *int       `json:"argentina_goals,omitempty"`
	OpponentGoals  *int       `json:"opponent_goals,omitempty"`
	ExternalID     string     `json:"external_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// SettlementResponse devuelve el resultado de un settlement.
type SettlementResponse struct {
	MatchID     string `json:"match_id"`
	Status      string `json:"status"`
	TotalPreds  int    `json:"total_predictions"`
	Correct     int    `json:"correct"`
}

// RewardRetryResponse devuelve el resultado de reintentar premios pendientes.
type RewardRetryResponse struct {
	Processed int `json:"processed"`
	Assigned  int `json:"assigned"`
	Failed    int `json:"failed"`
	Remaining int `json:"pending"`
}
