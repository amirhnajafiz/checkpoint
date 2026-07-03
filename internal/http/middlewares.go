package http

import (
	"database/sql"
	"errors"
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

// errorHandler renders every error as a JSON errorResponse, mapping common
// database and framework errors to appropriate status codes.
func errorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status := http.StatusInternalServerError
	message := "internal server error"

	var he *echo.HTTPError
	switch {
	case errors.Is(err, sql.ErrNoRows):
		status = http.StatusNotFound
		message = "resource not found"
	case errors.As(err, &he):
		status = he.Code
		if m, ok := he.Message.(string); ok {
			message = m
		} else {
			message = http.StatusText(he.Code)
		}
	}

	_ = c.JSON(status, errorResponse{Message: message})
}
