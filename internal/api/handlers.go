package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/restful-otaku/j3/internal/auth"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.PingContext(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable", "database": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}

func (s *Server) versionInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name": "J3 venue server", "version": s.Version,
	})
}

func (s *Server) roles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"roles": auth.Roles()})
}

// demo reports the seeded data so the local QA page can show a live snapshot.
func (s *Server) demo(w http.ResponseWriter, r *http.Request) {
	counts, err := s.counts(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"counts": counts})
}

func (s *Server) counts(r *http.Request) (map[string]int, error) {
	counts := map[string]int{}
	for name, query := range map[string]string{
		"tenants":     `SELECT COUNT(*) FROM tenants`,
		"clubs":       `SELECT COUNT(*) FROM clubs`,
		"users":       `SELECT COUNT(*) FROM users`,
		"fighters":    `SELECT COUNT(*) FROM fighters`,
		"beltRanks":   `SELECT COUNT(*) FROM belt_ranks`,
		"events":      `SELECT COUNT(*) FROM events`,
		"categories":  `SELECT COUNT(*) FROM categories`,
		"entrants":    `SELECT COUNT(*) FROM entrants`,
		"mats":        `SELECT COUNT(*) FROM mats`,
		"fights":      `SELECT COUNT(*) FROM fights`,
		"fightEvents": `SELECT COUNT(*) FROM fight_events`,
	} {
		var n int
		if err := s.Store.QueryRowContext(r.Context(), query).Scan(&n); err != nil {
			return nil, err
		}
		counts[name] = n
	}
	return counts, nil
}

// defaultEventID resolves the active event from the database instead of a
// hard-coded demo identifier, so newly created events work with every handler.
func (s *Server) defaultEventID(r *http.Request) string {
	var id string
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT id FROM events ORDER BY created_at DESC LIMIT 1`).Scan(&id); err == nil {
		return id
	}
	return "ev-okami"
}

// eventIDFromRequest honours an explicit eventId but otherwise follows the
// active event in the database. Display handlers must not assume a particular
// seed produced a particular event ID.
func (s *Server) eventIDFromRequest(r *http.Request) string {
	if id := strings.TrimSpace(r.URL.Query().Get("eventId")); id != "" {
		return id
	}
	return s.defaultEventID(r)
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	log.Printf("request error: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func (s *Server) root(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name": "J3 venue server",
		"docs": []string{"/healthz", "/readyz", "/api/v1/version", "/api/v1/roles", "/api/v1/demo"},
	})
}
