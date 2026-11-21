package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// demoLogin is only registered/active when J3_DEMO=1. It gives QA a runner
// session without going through an external identity provider.
func (s *Server) demoLogin(w http.ResponseWriter, r *http.Request) {
	if !demoMode() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO sessions (token, user_id, role, expires_at)
		VALUES (?, 'u-tony', 'tournament_runner', strftime('%Y-%m-%dT%H:%M:%fZ','now','+12 hours'))`, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "email": "tony@okami.j3", "role": "runner",
	})
}
