package validate

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
)

var engine = validator.New()

func BindAndValidate(c *gin.Context, data any, overrides ...func()) error {
	// 1. Bind JSON body
	if err := c.ShouldBindJSON(data); err != nil {
		var validationErrs validator.ValidationErrors

		if errors.As(err, &validationErrs) {
			return FormatValidationErrors(validationErrs)
		}

		return err
	}
	for _, override := range overrides {
		override()
	}
	if err := engine.Struct(data); err != nil {
		var validationErrs validator.ValidationErrors

		if errors.As(err, &validationErrs) {
			return FormatValidationErrors(validationErrs)
		}

		return err
	}

	return nil
}

func FormatValidationErrors(errs validator.ValidationErrors) error {
	errors := make(map[string]string)

	for _, err := range errs {
		field := err.Field()
		switch err.Tag() {
		case "required":
			errors[field] = "is required"
		case "email":
			errors[field] = "must be a valid email"
		case "min":
			errors[field] = fmt.Sprintf(
				"must be at least %s",
				err.Param(),
			)
		case "max":
			errors[field] = fmt.Sprintf(
				"must be at most %s",
				err.Param(),
			)
		case "len":
			errors[field] = fmt.Sprintf(
				"must be %s characters",
				err.Param(),
			)
		case "uuid":
			errors[field] = "must be a valid UUID"

		default:
			errors[field] = "is invalid"
		}
	}

	return fmt.Errorf("validation failed: %v", errors)
}
