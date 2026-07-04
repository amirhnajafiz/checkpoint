package http

import (
	"database/sql"
	"errors"
	"net/http"

	oauth "github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// contextUserEmail is the echo context key holding the authenticated user email.
const contextUserEmail = "user_email"

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

// authMiddleware requires a valid user JWT in the Authorization header and
// stores the caller's email in the request context.
func (h *Handler) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := bearerToken(c)
		if raw == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
		}

		claims, err := h.jwtManager.Parse(raw)
		if err != nil || claims.JWTKind != oauth.JWTKindUser || claims.Subject == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		c.Set(contextUserEmail, claims.Subject)
		return next(c)
	}
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
