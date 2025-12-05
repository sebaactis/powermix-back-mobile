package validations

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	return &Validator{v: v}
}

func (val *Validator) ValidateStruct(dst any) (map[string]string, bool) {
	if err := val.v.Struct(dst); err != nil {
		fieldErrs := map[string]string{}
		for _, fe := range err.(validator.ValidationErrors) {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				fieldErrs[field] = "Es requerido"
			case "email":
				fieldErrs[field] = "Tenes que ingresar un mail válido"
			case "gt":
				fieldErrs[field] = "Tiene que ser mayor a: " + fe.Param()
			case "min":
				fieldErrs[field] = "Los caracteres deben ser al menos: " + fe.Param()
			case "max":
				fieldErrs[field] = "Los caracteres deben ser maximo de: " + fe.Param()
			case "nefield":
				fieldErrs[field] = "Tiene que ser diferente de " + fe.Param()
			case "eqfield":
				fieldErrs[field] = "Tiene que ser igual que " + fe.Param()
			default:
				fieldErrs[field] = strings.ToLower(fe.Tag())
			}
		}
		return fieldErrs, false
	}
	return nil, true
}

func (val *Validator) BindAndValidateJSON(r *http.Request, dst any) (map[string]string, bool) {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return map[string]string{"body": "invalid JSON"}, false
	}
	return val.ValidateStruct(dst)
}
