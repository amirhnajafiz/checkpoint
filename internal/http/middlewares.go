package http

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
)

// contextUserEmail is the echo context key holding the authenticated user email.
const contextUserEmail = "user_email"

// authMiddleware requires a valid user JWT in the Authorization header and
// stores the caller's email in the request context.
func (h *Handler) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := bearerToken(c)
		if raw == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
		}

		claims, err := h.jwtManager.Parse(raw)
		if err != nil || claims.JWTKind != auth.JWTKindUser || claims.Subject == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		c.Set(contextUserEmail, claims.Subject)
		return next(c)
	}
}

// errorHandler renders every error as a JSON errorResponse, mapping common
// database and framework errors to appropriate status codes. It logs the full
// underlying error server-side (choosing a level by status class) while the
// response body carries only a sanitized message, so internal details are never
// exposed to end users.
func (h *Handler) errorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status := http.StatusInternalServerError
	message := "internal server error"

	// detail is the error logged server-side; it defaults to err but is
	// replaced by an HTTPError's internal cause when one is attached, so the
	// real reason survives in the logs even though the client sees only message.
	detail := err

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
		if he.Internal != nil {
			detail = he.Internal
		}
	}

	req := c.Request()
	fields := []zap.Field{
		zap.String("method", req.Method),
		zap.String("uri", req.RequestURI),
		zap.Int("status", status),
		zap.Error(detail),
	}

	// 5xx are unexpected server-side failures worth an error-level alert; 4xx
	// are expected client mistakes and only warrant an info-level record.
	if status >= http.StatusInternalServerError {
		h.logger.Error("request failed", fields...)
	} else {
		h.logger.Info("request rejected", fields...)
	}

	_ = c.JSON(status, errorResponse{Message: message})
}
