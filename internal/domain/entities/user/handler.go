package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type HTTPHandler struct {
	service *Service
	JWT     *jwtx.JWT
}

func NewHTTPHandler(service *Service, JWT *jwtx.JWT) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		JWT:     JWT,
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

	u, err := h.service.GetByID(r.Context(), uuid.MustParse(idStr))

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Id inválido", map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(ToResponse(u))
}

func (h *HTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.WriteError(w, http.StatusBadRequest, "No hay token informado",
			map[string]string{"error": "No hay token informado"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.WriteError(w, http.StatusBadRequest, "Formato de token inválido",
			map[string]string{"error": "Authorization debe ser 'Bearer <token>'"})
		return
	}

	tokenStr := parts[1]

	userID, _, _, err := h.JWT.Parse(tokenStr, jwtx.TokenTypeAccess)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Token inválido",
			map[string]string{"error": err.Error()})
		return
	}

	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Usuario no encontrado",
			map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(ToResponse(user))
}
