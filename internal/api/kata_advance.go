package api

import (
	"net/http"
	"strconv"
)

// advanceKata promotes the top performers of the current round into the next
// round, retaining their previous performance for the audit trail.
func (s *Server) advanceKata(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "fights", "edit") {
		return
	}
	var body struct {
		AdvanceCount int `json:"advanceCount"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.AdvanceCount < 1 {
		body.AdvanceCount = 8
	}

	var currentRound int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT COALESCE(MAX(round_number),0) FROM kata_performances WHERE event_id = 'ev-demo'`).
		Scan(&currentRound); err != nil {
		writeError(w, err)
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT fighter_id, kata_name, panel_size, bunkai_score
		FROM kata_performances
		WHERE event_id = 'ev-demo' AND round_number = ? AND status = 'complete'
		ORDER BY total DESC LIMIT ?`, currentRound, body.AdvanceCount)
	if err != nil {
		writeError(w, err)
		return
	}
	type promotion struct {
		fighter, kata string
		panel         int
		bunkai        float64
	}
	promotions := []promotion{}
	for rows.Next() {
		var item promotion
		if err := rows.Scan(&item.fighter, &item.kata, &item.panel, &item.bunkai); err != nil {
			rows.Close()
			writeError(w, err)
			return
		}
		promotions = append(promotions, item)
	}
	rows.Close()

	for _, item := range promotions {
		if _, err := s.Store.ExecContext(r.Context(), `
			INSERT INTO kata_performances
			 (id, event_id, fighter_id, kata_name, panel_size, round_number, kata_type, bunkai_score)
			 VALUES (?, 'ev-demo', ?, ?, ?, ?, 'shitei', ?)`,
			newID("kata"), item.fighter, item.kata, item.panel, currentRound+1, item.bunkai); err != nil {
			writeError(w, err)
			return
		}
	}
	s.audit(r, "kata.advance", "event", "ev-demo",
		`{"from":`+strconv.Itoa(currentRound)+`,"to":`+strconv.Itoa(currentRound+1)+`,"count":`+strconv.Itoa(len(promotions))+`}`)
	items, err := s.listKata(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"advanced": len(promotions), "round": currentRound + 1, "performances": items,
	})
}
