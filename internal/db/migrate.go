package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// migrationsFS embeds the SQL files so migrations ship inside the binary and
// run without the source tree being present at deploy time.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// bootstrapMigration creates the schema_migrations bookkeeping table. It is a
// regular migration file (version 0) so no schema lives as a hardcoded query in
// Go; it is run first, before applied versions can be read, and is idempotent.
const bootstrapMigration = "000000_schema_migrations.up.sql"

// migration is a single versioned step parsed from the migrations directory.
type migration struct {
	version int64
	name    string
	up      string
}

// Migrate applies every pending "up" migration in ascending version order.
func Migrate(conn *sql.DB) error {
	if err := ensureMigrationsTable(conn); err != nil {
		return err
	}

	applied, err := appliedVersions(conn)
	if err != nil {
		return err
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := applyMigration(conn, m); err != nil {
			return fmt.Errorf("apply migration %06d_%s: %w", m.version, m.name, err)
		}
	}

	return nil
}

func ensureMigrationsTable(conn *sql.DB) error {
	content, err := fs.ReadFile(migrationsFS, "migrations/"+bootstrapMigration)
	if err != nil {
		return fmt.Errorf("read bootstrap migration: %w", err)
	}
	if _, err := conn.Exec(string(content)); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func appliedVersions(conn *sql.DB) (map[int64]bool, error) {
	rows, err := conn.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyMigration(conn *sql.DB, m migration) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.up); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m.version); err != nil {
		return err
	}

	return tx.Commit()
}

// loadMigrations reads and parses every *.up.sql file from the embedded FS,
// returning them sorted by version.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migrations []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}

		version, name, err := parseName(e.Name())
		if err != nil {
			return nil, err
		}

		content, err := fs.ReadFile(migrationsFS, "migrations/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}

		migrations = append(migrations, migration{
			version: version,
			name:    name,
			up:      string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

// parseName splits a "000001_schema.up.sql" filename into version and name.
func parseName(filename string) (int64, string, error) {
	base := strings.TrimSuffix(filename, ".up.sql")

	idx := strings.Index(base, "_")
	if idx <= 0 {
		return 0, "", fmt.Errorf("malformed migration filename %q", filename)
	}

	version, err := strconv.ParseInt(base[:idx], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse version from %q: %w", filename, err)
	}

	return version, base[idx+1:], nil
}
