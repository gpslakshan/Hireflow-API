package handler

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// extractUserID pulls the user_id claim injected by AuthMiddleware
// from the Gin request context.
func extractUserID(c *gin.Context) (uuid.UUID, error) {
	raw, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}
	return uuid.Parse(raw.(string))
}

// formatValidationErrors converts go-playground/validator errors into a
// clean map keyed by JSON field name with human-readable messages.
//
// Input:  validator.ValidationErrors (raw library output)
// Output: map[string]string{"email": "must be a valid email address", ...}
func formatValidationErrors(err error) map[string]string {
	result := make(map[string]string)

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		// Not a validation error — return generic message
		result["error"] = "invalid input"
		return result
	}

	for _, fe := range validationErrs {
		// fe.Field()    → Go struct field name  e.g. "FullName"
		// fe.Tag()      → validation rule name  e.g. "min"
		// fe.Param()    → rule parameter        e.g. "2" (for min=2)
		// fe.Value()    → the submitted value
		//
		// We use fe.Field() to look up the json tag name via a separate
		// map so the response uses "full_name" not "FullName"
		field := toSnakeCase(fe.Field())
		result[field] = humanMessage(fe)
	}

	return result
}

// humanMessage converts a single FieldError into a readable sentence.
func humanMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "url":
		return "must be a valid URL"
	case "uuid4":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed validation on '%s'", fe.Tag())
	}
}

// toSnakeCase converts a PascalCase Go field name to snake_case
// so responses use "full_name" instead of "FullName".
func toSnakeCase(s string) string {
	result := []rune{}
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32) // convert to lowercase
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
