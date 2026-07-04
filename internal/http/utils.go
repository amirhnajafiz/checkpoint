package http

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

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

// ttlDuration converts a stored per-account TTL (seconds) into a Duration,
// returning ok=false when unset so callers can fall back to the default TTL.
func ttlDuration(ttlSeconds sql.NullInt64) (time.Duration, bool) {
	if ttlSeconds.Valid && ttlSeconds.Int64 > 0 {
		return time.Duration(ttlSeconds.Int64) * time.Second, true
	}
	return 0, false
}

// ttlDisplay renders a stored per-account TTL for API responses, returning ""
// when the account uses the default TTL.
func ttlDisplay(ttlSeconds sql.NullInt64) string {
	if d, ok := ttlDuration(ttlSeconds); ok {
		return d.String()
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
