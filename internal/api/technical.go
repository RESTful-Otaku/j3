package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/restful-otaku/j3/internal/auth"
)

// technicalReasons are the on-the-day outcomes officials may call without
// editing the event configuration. DQ-style reasons lock a disqualification
// outcome so rankings and displays agree with the audit trail.
var technicalReasons = map[string]string{
	"injury":     "Injury",
	"forfeit":    "Forfeit",
	"cheating":   "Cheating / unsportsmanlike conduct",
	"dq":         "Disqualification",
	"no_show":    "No show",
	"withdrawal": "Withdrawal",
	"referee":    "Referee decision",
	"other":      "Other recorded reason",
}

var technicalDQReasons = map[string]bool{"cheating": true, "dq": true}

type technicalResultRequest struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

// callTechnicalResult records a technical win/loss (injury, forfeit, cheating,
// no-show) with the same assignment guard as live scoring, then locks the bout.
func (s *Server) callTechnicalResult(w http.ResponseWriter, r *http.Request) {
	fightID := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, fightID) {
		return
	}
	var body technicalResultRequest
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	body.Winner = strings.ToLower(strings.TrimSpace(body.Winner))
	body.Reason = strings.ToLower(strings.TrimSpace(body.Reason))
	if body.Winner != "red" && body.Winner != "blue" {
		writeBadRequest(w, "winner must be red or blue")
		return
	}
	reasonLabel, ok := technicalReasons[body.Reason]
	if !ok {
		writeBadRequest(w, "unknown technical result reason")
		return
	}
	outcome := "win_" + body.Winner
	if technicalDQReasons[body.Reason] {
		outcome = "dq_" + body.Winner
	}
	technique := "Technical result · " + reasonLabel
	if strings.TrimSpace(body.Detail) != "" {
		technique += " · " + strings.TrimSpace(body.Detail)
	}

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(), `
		UPDATE fights SET state = 'forfeit', outcome = ?, version = version + 1,
		  completed_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id = ?`, outcome, fightID); err != nil {
		writeError(w, err)
		return
	}
	identity, _ := auth.FromContext(r.Context())
	actor := identity.UserID
	if actor == "" {
		actor = "system"
	}
	payload, _ := json.Marshal(map[string]string{
		"winner": body.Winner, "reason": body.Reason, "detail": body.Detail, "outcome": outcome,
	})
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO score_events
		 (fight_id, seq, kind, corner, points, technique, card, actor_user_id, payload_json)
		SELECT ?, COALESCE(MAX(seq),0)+1, 'technical', ?, 0, ?, '', ?, ?
		FROM score_events WHERE fight_id = ?`,
		fightID, body.Winner, technique, actor, string(payload), fightID); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE fight_state SET running = 0, ends_at = '' WHERE fight_id = ?`, fightID); err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "fight.technical", "fight", fightID, fmt.Sprintf(`{"outcome":%q,"reason":%q}`, outcome, body.Reason))
	s.broadcastFight(r, fightID, "fight.technical")
	s.getFightState(w, r)
}
