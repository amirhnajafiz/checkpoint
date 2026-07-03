package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// Store is the single entry point future components (HTTP handlers, services,
// ...) use to talk to the database. It embeds the sqlc-generated *models.Queries
// so every CRUD method is available directly on the Store, while keeping the
// underlying *sql.DB around so multiple queries can be composed atomically.
type Store struct {
	*models.Queries
	db *sql.DB
}

// NewStore builds a Store on top of an existing database connection pool (the
// one returned by New).
func NewStore(conn *sql.DB) *Store {
	return &Store{
		Queries: models.New(conn),
		db:      conn,
	}
}

// ExecTx runs fn inside a single database transaction. The *models.Queries
// handed to fn is scoped to that transaction, so every call made through it is
// part of the same atomic unit. The transaction is committed when fn returns
// nil and rolled back otherwise.
func (s *Store) ExecTx(ctx context.Context, fn func(*models.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(s.Queries.WithTx(tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
