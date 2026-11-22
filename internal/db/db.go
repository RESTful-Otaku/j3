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
	// Pragmas belong in the DSN so every pooled connection gets them; executing
	// PRAGMA on *sql.DB only configures whichever connection handles that call.
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// In-memory databases must stay on one connection; file-backed venues use a
	// small pool with WAL so slow reads cannot block other requests.
	maxOpen := 8
	if path == ":memory:" {
		maxOpen = 1
	}
	conn.SetMaxOpenConns(maxOpen)
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connect sqlite: %w", err)
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
