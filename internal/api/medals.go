package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type medalView struct {
	ID         string `json:"id"`
	FighterID  string `json:"fighterId"`
	Name       string `json:"name"`
	Club       string `json:"club"`
	ClubColour string `json:"clubColour"`
	Kind       string `json:"kind"`
}

func (s *Server) awardMedals(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "fights", "edit") {
		return
	}
	var body struct {
		Gold     string   `json:"gold"`
		Silver   string   `json:"silver"`
		Bronze   []string `json:"bronze"`
		Category string   `json:"category"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	awards := []struct{ id, kind string }{}
	if body.Gold != "" {
		awards = append(awards, struct{ id, kind string }{body.Gold, "gold"})
	}
	if body.Silver != "" {
		awards = append(awards, struct{ id, kind string }{body.Silver, "silver"})
	}
	for _, id := range body.Bronze {
		if id != "" {
			awards = append(awards, struct{ id, kind string }{id, "bronze"})
		}
	}
	if len(awards) == 0 {
		writeBadRequest(w, "at least one medal is required")
		return
	}
	eventID := s.defaultEventID(r)
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	for _, award := range awards {
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO medals (id, event_id, category_id, fighter_id, kind)
			VALUES (?, ?, NULLIF(?, ''), ?, ?)`,
			newID("medal"), eventID, body.Category, award.id, award.kind); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "medal already awarded"})
				return
			}
			writeError(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	encoded, _ := json.Marshal(body)
	s.audit(r, "medals.award", "event", eventID, string(encoded))
	s.listMedalsHandler(w, r)
}

func (s *Server) listMedalsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT m.id, m.fighter_id, COALESCE(f.display_name, m.fighter_id),
		       COALESCE(c.name,''), COALESCE(c.primary_colour,'#64748b'), m.kind
		FROM medals m
		LEFT JOIN fighters f ON f.id = m.fighter_id
		LEFT JOIN clubs c ON c.id = f.club_id
		WHERE m.event_id = ? ORDER BY m.awarded_at DESC`, s.eventIDFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	medals := []medalView{}
	for rows.Next() {
		var medal medalView
		if err := rows.Scan(&medal.ID, &medal.FighterID, &medal.Name, &medal.Club, &medal.ClubColour, &medal.Kind); err != nil {
			writeError(w, err)
			return
		}
		medals = append(medals, medal)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"medals": medals})
}
