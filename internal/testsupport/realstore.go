//go:build integration

package testsupport

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/amirhnajafiz/mayigoo/internal/db"
)

// One Postgres container is shared across every test in a package (spinning one
// per test would be painfully slow). Each test still gets a clean slate because
// NewRealStore truncates all tables in t.Cleanup. The container is reaped by
// testcontainers' ryuk sidecar when the test process exits, so no explicit
// teardown is needed.
var (
	pgOnce sync.Once
	pgDSN  string
	pgErr  error
)

func startPostgres() {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("mayigoo_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		pgErr = err
		return
	}

	pgDSN, pgErr = container.ConnectionString(ctx, "sslmode=disable")
}

// NewRealStore returns a *db.Store backed by a real, throwaway PostgreSQL
// running in a container, with your actual migrations applied. Each call hands
// back a clean database (all tables truncated, identity sequences reset), so
// tests can insert, delete, and query rows against real state:
//
//	store := testsupport.NewRealStore(t)
//	u, _ := store.UpsertUser(ctx, "a@b.io")
//	sa, _ := store.CreateServiceAccount(ctx, models.CreateServiceAccountParams{...})
//	_ = store.DeleteServiceAccount(ctx, sa.ID)
//	_, err := store.GetServiceAccount(ctx, sa.ID) // err == sql.ErrNoRows
func NewRealStore(t *testing.T) *db.Store {
	t.Helper()

	pgOnce.Do(startPostgres)
	if pgErr != nil {
		t.Fatalf("start postgres container: %v", pgErr)
	}

	conn, err := sql.Open("pgx", pgDSN)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		t.Fatalf("ping db: %v", err)
	}

	// Apply the real migrations (idempotent — only the first test in a package
	// actually creates the schema; later ones find it already present).
	if err := db.Migrate(conn); err != nil {
		_ = conn.Close()
		t.Fatalf("migrate: %v", err)
	}

	// Give this test a clean slate, and leave one behind when it finishes.
	truncateAll(t, conn)
	t.Cleanup(func() {
		truncateAll(t, conn)
		_ = conn.Close()
	})

	return db.NewStore(conn)
}

// truncateAll wipes every application table and resets identity sequences so
// SERIAL ids start from 1, making generated ids predictable across tests.
func truncateAll(t *testing.T, conn *sql.DB) {
	t.Helper()

	const stmt = `TRUNCATE service_account_kv, service_account_meta, service_accounts, users RESTART IDENTITY CASCADE`
	if _, err := conn.Exec(stmt); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
