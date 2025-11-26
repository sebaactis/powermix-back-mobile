package proof

import (
    "encoding/json"
    "net/http"

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
    var req ProofRequest

    userId, ok := middlewares.UserIDFromContext(r.Context())
    if !ok {
        utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el userId del contexto", nil)
        return
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.WriteError(w, http.StatusBadRequest, "Error al intentar parsear el request, por favor validar el mismo", nil)
        return
    }

    req.UserID = userId

    proof, err := h.service.Create(r.Context(), &req)
    if err != nil {
        utils.WriteError(w, http.StatusBadRequest, "Error al agregar el comprobante de pago", map[string]string{"error": err.Error()})
        return
    }

    utils.WriteJSON(w, http.StatusCreated, proof)
}

func (h *HTTPHandler) CreateFromOthers(w http.ResponseWriter, r *http.Request) {
    var req ProofOthersRequest

    userId, ok := middlewares.UserIDFromContext(r.Context())
    if !ok {
        utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el userId del contexto", nil)
        return
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.WriteError(w, http.StatusBadRequest, "Error al intentar parsear el request, por favor validar el mismo", nil)
        return
    }

    req.UserID = userId

    proof, err := h.service.CreateFromOthers(r.Context(), &req)
    if err != nil {
        utils.WriteError(w, http.StatusBadRequest, "Error al agregar el comprobante de pago (otros medios)", map[string]string{"error": err.Error()})
        return
    }

    utils.WriteJSON(w, http.StatusCreated, proof)
}

func (h *HTTPHandler) GetAllByUserId(w http.ResponseWriter, r *http.Request) {
    userId, ok := middlewares.UserIDFromContext(r.Context())
    if !ok {
        utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el userId del contexto", nil)
        return
    }

    proofs, err := h.service.GetAllByUserId(r.Context(), userId)
    if err != nil {
        utils.WriteError(w, http.StatusBadRequest, "No se pudieron recuperar los comprobantes del usuario", nil)
        return
    }

    utils.WriteJSON(w, http.StatusOK, proofs)
}

func (h *HTTPHandler) GetLastThreeByUserId(w http.ResponseWriter, r *http.Request) {
    userId, ok := middlewares.UserIDFromContext(r.Context())
    if !ok {
        utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el userId del contexto", nil)
        return
    }

    proofs, err := h.service.GetLastThreeByUserId(r.Context(), userId)
    if err != nil {
        utils.WriteError(w, http.StatusBadRequest, "No se pudieron recuperar los comprobantes del usuario", nil)
        return
    }

    utils.WriteJSON(w, http.StatusOK, proofs)
}


func (h *HTTPHandler) GetById(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    if id == "" {
        utils.WriteError(w, http.StatusBadRequest, "El id del comprobante de pago es requerido", nil)
        return
    }

    proof, err := h.service.GetById(r.Context(), id)
    if err != nil {
        utils.WriteError(w, http.StatusBadRequest, "No se pudo recuperar el comprobante de pago", nil)
        return
    }

    utils.WriteJSON(w, http.StatusOK, proof)
}
