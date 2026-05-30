// Package token guarda refresh/reset tokens.
// Los endpoints HTTP que usan este servicio viven en internal/security/auth
// (por ejemplo RefreshToken en POST /api/v1/refreshToken).
package token

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

