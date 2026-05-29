package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
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
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "Error al intentar parsear el request, por favor validar el mismo",
		})
		return
	}

	user, err := h.service.Create(r.Context(), &req)

	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
				Code:    utils.ErrCodeValidation,
				Message: "Error de validación",
				Fields:  fields,
			})
			return
		}

		if errors.Is(err, ErrDuplicateEmail) {
			utils.WriteError(w, http.StatusConflict, utils.WriteErrorOpts{
				Code:    utils.ErrCodeDuplicateEntry,
				Message: "El email ya se encuentra en uso",
			})
			return
		}

		slog.Error("error al crear usuario", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error en el servidor",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, ToResponse(user))
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "Id inválido",
		})
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
				Code:    utils.ErrCodeNotFound,
				Message: "Usuario no encontrado",
			})
			return
		}
		slog.Error("error al obtener usuario", "user_id", id, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error en el servidor",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(user))
}

func (h *HTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Token inválido",
		})
		return
	}

	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
				Code:    utils.ErrCodeNotFound,
				Message: "Usuario no encontrado",
			})
			return
		}
		slog.Error("error al obtener usuario actual", "user_id", userID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error en el servidor",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(user))
}

func (h *HTTPHandler) Update(w http.ResponseWriter, r *http.Request) {

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Token inválido",
		})
		return
	}

	var req UserUpdate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "No se pudo parsear la request, por favor revise los datos enviados",
		})
		return
	}

	userUpdate, err := h.service.Update(r.Context(), userID, req)

	if err != nil {
		if errors.Is(err, ErrSameName) {
			utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
				Code:    utils.ErrCodeValidation,
				Message: "El nombre no puede ser igual al actual",
			})
			return
		}
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
				Code:    utils.ErrCodeValidation,
				Message: "Error de validación",
				Fields:  fields,
			})
			return
		}
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
				Code:    utils.ErrCodeNotFound,
				Message: "Usuario no encontrado",
			})
			return
		}
		slog.Error("error al actualizar usuario", "user_id", userID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "No se pudo actualizar el usuario",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(userUpdate))
}

func (h *HTTPHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Token inválido",
		})
		return
	}

	var req UserChangePassword

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "No se pudo parsear la request, por favor revise los datos enviados",
		})
		return
	}

	userUpdate, err := h.service.UpdatePassword(r.Context(), userID, req)

	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
				Code:    utils.ErrCodeValidation,
				Message: "Error de validación",
				Fields:  fields,
			})
			return
		}
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
				Code:    utils.ErrCodeNotFound,
				Message: "Usuario no encontrado",
			})
			return
		}
		slog.Error("error al actualizar contraseña", "user_id", userID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "No se pudo actualizar la contraseña",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, ToResponse(userUpdate))

}

func (h *HTTPHandler) SendEmailContact(w http.ResponseWriter, r *http.Request) {
	var req *mailer.ContactRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "No se pudo parsear la request, por favor revise los datos enviados",
		})
		return
	}

	_, err := h.getUserIDFromRequest(r)

	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Token inválido",
		})
		return
	}

	emailSend, err := h.service.SendEmailContact(r.Context(), *req)

	if err != nil {
		slog.Error("error al enviar consulta de contacto", "error", err)
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "Error al intentar enviar su consulta",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, emailSend)
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
