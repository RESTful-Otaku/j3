package db_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/restful-otaku/j3/internal/db"
)

func TestPragmasApplyToEveryPooledConnection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pragmas.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	store.SetMaxOpenConns(4)
	for i := 0; i < 8; i++ {
		var enabled int
		if err := store.QueryRowContext(context.Background(), `PRAGMA foreign_keys`).Scan(&enabled); err != nil {
			t.Fatal(err)
		}
		if enabled != 1 {
			t.Fatalf("foreign_keys disabled on connection %d", i)
		}
	}
}
