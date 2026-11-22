package auth

import "testing"

func TestInviteChain(t *testing.T) {
	cases := []struct {
		name                  string
		actor, target         Role
		actorClub, targetClub string
		want                  bool
	}{
		{"admin invites admin", RoleAdmin, RoleAdmin, "", "", true},
		{"admin invites runner", RoleAdmin, RoleTournamentRunner, "", "", true},
		{"runner invites official", RoleTournamentRunner, RoleOfficial, "", "", true},
		{"runner invites club owner", RoleTournamentRunner, RoleClubLeader, "", "", true},
		{"runner invites participant", RoleTournamentRunner, RoleFighter, "", "", true},
		{"runner cannot invite admin", RoleTournamentRunner, RoleAdmin, "", "", false},
		{"runner cannot invite runner", RoleTournamentRunner, RoleTournamentRunner, "", "", false},
		{"official invites club owner", RoleOfficial, RoleClubLeader, "", "", true},
		{"official invites participant", RoleOfficial, RoleFighter, "", "", true},
		{"official cannot invite official", RoleOfficial, RoleOfficial, "", "", false},
		{"club owner invites own participant", RoleClubLeader, RoleFighter, "club-a", "club-a", true},
		{"club owner cannot invite other club", RoleClubLeader, RoleFighter, "club-a", "club-b", false},
		{"club owner cannot invite club owner", RoleClubLeader, RoleClubLeader, "club-a", "club-a", false},
		{"fighter invites nobody", RoleFighter, RoleFighter, "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanInvite(tc.actor, tc.target, tc.actorClub, tc.targetClub); got != tc.want {
				t.Fatalf("CanInvite = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanonicalAndVirtualRole(t *testing.T) {
	for input, want := range map[string]Role{
		"admin": RoleAdmin, "owner": RoleAdmin, "organiser": RoleTournamentRunner,
		"moderator": RoleOfficial, "club_owner": RoleClubLeader,
		"participant": RoleFighter, "spectator": RoleSpectator, "unknown": "",
	} {
		if got := CanonicalRole(input); got != want {
			t.Errorf("CanonicalRole(%q) = %q, want %q", input, got, want)
		}
	}
	if got := VirtualRole("tournament_runner", true); got != RoleAdmin {
		t.Fatalf("VirtualRole admin = %q", got)
	}
	if got := VirtualRole("official", false); got != RoleOfficial {
		t.Fatalf("VirtualRole official = %q", got)
	}
	role, admin := PersistedRole(RoleAdmin)
	if role != "tournament_runner" || !admin {
		t.Fatalf("PersistedRole admin = %q, %v", role, admin)
	}
}

func TestSelfSignupAndApproval(t *testing.T) {
	if !SelfSignupRoles(RoleFighter) || !SelfSignupRoles(RoleSpectator) {
		t.Fatal("participants and spectators must be able to self-register")
	}
	for _, role := range []Role{RoleAdmin, RoleTournamentRunner, RoleOfficial, RoleClubLeader} {
		if SelfSignupRoles(role) {
			t.Errorf("role %q must require an invitation", role)
		}
	}
	if !CanApprove(RoleAdmin, RoleTournamentRunner) || CanApprove(RoleClubLeader, RoleTournamentRunner) {
		t.Fatal("only admins can approve runner requests")
	}
}
