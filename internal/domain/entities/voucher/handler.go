package voucher

import (
	"encoding/json"
	"net/http"

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

	v, err := h.service.Create(r.Context(), &req)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error en el servidor al intentar crear el voucher", map[string]string{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, v)
}
