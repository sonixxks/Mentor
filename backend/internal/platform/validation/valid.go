package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name != "" {
			return name
		}
		return f.Name
	})
	return v
}

func ValidationMessage(fe validator.FieldError) string {
	switch fe.ActualTag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param() + " characters long"
	case "max":
		return "must be at most " + fe.Param() + " characters long"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "field is not valid"
	}
}
