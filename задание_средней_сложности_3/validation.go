package main

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError представляет ошибку валидации с читаемым сообщением.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// formatValidationErrors преобразует ошибки validator в понятные сообщения.
func formatValidationErrors(err error) []ValidationError {
	var errs []ValidationError

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []ValidationError{{Field: "body", Message: err.Error()}}
	}

	for _, fe := range validationErrors {
		field := strings.ToLower(fe.Field())
		errs = append(errs, ValidationError{
			Field:   field,
			Message: buildMessage(fe),
		})
	}
	return errs
}

func buildMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("поле '%s' обязательно", strings.ToLower(fe.Field()))
	case "email":
		return "некорректный формат email"
	case "min":
		if fe.Type().Kind().String() == "string" {
			return fmt.Sprintf("минимальная длина — %s символов", fe.Param())
		}
		return fmt.Sprintf("минимальное значение — %s", fe.Param())
	case "max":
		if fe.Type().Kind().String() == "string" {
			return fmt.Sprintf("максимальная длина — %s символов", fe.Param())
		}
		return fmt.Sprintf("максимальное значение — %s", fe.Param())
	default:
		return fmt.Sprintf("не прошло валидацию: %s", fe.Tag())
	}
}

// ValidateStruct запускает валидацию структуры и возвращает срез ошибок.
func ValidateStruct(s interface{}) []ValidationError {
	if err := validate.Struct(s); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}
