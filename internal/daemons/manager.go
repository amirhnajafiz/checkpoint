package daemons

import (
	"context"
	"log"

	"golang.org/x/sync/errgroup"
)

// Manager runs a set of daemons as background goroutines and stops them together
// when the context is cancelled (e.g. on a termination signal).
type Manager struct {
	daemons []Daemon
}

// NewManager builds a Manager for the given daemons.
func NewManager(daemons ...Daemon) *Manager {
	return &Manager{daemons: daemons}
}

// Run starts every daemon under an errgroup and blocks until the context is
// cancelled or a daemon returns an error. When one daemon fails, the group's
// context is cancelled so the rest shut down too; Run then returns that error.
func (m *Manager) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, d := range m.daemons {
		d := d
		g.Go(func() error {
			log.Printf("daemon %q started", d.Name())
			err := d.Run(ctx)
			log.Printf("daemon %q stopped", d.Name())
			return err
		})
	}

	return g.Wait()
}
