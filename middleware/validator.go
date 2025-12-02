package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function to use json tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) map[string]string {
	errors := make(map[string]string)

	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			errors[field] = formatValidationError(err)
		}
	}

	return errors
}

// formatValidationError formats validation error messages
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param())
	case "uuid":
		return "Invalid UUID format"
	case "url":
		return "Invalid URL format"
	case "numeric":
		return "Must be a number"
	case "alpha":
		return "Must contain only letters"
	case "alphanum":
		return "Must contain only letters and numbers"
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

// ValidationErrorResponse sends validation error response
func ValidationErrorResponse(w http.ResponseWriter, errors map[string]string, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":      "Validation failed",
		"errors":     errors,
		"request_id": requestID,
	})
}
