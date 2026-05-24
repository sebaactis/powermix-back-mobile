package prode

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// ListMatches devuelve los partidos visibles con la predicción del usuario autenticado.
func (h *HTTPHandler) ListMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := h.service.ListMatches(r.Context())
	if err != nil {
		slog.Error("error al listar partidos", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al obtener los partidos", nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, matches)
}

// GetMatch devuelve un partido con la predicción del usuario autenticado.
func (h *HTTPHandler) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID de partido inválido", nil)
		return
	}

	match, err := h.service.GetMatch(r.Context(), matchID)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			utils.WriteError(w, http.StatusNotFound, "Partido no encontrado", nil)
			return
		}
		slog.Error("error al obtener partido", "match_id", matchID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al obtener el partido", nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, match)
}

// CreateOrUpdatePrediction crea o actualiza la predicción del usuario autenticado.
func (h *HTTPHandler) CreateOrUpdatePrediction(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID de partido inválido", nil)
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		slog.Error("usuario no autenticado al crear predicción")
		utils.WriteError(w, http.StatusUnauthorized, "Usuario no autenticado", nil)
		return
	}

	var req PredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al parsear el request, por favor validar el mismo", nil)
		return
	}

	pred, err := h.service.CreateOrUpdatePrediction(r.Context(), userID, matchID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrCutoffPassed):
			utils.WriteError(w, http.StatusConflict, "La hora límite para predecir este partido ya pasó", nil)
		case errors.Is(err, ErrInvalidScore), errors.Is(err, ErrScoreOutOfRange):
			utils.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		case errors.Is(err, ErrMatchNotFound):
			utils.WriteError(w, http.StatusNotFound, "Partido no encontrado", nil)
		case errors.Is(err, ErrMatchNotOpen):
			utils.WriteError(w, http.StatusConflict, "El partido no está abierto para predicciones", nil)
		default:
			slog.Error("error al guardar predicción", "match_id", matchID, "user_id", userID, "error", err)
			utils.WriteError(w, http.StatusInternalServerError, "Error al guardar la predicción", nil)
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, pred)
}

// GetMyPredictions devuelve todas las predicciones del usuario autenticado.
func (h *HTTPHandler) GetMyPredictions(w http.ResponseWriter, r *http.Request) {
	predictions, err := h.service.GetMyPredictions(r.Context())
	if err != nil {
		slog.Error("error al obtener predicciones del usuario", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al obtener las predicciones", nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, predictions)
}

// ---- Admin handlers ----

// AdminCreateMatch crea un nuevo partido.
func (h *HTTPHandler) AdminCreateMatch(w http.ResponseWriter, r *http.Request) {
	var req CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al parsear el request, por favor validar el mismo", nil)
		return
	}

	if req.Stage == "" || req.Opponent == "" || req.KickoffAt.IsZero() {
		utils.WriteError(w, http.StatusBadRequest, "Faltan campos obligatorios: stage, opponent, kickoff_at", nil)
		return
	}

	match, err := h.service.CreateMatch(r.Context(), req)
	if err != nil {
		slog.Error("error al crear partido", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al crear el partido", nil)
		return
	}

	slog.Info("partido creado por admin", "match_id", match.ID, "opponent", match.Opponent)
	utils.WriteSuccess(w, http.StatusCreated, match)
}

// AdminUpdateMatch actualiza los campos de un partido.
func (h *HTTPHandler) AdminUpdateMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID de partido inválido", nil)
		return
	}

	var req UpdateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al parsear el request", nil)
		return
	}

	match, err := h.service.UpdateMatch(r.Context(), matchID, req)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			utils.WriteError(w, http.StatusNotFound, "Partido no encontrado", nil)
			return
		}
		slog.Error("error al actualizar partido", "match_id", matchID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al actualizar el partido", nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, match)
}

// AdminRecordResult registra el resultado de 90 minutos de un partido.
func (h *HTTPHandler) AdminRecordResult(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID de partido inválido", nil)
		return
	}

	var req RecordResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al parsear el request", nil)
		return
	}

	match, err := h.service.RecordResult(r.Context(), matchID, req)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			utils.WriteError(w, http.StatusNotFound, "Partido no encontrado", nil)
			return
		}
		if errors.Is(err, ErrInvalidScore) {
			utils.WriteError(w, http.StatusBadRequest, "El resultado debe ser no negativo", nil)
			return
		}
		slog.Error("error al registrar resultado", "match_id", matchID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al registrar el resultado", nil)
		return
	}

	slog.Info("resultado registrado por admin",
		"match_id", matchID,
		"argentina", req.ArgentinaGoals,
		"opponent", req.OpponentGoals,
	)
	utils.WriteSuccess(w, http.StatusOK, match)
}

// AdminSettleMatch ejecuta el settlement de un partido.
func (h *HTTPHandler) AdminSettleMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID de partido inválido", nil)
		return
	}

	result, err := h.service.SettleMatch(r.Context(), matchID)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			utils.WriteError(w, http.StatusNotFound, "Partido no encontrado", nil)
			return
		}
		if errors.Is(err, ErrResultMissing) {
			utils.WriteError(w, http.StatusConflict, "El partido aún no tiene resultado cargado", nil)
			return
		}
		slog.Error("error al ejecutar settlement", "match_id", matchID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al procesar las predicciones", nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, result)
}

// AdminRetryPendingRewards reintenta asignar vouchers a premios pendientes.
func (h *HTTPHandler) AdminRetryPendingRewards(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.RetryPendingRewards(r.Context())
	if err != nil {
		slog.Error("error al reintentar premios pendientes", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Error al procesar los premios pendientes", nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, result)
}
