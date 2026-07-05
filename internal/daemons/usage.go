package daemons

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/amirhnajafiz/mayigoo/internal/db"
	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// UsageDaemon batches service-account validation events and flushes the
// aggregated counts to the database on a fixed interval, keeping the hot
// validate path off the database.
type UsageDaemon struct {
	store    *db.Store
	logger   *zap.Logger
	interval time.Duration
	events   chan int32
}

// NewUsageDaemon builds a UsageDaemon that flushes every interval. buffer is the
// capacity of the event channel; a full channel drops events rather than
// blocking callers.
func NewUsageDaemon(store *db.Store, logger *zap.Logger, interval time.Duration, buffer int) *UsageDaemon {
	if buffer <= 0 {
		buffer = 1
	}

	return &UsageDaemon{
		store:    store,
		logger:   logger,
		interval: interval,
		events:   make(chan int32, buffer),
	}
}

// Name identifies the daemon.
func (d *UsageDaemon) Name() string { return "usage" }

// Record reports that a service account's token was validated. It never blocks:
// if the buffer is full the event is dropped rather than stalling the caller.
func (d *UsageDaemon) Record(accountID int32) {
	select {
	case d.events <- accountID:
	default:
	}
}

// Run accumulates events and flushes them on each tick until ctx is cancelled,
// at which point it performs a final best-effort flush.
func (d *UsageDaemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	pending := make(map[int32]int64)
	for {
		select {
		case <-ctx.Done():
			// The parent context is gone; flush what we have with a fresh,
			// bounded context so counts are not lost on shutdown.
			flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			d.flush(flushCtx, pending)
			cancel()
			return nil
		case id := <-d.events:
			pending[id]++
		case <-ticker.C:
			d.flush(ctx, pending)
		}
	}
}

// flush writes each account's accumulated count and removes the entries it
// succeeds on. Failures are logged, not returned, so a transient database error
// doesn't kill the daemon; unflushed counts are retried on the next tick.
func (d *UsageDaemon) flush(ctx context.Context, pending map[int32]int64) {
	for id, count := range pending {
		if err := d.store.AddServiceAccountUsage(ctx, models.AddServiceAccountUsageParams{
			AccountID: id,
			Usage:     count,
		}); err != nil {
			d.logger.Warn("usage flush failed",
				zap.Int32("account_id", id), zap.Error(err))
			continue
		}
		delete(pending, id)
	}
}
