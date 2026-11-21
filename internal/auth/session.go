// Package auth holds session identity and authorization primitives.
package auth

import "context"

type contextKey struct{}

// Identity is the verified caller attached to the request context.
type Identity struct {
	UserID string
	Role   Role
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

func FromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}

// CanonicalRole maps UI role names to the persisted RBAC role names.
func CanonicalRole(value string) Role {
	switch value {
	case "runner", "tournament_runner":
		return RoleTournamentRunner
	case "official", "moderator":
		return RoleOfficial
	case "club_leader":
		return RoleClubLeader
	case "fighter", "participant":
		return RoleFighter
	case "spectator":
		return RoleSpectator
	default:
		return ""
	}
}
