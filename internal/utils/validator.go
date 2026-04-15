package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"swap/iguti/swap-service/internal/dto"
)

// Validator es una instancia global de validator
type Validator struct {
	validate *validator.Validate
}

// NewValidator crea una nueva instancia del validador con validaciones personalizadas
func NewValidator() *Validator {
	validate := validator.New()

	// Registrar validaciones personalizadas
	validate.RegisterValidation("latitude", validateLatitude)
	validate.RegisterValidation("longitude", validateLongitude)
	validate.RegisterValidation("not_zero", validateNotZero)

	return &Validator{
		validate: validate,
	}
}

// validateLatitude valida que el valor sea una latitud válida (-90 a 90)
func validateLatitude(fl validator.FieldLevel) bool {
	lat, ok := fl.Field().Interface().(float64)
	if !ok {
		return false
	}
	return lat >= -90 && lat <= 90
}

// validateLongitude valida que el valor sea una longitud válida (-180 a 180)
func validateLongitude(fl validator.FieldLevel) bool {
	lng, ok := fl.Field().Interface().(float64)
	if !ok {
		return false
	}
	return lng >= -180 && lng <= 180
}

// validateNotZero valida que el valor no sea cero
func validateNotZero(fl validator.FieldLevel) bool {
	lat, ok := fl.Field().Interface().(float64)
	if !ok {
		return true
	}
	return lat != 0
}

// ValidateStruct valida una estructura y retorna un error formateado
func (v *Validator) ValidateStruct(s interface{}) *dto.APIResponse {
	if err := v.validate.Struct(s); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return v.formatValidationErrors(validationErrors)
	}
	return nil
}

// formatValidationErrors formatea los errores de validación en un mensaje legible
func (v *Validator) formatValidationErrors(errors validator.ValidationErrors) *dto.APIResponse {
	var messages []string
	for _, err := range errors {
		field := err.Field()
		tag := err.Tag()
		param := err.Param()

		msg := v.getErrorMessage(field, tag, param)
		messages = append(messages, msg)
	}

	return &dto.APIResponse{
		Success: false,
		Status:  400,
		Message: "Error de validación en los datos enviados",
		Errors:  strings.Join(messages, "; "),
	}
}

// getErrorMessage retorna un mensaje de error legible según el tipo de validación
func (v *Validator) getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("El campo '%s' es obligatorio", field)
	case "email":
		return fmt.Sprintf("El campo '%s' debe ser un email válido", field)
	case "min":
		return fmt.Sprintf("El campo '%s' debe tener al menos %s caracteres", field, param)
	case "max":
		return fmt.Sprintf("El campo '%s' no debe exceder %s caracteres", field, param)
	case "uuid":
		return fmt.Sprintf("El campo '%s' debe ser un UUID válido", field)
	case "url":
		return fmt.Sprintf("El campo '%s' debe ser una URL válida", field)
	case "oneof":
		return fmt.Sprintf("El campo '%s' debe ser uno de: %s", field, param)
	case "latitude":
		return fmt.Sprintf("El campo '%s' debe ser una latitud válida entre -90 y 90", field)
	case "longitude":
		return fmt.Sprintf("El campo '%s' debe ser una longitud válida entre -180 y 180", field)
	case "not_zero":
		return fmt.Sprintf("El campo '%s' no puede ser cero", field)
	default:
		return fmt.Sprintf("El campo '%s' no cumple con la validación '%s'", field, tag)
	}
}

// ValidateAndParse valida y parsea el body JSON en la estructura proporcionada
func ValidateAndParse(body interface{}, target interface{}) *dto.APIResponse {
	validator := NewValidator()
	return validator.ValidateStruct(target)
}
