package utils

import "github.com/go-playground/validator/v10"

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

func ValidateStruct(m interface{}) error {
	var errors []*ErrorResponse
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
		return &ValidationError{Errors: errors}
	}
	return nil
}

type ValidationError struct {
	Errors []*ErrorResponse
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
