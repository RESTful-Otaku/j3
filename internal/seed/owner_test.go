package seed

import (
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/config"
	"github.com/restful-otaku/j3/internal/db"
)

func TestEnsureOwnerGeneratesAndReusesHostedAdmin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "owner.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	cfg := config.Config{OwnerEmail: "j3@local", OwnerName: "J3 Owner"}
	generated, err := EnsureOwner(store, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(generated) < 16 {
		t.Fatalf("generated password = %q", generated)
	}
	var id, email string
	var admin, mustChange int
	if err := store.QueryRow(
		`SELECT id, email, is_admin, must_change_password FROM users WHERE lower(email) = 'j3@local'`).
		Scan(&id, &email, &admin, &mustChange); err != nil {
		t.Fatal(err)
	}
	if admin != 1 || mustChange != 1 {
		t.Fatalf("owner flags admin=%d mustChange=%d", admin, mustChange)
	}
	again, err := EnsureOwner(store, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if again != "" {
		t.Fatalf("second bootstrap generated %q", again)
	}
	var count int
	if err := store.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = 1`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("admin count = %d, want 1", count)
	}
}

func TestEnsureOwnerUsesSharedDemoPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "demo-owner.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	generated, err := EnsureOwner(store, config.Config{OwnerEmail: "j3@local"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if generated != "" {
		t.Fatalf("demo bootstrap generated %q", generated)
	}
	var hash string
	var mustChange int
	if err := store.QueryRow(
		`SELECT password_hash, must_change_password FROM users WHERE lower(email) = 'j3@local'`).
		Scan(&hash, &mustChange); err != nil {
		t.Fatal(err)
	}
	if mustChange != 0 || bcrypt.CompareHashAndPassword([]byte(hash), []byte("demo")) != nil {
		t.Fatal("demo owner must use the shared demo password without forced change")
	}
}
