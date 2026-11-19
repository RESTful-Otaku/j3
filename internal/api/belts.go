package api

import "net/http"

type beltOption struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Colour    string `json:"colour"`
	SortOrder int    `json:"sortOrder"`
	GradeKind string `json:"gradeKind"`
	IsCustom  bool   `json:"isCustom"`
}

// belts returns the full JJJ grade spectrum so sign-up can render a colour
// block beside each name, including tenant-defined custom grades.
func (s *Server) belts(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT id, label, colour, sort_order, grade_kind, is_custom
		FROM belt_ranks WHERE tenant_id = 'demo-tenant' ORDER BY sort_order`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	belts := []beltOption{}
	for rows.Next() {
		var belt beltOption
		var custom int
		if err := rows.Scan(&belt.ID, &belt.Label, &belt.Colour, &belt.SortOrder, &belt.GradeKind, &custom); err != nil {
			writeError(w, err)
			return
		}
		belt.IsCustom = custom == 1
		belts = append(belts, belt)
	}
	writeJSON(w, http.StatusOK, map[string]any{"belts": belts})
}
