package http

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// bindAndValidate binds the JSON body into req and runs struct validation,
// returning a 400 HTTPError on failure.
func bindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	return nil
}

// pathID reads an integer path parameter (e.g. /:id) as an int32.
func pathID(c echo.Context, name string) (int32, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 32)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return int32(v), nil
}

// boolOrDefault dereferences an optional request bool, using def when nil.
func boolOrDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

// nullTime unwraps a sql.NullTime, returning the zero time when not valid.
func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(c echo.Context) string {
	const prefix = "Bearer "
	header := c.Request().Header.Get("Authorization")
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return header[len(prefix):]
	}
	return ""
}

// randomState generates a random CSRF state value for the OAuth flow.
func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// userEmail returns the authenticated user's email from the request context.
func userEmail(c echo.Context) string {
	email, _ := c.Get(contextUserEmail).(string)
	return email
}

// ownedAccount fetches a service account and confirms it belongs to the caller,
// returning 404 otherwise so account existence is not leaked across users.
func (h *Handler) ownedAccount(c echo.Context, id int32) (models.ServiceAccount, error) {
	account, err := h.store.GetServiceAccount(c.Request().Context(), id)
	if err != nil {
		return models.ServiceAccount{}, err
	}

	if account.UserEmail != userEmail(c) {
		return models.ServiceAccount{}, echo.NewHTTPError(http.StatusNotFound, "resource not found")
	}

	return account, nil
}
