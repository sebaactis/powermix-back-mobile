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
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	u, err := h.service.Create(r.Context(), &req)

	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, "validation error", fields)
			return
		}

		if errors.Is(err, ErrDuplicateEmail) {
			utils.WriteError(w, http.StatusConflict, "email already exists", nil)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToResponse(u))
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	u, err := h.service.GetByID(r.Context(), uint(id))

	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(ToResponse(u))
}

func (h *HTTPHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.FindAll(r.Context())

	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "users not found", nil)
		return
	}

	json.NewEncoder(w).Encode(ToResponseMany(users))
}
