package api

import "net/http"

type kataCatalogueItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Style      string `json:"style"`
	Level      string `json:"level"`
	Discipline string `json:"discipline"`
	IsWeapon   bool   `json:"isWeapon"`
}

func (s *Server) kataCatalogue(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT id, name, style, level, discipline, is_weapon
		FROM kata_catalogue ORDER BY discipline, name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	items := []kataCatalogueItem{}
	for rows.Next() {
		var item kataCatalogueItem
		var weapon int
		if err := rows.Scan(&item.ID, &item.Name, &item.Style, &item.Level, &item.Discipline, &weapon); err != nil {
			writeError(w, err)
			return
		}
		item.IsWeapon = weapon == 1
		if filter := r.URL.Query().Get("discipline"); filter == "" || filter == item.Discipline {
			items = append(items, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"kata": items})
}

type trainingSessionView struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Kind     string `json:"kind"`
	Leader   string `json:"leader"`
	StartsAt string `json:"startsAt"`
	Duration int    `json:"duration"`
	Status   string `json:"status"`
}

func (s *Server) listTraining(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT id, title, kind, leader, starts_at, duration_seconds, status
		FROM training_sessions WHERE event_id = 'ev-demo' ORDER BY starts_at`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	items := []trainingSessionView{}
	for rows.Next() {
		var item trainingSessionView
		if err := rows.Scan(&item.ID, &item.Title, &item.Kind, &item.Leader,
			&item.StartsAt, &item.Duration, &item.Status); err != nil {
			writeError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": items})
}

func (s *Server) createTraining(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title, Kind, Leader, StartsAt string
		Duration                      int
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.Title == "" {
		writeBadRequest(w, "title is required")
		return
	}
	if body.Kind == "" {
		body.Kind = "drill"
	}
	if body.Duration < 60 {
		body.Duration = 1800
	}
	id := newID("session")
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO training_sessions
		 (id, event_id, title, kind, leader, starts_at, duration_seconds)
		 VALUES (?, 'ev-demo', ?, ?, ?, ?, ?)`,
		id, body.Title, body.Kind, body.Leader, body.StartsAt, body.Duration)
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "training.create", "training_session", id, body.Title)
	s.listTraining(w, r)
}
