// Package api exposes the venue server's HTTP interface.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/restful-otaku/j3/internal/db"
)

// Server holds dependencies for the venue HTTP API.
type Server struct {
	Store   *db.DB
	Version string
	handler http.Handler
}

// New builds a server with its middleware chain.
func New(store *db.DB, version string) *Server {
	s := &Server{Store: store, Version: version}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("GET /api/v1/version", s.versionInfo)
	mux.HandleFunc("GET /api/v1/roles", s.roles)
	mux.HandleFunc("GET /api/v1/demo", s.demo)
	mux.HandleFunc("GET /api/v1/demo/accounts", s.demoAccounts)
	mux.HandleFunc("GET /api/v1/demo/view", s.demoView)
	mux.HandleFunc("GET /api/v1/demo/fights", s.demoFights)
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
