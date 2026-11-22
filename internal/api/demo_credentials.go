package api

import (
	"net/http"

	"github.com/restful-otaku/j3/internal/auth"
)

// Demo builds use one shared password so QA can test every role without
// signing up or resetting credentials. This endpoint is never registered
// outside demo mode.
const demoPassword = "demo"

type demoCredential struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	RoleLabel string `json:"roleLabel"`
	Club      string `json:"club"`
	Password  string `json:"password"`
	Status    string `json:"status"`
}

// demoCredentials is the in-app account sheet shown on the sign-in and demo
// accounts screens. It deliberately mirrors the seeded database so a DB reset
// can never leave the published credentials stale.
func (s *Server) demoCredentials(w http.ResponseWriter, r *http.Request) {
	if !demoMode() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role, u.is_admin,
		       COALESCE(c.name, ''), u.status
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		ORDER BY u.is_admin DESC, CASE u.role
		  WHEN 'tournament_runner' THEN 0 WHEN 'official' THEN 1
		  WHEN 'club_leader' THEN 2 WHEN 'fighter' THEN 3 ELSE 4 END,
		  u.display_name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	accounts := []demoCredential{}
	for rows.Next() {
		var item demoCredential
		var persisted string
		var admin int
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &persisted, &admin,
			&item.Club, &item.Status); err != nil {
			writeError(w, err)
			return
		}
		role := effectiveStoredRole(persisted, admin)
		item.Role = string(role)
		item.RoleLabel = roleLabel(role)
		item.Password = demoPassword
		accounts = append(accounts, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}

	// One representative account per role keeps the sign-in sheet readable;
	// the full list stays available for club/participant QA.
	quick := []demoCredential{}
	preferred := map[auth.Role]string{
		auth.RoleAdmin: "u-owner", auth.RoleTournamentRunner: "u-tony",
		auth.RoleOfficial: "u-official", auth.RoleClubLeader: "u-leader-1",
		auth.RoleFighter: "u-fighter-01", auth.RoleSpectator: "u-spectator",
	}
	for _, role := range []auth.Role{auth.RoleAdmin, auth.RoleTournamentRunner,
		auth.RoleOfficial, auth.RoleClubLeader, auth.RoleFighter, auth.RoleSpectator} {
		var selected demoCredential
		for _, account := range accounts {
			if auth.Role(account.Role) == role {
				selected = account
				if account.ID == preferred[role] {
					break
				}
			}
		}
		if selected.ID != "" && selected.Status == "active" {
			quick = append(quick, selected)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"password": demoPassword, "accounts": accounts,
		"quickAccounts": quick, "total": len(accounts),
	})
}
