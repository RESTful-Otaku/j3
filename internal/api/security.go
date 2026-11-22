package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
)

const sessionTimeLayout = "2006-01-02T15:04:05.000Z"

var publicGetRoutes = map[string]bool{
	"/healthz": true, "/readyz": true, "/api/v1/version": true, "/api/v1/roles": true,
	"/api/v1/demo": true, "/api/v1/demo/fights": true, "/api/v1/demo/standings": true,
	"/api/v1/demo/identities": true, "/api/v1/demo/credentials": true,
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
	if r.Method == http.MethodPost && r.URL.Path == "/api/v1/invites/accept" {
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
	var userID, persisted, status, expires string
	var admin, mustChange int
	// The session row is only a bearer token. Current authority, account status
	// and the one-time-password flag must come from the users table on every
	// request, otherwise demotions, suspensions and forced password changes
	// would not take effect until the token expired.
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT s.user_id, u.role, u.status, s.expires_at, u.is_admin, u.must_change_password
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token = ?`, token).
		Scan(&userID, &persisted, &status, &expires, &admin, &mustChange)
	if err != nil {
		return auth.Identity{}, false
	}
	expiry, err := parseSessionExpiry(expires)
	if err != nil || time.Now().After(expiry) || status != "active" {
		return auth.Identity{}, false
	}
	role := auth.VirtualRole(persisted, admin == 1)
	if role == "" {
		return auth.Identity{}, false
	}
	return auth.Identity{UserID: userID, Role: role, MustChangePassword: mustChange == 1}, true
}

// parseSessionExpiry accepts every timestamp shape written by the login flows,
// including SQLite's millisecond strftime format and RFC3339 variants with or
// without fractional seconds. A token is never rejected just because its
// expiry was stored at a different precision.
func parseSessionExpiry(value string) (time.Time, error) {
	for _, layout := range []string{sessionTimeLayout, time.RFC3339Nano, time.RFC3339} {
		if expiry, err := time.Parse(layout, value); err == nil {
			return expiry, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised session expiry %q", value)
}

func (s *Server) requireAction(w http.ResponseWriter, r *http.Request, resource, action string) bool {
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return false
	}
	role, ok := s.effectiveRole(r)
	if !ok {
		role = identity.Role
	}
	column := map[string]string{"view": "can_view", "create": "can_create", "edit": "can_edit", "delete": "can_delete"}[action]
	if column == "" {
		http.Error(w, `{"error":"invalid action"}`, http.StatusBadRequest)
		return false
	}
	var allowed int
	err := s.Store.QueryRowContext(r.Context(),
		`SELECT `+column+` FROM role_permissions WHERE role = ? AND resource = ?`, role, resource).
		Scan(&allowed)
	if err == sql.ErrNoRows {
		// Admins and runners may use capabilities not yet configured.
		allowed = 0
		if role == auth.RoleAdmin || role == auth.RoleTournamentRunner {
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
