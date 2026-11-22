// Package api exposes the venue server's HTTP interface.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/restful-otaku/j3/internal/auth"
	"github.com/restful-otaku/j3/internal/db"
)

// Server holds dependencies for the venue HTTP API.
type Server struct {
	Store     *db.DB
	Version   string
	Hub       *LiveHub
	testToken string // enabled only from _test.go; bypasses bearer lookup in in-process tests
	handler   http.Handler
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
	mux.HandleFunc("GET /api/v1/demo/identities", s.demoIdentities)
	mux.HandleFunc("GET /api/v1/demo/credentials", s.demoCredentials)
	mux.HandleFunc("GET /api/v1/demo/accounts", s.demoAccounts)
	mux.HandleFunc("GET /api/v1/demo/view", s.demoView)
	mux.HandleFunc("GET /api/v1/demo/fights", s.demoFights)
	mux.HandleFunc("GET /api/v1/demo/my-contests", s.myContests)
	mux.HandleFunc("GET /api/v1/demo/role-requests", s.roleRequests)
	mux.HandleFunc("POST /api/v1/demo/role-requests/{id}/approve", s.approveRoleRequest)
	mux.HandleFunc("GET /api/v1/demo/pools", s.listPoolsHandler)
	mux.HandleFunc("POST /api/v1/demo/pools/generate", s.generatePools)
	mux.HandleFunc("GET /api/v1/demo/brackets", s.listBracketsHandler)
	mux.HandleFunc("POST /api/v1/demo/brackets/build", s.buildBracket)
	mux.HandleFunc("GET /api/v1/demo/medals", s.listMedalsHandler)
	mux.HandleFunc("POST /api/v1/demo/medals/award", s.awardMedals)
	mux.HandleFunc("GET /api/v1/demo/kata", s.listKataHandler)
	mux.HandleFunc("POST /api/v1/demo/kata/advance", s.advanceKata)
	mux.HandleFunc("GET /api/v1/demo/kata/catalogue", s.kataCatalogue)
	mux.HandleFunc("GET /api/v1/demo/training", s.listTraining)
	mux.HandleFunc("POST /api/v1/demo/training", s.createTraining)
	mux.HandleFunc("POST /api/v1/demo/rotation", s.buildRotation)
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
	mux.HandleFunc("POST /api/v1/demo/fights/{id}/officials", s.assignOfficial)
	mux.HandleFunc("GET /api/v1/demo/standings", s.standings)
	mux.HandleFunc("GET /api/v1/events/stream", s.streamEvents)
	mux.HandleFunc("GET /api/v1/fights/{id}/state", s.getFightState)
	mux.HandleFunc("POST /api/v1/fights/{id}/start", s.startFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/pause", s.pauseFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/resume", s.resumeFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/event", s.addFightEvent)
	mux.HandleFunc("POST /api/v1/fights/{id}/hold", s.startHold)
	mux.HandleFunc("POST /api/v1/fights/{id}/undo", s.undoFightEvent)
	mux.HandleFunc("POST /api/v1/fights/{id}/submit", s.submitFight)
	mux.HandleFunc("POST /api/v1/fights/{id}/technical", s.callTechnicalResult)
	mux.HandleFunc("GET /api/v1/demo/belts", s.belts)
	mux.HandleFunc("GET /api/v1/demo/simulate", s.simulate)
	mux.HandleFunc("GET /api/v1/demo/stats", s.stats)
	mux.HandleFunc("POST /api/v1/tournaments", s.createTournament)
	mux.HandleFunc("POST /api/v1/demo/auto-run", s.autoRun)
	mux.HandleFunc("GET /api/v1/demo/readiness", s.readiness)
	mux.HandleFunc("GET /api/v1/demo/event-config", s.eventConfig)
	mux.HandleFunc("GET /api/v1/demo/attendance", s.attendance)
	mux.HandleFunc("GET /api/v1/demo/techniques", s.listTechniques)
	mux.HandleFunc("POST /api/v1/auth/login", s.startLogin)
	mux.HandleFunc("POST /api/v1/auth/verify", s.verifyLogin)
	mux.HandleFunc("POST /api/v1/auth/demo", s.demoLogin)
	mux.HandleFunc("POST /api/v1/auth/reset", s.completeReset)
	mux.HandleFunc("POST /api/v1/auth/reset/request", s.requestReset)
	mux.HandleFunc("POST /api/v1/auth/password", s.changeOwnPassword)
	mux.HandleFunc("GET /api/v1/invites", s.listInvites)
	mux.HandleFunc("POST /api/v1/invites", s.createInvite)
	mux.HandleFunc("POST /api/v1/invites/accept", s.acceptInvite)
	mux.HandleFunc("POST /api/v1/invites/{id}/revoke", s.revokeInvite)
	mux.HandleFunc("POST /api/v1/demo/announcements", s.sendAnnouncement)
	mux.HandleFunc("GET /api/v1/demo/announcements", s.listAnnouncements)
	mux.HandleFunc("GET /api/v1/demo/suggestions", s.suggestions)
	mux.HandleFunc("GET /api/v1/demo/admin/users", s.adminUsers)
	mux.HandleFunc("POST /api/v1/demo/admin/users/{id}/status", s.setUserStatus)
	mux.HandleFunc("POST /api/v1/demo/admin/users/{id}/reset", s.resetPassword)
	mux.HandleFunc("GET /api/v1/demo/permissions", s.permissions)
	mux.HandleFunc("POST /api/v1/demo/permissions", s.updatePermission)
	mux.HandleFunc("POST /api/v1/demo/entrants/{id}/checkin", s.checkIn)
	mux.HandleFunc("GET /api/v1/demo/team-rankings", s.teamRankings)
	mux.HandleFunc("GET /api/v1/demo/export", s.exportCSV)
	mux.HandleFunc("POST /api/v1/demo/import", s.importEntries)
	mux.HandleFunc("GET /api/v1/templates", s.templates)
	mux.HandleFunc("POST /api/v1/templates/{id}/import", s.importTemplate)
	mux.HandleFunc("POST /api/v1/templates/{id}/import-people", s.importPeople)
	mux.HandleFunc("GET /api/v1/demo/audit", s.auditLog)
	mux.HandleFunc("GET /api/v1/query", s.query)
	mux.HandleFunc("POST /api/v1/demo/fights/{id}/forfeit", s.forfeitFight)
	mux.HandleFunc("POST /api/v1/demo/entrants/{id}/status", s.entrantStatus)
	mux.HandleFunc("POST /api/v1/uploads", s.upload)
	_ = os.MkdirAll("data/uploads", 0o755)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("data/uploads"))))
	mux.HandleFunc("GET /", s.root)
	s.handler = s.middleware(mux)
	return s
}

// Handler returns the root handler, suitable for http.Server or httptest.
func (s *Server) Handler() http.Handler { return s.handler }

// allowedCORSOrigin echoes a configured origin (or a comma-separated list)
// only when the request actually comes from one of them. `*` keeps the old
// demo behaviour of reflecting any single origin.
func allowedCORSOrigin(r *http.Request) string {
	requestOrigin := r.Header.Get("Origin")
	if requestOrigin == "" {
		return ""
	}
	for _, configured := range strings.Split(os.Getenv("J3_CORS_ORIGIN"), ",") {
		configured = strings.TrimSpace(configured)
		if configured == "*" || configured == requestOrigin {
			return requestOrigin
		}
	}
	return ""
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A handler bug must never take down a live venue or leak a stack trace
		// to a display. Log it and return a predictable JSON error instead.
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic serving %s %s: %v\n%s", r.Method, r.URL.Path, recovered, debug.Stack())
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'")
		if origin := allowedCORSOrigin(r); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}
		if id := requestID(); id != "" {
			w.Header().Set("X-Request-ID", id)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if isPublic(r) {
			// Public endpoints may still personalise themselves (for example the
			// demo identity directory). A bad token never blocks public data.
			if r.Header.Get("Authorization") != "" {
				if identity, ok := s.identityFromRequest(r); ok {
					r = r.WithContext(auth.WithIdentity(r.Context(), identity))
				}
			}
		} else {
			var identity auth.Identity
			var ok bool
			if s.testToken != "" && r.Header.Get("Authorization") == "" {
				identity, ok = auth.Identity{UserID: "u-tony", Role: auth.RoleTournamentRunner}, true
			} else {
				identity, ok = s.identityFromRequest(r)
			}
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"authentication required"}`))
				return
			}
			r = r.WithContext(auth.WithIdentity(r.Context(), identity))
			// A one-time/generated password grants access to the change form
			// only. Public boards remain readable so the login page still works.
			if identity.MustChangePassword && r.URL.Path != "/api/v1/auth/password" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error":   "password_change_required",
					"message": "Set a new password before using this account.",
				})
				return
			}
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
