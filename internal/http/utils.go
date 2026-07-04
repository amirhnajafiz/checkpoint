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

// randomState generates a random CSRF state value for the OAuth flow.
func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
