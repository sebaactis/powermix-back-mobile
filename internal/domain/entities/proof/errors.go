package proof

import "errors"

var (
	ErrInternal              = errors.New("proof: error interno de persistencia")
	ErrProofDuplicateID      = errors.New("proof: ya tenes guardado un comprobante con este ID")
	ErrProofNotFoundID       = errors.New("proof: el comprobante no existe, verifique los datos")
	ErrPaymentNotFound       = errors.New("proof: no se encontró un pago que coincida")
	ErrProofDuplicateMP      = errors.New("proof: ya guardaste un comprobante con este pago de Mercado Pago")
	ErrProofIDRequired       = errors.New("proof: id es requerido")
)
