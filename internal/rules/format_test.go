package rules_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/restful-otaku/j3/internal/rules"
)

func TestSuggestFormat(t *testing.T) {
	cases := []struct {
		style           string
		entrants, pools int
		want            string
	}{
		{"Single elimination", 1, 1, "no_contest"},
		{"Single elimination", 3, 1, "round_robin"},
		{"Double elimination", 12, 4, "round_robin"},
		{"Double elimination", 20, 4, "Double elimination"},
		{"Fight pools", 12, 4, "Fight pools"},
		{"Single elimination", 10, 0, "round_robin"},
	}
	for _, c := range cases {
		if got := rules.SuggestFormat(c.style, c.entrants, c.pools); got != c.want {
			t.Errorf("SuggestFormat(%q,%d,%d) = %q, want %q", c.style, c.entrants, c.pools, got, c.want)
		}
	}
}

func TestThirdPlacePlan(t *testing.T) {
	if got := rules.ThirdPlacePlan(true, true); got != "dual_bronze" {
		t.Fatalf("rematch after prior meeting: %s", got)
	}
	if got := rules.ThirdPlacePlan(false, true); got != "rematch" {
		t.Fatalf("expected rematch: %s", got)
	}
	if got := rules.ThirdPlacePlan(false, false); got != "dual_bronze" {
		t.Fatalf("rematch disabled: %s", got)
	}
}

func TestCaps(t *testing.T) {
	if !rules.CanAccept(4, 3) || rules.CanAccept(4, 4) || !rules.CanAccept(0, 999) {
		t.Fatal("capacity check is wrong")
	}
	if rules.CapRemaining(4, 2) != 2 || rules.CapRemaining(4, 9) != 0 || rules.CapRemaining(0, 9) != -1 {
		t.Fatal("remaining capacity is wrong")
	}
}

func TestUnevenPoolsAreBalanced(t *testing.T) {
	input := make([]rules.Entrant, 13)
	for i := range input {
		input[i] = rules.Entrant{ID: fmt.Sprintf("f%02d", i), ClubID: fmt.Sprintf("club-%d", i%3), Seed: i + 1}
	}
	pools := rules.GeneratePools(input, rules.PoolConstraints{
		TargetPoolSize: 5, MaxPools: 4, SameClubSeparation: true, Seed: 3,
	})
	total := 0
	min, max := 99, 0
	for _, pool := range pools {
		total += len(pool.Members)
		if len(pool.Members) < min {
			min = len(pool.Members)
		}
		if len(pool.Members) > max {
			max = len(pool.Members)
		}
		if len(pool.Members) > 5 {
			t.Fatalf("pool %s exceeds target size", pool.Name)
		}
	}
	if total != 13 || max-min > 1 {
		t.Fatalf("unbalanced uneven pools: total=%d min=%d max=%d", total, min, max)
	}
}

func TestMissingCategoriesAndVoluntaryPromotion(t *testing.T) {
	stats := rules.Stats{DOB: "1995-01-01", WeightKg: 92, HeightCm: 180, BeltOrder: 6, Gender: "male"}
	on := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	if got := rules.Compatible(nil, stats, on, true); len(got) != 0 {
		t.Fatalf("empty categories produced %d offers", len(got))
	}
	cats := []rules.Category{
		{"light", "Light", "weight", "open", 60, 75, ""},
		{"heavy", "Heavy", "weight", "open", 95, 120, ""},
	}
	strict := rules.Compatible(cats, stats, on, false)
	for _, item := range strict {
		if item.MoveUp {
			t.Fatal("promotion offered without consent")
		}
	}
	promoted := rules.Compatible(cats, stats, on, true)
	found := false
	for _, item := range promoted {
		if item.CategoryID == "heavy" && item.MoveUp {
			found = true
		}
	}
	if !found {
		t.Fatal("voluntary promotion to heavier category not offered")
	}
}
