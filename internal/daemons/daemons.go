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
