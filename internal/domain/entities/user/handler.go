package user

import (
	"encoding/json"
	"errors"
	"fmt"
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

	utils.WriteSuccess(w, http.StatusCreated, ToResponse(u))
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	u, err := h.service.GetByID(r.Context(), uuid.MustParse(idStr))

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Id inválido", map[string]string{"error": err.Error()})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(u))
}

func (h *HTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
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

	utils.WriteSuccess(w, http.StatusOK, ToResponse(user))
}

func (h *HTTPHandler) Update(w http.ResponseWriter, r *http.Request) {

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Token invalido", map[string]string{"error": err.Error()})
		return
	}

	var req UserUpdate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "No se pudo parsear la respuesta, por favor revise los datos enviados", map[string]string{"error": err.Error()})
		return
	}

	userUpdate, err := h.service.Update(r.Context(), userID, req)

	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, "Error de validación", fields)
			return
		}

		if errors.Is(err, ErrSameName) {
			utils.WriteError(w, http.StatusBadRequest, "El nombre no puede ser igual al actual", nil)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "No se pudo actualizar el usuario",
			map[string]string{"error": err.Error()})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(userUpdate))
}

func (h *HTTPHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Token invalido", map[string]string{"error": err.Error()})
		return
	}

	var req UserChangePassword

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "No se pudo parsear la respuesta, por favor revise los datos enviados", map[string]string{"error": err.Error()})
		return
	}

	userUpdate, err := h.service.UpdatePassword(r.Context(), userID, req)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "No se pudo actualizar el usuario",
			map[string]string{"error": err.Error()})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(userUpdate))

}

// Helper privado
func (h *HTTPHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return uuid.Nil, fmt.Errorf("no hay token informado")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return uuid.Nil, fmt.Errorf("formato de token inválido, debe ser 'Bearer <token>'")
	}

	tokenStr := parts[1]

	userID, _, _, err := h.JWT.Parse(tokenStr, jwtx.TokenTypeAccess)
	if err != nil {
		return uuid.Nil, fmt.Errorf("token inválido: %w", err)
	}

	return userID, nil
}
