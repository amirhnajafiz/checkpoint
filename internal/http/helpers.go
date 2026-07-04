package http

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(c echo.Context) string {
	const prefix = "Bearer "
	header := c.Request().Header.Get("Authorization")
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return header[len(prefix):]
	}
	return ""
}

// userEmail returns the authenticated user's email from the request context.
func userEmail(c echo.Context) string {
	email, _ := c.Get(contextUserEmail).(string)
	return email
}

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

// parseTTL converts an optional duration string (e.g. "24h", "90m") from a
// request into a nullable seconds value for storage. An empty string yields a
// NULL, meaning the account falls back to the default TTL.
func parseTTL(s string) (sql.NullInt64, error) {
	if strings.TrimSpace(s) == "" {
		return sql.NullInt64{}, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return sql.NullInt64{}, echo.NewHTTPError(http.StatusBadRequest, "invalid ttl; use a positive duration like 24h or 90m")
	}

	return sql.NullInt64{Int64: int64(d.Seconds()), Valid: true}, nil
}

// pathID reads an integer path parameter (e.g. /:id) as an int32.
func pathID(c echo.Context, name string) (int32, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 32)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return int32(v), nil
}

// setAccountKV replaces an account's key/value labels with the given map.
func setAccountKV(ctx context.Context, q *models.Queries, accountID int32, kv map[string]string) error {
	if err := q.DeleteServiceAccountKVByAccount(ctx, accountID); err != nil {
		return err
	}
	for k, v := range kv {
		if _, err := q.SetServiceAccountKV(ctx, models.SetServiceAccountKVParams{
			AccountID: accountID,
			Xkey:      k,
			Xvalue:    v,
		}); err != nil {
			return err
		}
	}
	return nil
}
