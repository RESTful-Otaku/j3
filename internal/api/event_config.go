package api

import (
	"encoding/json"
	"net/http"
)

// eventConfig exposes the full organiser configuration for public review and
// for the different role-specific views.
func (s *Server) eventConfig(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		eventID = "ev-okami"
	}
	var info struct {
		Name         string `json:"name"`
		Organiser    string `json:"organiser"`
		Venue        string `json:"venue"`
		Costs        string `json:"costs"`
		Refreshments string `json:"refreshments"`
		Parking      string `json:"parking"`
		Hotel        string `json:"hotel"`
		Evening      string `json:"evening"`
		Course       string `json:"course"`
		Attendance   string `json:"attendanceRule"`
	}
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT e.name, COALESCE(i.organiser,''), COALESCE(i.venue,''),
		       COALESCE(i.costs_json,'[]'), COALESCE(i.refreshments,''),
		       COALESCE(i.parking,''), COALESCE(i.hotel_json,'[]'),
		       COALESCE(i.evening_json,'{}'), COALESCE(i.course_json,'{}'),
		       COALESCE(i.attendance_rule,'')
		FROM events e LEFT JOIN event_info i ON i.event_id = e.id
		WHERE e.id = ?`, eventID).
		Scan(&info.Name, &info.Organiser, &info.Venue, &info.Costs, &info.Refreshments,
			&info.Parking, &info.Hotel, &info.Evening, &info.Course, &info.Attendance)
	if err != nil {
		writeError(w, err)
		return
	}

	kit := []struct{ Item, Note string }{}
	kitRows, err := s.Store.QueryContext(r.Context(),
		`SELECT item, note FROM kit_requirements WHERE event_id = ? ORDER BY id`, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	for kitRows.Next() {
		var item, note string
		if kitRows.Scan(&item, &note) == nil {
			kit = append(kit, struct{ Item, Note string }{item, note})
		}
	}
	kitRows.Close()

	rules := []struct {
		Name, Kind, Ages, Rules string
	}{}
	ruleRows, err := s.Store.QueryContext(r.Context(),
		`SELECT name, kind, COALESCE(notes,''), rules_json FROM discipline_rules WHERE event_id = ? ORDER BY id`, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	for ruleRows.Next() {
		var item struct{ Name, Kind, Ages, Rules string }
		if ruleRows.Scan(&item.Name, &item.Kind, &item.Ages, &item.Rules) == nil {
			rules = append(rules, item)
		}
	}
	ruleRows.Close()

	staff := []struct{ Mat, Role, Name string }{}
	staffRows, err := s.Store.QueryContext(r.Context(), `
		SELECT ms.mat_name, ms.role, COALESCE(u.display_name,'Unassigned')
		FROM mat_staff ms LEFT JOIN users u ON u.id = ms.user_id
		WHERE ms.event_id = ? ORDER BY ms.mat_name, ms.sort_order`, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	for staffRows.Next() {
		var item struct{ Mat, Role, Name string }
		if staffRows.Scan(&item.Mat, &item.Role, &item.Name) == nil {
			staff = append(staff, item)
		}
	}
	staffRows.Close()

	decode := func(raw string) any {
		var value any
		if json.Unmarshal([]byte(raw), &value) == nil {
			return value
		}
		return nil
	}
	eventView := map[string]any{
		"name": info.Name, "organiser": info.Organiser, "venue": info.Venue,
		"costs": decode(info.Costs), "refreshments": info.Refreshments,
		"parking": info.Parking, "hotel": decode(info.Hotel),
		"evening": decode(info.Evening), "course": decode(info.Course),
		"attendanceRule": info.Attendance,
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"event": eventView, "kit": kit, "rules": rules, "staff": staff,
	})
}

// attendanceList supports the "must attend the morning course" rule.
func (s *Server) attendance(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT ca.entrant_id, f.display_name, ca.attended, ca.excluded
		FROM course_attendance ca JOIN entrants e ON e.id = ca.entrant_id
		JOIN fighters f ON f.id = e.fighter_id
		WHERE e.event_id = ? ORDER BY f.display_name`, "ev-okami")
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var entrantID, name string
		var attended, excluded int
		if err := rows.Scan(&entrantID, &name, &attended, &excluded); err != nil {
			writeError(w, err)
			return
		}
		items = append(items, map[string]any{
			"entrantId": entrantID, "name": name,
			"attended": attended == 1, "excluded": excluded == 1,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"attendance": items})
}
