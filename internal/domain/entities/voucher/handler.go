package voucher

import (
	"encoding/json"
	"errors"
	"log/slog"
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
		writeVoucherValidation(w, "Error al intentar parsear el request, por favor validar el mismo", nil)
		return
	}

	v, err := h.service.AssignNextVoucher(r.Context(), &req)

	if err != nil {
		if errors.Is(err, ErrNoAvailableVouchers) {
			writeVoucherConflict(w, "No hay vouchers disponibles en este momento")
			return
		}
		slog.Error("error al asignar voucher", "error", err)
		writeVoucherInternal(w, "Error en el servidor al intentar crear el voucher")
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, v)
}

func (h *HTTPHandler) GetAllByUserID(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	userID, ok := middlewares.UserIDFromContext(ctx)

	if !ok {
		writeVoucherUnauthorized(w)
		return
	}

	vouchers, err := h.service.GetAllByUserID(ctx, userID)

	if err != nil {
		slog.Error("error al recuperar vouchers del usuario", "user_id", userID, "error", err)
		writeVoucherInternal(w, "Error al recuperar los vouchers")
		return
	}
	utils.WriteSuccess(w, http.StatusOK, vouchers)

}

func (h *HTTPHandler) DeleteVoucher(w http.ResponseWriter, r *http.Request) {
	voucherIDStr := r.PathValue("id")

	if voucherIDStr == "" {
		writeVoucherValidation(w, "El id del voucher es requerido", nil)
		return
	}

	voucherID, err := uuid.Parse(voucherIDStr)
	if err != nil {
		writeVoucherValidation(w, "El formato del id del voucher es inválido", nil)
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		writeVoucherUnauthorized(w)
		return
	}

	if err := h.service.DeleteVoucher(r.Context(), voucherID, userID); err != nil {
		var (
			voucherNotFound         = errors.Is(err, ErrVoucherNotFound)
			voucherNotBelongsToUser = errors.Is(err, ErrVoucherNotBelongsToUser)
			voucherNotUsed          = errors.Is(err, ErrVoucherNotUsed)
		)

		if voucherNotFound {
			writeVoucherNotFound(w, "El voucher no existe")
			return
		}
		if voucherNotBelongsToUser {
			writeVoucherForbidden(w, "No tienes permiso para eliminar este voucher")
			return
		}
		if voucherNotUsed {
			writeVoucherValidation(w, "Solo se pueden eliminar vouchers que ya han sido usados", nil)
			return
		}

		slog.Error("error al eliminar voucher", "voucher_id", voucherID, "user_id", userID, "error", err)
		writeVoucherInternal(w, "Error al eliminar el voucher")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Voucher eliminado correctamente"})
}

func (h *HTTPHandler) GetAvailableCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.GetAvailableCount(r.Context())
	if err != nil {
		slog.Error("error al obtener cantidad de vouchers disponibles", "error", err)
		writeVoucherInternal(w, "Error al obtener la cantidad de vouchers disponibles")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]int64{"available": count})
}

func writeVoucherUnauthorized(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
		Code:    utils.ErrCodeUnauthorized,
		Message: "No se pudo recuperar el usuario de la sesión",
	})
}

func writeVoucherValidation(w http.ResponseWriter, message string, fields interface{}) {
	utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
		Code:    utils.ErrCodeValidation,
		Message: message,
		Fields:  fields,
	})
}

func writeVoucherNotFound(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusNotFound, utils.WriteErrorOpts{
		Code:    utils.ErrCodeNotFound,
		Message: message,
	})
}

func writeVoucherForbidden(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusForbidden, utils.WriteErrorOpts{
		Code:    utils.ErrCodeUnauthorized,
		Message: message,
	})
}

func writeVoucherConflict(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusConflict, utils.WriteErrorOpts{
		Code:    utils.ErrCodeConflict,
		Message: message,
	})
}

func writeVoucherInternal(w http.ResponseWriter, message string) {
	utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
		Code:    utils.ErrCodeInternal,
		Message: message,
	})
}
