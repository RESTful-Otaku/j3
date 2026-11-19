// Package db owns the SQLite store and schema migrations for the venue server.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the venue database.
type DB struct {
	*sql.DB
	Path string
}

// Open creates (if needed) and configures the venue database, then migrates it.
func Open(path string) (*DB, error) {
	if path != ":memory:" {
		if dir := filepath.Dir(path); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create data dir: %w", err)
			}
		}
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1) // WAL + single writer keeps event-day writes deterministic.
	for _, pragma := range []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA synchronous = NORMAL`,
	} {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("%s: %w", pragma, err)
		}
	}
	store := &DB{DB: conn, Path: path}
	if err := store.Migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return store, nil
}

// Migrate applies all embedded migrations not yet recorded.
func (d *DB) Migrate() error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Bootstrap the tracking table before checking whether migration 1 was applied.
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	)`); err != nil {
		return fmt.Errorf("bootstrap migrations table: %w", err)
	}
	for _, m := range migrations {
		if err := migrate(tx, m); err != nil {
			return err
		}
	}
	return tx.Commit()
}
