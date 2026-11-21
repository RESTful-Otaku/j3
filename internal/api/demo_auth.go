package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
)

const demoSessionLifetime = 12 * time.Hour

type demoRunnerAccount struct {
	ID    string
	Email string
	Role  string
}

// demoLogin is only active when J3_DEMO=1. It gives QA a runner session without
// going through an external identity provider. The runner is resolved from the
// database (and created on a completely empty venue DB) so the endpoint cannot
// fail with a foreign-key error when no Okami seed is present.
func (s *Server) demoLogin(w http.ResponseWriter, r *http.Request) {
	if !demoMode() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	account, err := s.demoRunner(r)
	if err != nil {
		writeError(w, err)
		return
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		writeError(w, err)
		return
	}
	token := hex.EncodeToString(buf)
	expiresAt := time.Now().UTC().Add(demoSessionLifetime).Format(sessionTimeLayout)
	_, err = s.Store.ExecContext(r.Context(), `
		INSERT INTO sessions (token, user_id, role, expires_at)
		VALUES (?, ?, ?, ?)`, token, account.ID, account.Role, expiresAt)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "email": account.Email, "role": "runner",
		"userId": account.ID, "expiresAt": expiresAt,
	})
}

// demoRunner prefers the seeded Okami/base runner, then any active runner, and
// finally bootstraps a minimal tenant + runner on an empty venue database.
func (s *Server) demoRunner(r *http.Request) (demoRunnerAccount, error) {
	var account demoRunnerAccount
	lookup := func() error {
		return s.Store.QueryRowContext(r.Context(), `
			SELECT id, email, role FROM users
			WHERE role = 'tournament_runner' AND status = 'active'
			ORDER BY CASE id WHEN 'u-tony' THEN 0 WHEN 'u-runner' THEN 1 ELSE 2 END, id
			LIMIT 1`).Scan(&account.ID, &account.Email, &account.Role)
	}
	err := lookup()
	if err == nil {
		account.Role = string(auth.CanonicalRole(account.Role))
		return account, nil
	}
	if err != sql.ErrNoRows {
		return demoRunnerAccount{}, err
	}
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		return demoRunnerAccount{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(), `
		INSERT OR IGNORE INTO tenants (id, name, slug, locale)
		VALUES ('demo-tenant', 'J3 Demo Dojo', 'demo', 'en')`); err != nil {
		return demoRunnerAccount{}, err
	}
	if _, err := tx.ExecContext(r.Context(), `
		INSERT OR IGNORE INTO users (id, tenant_id, role, display_name, email, status)
		VALUES ('u-demo-runner', 'demo-tenant', 'tournament_runner',
		        'Demo Runner', 'runner@demo.j3', 'active')`); err != nil {
		return demoRunnerAccount{}, err
	}
	if err := tx.Commit(); err != nil {
		return demoRunnerAccount{}, err
	}
	if err := lookup(); err != nil {
		return demoRunnerAccount{}, err
	}
	account.Role = string(auth.CanonicalRole(account.Role))
	return account, nil
}
