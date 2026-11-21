package api

import (
	"net/http"
	"strconv"
	"strings"
)

// actorFromRequest resolves the demo actor from request headers. Defaults keep
// the local QA experience usable without a real identity provider.
func actorFromRequest(r *http.Request) (string, string) {
	userID := r.Header.Get("X-J3-User")
	if userID == "" {
		userID = "u-official"
	}
	role := r.Header.Get("X-J3-Role")
	if role == "" {
		role = "official"
	}
	return userID, role
}

type auditEntry struct {
	ID          int    `json:"id"`
	ActorUserID string `json:"actorUserId"`
	ActorRole   string `json:"actorRole"`
	Action      string `json:"action"`
	TargetType  string `json:"targetType"`
	TargetID    string `json:"targetId"`
	BeforeJSON  string `json:"beforeJson"`
	AfterJSON   string `json:"afterJson"`
	CreatedAt   string `json:"createdAt"`
}

// auditLog exposes the accountability trail to runners and officials.
func (s *Server) auditLog(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "view") {
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 1000 {
		limit = 250
	}
	query := `SELECT id, COALESCE(actor_user_id,''), actor_role, action, target_type,
		target_id, before_json, after_json, created_at
		FROM audit_log WHERE tenant_id = 'demo-tenant'`
	args := []any{}
	if actor := r.URL.Query().Get("actor"); actor != "" {
		query += ` AND actor_user_id = ?`
		args = append(args, actor)
	}
	if action := r.URL.Query().Get("action"); action != "" {
		query += ` AND action LIKE ? ESCAPE '\'`
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(action)
		args = append(args, "%"+escaped+"%")
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.Store.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	entries := []auditEntry{}
	for rows.Next() {
		var entry auditEntry
		if err := rows.Scan(&entry.ID, &entry.ActorUserID, &entry.ActorRole, &entry.Action,
			&entry.TargetType, &entry.TargetID, &entry.BeforeJSON, &entry.AfterJSON,
			&entry.CreatedAt); err != nil {
			writeError(w, err)
			return
		}
		entries = append(entries, entry)
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}
