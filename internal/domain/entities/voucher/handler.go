package voucher

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
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
	var req VoucherRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al intentar parsear el request, por favor validar el mismo", nil)
		return
	}

	v, err := h.service.AssignNextVoucher(r.Context(), &req)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error en el servidor al intentar crear el voucher", map[string]string{"error": err.Error()})
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, v)
}

func (h *HTTPHandler) GetAllByUserId(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	userId, ok := middlewares.UserIDFromContext(ctx)

	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el userId del contexto", nil)
		return
	}

	vouchers, err := h.service.GetAllByUserId(ctx, userId)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Error al recuperar los vouchers", err)
		return
	}
	utils.WriteSuccess(w, http.StatusOK, vouchers)

}

func (h *HTTPHandler) DeleteVoucher(w http.ResponseWriter, r *http.Request) {
	voucherIDStr := r.PathValue("id")

	if voucherIDStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "El id del voucher es requerido", nil)
		return
	}

	voucherID, err := uuid.Parse(voucherIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "El formato del id del voucher es inválido", nil)
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el userId del contexto", nil)
		return
	}

	if err := h.service.DeleteVoucher(r.Context(), voucherID, userID); err != nil {
		var (
			voucherNotFound         = errors.Is(err, ErrVoucherNotFound)
			voucherNotBelongsToUser = errors.Is(err, ErrVoucherNotBelongsToUser)
			voucherNotUsed          = errors.Is(err, ErrVoucherNotUsed)
		)

		if voucherNotFound {
			utils.WriteError(w, http.StatusNotFound, "El voucher no existe", nil)
			return
		}
		if voucherNotBelongsToUser {
			utils.WriteError(w, http.StatusForbidden, "No tienes permiso para eliminar este voucher", nil)
			return
		}
		if voucherNotUsed {
			utils.WriteError(w, http.StatusBadRequest, "Solo se pueden eliminar vouchers que ya han sido usados", nil)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "Error al eliminar el voucher", map[string]string{"error": err.Error()})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Voucher eliminado correctamente"})
}
