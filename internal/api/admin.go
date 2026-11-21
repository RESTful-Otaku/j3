package api

import (
	"fmt"
	"net/http"
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
	if !s.requireAction(w, r, "participants", "view") {
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT u.id, u.display_name, u.email, u.role, u.status,
		       COALESCE(c.name,''), COALESCE(u.role_request,'')
		FROM users u LEFT JOIN clubs c ON c.id = u.club_id
		WHERE u.tenant_id='demo-tenant' ORDER BY u.role, u.display_name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	users := []adminUser{}
	for rows.Next() {
		var user adminUser
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role,
			&user.Status, &user.Club, &user.RoleRequest); err != nil {
			writeError(w, err)
			return
		}
		users = append(users, user)
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (s *Server) setUserStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "edit") {
		return
	}
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	allowed := map[string]bool{"active": true, "locked": true, "suspended": true}
	if !allowed[body.Status] {
		writeBadRequest(w, "invalid status")
		return
	}
	if _, err := s.Store.ExecContext(r.Context(),
		`UPDATE users SET status=?, updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=?`, body.Status, id); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "admin.user.status", "user", id, fmt.Sprintf(`{"status":%q}`, body.Status))
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": body.Status})
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "edit") {
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
	s.audit(r, "admin.password_reset", "user", id, `{"sent":true}`)
	writeJSON(w, http.StatusOK, map[string]any{"resetToken": token, "emailSent": true})
}
