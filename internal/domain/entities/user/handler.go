package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	var req UserCreate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al intentar parsear el request, por favor validar el mismo", nil)
		return
	}

	u, err := h.service.Create(r.Context(), &req)

	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, "Error de validacion", fields)
			return
		}

		if errors.Is(err, ErrDuplicateEmail) {
			utils.WriteError(w, http.StatusConflict, "El email ya se encuentra en uso", nil)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "Error en el servidor", map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToResponse(u))
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "Id inválido", map[string]string{"error": err.Error()})
		return
	}

	u, err := h.service.GetByID(r.Context(), uint(id))

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Id inválido", map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(ToResponse(u))
}
