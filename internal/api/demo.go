package api

import (
	"net/http"

	"github.com/restful-otaku/j3/internal/auth"
)

// NOTE: the endpoints in this file are for the local QA demo only. Real
// authentication and per-request actor resolution will replace "act as" before
// the venue server is exposed publicly.

type demoAccount struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Email     string        `json:"email"`
	Role      auth.Role     `json:"role"`
	RoleLabel string        `json:"roleLabel"`
	ClubName  string        `json:"clubName"`
	Source    string        `json:"source"`
	Status    string        `json:"status"`
	Actions   []auth.Action `json:"actions"`
}

func roleLabel(role auth.Role) string {
	for _, info := range auth.Roles() {
		if info.ID == role {
			return info.Label
		}
	}
	return string(role)
}

// demoAccounts lists every seeded account so QA can switch roles instantly.
func (s *Server) demoAccounts(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role,
		       COALESCE(c.name, ''), u.source, u.status
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		ORDER BY CASE u.role
		  WHEN 'tournament_runner' THEN 0 WHEN 'official' THEN 1
		  WHEN 'club_leader' THEN 2 WHEN 'fighter' THEN 3 ELSE 4 END,
		  u.display_name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	accounts := []demoAccount{}
	for rows.Next() {
		var a demoAccount
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.Role, &a.ClubName, &a.Source, &a.Status); err != nil {
			writeError(w, err)
			return
		}
		a.RoleLabel = roleLabel(a.Role)
		a.Actions = auth.Actions(a.Role)
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": accounts})
}

// demoView resolves the "acting as" account and returns what that user can do.
func (s *Server) demoView(w http.ResponseWriter, r *http.Request) {
	actorID := r.URL.Query().Get("actor")
	if actorID == "" {
		actorID = r.Header.Get("X-J3-Actor")
	}
	if actorID == "" {
		actorID = "u-runner"
	}
	var a demoAccount
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role,
		       COALESCE(c.name, ''), u.source, u.status
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		WHERE u.id = ?`, actorID).
		Scan(&a.ID, &a.Name, &a.Email, &a.Role, &a.ClubName, &a.Source, &a.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	a.RoleLabel = roleLabel(a.Role)
	a.Actions = auth.Actions(a.Role)
	counts, err := s.counts(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"actor": a, "actions": a.Actions, "counts": counts,
	})
}
