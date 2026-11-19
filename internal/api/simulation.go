package api

import (
	"net/http"
	"strconv"

	"github.com/restful-otaku/j3/internal/simulator"
)

func simulationParams(r *http.Request) (int64, int) {
	seed, err := strconv.ParseInt(r.URL.Query().Get("seed"), 10, 64)
	if err != nil {
		seed = 42
	}
	count, err := strconv.Atoi(r.URL.Query().Get("fighters"))
	if err != nil || count < 8 {
		count = 50
	}
	if count > 200 {
		count = 200
	}
	return seed, count
}

// simulate runs a complete deterministic tournament for QA and what-if planning.
func (s *Server) simulate(w http.ResponseWriter, r *http.Request) {
	seed, count := simulationParams(r)
	writeJSON(w, http.StatusOK, simulator.Run(seed, count))
}

// stats powers the spectator charts: moves, club points, medals and rankings.
func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	seed, count := simulationParams(r)
	result := simulator.Run(seed, count)
	writeJSON(w, http.StatusOK, map[string]any{
		"topMoves": result.TopMoves, "clubPoints": result.ClubPoints,
		"clubMedals": result.ClubMedals, "fighterMedals": result.FighterMedals,
		"standings": result.Standings, "medals": result.Medals,
		"boutCount": len(result.Bouts),
	})
}
