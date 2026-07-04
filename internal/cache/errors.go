package cache

import "errors"

// ErrNotFound is returned by GetServiceToken when no token is cached for the
// account (missing or expired).
var ErrNotFound = errors.New("cache: token not found")
