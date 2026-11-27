package api

import (
	"database/sql"
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
	PhotoPath string        `json:"photoPath"`
	AvatarPath string       `json:"avatarPath"`
	ClubLogo  string        `json:"clubLogo"`
	ClubColour string       `json:"clubColour"`
}

func roleLabel(role auth.Role) string {
	for _, info := range auth.Roles() {
		if info.ID == role {
			return info.Label
		}
	}
	return string(role)
}

func effectiveStoredRole(persisted string, admin int) auth.Role {
	return auth.VirtualRole(persisted, admin == 1)
}

type demoIdentity struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Role      string        `json:"role"`
	RoleLabel string        `json:"roleLabel"`
	ClubName  string        `json:"clubName"`
	Actions   []auth.Action `json:"actions"`
}

// demoIdentities is the non-sensitive QA switcher directory. It is public only
// in demo mode and deliberately omits email, source, status and audit data, so
// switching identity never requires (or leaks) private participant records.
func (s *Server) demoIdentities(w http.ResponseWriter, r *http.Request) {
	if !demoMode() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT u.id, u.display_name, u.role, u.is_admin, COALESCE(c.name, '')
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		WHERE u.status = 'active'
		ORDER BY u.is_admin DESC, CASE u.role
		  WHEN 'tournament_runner' THEN 0 WHEN 'official' THEN 1
		  WHEN 'club_leader' THEN 2 WHEN 'fighter' THEN 3 ELSE 4 END,
		  u.display_name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	identities := []demoIdentity{}
	for rows.Next() {
		var item demoIdentity
		var persisted string
		var admin int
		if err := rows.Scan(&item.ID, &item.Name, &persisted, &admin, &item.ClubName); err != nil {
			writeError(w, err)
			return
		}
		role := effectiveStoredRole(persisted, admin)
		item.Role = string(role)
		item.RoleLabel = roleLabel(role)
		item.Actions = auth.Actions(role)
		identities = append(identities, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}

	var current *demoIdentity
	if identity, ok := auth.FromContext(r.Context()); ok {
		var item demoIdentity
		var persisted string
		var admin int
		err := s.Store.QueryRowContext(r.Context(), `
			SELECT u.id, u.display_name, u.role, u.is_admin, COALESCE(c.name, '')
			FROM users u LEFT JOIN clubs c ON c.id = u.club_id WHERE u.id = ?`, identity.UserID).
			Scan(&item.ID, &item.Name, &persisted, &admin, &item.ClubName)
		if err == nil {
			role := effectiveStoredRole(persisted, admin)
			item.Role = string(role)
			item.RoleLabel = roleLabel(role)
			item.Actions = auth.Actions(role)
			current = &item
		} else if err != sql.ErrNoRows {
			writeError(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"identities": identities, "current": current})
}

// demoAccounts lists every seeded account so QA can switch roles instantly.
func (s *Server) demoAccounts(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "settings", "view") {
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role, u.is_admin,
		       COALESCE(c.name, ''), u.source, u.status,
		       COALESCE(u.photo_path,''), COALESCE(c.logo_path,''), COALESCE(c.primary_colour,'#64748b')
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
	accounts := []demoAccount{}
	for rows.Next() {
		var a demoAccount
		var persisted string
		var admin int
		var photo, clubLogo, clubColour string
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &persisted, &admin,
			&a.ClubName, &a.Source, &a.Status, &photo, &clubLogo, &clubColour); err != nil {
			writeError(w, err)
			return
		}
		a.Role = effectiveStoredRole(persisted, admin)
		a.RoleLabel = roleLabel(a.Role)
		a.Actions = auth.Actions(a.Role)
		a.PhotoPath = photo; a.AvatarPath = photo; a.ClubLogo = clubLogo; a.ClubColour = clubColour
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": accounts})
}

// demoView returns the verified session account and what that user can do.
// The "actor" query parameter is intentionally ignored: identity comes only
// from the bearer session, exactly as it does for every protected endpoint.
func (s *Server) demoView(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "view") {
		return
	}
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	actorID := identity.UserID
	var a demoAccount
	var persisted string
	var admin int
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role, u.is_admin,
		       COALESCE(c.name, ''), u.source, u.status,
		       COALESCE(u.photo_path,''), COALESCE(c.logo_path,''), COALESCE(c.primary_colour,'#64748b')
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		WHERE u.id = ?`, actorID).
		Scan(&a.ID, &a.Name, &a.Email, &persisted, &admin, &a.ClubName, &a.Source, &a.Status,
			&a.PhotoPath, &a.ClubLogo, &a.ClubColour)
	if err != nil {
		writeError(w, err)
		return
	}
	a.Role = effectiveStoredRole(persisted, admin)
	a.RoleLabel = roleLabel(a.Role)
	a.Actions = auth.Actions(a.Role)
	a.AvatarPath = a.PhotoPath
	counts, err := s.counts(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"actor": a, "actions": a.Actions, "counts": counts,
	})
}
