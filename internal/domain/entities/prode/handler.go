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
		slog.ErrorContext(r.Context(), "error al listar partidos", "error", err)
		writeProdeInternal(w, "Error al obtener los partidos")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, matches)
}

// GetMatch devuelve un partido con la predicción del usuario autenticado.
func (h *HTTPHandler) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		writeProdeValidation(w, "ID de partido inválido", nil)
		return
	}

	match, err := h.service.GetMatch(r.Context(), matchID)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			writeProdeNotFound(w, "Partido no encontrado")
			return
		}
		slog.ErrorContext(r.Context(), "error al obtener partido", "match_id", matchID, "error", err)
		writeProdeInternal(w, "Error al obtener el partido")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, match)
}

// CreateOrUpdatePrediction crea o actualiza la predicción del usuario autenticado.
func (h *HTTPHandler) CreateOrUpdatePrediction(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		writeProdeValidation(w, "ID de partido inválido", nil)
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeProdeUnauthorized(w)
		return
	}

	var req PredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProdeValidation(w, "Error al parsear el request, por favor validar el mismo", nil)
		return
	}

	pred, err := h.service.CreateOrUpdatePrediction(r.Context(), userID, matchID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrCutoffPassed):
			writeProdeConflict(w, "La hora límite para predecir este partido ya pasó")
		case errors.Is(err, ErrInvalidScore):
			writeProdeValidation(w, "El resultado debe ser no negativo", nil)
		case errors.Is(err, ErrScoreOutOfRange):
			writeProdeValidation(w, "El resultado excede el rango permitido", nil)
		case errors.Is(err, ErrMatchNotFound):
			writeProdeNotFound(w, "Partido no encontrado")
		case errors.Is(err, ErrMatchNotOpen):
			writeProdeConflict(w, "El partido no está abierto para predicciones")
		default:
			slog.ErrorContext(r.Context(), "error al guardar predicción", "match_id", matchID, "user_id", userID, "error", err)
			writeProdeInternal(w, "Error al guardar la predicción")
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, pred)
}

// GetMyPredictions devuelve todas las predicciones del usuario autenticado.
func (h *HTTPHandler) GetMyPredictions(w http.ResponseWriter, r *http.Request) {
	predictions, err := h.service.GetMyPredictions(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "error al obtener predicciones del usuario", "error", err)
		writeProdeInternal(w, "Error al obtener las predicciones")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, predictions)
}

// ---- Admin handlers ----

// AdminCreateMatch crea un nuevo partido.
func (h *HTTPHandler) AdminCreateMatch(w http.ResponseWriter, r *http.Request) {
	var req CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProdeValidation(w, "Error al parsear el request, por favor validar el mismo", nil)
		return
	}

	if req.Stage == "" || req.Opponent == "" || req.KickoffAt.IsZero() {
		writeProdeValidation(w, "Faltan campos obligatorios: stage, opponent, kickoff_at", nil)
		return
	}

	match, err := h.service.CreateMatch(r.Context(), req)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al crear partido", "error", err)
		writeProdeInternal(w, "Error al crear el partido")
		return
	}

	slog.InfoContext(r.Context(), "partido creado por admin", "match_id", match.ID, "opponent", match.Opponent)
	utils.WriteSuccess(w, http.StatusCreated, match)
}

// AdminUpdateMatch actualiza los campos de un partido.
func (h *HTTPHandler) AdminUpdateMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		writeProdeValidation(w, "ID de partido inválido", nil)
		return
	}

	var req UpdateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProdeValidation(w, "Error al parsear el request", nil)
		return
	}

	match, err := h.service.UpdateMatch(r.Context(), matchID, req)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			writeProdeNotFound(w, "Partido no encontrado")
			return
		}
		slog.ErrorContext(r.Context(), "error al actualizar partido", "match_id", matchID, "error", err)
		writeProdeInternal(w, "Error al actualizar el partido")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, match)
}

// AdminRecordResult registra el resultado de 90 minutos de un partido.
func (h *HTTPHandler) AdminRecordResult(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		writeProdeValidation(w, "ID de partido inválido", nil)
		return
	}

	var req RecordResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProdeValidation(w, "Error al parsear el request", nil)
		return
	}

	match, err := h.service.RecordResult(r.Context(), matchID, req)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			writeProdeNotFound(w, "Partido no encontrado")
			return
		}
		if errors.Is(err, ErrInvalidScore) {
			writeProdeValidation(w, "El resultado debe ser no negativo", nil)
			return
		}
		slog.ErrorContext(r.Context(), "error al registrar resultado", "match_id", matchID, "error", err)
		writeProdeInternal(w, "Error al registrar el resultado")
		return
	}

	slog.InfoContext(r.Context(), "resultado registrado por admin", "match_id", matchID,
		"argentina", req.ArgentinaGoals,
		"opponent", req.OpponentGoals,)
	utils.WriteSuccess(w, http.StatusOK, match)
}

// AdminSettleMatch ejecuta el settlement de un partido.
func (h *HTTPHandler) AdminSettleMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		writeProdeValidation(w, "ID de partido inválido", nil)
		return
	}

	result, err := h.service.SettleMatch(r.Context(), matchID)
	if err != nil {
		if errors.Is(err, ErrMatchNotFound) {
			writeProdeNotFound(w, "Partido no encontrado")
			return
		}
		if errors.Is(err, ErrResultMissing) {
			writeProdeConflict(w, "El partido aún no tiene resultado cargado")
			return
		}
		slog.ErrorContext(r.Context(), "error al ejecutar settlement", "match_id", matchID, "error", err)
		writeProdeInternal(w, "Error al procesar las predicciones")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, result)
}

// AdminRetryPendingRewards reintenta asignar vouchers a premios pendientes.
func (h *HTTPHandler) AdminRetryPendingRewards(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.RetryPendingRewards(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "error al reintentar premios pendientes", "error", err)
		writeProdeInternal(w, "Error al procesar los premios pendientes")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, result)
}

func writeProdeUnauthorized(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
		Code:    utils.ErrCodeUnauthorized,
		Message: "Usuario no autenticado",
	})
}

func writeProdeValidation(w http.ResponseWriter, message string, fields interface{}) {
	utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
		Code:    utils.ErrCodeValidation,
		Message: message,
		Fields:  fields,
	})
}

func writeProdeNotFound(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
		Code:    utils.ErrCodeNotFound,
		Message: message,
	})
}

func writeProdeConflict(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusConflict, utils.WriteErrorOpts{
		Code:    utils.ErrCodeConflict,
		Message: message,
	})
}

func writeProdeInternal(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
		Code:    utils.ErrCodeInternal,
		Message: message,
	})
}
