// Package auth holds session identity and authorization primitives.
package auth

import (
	"context"
	"strings"
)

type contextKey struct{}

// Identity is the verified caller attached to the request context.
type Identity struct {
	UserID             string
	Role               Role
	MustChangePassword bool
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

func FromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}

// CanonicalRole maps UI role names to the RBAC role names. Admins are stored
// as tournament runners with an is_admin flag (see VirtualRole) so existing
// databases keep their CHECK constraint while the API exposes a true admin
// role in the invite hierarchy.
func CanonicalRole(value string) Role {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "admin", "owner":
		return RoleAdmin
	case "runner", "organiser", "organizer", "tournament_runner":
		return RoleTournamentRunner
	case "official", "moderator":
		return RoleOfficial
	case "club_owner", "club_leader":
		return RoleClubLeader
	case "fighter", "participant":
		return RoleFighter
	case "spectator":
		return RoleSpectator
	default:
		return ""
	}
}

// VirtualRole combines the persisted role with the admin flag into the role
// exposed to sessions, UI and invitation checks.
func VirtualRole(persisted string, isAdmin bool) Role {
	if isAdmin {
		return RoleAdmin
	}
	return CanonicalRole(persisted)
}

// PersistedRole is the inverse of VirtualRole: an admin is a flagged runner in
// storage, every other role is stored verbatim.
func PersistedRole(role Role) (string, bool) {
	if role == RoleAdmin {
		return string(RoleTournamentRunner), true
	}
	return string(role), false
}
