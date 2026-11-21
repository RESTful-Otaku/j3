package api

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
)

var publicGetRoutes = map[string]bool{
	"/healthz": true, "/readyz": true, "/api/v1/version": true, "/api/v1/roles": true,
	"/api/v1/demo": true, "/api/v1/demo/fights": true, "/api/v1/demo/standings": true,
	"/api/v1/demo/event": true, "/api/v1/demo/belts": true, "/api/v1/demo/simulate": true,
	"/api/v1/demo/stats": true, "/api/v1/demo/kata/catalogue": true,
	"/api/v1/demo/brackets": true, "/api/v1/demo/pools": true, "/api/v1/demo/medals": true,
	"/api/v1/demo/schedule": true, "/api/v1/demo/event-config": true,
	"/api/v1/demo/team-rankings": true, "/api/v1/demo/techniques": true,
	"/api/v1/demo/announcements": true, "/api/v1/demo/suggestions": true,
	"/api/v1/templates":     true,
	"/api/v1/events/stream": true,
}

var publicPostPrefixes = []string{"/api/v1/auth/", "/api/v1/signup"}

func demoMode() bool { return os.Getenv("J3_DEMO") == "1" }

func isPublic(r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/") && !strings.HasPrefix(r.URL.Path, "/uploads/") {
		return true
	}
	if r.Method == http.MethodGet && publicGetRoutes[r.URL.Path] {
		return true
	}
	if r.Method == http.MethodPost {
		for _, prefix := range publicPostPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}
		if r.URL.Path == "/api/v1/auth/demo" && demoMode() {
			return true
		}
	}
	return false
}

func (s *Server) identityFromRequest(r *http.Request) (auth.Identity, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return auth.Identity{}, false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	var userID, role string
	var expires string
	err := s.Store.QueryRowContext(context.Background(),
		`SELECT user_id, role, expires_at FROM sessions WHERE token = ?`, token).
		Scan(&userID, &role, &expires)
	if err != nil || expires < nowString() {
		return auth.Identity{}, false
	}
	canonical := auth.CanonicalRole(role)
	if canonical == "" {
		return auth.Identity{}, false
	}
	return auth.Identity{UserID: userID, Role: canonical}, true
}

func (s *Server) requireAction(w http.ResponseWriter, r *http.Request, resource, action string) bool {
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return false
	}
	column := map[string]string{"view": "can_view", "create": "can_create", "edit": "can_edit", "delete": "can_delete"}[action]
	if column == "" {
		http.Error(w, `{"error":"invalid action"}`, http.StatusBadRequest)
		return false
	}
	var allowed int
	err := s.Store.QueryRowContext(r.Context(),
		`SELECT `+column+` FROM role_permissions WHERE role = ? AND resource = ?`, identity.Role, resource).
		Scan(&allowed)
	if err == sql.ErrNoRows {
		// Only the runner may use capabilities not yet configured.
		allowed = 0
		if identity.Role == auth.RoleTournamentRunner {
			allowed = 1
		}
	} else if err != nil {
		http.Error(w, `{"error":"authorization unavailable"}`, http.StatusServiceUnavailable)
		return false
	}
	if allowed != 1 {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return false
	}
	return true
}

func nowString() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
}
