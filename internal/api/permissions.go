package api

import (
	"net/http"
	"strconv"
)

type permissionView struct {
	Role      string `json:"role"`
	Resource  string `json:"resource"`
	CanView   bool   `json:"canView"`
	CanCreate bool   `json:"canCreate"`
	CanEdit   bool   `json:"canEdit"`
	CanDelete bool   `json:"canDelete"`
}

func (s *Server) permissions(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "settings", "view") {
		return
	}
	var list []permissionView
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT role, resource, can_view, can_create, can_edit, can_delete
		FROM role_permissions ORDER BY role, resource`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	list = []permissionView{}
	for rows.Next() {
		var item permissionView
		var view, create, edit, del int
		if err := rows.Scan(&item.Role, &item.Resource, &view, &create, &edit, &del); err != nil {
			writeError(w, err)
			return
		}
		item.CanView, item.CanCreate, item.CanEdit, item.CanDelete = view == 1, create == 1, edit == 1, del == 1
		list = append(list, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"permissions": list})
}

func (s *Server) updatePermission(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "settings", "edit") {
		return
	}
	var body struct {
		Role, Resource, Flag string
		Value                bool
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	columns := map[string]string{
		"view": "can_view", "create": "can_create", "edit": "can_edit", "delete": "can_delete",
	}
	column, ok := columns[body.Flag]
	if !ok || body.Role == "" || body.Resource == "" {
		writeBadRequest(w, "role, resource and a valid flag are required")
		return
	}
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO role_permissions (role, resource, `+column+`)
		VALUES (?, ?, ?)
		ON CONFLICT(role, resource) DO UPDATE SET `+column+` = excluded.`+column,
		body.Role, body.Resource, strconv.FormatBool(body.Value) == "true")
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "role.permission", "role", body.Role,
		`{"resource":"`+body.Resource+`","flag":"`+body.Flag+`","value":`+strconv.FormatBool(body.Value)+`}`)
	writeJSON(w, http.StatusOK, map[string]any{"role": body.Role, "resource": body.Resource, "flag": body.Flag, "value": body.Value})
}
