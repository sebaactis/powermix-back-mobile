package prode

import "errors"

var (
	ErrMatchNotFound           = errors.New("prode: partido no encontrado")
	ErrPredictionNotFound      = errors.New("prode: predicción no encontrada")
	ErrCutoffPassed            = errors.New("prode: la hora límite de predicción ya pasó")
	ErrInvalidScore            = errors.New("prode: el resultado debe ser no negativo")
	ErrScoreOutOfRange         = errors.New("prode: el resultado excede el rango permitido")
	ErrMatchNotOpen            = errors.New("prode: el partido no está abierto para predicciones")
	ErrResultMissing           = errors.New("prode: el resultado del partido aún no fue cargado")
	ErrDuplicatePrediction     = errors.New("prode: el usuario ya tiene una predicción para este partido")
	ErrPredictionAlreadyLocked = errors.New("prode: la predicción está bloqueada y no puede editarse")
	ErrRewardAlreadyFulfilled  = errors.New("prode: el premio ya fue otorgado")
	ErrRewardNotFound          = errors.New("prode: premio no encontrado")
	ErrMaintenanceDisabled     = errors.New("prode: el modo mantenimiento está deshabilitado")
	ErrInvalidAdminKey         = errors.New("prode: clave de administración inválida")
	ErrProdeDisabled           = errors.New("prode: la funcionalidad está deshabilitada")
)
