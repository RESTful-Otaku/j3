package rules_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/restful-otaku/j3/internal/rules"
)

type simFighter struct{ id, club string }
type simBout struct {
	a, b, outcome, stage string
}

type simStats struct {
	wins, losses, draws, submissions, forfeits int
}

// simulateTournament runs a complete event: pools -> round robin -> knockout
// -> bronze decider, with deterministic randomness from the given seed.
func simulateTournament(seed int64) ([]simBout, map[string]simStats, []string) {
	rng := rand.New(rand.NewSource(seed))
	fighters := make([]simFighter, 48)
	for i := range fighters {
		fighters[i] = simFighter{id: fmt.Sprintf("f%02d", i), club: fmt.Sprintf("club-%d", i%4)}
	}
	entrants := make([]rules.Entrant, len(fighters))
	for i, f := range fighters {
		entrants[i] = rules.Entrant{ID: f.id, ClubID: f.club, Seed: i + 1}
	}
	pools := rules.GeneratePools(entrants, rules.PoolConstraints{
		TargetPoolSize: 6, MaxPools: 8, SameClubSeparation: true, Seed: seed,
	})

	stats := map[string]simStats{}
	bouts := []simBout{}
	register := func(a, b, outcome, stage string) {
		bouts = append(bouts, simBout{a, b, outcome, stage})
		sa, sb := stats[a], stats[b]
		switch outcome {
		case "A":
			sa.wins++
			sb.losses++
		case "B":
			sb.wins++
			sa.losses++
		case "DRAW":
			sa.draws++
			sb.draws++
		case "SA":
			sa.wins++
			sa.submissions++
			sb.losses++
		case "SB":
			sb.wins++
			sb.submissions++
			sa.losses++
		case "FA":
			sb.wins++
			sb.forfeits++
			sa.losses++
		case "FB":
			sa.wins++
			sa.forfeits++
			sb.losses++
		}
		stats[a], stats[b] = sa, sb
	}

	// Round robin inside every pool.
	advance := []simFighter{}
	for _, pool := range pools {
		ids := pool.Members
		for i := range ids {
			for j := i + 1; j < len(ids); j++ {
				outcomes := []string{"A", "B", "DRAW", "SA", "SB", "FA", "FB"}
				register(ids[i], ids[j], outcomes[rng.Intn(len(outcomes))], "pool")
			}
		}
		ranked := append([]string(nil), ids...)
		sortByPoints(ranked, stats)
		if len(ranked) > 0 {
			advance = append(advance, simFighter{id: ranked[0], club: clubOf(fighters, ranked[0])})
		}
		if len(ranked) > 1 {
			advance = append(advance, simFighter{id: ranked[1], club: clubOf(fighters, ranked[1])})
		}
	}

	// Knockout stage with a third-place decider.
	if len(advance) > 8 {
		advance = advance[:8]
	}
	semiLosers := []simFighter{}
	for len(advance) > 1 {
		next := []simFighter{}
		for i := 0; i+1 < len(advance); i += 2 {
			outcome := []string{"A", "B", "SA", "SB", "FA", "FB"}[rng.Intn(6)]
			stage := "quarter"
			if len(advance) == 4 {
				stage = "semi"
			}
			if len(advance) == 2 {
				stage = "final"
			}
			register(advance[i].id, advance[i+1].id, outcome, stage)
			winner := advance[i]
			loser := advance[i+1]
			if outcome == "B" || outcome == "SB" || outcome == "FB" {
				winner, loser = advance[i+1], advance[i]
			}
			next = append(next, winner)
			if len(advance) == 4 { // semi-final losers feed the bronze decider
				semiLosers = append(semiLosers, loser)
			}
		}
		advance = next
	}

	medals := []string{}
	if len(advance) > 0 {
		medals = append(medals, advance[0].id)
	}
	if len(semiLosers) == 2 {
		if hasMet(bouts, semiLosers[0].id, semiLosers[1].id) {
			// Configurable "no rematch for bronze" rule: pick an alternative
			// loser from earlier rounds when one is available.
			for _, bout := range bouts {
				candidate := bout.a
				if candidate == semiLosers[0].id || candidate == semiLosers[1].id {
					candidate = bout.b
				}
				if candidate != semiLosers[0].id && candidate != semiLosers[1].id &&
					!hasMet(bouts, semiLosers[0].id, candidate) {
					semiLosers[1] = simFighter{id: candidate}
					break
				}
			}
		}
		outcome := []string{"A", "B", "DRAW"}[rng.Intn(3)]
		register(semiLosers[0].id, semiLosers[1].id, outcome, "bronze")
		if outcome == "DRAW" {
			medals = append(medals, semiLosers[0].id, semiLosers[1].id)
		} else {
			medals = append(medals, semiLosers[0].id)
		}
	}
	return bouts, stats, medals
}

func sortByPoints(ids []string, stats map[string]simStats) {
	for i := 1; i < len(ids); i++ {
		for j := i; j > 0; j-- {
			left, right := stats[ids[j-1]], stats[ids[j]]
			leftPts := left.wins*3 + left.draws + left.submissions
			rightPts := right.wins*3 + right.draws + right.submissions
			if leftPts >= rightPts {
				break
			}
			ids[j-1], ids[j] = ids[j], ids[j-1]
		}
	}
}

func clubOf(fighters []simFighter, id string) string {
	for _, f := range fighters {
		if f.id == id {
			return f.club
		}
	}
	return ""
}

func hasMet(bouts []simBout, a, b string) bool {
	for _, bout := range bouts {
		if (bout.a == a && bout.b == b) || (bout.a == b && bout.b == a) {
			return true
		}
	}
	return false
}

func TestFullTournamentSimulation(t *testing.T) {
	for seed := int64(1); seed <= 120; seed++ {
		bouts, stats, medals := simulateTournament(seed)
		if len(bouts) == 0 {
			t.Fatalf("seed %d: no bouts", seed)
		}
		seenBouts := map[string]bool{}
		for _, bout := range bouts {
			if bout.a == bout.b {
				t.Fatalf("seed %d: fighter fought themselves", seed)
			}
			key := bout.stage + ":" + bout.a + "|" + bout.b
			if seenBouts[key] {
				t.Fatalf("seed %d: duplicate pairing %s", seed, key)
			}
			seenBouts[key] = true
			if _, ok := stats[bout.a]; !ok {
				t.Fatalf("seed %d: missing stats for %s", seed, bout.a)
			}
		}
		if len(medals) < 2 || len(medals) > 3 {
			t.Fatalf("seed %d: unexpected medal count %d", seed, len(medals))
		}
		medalSeen := map[string]bool{}
		for _, id := range medals {
			if medalSeen[id] {
				t.Fatalf("seed %d: duplicate medal for %s", seed, id)
			}
			medalSeen[id] = true
		}
		// Deterministic replay.
		bouts2, stats2, medals2 := simulateTournament(seed)
		if fmt.Sprint(bouts) != fmt.Sprint(bouts2) ||
			fmt.Sprint(stats) != fmt.Sprint(stats2) ||
			fmt.Sprint(medals) != fmt.Sprint(medals2) {
			t.Fatalf("seed %d: simulation is not deterministic", seed)
		}
	}
}

func TestSimulationEdgeCases(t *testing.T) {
	// Sparse fields and tiny pools should still produce coherent results.
	for seed := int64(0); seed < 20; seed++ {
		_, stats, medals := simulateTournament(seed * 7)
		for id, s := range stats {
			if s.wins < 0 || s.losses < 0 || s.draws < 0 || id == "" {
				t.Fatalf("invalid stats for %s: %+v", id, s)
			}
		}
		if len(medals) == 0 {
			t.Fatal("no medals awarded")
		}
	}
}
