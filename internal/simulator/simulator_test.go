package simulator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/restful-otaku/j3/internal/simulator"
)

func TestSimulationDeterministicAcrossSeedsAndSizes(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		for _, size := range []int{8, 17, 50, 96} {
			first := simulator.Run(seed, size)
			second := simulator.Run(seed, size)
			if fmt.Sprint(first) != fmt.Sprint(second) {
				t.Fatalf("seed %d size %d is not deterministic", seed, size)
			}
			if len(first.Bouts) == 0 {
				t.Fatalf("seed %d size %d produced no bouts", seed, size)
			}
			if len(first.Medals) < 2 {
				t.Fatalf("seed %d size %d produced too few medals", seed, size)
			}
			kinds := map[string]bool{}
			for _, medal := range first.Medals {
				kinds[medal.Kind] = true
			}
			if !kinds["gold"] || !kinds["silver"] || !kinds["bronze"] {
				t.Fatalf("seed %d size %d missing medal kinds: %+v", seed, size, first.Medals)
			}
			seen := map[string]bool{}
			for _, bout := range first.Bouts {
				if bout.RedID == bout.BlueID {
					t.Fatal("self bout")
				}
				if bout.RedID == "" || bout.BlueID == "" {
					t.Fatal("empty slot in bout")
				}
				seen[bout.ID] = true
			}
			if len(seen) != len(first.Bouts) {
				t.Fatal("duplicate bout id")
			}
		}
	}
}

func TestSimulationUsesRealisticScores(t *testing.T) {
	result := simulator.Run(42, 50)
	submissions, decisions := 0, 0
	for _, bout := range result.Bouts {
		if strings.Contains(bout.Outcome, "forfeit") {
			if bout.PointsRed != 0 || bout.PointsBlue != 0 {
				t.Fatalf("forfeit has points %d-%d", bout.PointsRed, bout.PointsBlue)
			}
			continue
		}
		if bout.Outcome == "draw" {
			if bout.PointsRed < 10 || bout.PointsRed > 40 || bout.PointsRed != bout.PointsBlue {
				t.Fatalf("unrealistic draw %d-%d", bout.PointsRed, bout.PointsBlue)
			}
			continue
		}
		for _, points := range []int{bout.PointsRed, bout.PointsBlue} {
			if points < 4 || points > 40 {
				t.Fatalf("bout %s has unrealistic score %d (method %s)", bout.ID, points, bout.Method)
			}
		}
		if strings.Contains(bout.Outcome, "submission") {
			submissions++
			if bout.Method == "" || len(bout.Techniques) == 0 {
				t.Fatalf("submission bout %s missing technique detail", bout.ID)
			}
			continue
		}
		decisions++
		winner, loser := bout.PointsRed, bout.PointsBlue
		if strings.Contains(bout.Outcome, "blue") {
			winner, loser = bout.PointsBlue, bout.PointsRed
		}
		if winner < 10 || winner > 40 || winner <= loser {
			t.Fatalf("bout %s decision score %d-%d", bout.ID, bout.PointsRed, bout.PointsBlue)
		}
	}
	if submissions == 0 || decisions == 0 {
		t.Fatalf("expected submissions and decisions, got %d and %d", submissions, decisions)
	}
}

func TestSimulationStats(t *testing.T) {
	result := simulator.Run(7, 50)
	if len(result.TopMoves) == 0 || len(result.ClubPoints) == 0 || len(result.ClubMedals) == 0 {
		t.Fatal("chart data missing")
	}
	if len(result.Standings) != 50 {
		t.Fatalf("standings = %d, want 50", len(result.Standings))
	}
}
