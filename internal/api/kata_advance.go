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

	eventID := s.defaultEventID(r)
	var currentRound int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT COALESCE(MAX(round_number),0) FROM kata_performances WHERE event_id = ?`, eventID).
		Scan(&currentRound); err != nil {
		writeError(w, err)
		return
	}
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT fighter_id, kata_name, panel_size, bunkai_score
		FROM kata_performances
		WHERE event_id = ? AND round_number = ? AND status = 'complete'
		ORDER BY total DESC, fighter_id ASC LIMIT ?`, eventID, currentRound, body.AdvanceCount)
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

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	for _, item := range promotions {
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO kata_performances
			 (id, event_id, fighter_id, kata_name, panel_size, round_number, kata_type, bunkai_score)
			 VALUES (?, ?, ?, ?, ?, ?, 'shitei', ?)`,
			newID("kata"), eventID, item.fighter, item.kata, item.panel, currentRound+1, item.bunkai); err != nil {
			writeError(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "kata.advance", "event", eventID,
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
