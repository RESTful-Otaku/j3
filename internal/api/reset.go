package api

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// resetPassword completes a password reset using a token issued by the runner
// admin panel or the self-service request flow.
func (s *Server) completeReset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if len(body.Password) < 8 {
		writeBadRequest(w, "password must be at least 8 characters")
		return
	}
	var userID, expires string
	var used int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT user_id, expires_at, used FROM password_resets WHERE token = ?`, body.Token).
		Scan(&userID, &expires, &used); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired reset token"})
		return
	}
	expiry, err := time.Parse("2006-01-02T15:04:05.999Z", expires)
	if err != nil || used == 1 || time.Now().After(expiry) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired reset token"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, err)
		return
	}
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(),
		`UPDATE users SET password_hash = ?, must_change_password = 0,
		 updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`, string(hash), userID); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(),
		`UPDATE password_resets SET used = 1 WHERE token = ?`, body.Token); err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "password updated"})
}

// requestReset is the self-service counterpart to admin-initiated resets. It
// always reports success to avoid account enumeration; the QA token is only
// returned in demo mode.
func (s *Server) requestReset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	token := newID("reset")
	var userID string
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT id FROM users WHERE lower(email) = lower(?)`, body.Email).Scan(&userID); err == nil {
		_, _ = s.Store.ExecContext(r.Context(), `
			INSERT INTO password_resets (id, user_id, token, expires_at)
			VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now','+1 day'))`, token, userID, token)
	}
	response := map[string]any{"status": "if the account exists, a reset link has been sent"}
	if demoMode() && userID != "" {
		response["demoToken"] = token
	}
	writeJSON(w, http.StatusOK, response)
}
