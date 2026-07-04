// Package daemons runs long-lived background workers as goroutines under a
// single manager. Each daemon is driven by a context: cancelling it (e.g. on a
// termination signal) stops every daemon together.
package daemons

import "context"

// Daemon is a background worker managed by the Manager. Run blocks until ctx is
// cancelled (returning nil) or the daemon hits a fatal error.
type Daemon interface {
	// Name identifies the daemon in logs.
	Name() string
	// Run executes the daemon until ctx is cancelled.
	Run(ctx context.Context) error
}
