package daemons

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Manager runs a set of daemons as background goroutines and stops them together
// when the context is cancelled (e.g. on a termination signal).
type Manager struct {
	logger  *zap.Logger
	daemons []Daemon
}

// NewManager builds a Manager for the given daemons.
func NewManager(logger *zap.Logger, daemons ...Daemon) *Manager {
	return &Manager{logger: logger, daemons: daemons}
}

// Run starts every daemon under an errgroup and blocks until the context is
// cancelled or a daemon returns an error. When one daemon fails, the group's
// context is cancelled so the rest shut down too; Run then returns that error.
func (m *Manager) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, d := range m.daemons {
		g.Go(func() error {
			m.logger.Info("daemon started", zap.String("daemon", d.Name()))
			err := d.Run(ctx)
			m.logger.Info("daemon stopped", zap.String("daemon", d.Name()))
			return err
		})
	}

	return g.Wait()
}
