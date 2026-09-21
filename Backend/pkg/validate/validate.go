package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	validator "github.com/go-playground/validator/v10"
)

var engine = validator.New()

func Validate(value any) error {
	if value == nil {
		return errors.New("value is required")
	}
	if err := engine.Struct(value); err != nil {
		return formatError(err)
	}
	return nil
}

func BindAndValidate[T any](body io.Reader) (T, error) {
	var value T
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode request: %w", err)
	}
	if err := Validate(value); err != nil {
		return value, err
	}
	return value, nil
}

func formatError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("validate request: %w", err)
	}
	first := validationErrors[0]
	return fmt.Errorf("%s failed validation: %s", first.Field(), first.Tag())
}
