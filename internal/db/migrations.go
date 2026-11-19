package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var list []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration name %q", entry.Name())
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %q: %w", entry.Name(), err)
		}
		body, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, err
		}
		list = append(list, migration{version: version, name: entry.Name(), sql: string(body)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })
	return list, nil
}

func migrate(tx *sql.Tx, m migration) error {
	var applied int
	err := tx.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, m.version).Scan(&applied)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	if _, err := tx.Exec(m.sql); err != nil {
		return fmt.Errorf("apply %s: %w", m.name, err)
	}
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, m.version, m.name,
	); err != nil {
		return fmt.Errorf("record %s: %w", m.name, err)
	}
	return nil
}
