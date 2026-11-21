package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// checkIn records arrival and an official weigh-in, flagging out-of-class athletes.
func (s *Server) checkIn(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "edit") {
		return
	}
	entrantID := r.PathValue("id")
	var body struct {
		WeightKg float64 `json:"weightKg"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.WeightKg < 15 || body.WeightKg > 350 {
		writeBadRequest(w, "weight must be 15–350 kg")
		return
	}
	var categoryID, min, max string
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT COALESCE(e.category_id,''), COALESCE(CAST(c.min_value AS TEXT),''),
		       COALESCE(CAST(c.max_value AS TEXT),'')
		FROM entrants e LEFT JOIN categories c ON c.id = e.category_id WHERE e.id = ?`, entrantID).
		Scan(&categoryID, &min, &max)
	if err != nil {
		writeError(w, err)
		return
	}
	outOfClass := false
	if min != "" && body.WeightKg < parseFloat(min) {
		outOfClass = true
	}
	if max != "" && body.WeightKg > parseFloat(max) {
		outOfClass = true
	}
	if _, err := s.Store.ExecContext(r.Context(), `
		UPDATE entrants SET checked_in_at = strftime('%Y-%m-%dT%H:%M:%fZ','now'),
		  weighed_in_kg = ?, status = CASE WHEN ? THEN 'out_of_class' ELSE status END
		WHERE id = ?`, body.WeightKg, outOfClass, entrantID); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "entrant.check_in", "entrant", entrantID,
		fmt.Sprintf(`{"weightKg":%v,"outOfClass":%t}`, body.WeightKg, outOfClass))
	writeJSON(w, http.StatusOK, map[string]any{
		"entrantId": entrantID, "weightKg": body.WeightKg, "outOfClass": outOfClass,
	})
}

func parseFloat(value string) float64 {
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

type teamRanking struct {
	Club   string `json:"club"`
	Gold   int    `json:"gold"`
	Silver int    `json:"silver"`
	Bronze int    `json:"bronze"`
	Points int    `json:"points"`
}

// teamRankings converts medals into club/team points (3/2/1).
func (s *Server) teamRankings(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT COALESCE(c.name,'Independent'),
		  SUM(CASE WHEN m.kind='gold' THEN 1 ELSE 0 END),
		  SUM(CASE WHEN m.kind='silver' THEN 1 ELSE 0 END),
		  SUM(CASE WHEN m.kind='bronze' THEN 1 ELSE 0 END)
		FROM medals m JOIN fighters f ON f.id = m.fighter_id
		LEFT JOIN clubs c ON c.id = f.club_id
		WHERE m.event_id='ev-demo' GROUP BY c.name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	rankings := []teamRanking{}
	for rows.Next() {
		var item teamRanking
		if err := rows.Scan(&item.Club, &item.Gold, &item.Silver, &item.Bronze); err != nil {
			writeError(w, err)
			return
		}
		item.Points = item.Gold*3 + item.Silver*2 + item.Bronze
		rankings = append(rankings, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"teams": rankings})
}

// exportCSV provides a simple data export for spreadsheets and printed reports.
func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "view") {
		return
	}
	var query, header string
	switch r.URL.Query().Get("type") {
	case "entrants":
		query = `SELECT f.display_name, COALESCE(c.name,''), f.weight_kg, f.gender,
			COALESCE(b.label,''), COALESCE(e.status,'') FROM entrants e
			JOIN fighters f ON f.id=e.fighter_id
			LEFT JOIN clubs c ON c.id=f.club_id
			LEFT JOIN belt_ranks b ON b.id=f.belt_rank_id WHERE e.event_id='ev-demo'`
		header = "name,club,weightKg,gender,belt,status"
	case "fights":
		query = `SELECT id, round_label, state, outcome,
			COALESCE((SELECT display_name FROM fighters WHERE id=red_fighter_id),''),
			COALESCE((SELECT display_name FROM fighters WHERE id=blue_fighter_id),'')
			FROM fights WHERE event_id='ev-demo'`
		header = "id,round,state,outcome,red,blue"
	default:
		query = `SELECT display_name, COALESCE((SELECT name FROM clubs WHERE id=club_id),''),
			0, 0, 0 FROM fighters WHERE tenant_id='demo-tenant'`
		header = "name,club,wins,draws,losses"
	}
	rows, err := s.Store.QueryContext(r.Context(), query)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=j3-export.csv")
	writer := csv.NewWriter(w)
	_ = writer.Write(strings.Split(header, ","))
	columns, err := rows.Columns()
	if err != nil {
		writeError(w, err)
		return
	}
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err == nil {
			record := make([]string, len(values))
			for i, value := range values {
				record[i] = fmt.Sprint(value)
			}
			_ = writer.Write(record)
		}
	}
	writer.Flush()
}

// importEntries bulk-adds clubs and fighters from a CSV-like payload.
func (s *Server) importEntries(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "create") {
		return
	}
	var body struct {
		Rows []struct {
			Name, Club, Belt string
			WeightKg         float64
			DOB, Gender      string
		} `json:"rows"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if len(body.Rows) == 0 {
		writeBadRequest(w, "no rows supplied")
		return
	}
	clubIDs := map[string]string{}
	rows, err := s.Store.QueryContext(r.Context(), `SELECT id, name FROM clubs WHERE tenant_id='demo-tenant'`)
	if err != nil {
		writeError(w, err)
		return
	}
	for rows.Next() {
		var id, name string
		if rows.Scan(&id, &name) == nil {
			clubIDs[name] = id
		}
	}
	rows.Close()
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	for _, row := range body.Rows {
		clubID := ""
		if row.Club != "" {
			id, ok := clubIDs[row.Club]
			if !ok {
				id = newID("club")
				if _, err := tx.ExecContext(r.Context(), `
					INSERT INTO clubs (id, tenant_id, name) VALUES (?, 'demo-tenant', ?)`, id, row.Club); err != nil {
					writeError(w, err)
					return
				}
				clubIDs[row.Club] = id
			}
			clubID = id
		}
		gender := row.Gender
		if gender == "" {
			gender = "open"
		}
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO fighters (id, tenant_id, club_id, belt_rank_id, display_name, date_of_birth, weight_kg, gender)
			VALUES (?, 'demo-tenant', NULLIF(?,''), NULLIF(?,''), ?, ?, ?, ?)`,
			newID("f"), clubID, row.Belt, row.Name, row.DOB, row.WeightKg, gender); err != nil {
			writeError(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "import.entries", "event", "ev-demo", fmt.Sprintf(`{"rows":%d}`, len(body.Rows)))
	writeJSON(w, http.StatusCreated, map[string]any{"imported": len(body.Rows)})
}
