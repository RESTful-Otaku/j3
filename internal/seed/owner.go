package seed

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/config"
	"github.com/restful-otaku/j3/internal/db"
)

const ownerAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

// EnsureOwner idempotently creates the hosted owner/admin account. Local demo
// builds use the shared "demo" password; production either uses the configured
// password or generates a strong one that the caller must print once.
func EnsureOwner(store *db.DB, cfg config.Config, demo bool) (string, error) {
	email := strings.ToLower(strings.TrimSpace(cfg.OwnerEmail))
	if email == "" {
		email = "j3@local"
	}
	name := strings.TrimSpace(cfg.OwnerName)
	if name == "" {
		name = "J3 Owner"
	}

	tx, err := store.BeginTx(context.Background(), nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO tenants (id, name, slug, locale)
		VALUES (?, 'J3 Tournaments', 'j3', 'en')`, tenantID); err != nil {
		return "", err
	}
	var admins int
	if err := tx.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM users WHERE is_admin = 1 AND status = 'active'`).Scan(&admins); err != nil {
		return "", err
	}
	if admins > 0 {
		return "", tx.Commit()
	}

	var userID, storedHash, status string
	var isAdmin int
	err = tx.QueryRowContext(context.Background(), `
		SELECT id, password_hash, status, is_admin
		FROM users WHERE tenant_id = ? AND lower(email) = ?`, tenantID, email).
		Scan(&userID, &storedHash, &status, &isAdmin)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	generated := ""
	mustChange := 0
	password := cfg.OwnerPassword
	if demo {
		password = "demo"
	} else if password == "" {
		generated, err = randomOwnerPassword(24)
		if err != nil {
			return "", err
		}
		password = generated
		mustChange = 1
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err == nil && userID != "" {
		// Promote an existing account with this email instead of leaving a
		// second identity with owner rights.
		_, err = tx.ExecContext(context.Background(), `
			UPDATE users SET role = 'tournament_runner', is_admin = 1, status = 'active',
			       display_name = ?, password_hash = ?, must_change_password = ?,
			       updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
			WHERE id = ?`, name, string(hash), mustChange, userID)
	} else {
		userID = "u-owner"
		var exists int
		if err := tx.QueryRowContext(context.Background(),
			`SELECT 1 FROM users WHERE id = ?`, userID).Scan(&exists); err == nil {
			buf := make([]byte, 8)
			_, _ = rand.Read(buf)
			userID = "u-owner-" + hex.EncodeToString(buf)
		}
		_, err = tx.ExecContext(context.Background(), `
			INSERT INTO users (id, tenant_id, role, is_admin, display_name, email,
			                   password_hash, source, status, must_change_password)
			VALUES (?, ?, 'tournament_runner', 1, ?, ?, ?, 'independent', 'active', ?)`,
			userID, tenantID, name, email, string(hash), mustChange)
	}
	if err != nil {
		return "", err
	}
	after, _ := json.Marshal(map[string]any{"email": email, "name": name, "userId": userID})
	if _, err := tx.ExecContext(context.Background(), `
		INSERT INTO audit_log (tenant_id, actor_user_id, actor_role, action, target_type,
		                       target_id, after_json, created_at)
		VALUES (?, NULL, 'system', 'owner.bootstrap', 'user', ?, ?,
		       strftime('%Y-%m-%dT%H:%M:%fZ','now'))`, tenantID, userID, string(after)); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return generated, nil
}

func randomOwnerPassword(length int) (string, error) {
	var out strings.Builder
	for i := 0; i < length; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(ownerAlphabet))))
		if err != nil {
			return "", fmt.Errorf("generate owner password: %w", err)
		}
		out.WriteByte(ownerAlphabet[value.Int64()])
	}
	return out.String(), nil
}
