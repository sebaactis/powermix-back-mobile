package proof

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{
		service: service,
	}
}

func (h *HTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ProofRequest

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeProofUnauthorized(w)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProofValidation(w, "Error al intentar parsear el request, por favor validar el mismo", nil)
		return
	}

	req.UserID = userID

	proof, err := h.service.Create(r.Context(), &req)
	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			writeProofValidation(w, "Error de validación", fields)
			return
		}
		if writeProofServiceError(w, err, "No se pudo crear el comprobante", userID) {
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, proof)
}

func (h *HTTPHandler) CreateFromOthers(w http.ResponseWriter, r *http.Request) {
	var req ProofOthersRequest

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeProofUnauthorized(w)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProofValidation(w, "Error al intentar parsear el request, por favor validar el mismo", nil)
		return
	}

	req.UserID = userID

	proof, err := h.service.CreateFromOthers(r.Context(), &req)
	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			writeProofValidation(w, "Error de validación", fields)
			return
		}
		if writeProofServiceError(w, err, "No se pudo crear el comprobante", userID) {
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, proof)
}

func (h *HTTPHandler) GetAllByUserID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeProofUnauthorized(w)
		return
	}

	proofs, err := h.service.GetAllByUserID(r.Context(), userID)
	if err != nil {
		slog.Error("error al listar comprobantes del usuario", "user_id", userID, "error", err)
		writeProofInternal(w, "No se pudieron recuperar los comprobantes del usuario")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, proofs)
}

func (h *HTTPHandler) GetAllByUserIDPaginated(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeProofUnauthorized(w)
		return
	}

	q := r.URL.Query()
	pageStr := q.Get("page")
	pageSizeStr := q.Get("pageSize")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if v, err := strconv.Atoi(pageStr); err == nil && v > 0 {
			page = v
		}
	}

	if pageSizeStr != "" {
		if v, err := strconv.Atoi(pageSizeStr); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	var filters ProofFilters

	filters.IDMP = q.Get("id_mp")

	const dateLayout = "2006-01-02"

	if fromStr := q.Get("fromProofDate"); fromStr != "" {
		if t, err := time.Parse(dateLayout, fromStr); err == nil {
			filters.FromProofDate = &t
		}
	}

	if toStr := q.Get("toProofDate"); toStr != "" {
		if t, err := time.Parse(dateLayout, toStr); err == nil {
			filters.ToProofDate = &t
		}
	}

	if minStr := q.Get("minAmount"); minStr != "" {
		if v, err := strconv.ParseFloat(minStr, 64); err == nil {
			filters.MinAmount = &v
		}
	}

	if maxStr := q.Get("maxAmount"); maxStr != "" {
		if v, err := strconv.ParseFloat(maxStr, 64); err == nil {
			filters.MaxAmount = &v
		}
	}

	proofsPage, err := h.service.GetAllByUserIDPaginated(r.Context(), userID, page, pageSize, filters)
	if err != nil {
		slog.Error("error al listar comprobantes paginados", "user_id", userID, "error", err)
		writeProofInternal(w, "No se pudieron recuperar los comprobantes del usuario")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, proofsPage)
}

func (h *HTTPHandler) GetLastThreeByUserID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeProofUnauthorized(w)
		return
	}

	proofs, err := h.service.GetLastThreeByUserID(r.Context(), userID)
	if err != nil {
		slog.Error("error al listar últimos comprobantes", "user_id", userID, "error", err)
		writeProofInternal(w, "No se pudieron recuperar los comprobantes del usuario")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, proofs)
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		writeProofValidation(w, "El id del comprobante de pago es requerido", nil)
		return
	}

	proof, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProofIDRequired) {
			writeProofValidation(w, err.Error(), nil)
			return
		}
		slog.Error("error al obtener comprobante", "id", id, "error", err)
		writeProofInternal(w, "No se pudo recuperar el comprobante de pago")
		return
	}

	if proof == nil {
		writeProofNotFound(w, "Comprobante no encontrado")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, proof)
}

func writeProofUnauthorized(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
		Code:    utils.ErrCodeUnauthorized,
		Message: "No se pudo recuperar el usuario de la sesión",
	})
}

func writeProofValidation(w http.ResponseWriter, message string, fields interface{}) {
	utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
		Code:    utils.ErrCodeValidation,
		Message: message,
		Fields:  fields,
	})
}

func writeProofNotFound(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
		Code:    utils.ErrCodeNotFound,
		Message: message,
	})
}

func writeProofInternal(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
		Code:    utils.ErrCodeInternal,
		Message: message,
	})
}

// writeProofServiceError maps service errors to API responses. Returns true if handled.
func writeProofServiceError(w http.ResponseWriter, err error, internalMessage string, userID interface{}) bool {
	if errors.Is(err, ErrProofDuplicateID) ||
		errors.Is(err, ErrProofNotFoundID) ||
		errors.Is(err, ErrPaymentNotFound) ||
		errors.Is(err, ErrProofDuplicateMP) ||
		errors.Is(err, ErrProofIDRequired) {
		writeProofValidation(w, err.Error(), nil)
		return true
	}
	slog.Error(internalMessage, "user_id", userID, "error", err)
	writeProofInternal(w, internalMessage)
	return true
}
