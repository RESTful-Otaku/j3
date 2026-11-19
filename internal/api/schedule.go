package api

import (
	"fmt"
	"net/http"
	"time"
)

type scheduleItemView struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Area           string `json:"area"`
	PlannedStart   string `json:"plannedStart"`
	ProjectedStart string `json:"projectedStart"`
	ProjectedEnd   string `json:"projectedEnd"`
	Duration       int    `json:"duration"`
	Status         string `json:"status"`
	OffsetMinutes  int    `json:"offsetMinutes"`
	RestWarning    bool   `json:"restWarning"`
}

func (s *Server) schedule(r *http.Request, eventID string) ([]scheduleItemView, error) {
	offset := 0
	_ = s.Store.QueryRowContext(r.Context(),
		`SELECT offset_minutes FROM schedule_state WHERE event_id = ?`, eventID).Scan(&offset)
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT id, name, area, starts_at, duration_seconds, status
		FROM schedule_items WHERE event_id = ? ORDER BY sort_order`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []scheduleItemView{}
	var previousEnd time.Time
	for rows.Next() {
		var item scheduleItemView
		var planned string
		if err := rows.Scan(&item.ID, &item.Name, &item.Area, &planned, &item.Duration, &item.Status); err != nil {
			return nil, err
		}
		start, err := time.Parse(time.RFC3339, planned)
		if err != nil {
			return nil, err
		}
		projected := start.Add(time.Duration(offset) * time.Minute)
		item.PlannedStart = planned
		item.ProjectedStart = projected.Format(time.RFC3339)
		item.ProjectedEnd = projected.Add(time.Duration(item.Duration) * time.Second).Format(time.RFC3339)
		item.OffsetMinutes = offset
		if !previousEnd.IsZero() && projected.Before(previousEnd.Add(5*time.Minute)) {
			item.RestWarning = true
		}
		previousEnd = projected.Add(time.Duration(item.Duration) * time.Second)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Server) listSchedule(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		eventID = "ev-demo"
	}
	items, err := s.schedule(r, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) setDrift(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Minutes int `json:"minutes"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.Minutes < -480 || body.Minutes > 480 {
		writeBadRequest(w, "drift must be between -480 and 480 minutes")
		return
	}
	if _, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO schedule_state (event_id, offset_minutes) VALUES ('ev-demo', ?)
		ON CONFLICT(event_id) DO UPDATE SET offset_minutes = excluded.offset_minutes`, body.Minutes); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "schedule.drift", "event", "ev-demo", fmt.Sprintf(`{"minutes":%d}`, body.Minutes))
	items, err := s.schedule(r, "ev-demo")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) completeScheduleItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.Store.ExecContext(r.Context(), `
		UPDATE schedule_items SET status='complete',
		  actual_end = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`, id); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "schedule.complete", "schedule_item", id, `{"status":"complete"}`)
	items, err := s.schedule(r, "ev-demo")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
