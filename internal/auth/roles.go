// Package auth defines the RBAC hierarchy used across J3.
//
// Authority flows down the invite chain: admin -> tournament runners ->
// officials/moderators -> club leaders/owners -> fighters/participants. RBAC
// controls both what a user may do and what they may see (private vs public),
// which is central to integrity and fairness.
package auth

// Role identifies a user's authority in a tournament.
type Role string

const (
	RoleAdmin            Role = "admin"
	RoleTournamentRunner Role = "tournament_runner"
	RoleOfficial         Role = "official"
	RoleClubLeader       Role = "club_leader"
	RoleFighter          Role = "fighter"
	RoleSpectator        Role = "spectator"
)

// Action is a capability guarded by RBAC.
type Action string

const (
	ActionManageTenant   Action = "manage_tenant"
	ActionManageEvent    Action = "manage_event"
	ActionManageUsers    Action = "manage_users"
	ActionManageClub     Action = "manage_club"
	ActionInviteFighters Action = "invite_fighters"
	ActionScoreFight     Action = "score_fight"
	ActionRegisterSelf   Action = "register_self"
	ActionViewPrivate    Action = "view_private"
	ActionViewPublic     Action = "view_public"
)

// RoleInfo describes a role for UI and API consumers.
type RoleInfo struct {
	ID          Role     `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Rank        int      `json:"rank"`
	Actions     []Action `json:"actions"`
}

var roleInfos = []RoleInfo{
	{RoleAdmin, "Admin / owner", "Owns the platform, creates runners and can invite any role.", 0, []Action{
		ActionManageTenant, ActionManageEvent, ActionManageUsers, ActionManageClub,
		ActionInviteFighters, ActionScoreFight, ActionRegisterSelf, ActionViewPrivate, ActionViewPublic,
	}},
	{RoleTournamentRunner, "Tournament runner / organiser", "Creates events and tournament templates, manages officials and clubs.", 1, []Action{
		ActionManageTenant, ActionManageEvent, ActionManageUsers, ActionManageClub,
		ActionInviteFighters, ActionScoreFight, ActionRegisterSelf, ActionViewPrivate, ActionViewPublic,
	}},
	{RoleOfficial, "Official / moderator", "Scores and officiates assigned contests and can invite club owners.", 2, []Action{
		ActionScoreFight, ActionManageUsers, ActionInviteFighters, ActionRegisterSelf, ActionViewPrivate, ActionViewPublic,
	}},
	{RoleClubLeader, "Club owner / leader", "Manages their own club and its participants.", 3, []Action{
		ActionManageClub, ActionInviteFighters, ActionRegisterSelf, ActionViewPrivate, ActionViewPublic,
	}},
	{RoleFighter, "Fighter / participant", "Registers independently or via a club invitation.", 4, []Action{
		ActionRegisterSelf, ActionViewPublic,
	}},
	{RoleSpectator, "Spectator", "Sees only explicitly published information.", 5, []Action{
		ActionViewPublic,
	}},
}

// Roles returns all roles, highest authority first.
func Roles() []RoleInfo { return roleInfos }

// Actions returns the capabilities granted to a role.
func Actions(role Role) []Action {
	for _, info := range roleInfos {
		if info.ID == role {
			return info.Actions
		}
	}
	return nil
}

// Valid reports whether r is a known role.
func Valid(r Role) bool {
	for _, info := range roleInfos {
		if info.ID == r {
			return true
		}
	}
	return false
}

// Can reports whether role is permitted to perform action.
func Can(role Role, action Action) bool {
	for _, info := range roleInfos {
		if info.ID == role {
			for _, a := range info.Actions {
				if a == action {
					return true
				}
			}
			return false
		}
	}
	return false
}
