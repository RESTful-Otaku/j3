package api

import "net/http"

type officialAssignment struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// assignOfficial assigns a referee, judge or scorekeeper to a contest.
func (s *Server) assignOfficial(w http.ResponseWriter, r *http.Request) {
	fightID := r.PathValue("id")
	var body officialAssignment
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	allowed := map[string]bool{"referee": true, "judge": true, "scorekeeper": true}
	if !allowed[body.Role] || body.UserID == "" {
		writeBadRequest(w, "userId and a valid role are required")
		return
	}
	if _, err := s.Store.ExecContext(r.Context(), `
		INSERT OR REPLACE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, ?)`,
		fightID, body.UserID, body.Role); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "official.assign", "fight", fightID,
		`{"userId":"`+body.UserID+`","role":"`+body.Role+`"}`)
	writeJSON(w, http.StatusOK, map[string]any{"fightId": fightID, "userId": body.UserID, "role": body.Role, "status": "assigned"})
}

func (s *Server) isAssigned(r *http.Request, fightID, userID string) bool {
	var found string
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT role FROM fight_officials WHERE fight_id = ? AND user_id = ? LIMIT 1`,
		fightID, userID).Scan(&found)
	return err == nil
}
