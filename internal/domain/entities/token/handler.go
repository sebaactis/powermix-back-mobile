package token

import (
	"net/http"

	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.service.GetAll(r.Context())

	if err != nil {
		errors := make(map[string]string)
		errors["error"] = err.Error()
		utils.WriteError(w, http.StatusBadRequest, "El requests enviado es invalido", errors)
		return
	}

	utils.WriteJSON(w, http.StatusOK, ToResponseMany(tokens))

}
