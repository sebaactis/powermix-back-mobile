package token

import "errors"

var (
	ErrTokenNotFound = errors.New("token: no encontrado")
	ErrTokenInvalid  = errors.New("token: inválido o expirado")
	ErrInternal      = errors.New("token: error interno de persistencia")
)
