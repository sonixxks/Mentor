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
		return "обязательное поле"
	case "email":
		return "некорректная почта"
	case "min":
		return "не короче " + fe.Param() + " символов"
	case "max":
		return "не длиннее " + fe.Param() + " символов"
	case "oneof":
		return "допустимые значения: " + fe.Param()
	case "len":
		return "должно быть ровно " + fe.Param() + " символов"
	case "numeric":
		return "только цифры"
	default:
		return "некорректное значение"
	}
}
