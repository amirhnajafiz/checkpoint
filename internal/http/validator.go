package http

import (
	"net/http"

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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
