package api

import (
	"encoding/json"
	"net/http"
)

// eventConfig exposes the full organiser configuration for public review and
// for the different role-specific views.
func (s *Server) eventConfig(w http.ResponseWriter, r *http.Request) {
	eventID := s.eventIDFromRequest(r)
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
		// Public event configuration must render even with no event_info row or
		// missing optional data — never 500; serve a minimal shell.
		info.Name = "Tournament"
		info.Costs = "[]"; info.Hotel = "[]"; info.Evening = "{}"; info.Course = "{}"
	}

	kit := []struct {
		Item string `json:"item"`
		Note string `json:"note"`
	}{}
	kitRows, err := s.Store.QueryContext(r.Context(),
		`SELECT item, note FROM kit_requirements WHERE event_id = ? ORDER BY id`, eventID)
	if err != nil { kitRows = nil; err = nil }
	if kitRows != nil { for kitRows.Next() {
		var item, note string
		if kitRows.Scan(&item, &note) == nil {
			kit = append(kit, struct {
			Item string `json:"item"`
			Note string `json:"note"`
		}{item, note})
		}
	}
	}
	if kitRows != nil { _ = kitRows.Err(); kitRows.Close() }

	rules := []struct {
		Name  string `json:"name"`
		Kind  string `json:"kind"`
		Ages  string `json:"ages"`
		Rules string `json:"rules"`
	}{}
	ruleRows, err := s.Store.QueryContext(r.Context(),
		`SELECT name, kind, COALESCE(notes,''), rules_json FROM discipline_rules WHERE event_id = ? ORDER BY id`, eventID)
	if err != nil { ruleRows = nil; err = nil }
	if ruleRows != nil { for ruleRows.Next() {
		var item struct {
			Name  string `json:"name"`
			Kind  string `json:"kind"`
			Ages  string `json:"ages"`
			Rules string `json:"rules"`
		}
		if ruleRows.Scan(&item.Name, &item.Kind, &item.Ages, &item.Rules) == nil {
			rules = append(rules, item)
		}
	}
	}
	if ruleRows != nil { _ = ruleRows.Err(); ruleRows.Close() }

	staff := []struct {
		Mat  string `json:"mat"`
		Role string `json:"role"`
		Name string `json:"name"`
	}{}
	staffRows, err := s.Store.QueryContext(r.Context(), `
		SELECT ms.mat_name, ms.role, COALESCE(u.display_name,'Unassigned')
		FROM mat_staff ms LEFT JOIN users u ON u.id = ms.user_id
		WHERE ms.event_id = ? ORDER BY ms.mat_name, ms.sort_order`, eventID)
	if err != nil { staffRows = nil; err = nil }
	if staffRows != nil { for staffRows.Next() {
		var item struct {
			Mat  string `json:"mat"`
			Role string `json:"role"`
			Name string `json:"name"`
		}
		if staffRows.Scan(&item.Mat, &item.Role, &item.Name) == nil {
			staff = append(staff, item)
		}
	}
	}
	if staffRows != nil { _ = staffRows.Err(); staffRows.Close() }

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
	formats := []struct {
		Discipline, Format, Knockout, Third, Order string
		DualBronze, RetryTies                      bool
	}{}
	formatRows, err := s.Store.QueryContext(r.Context(), `
		SELECT COALESCE(NULLIF(dr.name,''), NULLIF(disc.name,''), df.discipline_id),
		       df.format, df.knockout, df.third_place, df.order_policy, df.dual_bronze, df.retry_ties
		FROM discipline_formats df
		LEFT JOIN discipline_rules dr ON dr.id = df.discipline_id AND dr.event_id = df.event_id
		LEFT JOIN disciplines disc ON disc.id = df.discipline_id AND disc.event_id = df.event_id
		WHERE df.event_id = ?`, eventID)
	if err != nil { formatRows = nil; err = nil }
	if formatRows != nil { for formatRows.Next() {
		var item struct {
			Discipline, Format, Knockout, Third, Order string
			DualBronze, RetryTies                      bool
		}
		var dual, retry int
		if formatRows.Scan(&item.Discipline, &item.Format, &item.Knockout, &item.Third,
			&item.Order, &dual, &retry) == nil {
			item.DualBronze, item.RetryTies = dual == 1, retry == 1
			formats = append(formats, item)
		}
	}
	}
	if formatRows != nil { _ = formatRows.Err(); formatRows.Close() }

	writeJSON(w, http.StatusOK, map[string]any{
		"event": eventView, "kit": kit, "rules": rules, "staff": staff, "formats": formats,
	})
}

// attendanceList supports the "must attend the morning course" rule.
func (s *Server) attendance(w http.ResponseWriter, r *http.Request) {
	// Attendance is best-effort for event-config displays; don't hard-fail.
	if !s.requireAction(w, r, "participants", "view") {
		writeJSON(w, http.StatusOK, map[string]any{"attendance": []any{}})
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT ca.entrant_id, f.display_name, ca.attended, ca.excluded
		FROM course_attendance ca JOIN entrants e ON e.id = ca.entrant_id
		JOIN fighters f ON f.id = e.fighter_id
		WHERE e.event_id = ? ORDER BY f.display_name`, s.eventIDFromRequest(r))
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"attendance": []any{}})
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
	if rows.Err() != nil { /* return what we have */ }
	writeJSON(w, http.StatusOK, map[string]any{"attendance": items})
}
