package api

import (
	"net/http"

	"github.com/restful-otaku/j3/internal/rules"
)

// buildRotation generates a round-robin free-sparring rotation with BYE handling.
func (s *Server) buildRotation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Fighters []string `json:"fighters"`
		Mats     []string `json:"mats"`
		Rounds   int      `json:"rounds"`
	}
	if err := decodeJSON(r, &body); err != nil { writeBadRequest(w, err.Error()); return }
	if len(body.Fighters) < 2 { writeBadRequest(w, "at least two fighters are required"); return }
	if len(body.Mats) == 0 { body.Mats = []string{"Mat 1"} }
	pairs := rules.BuildRotation(body.Fighters, body.Mats, body.Rounds)
	s.audit(r, "rotation.build", "event", "ev-demo", "")
	writeJSON(w, http.StatusOK, map[string]any{"pairs": pairs, "count": len(pairs)})
}
