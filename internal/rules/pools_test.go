package rules_test

import (
	"fmt"
	"testing"

	"github.com/restful-otaku/j3/internal/rules"
)

func entrants(n int) []rules.Entrant {
	out := make([]rules.Entrant, n)
	for i := range out {
		out[i] = rules.Entrant{
			ID:     fmt.Sprintf("f%02d", i+1),
			ClubID: fmt.Sprintf("club-%d", i%4+1),
			Seed:   n - i,
		}
	}
	return out
}

func TestGeneratePoolsBalancedAndSpread(t *testing.T) {
	constraints := rules.PoolConstraints{
		TargetPoolSize: 4, MaxPools: 4, SameClubSeparation: true, Seed: 42,
	}
	all := entrants(16)
	clubsByID := map[string]string{}
	for _, e := range all {
		clubsByID[e.ID] = e.ClubID
	}
	pools := rules.GeneratePools(all, constraints)
	if len(pools) != 4 {
		t.Fatalf("pools = %d, want 4", len(pools))
	}
	seen := map[string]bool{}
	for _, pool := range pools {
		if len(pool.Members) != 4 {
			t.Errorf("pool %s has %d members, want 4", pool.Name, len(pool.Members))
		}
		clubCounts := map[string]int{}
		for _, member := range pool.Members {
			if seen[member] {
				t.Errorf("duplicate member %s", member)
			}
			seen[member] = true
			clubCounts[clubsByID[member]]++
		}
		for _, n := range clubCounts {
			if n > 2 {
				t.Errorf("club has %d members in one pool", n)
			}
		}
	}
	if len(seen) != 16 {
		t.Fatalf("placed %d entrants, want 16", len(seen))
	}
}

func TestGeneratePoolsDeterministic(t *testing.T) {
	constraints := rules.PoolConstraints{TargetPoolSize: 4, MaxPools: 4, SameClubSeparation: true, Seed: 7}
	a := rules.GeneratePools(entrants(12), constraints)
	b := rules.GeneratePools(entrants(12), constraints)
	if fmt.Sprint(a) != fmt.Sprint(b) {
		t.Fatal("pool generation is not deterministic")
	}
}

func TestGeneratePoolsEmpty(t *testing.T) {
	if got := rules.GeneratePools(nil, rules.PoolConstraints{}); len(got) != 0 {
		t.Fatalf("expected no pools, got %d", len(got))
	}
}
