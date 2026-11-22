package auth

import "strings"

// Rank returns the authority rank: lower numbers mean more authority. Unknown
// roles rank below spectators so they can never invite anybody.
func Rank(role Role) int {
	for _, info := range roleInfos {
		if info.ID == role {
			return info.Rank
		}
	}
	return 99
}

// SelfSignupRoles are the only accounts a visitor may create without an
// invitation. Everyone else arrives through the invite chain.
func SelfSignupRoles(role Role) bool {
	return role == RoleFighter || role == RoleSpectator
}

// CanInvite enforces the platform authority chain.
//
// admin -> runners/organisers -> officials/moderators -> club owners ->
// participants. Admins may invite anybody. Runners may invite officials, club
// owners and participants. Officials may invite club owners and participants.
// Club owners may invite participants to their own club only.
func CanInvite(actor, target Role, actorClubID, targetClubID string) bool {
	if !Valid(actor) || !Valid(target) {
		return false
	}
	switch actor {
	case RoleAdmin:
		return true
	case RoleTournamentRunner:
		return target == RoleOfficial || target == RoleClubLeader ||
			target == RoleFighter || target == RoleSpectator
	case RoleOfficial:
		return target == RoleClubLeader || target == RoleFighter || target == RoleSpectator
	case RoleClubLeader:
		if strings.TrimSpace(actorClubID) == "" {
			return false
		}
		return (target == RoleFighter || target == RoleSpectator) &&
			strings.TrimSpace(targetClubID) == strings.TrimSpace(actorClubID)
	default:
		return false
	}
}

// CanApprove mirrors the invite chain for legacy role requests. A club owner
// can no longer approve a runner or admin request.
func CanApprove(actor, requested Role) bool {
	switch actor {
	case RoleAdmin:
		return true
	case RoleTournamentRunner:
		return requested == RoleOfficial || requested == RoleClubLeader || requested == RoleFighter
	case RoleOfficial:
		return requested == RoleClubLeader || requested == RoleFighter
	default:
		return false
	}
}
