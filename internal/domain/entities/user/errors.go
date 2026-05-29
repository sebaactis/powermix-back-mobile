package user

import "errors"

var (
	ErrNotFound       = errors.New("user: usuario no encontrado")
	ErrInternal       = errors.New("user: error interno de persistencia")
	ErrDuplicateEmail = errors.New("el email ya esta en uso")
)
