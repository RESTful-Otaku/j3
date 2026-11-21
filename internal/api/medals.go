package api

import (
	"encoding/json"
	"net/http"
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
	for _, award := range awards {
		if _, err := s.Store.ExecContext(r.Context(), `
			INSERT INTO medals (id, event_id, category_id, fighter_id, kind)
			VALUES (?, 'ev-demo', NULLIF(?, ''), ?, ?)`,
			newID("medal"), body.Category, award.id, award.kind); err != nil {
			writeError(w, err)
			return
		}
	}
	encoded, _ := json.Marshal(body)
	s.audit(r, "medals.award", "event", "ev-demo", string(encoded))
	s.listMedalsHandler(w, r)
}

func (s *Server) listMedalsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT m.id, m.fighter_id, COALESCE(f.display_name, m.fighter_id),
		       COALESCE(c.name,''), COALESCE(c.primary_colour,'#64748b'), m.kind
		FROM medals m
		LEFT JOIN fighters f ON f.id = m.fighter_id
		LEFT JOIN clubs c ON c.id = f.club_id
		WHERE m.event_id = 'ev-demo' ORDER BY m.awarded_at DESC`)
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
	writeJSON(w, http.StatusOK, map[string]any{"medals": medals})
}
