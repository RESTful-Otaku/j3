// Package api exposes the venue server's HTTP interface.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/restful-otaku/j3/internal/db"
)

// Server holds dependencies for the venue HTTP API.
type Server struct {
	Store   *db.DB
	Version string
	Hub     *LiveHub
	handler http.Handler
}

// New builds a server with its middleware chain.
func New(store *db.DB, version string) *Server {
	s := &Server{Store: store, Version: version, Hub: newLiveHub()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("GET /api/v1/version", s.versionInfo)
	mux.HandleFunc("GET /api/v1/roles", s.roles)
	mux.HandleFunc("GET /api/v1/demo", s.demo)
	mux.HandleFunc("GET /api/v1/demo/accounts", s.demoAccounts)
	mux.HandleFunc("GET /api/v1/demo/view", s.demoView)
	mux.HandleFunc("GET /api/v1/demo/fights", s.demoFights)
	mux.HandleFunc("GET /api/v1/demo/role-requests", s.roleRequests)
	mux.HandleFunc("POST /api/v1/demo/role-requests/{id}/approve", s.approveRoleRequest)
	mux.HandleFunc("GET /api/v1/demo/pools", s.listPoolsHandler)
	mux.HandleFunc("POST /api/v1/demo/pools/generate", s.generatePools)
	mux.HandleFunc("GET /api/v1/demo/brackets", s.listBracketsHandler)
	mux.HandleFunc("POST /api/v1/demo/brackets/build", s.buildBracket)
	mux.HandleFunc("GET /api/v1/demo/kata", s.listKataHandler)
	mux.HandleFunc("POST /api/v1/demo/kata/performances", s.createKataPerformance)
	mux.HandleFunc("POST /api/v1/demo/kata/{id}/scores", s.submitKataScore)
	mux.HandleFunc("GET /api/v1/demo/schedule", s.listSchedule)
	mux.HandleFunc("POST /api/v1/demo/schedule/drift", s.setDrift)
	mux.HandleFunc("POST /api/v1/demo/schedule/{id}/complete", s.completeScheduleItem)
	mux.HandleFunc("POST /api/v1/demo/brackets/slots/{id}/promote", s.promoteSlot)
	mux.HandleFunc("POST /api/v1/signup/preview", s.signupPreview)
	mux.HandleFunc("POST /api/v1/signup", s.signup)
	mux.HandleFunc("GET /api/v1/demo/event", s.eventDisplay)
	mux.HandleFunc("GET /api/v1/demo/officials", s.displayOfficials)
	mux.HandleFunc("GET /api/v1/demo/standings", s.standings)
	mux.HandleFunc("GET /api/v1/events/stream", s.streamEvents)
	mux.HandleFunc("GET /api/v1/fights/{id}/state", s.getFightState)
	mux.HandleFunc("POST /api/v1/fights/{id}/start", s.startFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/pause", s.pauseFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/resume", s.resumeFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/event", s.addFightEvent)
	mux.HandleFunc("POST /api/v1/fights/{id}/undo", s.undoFightEvent)
	mux.HandleFunc("POST /api/v1/fights/{id}/submit", s.submitFight)
	mux.HandleFunc("GET /api/v1/demo/belts", s.belts)
	mux.HandleFunc("GET /api/v1/demo/simulate", s.simulate)
	mux.HandleFunc("GET /api/v1/demo/stats", s.stats)
	mux.HandleFunc("POST /api/v1/tournaments", s.createTournament)
	mux.HandleFunc("GET /api/v1/templates", s.templates)
	mux.HandleFunc("GET /api/v1/demo/audit", s.auditLog)
	mux.HandleFunc("GET /api/v1/query", s.query)
	mux.HandleFunc("POST /api/v1/demo/fights/{id}/forfeit", s.forfeitFight)
	mux.HandleFunc("POST /api/v1/demo/entrants/{id}/status", s.entrantStatus)
	mux.HandleFunc("POST /api/v1/uploads", s.upload)
	_ = os.MkdirAll("data/uploads", 0o755)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("data/uploads"))))
	mux.HandleFunc("GET /", s.root)
	s.handler = middleware(mux)
	return s
}

// Handler returns the root handler, suitable for http.Server or httptest.
func (s *Server) Handler() http.Handler { return s.handler }

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if id := requestID(); id != "" {
			w.Header().Set("X-Request-ID", id)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
