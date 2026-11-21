package api

import "net/http"

type technique struct {
	ID      string `json:"id"`
	English string `json:"english"`
	Romaji  string `json:"romaji"`
	Points  int    `json:"points"`
	Kind    string `json:"kind"`
}

var techniques = []technique{
	{"strike", "Strike / punch", "Tsuki", 1, "score"},
	{"kick", "Kick", "Keri", 1, "score"},
	{"throw", "Clean throw", "Nage", 3, "score"},
	{"counter_throw", "Throw from defended kick", "Kaeshi-nage", 3, "score"},
	{"takedown", "Takedown", "Taoshi", 5, "score"},
	{"hold", "Hold down (3 sec)", "Osae-komi", 5, "score"},
	{"submission", "Submission / tap", "Tap-out", 0, "submission"},
	{"ippon", "Ippon", "Ippon", 3, "score"},
	{"wazaari", "Waza-ari", "Waza-ari", 1, "score"},
	{"yellow", "Yellow card / foul", "Shido", 0, "penalty"},
	{"red", "Red card / serious foul", "Hansoku", 0, "penalty"},
}

func (s *Server) listTechniques(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"techniques": techniques})
}
