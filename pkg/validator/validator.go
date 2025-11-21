package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	v := validator.New()

	// Daftarkan custom validation untuk username & slug
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("slug", validateSlug)

	return &CustomValidator{
		validator: v,
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (cv *CustomValidator) GetErrors(err error) []map[string]string {
	var errors []map[string]string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, map[string]string{
				"field":   e.Field(),
				"message": getErrorMessage(e),
			})
		}
	}

	return errors
}

// === CUSTOM VALIDATION: username ===
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Regex: hanya huruf, angka, underscore, dan titik (tapi titik tidak di awal/akhir atau berurutan)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([._-]?[a-zA-Z0-9]+)*$`, username)
	if !matched {
		return false
	}

	// Minimal 3 karakter (sudah ditangani oleh `min=3`)
	// Maksimal 50 karakter (sudah ditangani oleh `max=50`)
	return true
}

func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}

// === Update getErrorMessage untuk menangani 'username' ===
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return e.Field() + " must be a valid email"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "alphanum":
		return e.Field() + " must contain only alphanumeric characters"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	case "uuid":
		return e.Field() + " must be a valid UUID"
	case "url":
		return e.Field() + " must be a valid URL"
	case "username":
		return e.Field() + " can only contain letters, numbers, and single dots/underscores (not at start/end)"
	case "slug":
		return e.Field() + " must be lowercase letters, numbers, and hyphens only"
	default:
		return e.Field() + " is invalid"
	}
}