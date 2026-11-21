package api

import (
	"encoding/json"

	"fmt"
	"github.com/restful-otaku/j3/internal/auth"
	"net/http"
	"strings"
)

// forfeitFight records an injury, technical forfeit or disqualification and
// locks the result so downstream displays and standings stay consistent.
func (s *Server) forfeitFight(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "fights", "edit") {
		return
	}
	fightID := r.PathValue("id")
	var body struct {
		Reason string `json:"reason"`
		Winner string `json:"winner"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	outcome := ""
	switch strings.ToLower(body.Winner) {
	case "red":
		outcome = "win_red"
	case "blue":
		outcome = "win_blue"
	default:
		writeBadRequest(w, "winner must be red or blue")
		return
	}
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(), `
		UPDATE fights SET state = 'forfeit', outcome = ?, version = version + 1,
		  completed_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`,
		outcome, fightID); err != nil {
		writeError(w, err)
		return
	}
	payload, _ := json.Marshal(map[string]string{"reason": body.Reason, "winner": body.Winner})
	identity, _ := auth.FromContext(r.Context())
	actor := identity.UserID
	if actor == "" {
		actor = "system"
	}
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO score_events
		 (fight_id, seq, kind, corner, points, technique, card, actor_user_id, payload_json)
		SELECT ?, COALESCE(MAX(seq),0)+1, 'forfeit', ?, 0, ?, '', ?, ?
		FROM score_events WHERE fight_id = ?`,
		fightID, body.Winner, body.Reason, actor, string(payload), fightID); err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "fight.forfeit", "fight", fightID, string(payload))
	writeJSON(w, http.StatusOK, map[string]any{"fightId": fightID, "state": "forfeit", "outcome": outcome})
}

var allowedEntrantStatuses = map[string]bool{
	"active": true, "withdrawn": true, "no_show": true, "injured": true,
}

// entrantStatus handles dropouts and no-shows; pool generation automatically
// excludes inactive entrants on the next generate call.
func (s *Server) entrantStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "edit") {
		return
	}
	entrantID := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if !allowedEntrantStatuses[body.Status] {
		writeBadRequest(w, fmt.Sprintf("unknown status %q", body.Status))
		return
	}
	if _, err := s.Store.ExecContext(r.Context(),
		`UPDATE entrants SET status = ? WHERE id = ?`, body.Status, entrantID); err != nil {
		writeError(w, err)
		return
	}
	payload, _ := json.Marshal(map[string]string{"status": body.Status})
	s.audit(r, "entrant.status", "entrant", entrantID, string(payload))
	writeJSON(w, http.StatusOK, map[string]any{"entrantId": entrantID, "status": body.Status})
}
