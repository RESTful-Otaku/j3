package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
)

const demoSessionLifetime = 12 * time.Hour

type demoRunnerAccount struct {
	ID    string
	Email string
	Role  string
}

type demoLoginRequest struct {
	UserID string `json:"userId"`
}

// demoLogin is only active when J3_DEMO=1. With no body it signs QA in as a
// runner; with {"userId":"..."} it mints a real session for that seeded user so
// role switching enforces the same RBAC as a real login. The runner is resolved
// from the database (and created on an empty venue DB) so this endpoint cannot
// fail with a foreign-key error when a particular seed is absent.
func (s *Server) demoLogin(w http.ResponseWriter, r *http.Request) {
	if !demoMode() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	var body demoLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, err)
		return
	}
	var account demoRunnerAccount
	var err error
	if strings.TrimSpace(body.UserID) != "" {
		account, err = s.demoUser(r, strings.TrimSpace(body.UserID))
	} else {
		account, err = s.demoRunner(r)
	}
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
	canonical := auth.CanonicalRole(account.Role)
	_, err = s.Store.ExecContext(r.Context(), `
		INSERT INTO sessions (token, user_id, role, expires_at)
		VALUES (?, ?, ?, ?)`, token, account.ID, string(canonical), expiresAt)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "email": account.Email, "role": demoUIRole(canonical),
		"userId": account.ID, "expiresAt": expiresAt,
	})
}

// demoUIRole keeps the session payload compatible with the short role names
// used by the web UI while the database and middleware continue to use the
// canonical RBAC role names.
func demoUIRole(role auth.Role) string {
	switch role {
	case auth.RoleAdmin:
		return "admin"
	case auth.RoleTournamentRunner:
		return "runner"
	case auth.RoleOfficial:
		return "official"
	case auth.RoleClubLeader:
		return "club_leader"
	case auth.RoleFighter:
		return "fighter"
	default:
		return "spectator"
	}
}

// demoUser resolves any active seeded account for QA impersonation. This is
// deliberately demo-only: production sessions come from verified credentials.
func (s *Server) demoUser(r *http.Request, userID string) (demoRunnerAccount, error) {
	var account demoRunnerAccount
	var admin int
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT id, email, role, is_admin FROM users WHERE id = ? AND status = 'active'`, userID).
		Scan(&account.ID, &account.Email, &account.Role, &admin)
	if err != nil {
		return demoRunnerAccount{}, err
	}
	role := auth.VirtualRole(account.Role, admin == 1)
	if role == "" {
		return demoRunnerAccount{}, sql.ErrNoRows
	}
	account.Role = string(role)
	return account, nil
}

// demoRunner prefers the seeded Okami/base runner, then any active runner, and
// finally bootstraps a minimal tenant + runner on an empty venue database.
func (s *Server) demoRunner(r *http.Request) (demoRunnerAccount, error) {
	var account demoRunnerAccount
	var admin int
	lookup := func() error {
		return s.Store.QueryRowContext(r.Context(), `
			SELECT id, email, role, is_admin FROM users
			WHERE role = 'tournament_runner' AND status = 'active'
			ORDER BY is_admin DESC,
			         CASE id WHEN 'u-owner' THEN 0 WHEN 'u-tony' THEN 1 WHEN 'u-runner' THEN 2 ELSE 3 END, id
			LIMIT 1`).Scan(&account.ID, &account.Email, &account.Role, &admin)
	}
	err := lookup()
	if err == nil {
		account.Role = string(auth.VirtualRole(account.Role, admin == 1))
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
	account.Role = string(auth.VirtualRole(account.Role, admin == 1))
	return account, nil
}
