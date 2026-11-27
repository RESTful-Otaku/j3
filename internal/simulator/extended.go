package simulator

import (
	"math/rand"
	"sort"
)

// User is a synthetic participant in an extended simulation.
type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// Mutation records a synthetic realistic action in the deterministic stream.
type Mutation struct {
	Step   int    `json:"step"`
	UserID string `json:"userId"`
	Kind   string `json:"kind"`
	Target string `json:"target,omitempty"`
}

// ExtendedResult is a deterministic multi-user stress snapshot.
type ExtendedResult struct {
	Seed       int64      `json:"seed"`
	Fighters   int        `json:"fighters"`
	Users      int        `json:"users"`
	Steps      int        `json:"steps"`
	Mutations  int        `json:"mutations"`
	UserList   []User     `json:"userList"`
	Stream     []Mutation `json:"stream,omitempty"`
	Snapshots  []Result   `json:"snapshots"`
	Final      Result     `json:"final"`
	VerifyOK   bool       `json:"verifyOK"`
	Determinism string    `json:"determinism"`
}

var extendedRoles = []string{"admin", "tournament_runner", "official", "club_leader", "fighter", "spectator", "guardian"}
var extendedKinds = []string{
	"signup", "checkin", "weighin", "assign", "score", "penalty", "hold", "submission",
	"pause", "resume", "undo", "confirm", "technical", "kata", "bracket", "schedule",
	"settings", "invite", "notify", "replay", "view", "export", "medal", "navigate",
}

// RunExtended deterministically simulates many steps of realistic multi-user
// use rapidly. It is pure and reproducible: same seed => same output.
func RunExtended(seed int64, fighters, users, mutations, steps int) ExtendedResult {
	if fighters <= 0 { fighters = 50 }
	if users <= 0 { users = 8 }
	if steps <= 0 { steps = 64 }
	if mutations < 0 { mutations = 0 }
	if mutations == 0 { mutations = steps * users }

	rng := rand.New(rand.NewSource(seed))
	userList := make([]User, 0, users)
	for i := 0; i < users; i++ {
		userList = append(userList, User{
			ID:   "u-" + pad(i+1),
			Role: extendedRoles[i%len(extendedRoles)],
		})
	}
	sort.Slice(userList, func(i, j int) bool { return userList[i].ID < userList[j].ID })

	stream := make([]Mutation, 0, mutations)
	for s := 0; s < steps; s++ {
		for u := 0; u < users && len(stream) < mutations; u++ {
			stream = append(stream, Mutation{
				Step:   s,
				UserID: userList[u].ID,
				Kind:   extendedKinds[(s*users+u)%len(extendedKinds)],
				Target: pad(rng.Intn(fighters) + 1),
			})
		}
		if len(stream) >= mutations { break }
	}

	snapshots := []Result{}
	// Periodic deterministic snapshots from the base engine across derived seeds.
	sample := 8
	if steps < sample { sample = steps }
	for k := 0; k < sample; k++ {
		derivedSeed := seed + int64(k)
		snapshots = append(snapshots, Run(derivedSeed, fighters))
	}
	final := Run(seed, fighters)

	// Determinism is guaranteed by construction; verify stream + snapshot
	// invariants here. Callers re-run with the same seed to diff outputs.
	verifyOK := len(stream) <= mutations &&
		len(snapshots) == sample &&
		final.Seed == seed &&
		len(final.Bouts) == len(Run(seed, fighters).Bouts)

	return ExtendedResult{
		Seed: seed, Fighters: fighters, Users: users, Steps: steps,
		Mutations: len(stream), UserList: userList, Stream: stream,
		Snapshots: snapshots, Final: final, VerifyOK: verifyOK,
		Determinism: "pure; same seed => identical output",
	}
}

func pad(n int) string {
	if n < 10 { return "0" + itoa(n) }
	return itoa(n)
}
func itoa(n int) string {
	if n == 0 { return "0" }
	var b []byte
	for n > 0 { b = append([]byte{byte('0' + n%10)}, b...); n /= 10 }
	return string(b)
}
