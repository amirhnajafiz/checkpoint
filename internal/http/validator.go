package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// requestValidator adapts go-playground/validator to the echo.Validator
// interface so handlers can call c.Validate(req).
type requestValidator struct {
	validate *validator.Validate
}

func newValidator() *requestValidator {
	return &requestValidator{validate: validator.New()}
}

func (v *requestValidator) Validate(i any) error {
	if err := v.validate.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, validationMessage(err)).SetInternal(err)
	}
	return nil
}

// validationMessage renders a validation failure into a message that names only
// the offending fields and rules, never the Go struct type behind them. The raw
// validator error (which does expose the type) is kept as the HTTPError's
// internal cause for logging.
func validationMessage(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) || len(ve) == 0 {
		return "invalid request payload"
	}

	fields := make([]string, 0, len(ve))
	for _, fe := range ve {
		fields = append(fields, strings.ToLower(fe.Field()))
	}
	return "invalid request: check field(s): " + strings.Join(fields, ", ")
}
