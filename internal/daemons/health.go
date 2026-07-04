package daemons

import (
	"context"
	"time"
)

// Checker is a named health probe for a third-party dependency.
type Checker struct {
	Name  string
	Check func(ctx context.Context) error
}

// ComponentHealth is the latest probe result for a single dependency.
type ComponentHealth struct {
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}

// Health is a point-in-time snapshot of every dependency's health.
type Health struct {
	Healthy    bool                       `json:"healthy"`
	Components map[string]ComponentHealth `json:"components"`
	CheckedAt  time.Time                  `json:"checked_at"`
}

// HealthDaemon periodically probes third-party dependencies (database, cache,
// ...) and serves the most recent snapshot over a request channel, so its own
// goroutine remains the sole owner of the state (no locks).
type HealthDaemon struct {
	interval time.Duration
	timeout  time.Duration
	checkers []Checker
	requests chan chan Health
}

// NewHealthDaemon builds a HealthDaemon that probes the given checkers every
// interval.
func NewHealthDaemon(interval time.Duration, checkers ...Checker) *HealthDaemon {
	return &HealthDaemon{
		interval: interval,
		timeout:  5 * time.Second,
		checkers: checkers,
		requests: make(chan chan Health),
	}
}

// Name identifies the daemon.
func (d *HealthDaemon) Name() string { return "health" }

// Health returns the latest snapshot. The request travels over a channel to the
// daemon goroutine, which replies on a channel supplied by the caller.
func (d *HealthDaemon) Health(ctx context.Context) (Health, error) {
	reply := make(chan Health, 1)

	select {
	case d.requests <- reply:
	case <-ctx.Done():
		return Health{}, ctx.Err()
	}

	select {
	case h := <-reply:
		return h, nil
	case <-ctx.Done():
		return Health{}, ctx.Err()
	}
}

// Run probes on each tick and answers health requests until ctx is cancelled.
func (d *HealthDaemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	// Probe once up front so readers never observe an empty snapshot.
	latest := d.probe(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			latest = d.probe(ctx)
		case reply := <-d.requests:
			reply <- latest
		}
	}
}

// probe runs every checker (each with its own timeout) and aggregates the
// results into a snapshot.
func (d *HealthDaemon) probe(ctx context.Context) Health {
	h := Health{
		Healthy:    true,
		Components: make(map[string]ComponentHealth, len(d.checkers)),
		CheckedAt:  time.Now(),
	}

	for _, c := range d.checkers {
		cctx, cancel := context.WithTimeout(ctx, d.timeout)
		err := c.Check(cctx)
		cancel()

		comp := ComponentHealth{Healthy: err == nil}
		if err != nil {
			comp.Error = err.Error()
			h.Healthy = false
		}
		h.Components[c.Name] = comp
	}

	return h
}
