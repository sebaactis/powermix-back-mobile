package validations

type StructValidator interface {
	ValidateStruct(dst any) (map[string]string, bool)
}

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation error" }

func AsValidationError(err error) (map[string]string, bool) {
	if ve, ok := err.(*ValidationError); ok {
		return ve.Fields, true
	}
	return nil, false
}
