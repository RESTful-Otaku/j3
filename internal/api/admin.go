package api

import (
	"fmt"
	"net/http"

	"github.com/restful-otaku/j3/internal/auth"
)

type adminUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Club        string `json:"club"`
	RoleRequest string `json:"roleRequest"`
}

func (s *Server) adminUsers(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "settings", "view") {
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role, u.is_admin, u.status,
		       COALESCE(c.name,''), COALESCE(u.role_request,'')
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		WHERE u.tenant_id='demo-tenant'
		ORDER BY u.is_admin DESC, u.role, u.display_name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	users := []adminUser{}
	for rows.Next() {
		var user adminUser
		var persisted string
		var admin int
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &persisted, &admin,
			&user.Status, &user.Club, &user.RoleRequest); err != nil {
			writeError(w, err)
			return
		}
		user.Role = string(effectiveStoredRole(persisted, admin))
		users = append(users, user)
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (s *Server) managedTarget(r *http.Request) (auth.Role, string, string, bool) {
	_, actorClub, ok := s.actorScope(r)
	if !ok {
		return "", "", "", false
	}
	actorRole, ok := s.effectiveRole(r)
	if !ok {
		return "", "", "", false
	}
	var persisted, targetClub, status string
	var admin int
	if err := s.Store.QueryRowContext(r.Context(), `
		SELECT u.role, u.is_admin, COALESCE(u.club_id,''), u.status
		FROM users u WHERE u.id = ? AND u.tenant_id = 'demo-tenant'`, r.PathValue("id")).
		Scan(&persisted, &admin, &targetClub, &status); err != nil {
		return "", "", "", false
	}
	target := effectiveStoredRole(persisted, admin)
	if !canManageUser(actorRole, actorClub, target, targetClub) {
		return "", "", "", false
	}
	return target, status, targetClub, true
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "delete") { return }
	_, _, _, ok := s.managedTarget(r)
	if !ok { writeJSON(w, http.StatusForbidden, map[string]string{"error":"cannot manage this account"}); return }
	id := r.PathValue("id")
	if _, err := s.Store.ExecContext(r.Context(), `DELETE FROM users WHERE id = ?`, id); err != nil { writeError(w, err); return }
	s.audit(r, "account.delete", "user", id, `{"deleted":true}`)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *Server) setUserStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "edit") {
		return
	}
	target, status, _, ok := s.managedTarget(r)
	if !ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot manage this account"})
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	allowed := map[string]bool{"active": true, "suspended": true}
	if !allowed[body.Status] {
		writeBadRequest(w, "invalid status")
		return
	}
	actorRole, _ := s.effectiveRole(r)
	if status == "invited" && body.Status == "active" &&
		actorRole != auth.RoleAdmin && actorRole != auth.RoleTournamentRunner {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "invited accounts must accept their invitation to become active",
		})
		return
	}
	id := r.PathValue("id")
	if _, err := s.Store.ExecContext(r.Context(),
		`UPDATE users SET status=?, updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=?`, body.Status, id); err != nil {
		writeError(w, err)
		return
	}
	if body.Status == "suspended" {
		if _, err := s.Store.ExecContext(r.Context(), `DELETE FROM sessions WHERE user_id = ?`, id); err != nil {
			writeError(w, err)
			return
		}
	}
	s.audit(r, "admin.user.status", "user", id, fmt.Sprintf(`{"status":%q,"role":%q}`, body.Status, target))
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": body.Status})
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "edit") {
		return
	}
	target, _, _, ok := s.managedTarget(r)
	if !ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot manage this account"})
		return
	}
	id := r.PathValue("id")
	token := newID("reset")
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO password_resets (id, user_id, token, expires_at)
		VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now','+1 day'))`, token, id, token)
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "admin.password_reset", "user", id, fmt.Sprintf(`{"sent":true,"role":%q}`, target))
	writeJSON(w, http.StatusOK, map[string]any{"resetToken": token, "emailSent": true})
}
