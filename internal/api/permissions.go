package api

import (
	"net/http"
	"strconv"
)

var defaultPermissions = []struct {
	Role, Resource          string
	View, Create, Edit, Del int
}{
	{"runner", "participants", 1, 1, 1, 1},
	{"runner", "fights", 1, 1, 1, 1},
	{"runner", "settings", 1, 1, 1, 1},
	{"official", "participants", 1, 1, 1, 0},
	{"official", "fights", 1, 1, 1, 0},
	{"club_leader", "participants", 1, 1, 1, 0},
	{"club_leader", "fights", 1, 0, 0, 0},
	{"fighter", "participants", 1, 1, 1, 0},
	{"fighter", "fights", 1, 0, 0, 0},
	{"spectator", "participants", 0, 0, 0, 0},
	{"spectator", "fights", 1, 0, 0, 0},
}

type permissionView struct {
	Role      string `json:"role"`
	Resource  string `json:"resource"`
	CanView   bool   `json:"canView"`
	CanCreate bool   `json:"canCreate"`
	CanEdit   bool   `json:"canEdit"`
	CanDelete bool   `json:"canDelete"`
}

func (s *Server) permissions(w http.ResponseWriter, r *http.Request) {
	var list []permissionView
	if r.URL.Query().Get("init") == "1" {
		for _, p := range defaultPermissions {
			_, _ = s.Store.ExecContext(r.Context(), `
				INSERT OR IGNORE INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete)
				VALUES (?, ?, ?, ?, ?, ?)`, p.Role, p.Resource, p.View, p.Create, p.Edit, p.Del)
		}
	}
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
	if len(list) == 0 {
		for _, p := range defaultPermissions {
			list = append(list, permissionView{p.Role, p.Resource, p.View == 1, p.Create == 1, p.Edit == 1, p.Del == 1})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"permissions": list})
}

func (s *Server) updatePermission(w http.ResponseWriter, r *http.Request) {
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
