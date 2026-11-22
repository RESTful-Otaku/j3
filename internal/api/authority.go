package api

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/restful-otaku/j3/internal/auth"
)

// userRole reads the effective role, including the admin flag that is stored
// alongside the runner role for backwards-compatible SQLite CHECK constraints.
func (s *Server) userRole(ctx context.Context, userID string) (auth.Role, error) {
	var persisted string
	var admin int
	err := s.Store.QueryRowContext(ctx,
		`SELECT role, is_admin FROM users WHERE id = ?`, userID).Scan(&persisted, &admin)
	if err != nil {
		return "", err
	}
	return auth.VirtualRole(persisted, admin == 1), nil
}

func (s *Server) userClub(ctx context.Context, userID string) (string, error) {
	var club sql.NullString
	err := s.Store.QueryRowContext(ctx,
		`SELECT club_id FROM users WHERE id = ?`, userID).Scan(&club)
	if err != nil {
		return "", err
	}
	return club.String, nil
}

// actorScope resolves the verified session identity and the club it belongs
// to. Invitation handlers use this to prevent club owners acting outside their
// own club.
func (s *Server) actorScope(r *http.Request) (auth.Identity, string, bool) {
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		return auth.Identity{}, "", false
	}
	club, err := s.userClub(r.Context(), identity.UserID)
	if err != nil && err != sql.ErrNoRows {
		return auth.Identity{}, "", false
	}
	return identity, club, true
}

// requireOrgRole guards platform-level configuration: admins and runners own
// tournament templates and role permissions; club owners may brand their own
// club but cannot change the authority matrix.
func (s *Server) requireOrgRole(w http.ResponseWriter, r *http.Request) bool {
	role, ok := s.effectiveRole(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return false
	}
	if role != auth.RoleAdmin && role != auth.RoleTournamentRunner {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only admins and runners may do this"})
		return false
	}
	return true
}

// requireAdmin guards changes to the authority matrix itself. Runners may use
// permissions but must not be able to rewrite what any role may do.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	role, ok := s.effectiveRole(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return false
	}
	if role != auth.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only admins may change permissions"})
		return false
	}
	return true
}

// canManageUser scopes administrative user operations. Club owners may manage
// only their own fighters; officials may manage participants; runners and
// admins may manage everybody.
func canManageUser(actor auth.Role, actorClub string, target auth.Role, targetClub string) bool {
	switch actor {
	case auth.RoleAdmin, auth.RoleTournamentRunner:
		return true
	case auth.RoleOfficial:
		return target == auth.RoleFighter || target == auth.RoleSpectator
	case auth.RoleClubLeader:
		return actorClub != "" && (target == auth.RoleFighter || target == auth.RoleSpectator) && targetClub == actorClub
	default:
		return false
	}
}

// effectiveRole refreshes the session role from the database. Role promotions
// and demotions are audited material changes, so a stale session must not keep
// an old authority level.
func (s *Server) effectiveRole(r *http.Request) (auth.Role, bool) {
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		return "", false
	}
	role, err := s.userRole(r.Context(), identity.UserID)
	if err != nil {
		return identity.Role, true
	}
	return role, true
}
