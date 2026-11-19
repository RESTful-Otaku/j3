package rules_test

import (
	"fmt"
	"testing"

	"github.com/restful-otaku/j3/internal/rules"
)

// Property-style checks across many seeds and field sizes. The generator must
// never lose or duplicate a fighter, and pools must stay balanced.
func TestGeneratePoolsProperties(t *testing.T) {
	for seed := int64(0); seed < 60; seed++ {
		for _, n := range []int{1, 2, 3, 4, 5, 7, 11, 16, 23, 40, 60} {
			input := make([]rules.Entrant, n)
			for i := range input {
				input[i] = rules.Entrant{
					ID:     fmt.Sprintf("f%03d", i),
					ClubID: fmt.Sprintf("club-%d", (i+int(seed))%4),
					Seed:   (i * 7) % 13,
				}
			}
			constraints := rules.PoolConstraints{
				TargetPoolSize: 4, SameClubSeparation: seed%2 == 0, Seed: seed,
			}
			pools := rules.GeneratePools(input, constraints)
			wantPools := (n + 3) / 4
			if len(pools) != wantPools {
				t.Fatalf("seed=%d n=%d: got %d pools, want %d", seed, n, len(pools), wantPools)
			}
			seen := map[string]bool{}
			minSize, maxSize := n, 0
			for _, pool := range pools {
				size := len(pool.Members)
				if size < minSize {
					minSize = size
				}
				if size > maxSize {
					maxSize = size
				}
				for _, id := range pool.Members {
					if seen[id] {
						t.Fatalf("seed=%d n=%d: duplicate fighter %s", seed, n, id)
					}
					seen[id] = true
				}
			}
			if len(seen) != n {
				t.Fatalf("seed=%d n=%d: placed %d of %d", seed, n, len(seen), n)
			}
			if maxSize-minSize > 1 {
				t.Fatalf("seed=%d n=%d: unbalanced pools %d..%d", seed, n, minSize, maxSize)
			}
			again := rules.GeneratePools(input, constraints)
			if fmt.Sprint(pools) != fmt.Sprint(again) {
				t.Fatalf("seed=%d n=%d: non-deterministic output", seed, n)
			}
		}
	}
}

func TestGeneratePoolsMaxPoolsCap(t *testing.T) {
	input := entrants(30)
	pools := rules.GeneratePools(input, rules.PoolConstraints{
		TargetPoolSize: 4, MaxPools: 3, SameClubSeparation: true, Seed: 1,
	})
	if len(pools) != 3 {
		t.Fatalf("pools = %d, want cap of 3", len(pools))
	}
	total := 0
	for _, pool := range pools {
		total += len(pool.Members)
	}
	if total != len(input) {
		t.Fatalf("placed %d of %d", total, len(input))
	}
}

func TestGeneratePoolsSingleEntrant(t *testing.T) {
	pools := rules.GeneratePools(entrants(1), rules.PoolConstraints{TargetPoolSize: 4, Seed: 5})
	if len(pools) != 1 || len(pools[0].Members) != 1 {
		t.Fatalf("unexpected single-entrant pools: %+v", pools)
	}
}
