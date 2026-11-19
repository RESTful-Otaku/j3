package simulator_test

import (
	"fmt"
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

func TestSimulationStats(t *testing.T) {
	result := simulator.Run(7, 50)
	if len(result.TopMoves) == 0 || len(result.ClubPoints) == 0 || len(result.ClubMedals) == 0 {
		t.Fatal("chart data missing")
	}
	if len(result.Standings) != 50 {
		t.Fatalf("standings = %d, want 50", len(result.Standings))
	}
}
