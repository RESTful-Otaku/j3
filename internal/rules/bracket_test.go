package rules_test

import (
	"fmt"
	"testing"

	"github.com/restful-otaku/j3/internal/rules"
)

func TestSeedEntrantsSpreadsClubs(t *testing.T) {
	input := make([]rules.Entrant, 12)
	for i := range input {
		input[i] = rules.Entrant{ID: fmt.Sprintf("f%02d", i), ClubID: fmt.Sprintf("club-%d", i%3), Seed: 12 - i}
	}
	ordered := rules.SeedEntrants(input, true)
	seen := map[string]bool{}
	for _, entrant := range ordered {
		if seen[entrant.ID] {
			t.Fatal("duplicate entrant")
		}
		seen[entrant.ID] = true
	}
	if len(seen) != len(input) {
		t.Fatal("lost entrants")
	}
	// Top three should come from three different clubs where possible.
	clubs := map[string]bool{}
	for _, entrant := range ordered[:3] {
		clubs[entrant.ClubID] = true
	}
	if len(clubs) < 3 {
		t.Fatalf("clubs clustered at the top: %v", clubs)
	}
}

func TestWithByes(t *testing.T) {
	got := rules.WithByes([]string{"a", "b", "c"})
	if len(got) != 4 || got[3] != "BYE" {
		t.Fatalf("unexpected byes: %v", got)
	}
	if got := rules.WithByes(nil); len(got) != 0 {
		t.Fatalf("empty input: %v", got)
	}
}

func TestTieBreakChain(t *testing.T) {
	a := rules.TieBreakStats{Wins: 2, Submissions: 0, PointsScored: 9, PointsConceded: 6}
	b := rules.TieBreakStats{Wins: 2, Submissions: 1, PointsScored: 7, PointsConceded: 4}
	if got := rules.TieBreak(a, b, []string{"wins", "submissions"}); got != 1 {
		t.Fatalf("submissions should decide: %d", got)
	}
	if got := rules.TieBreak(a, b, []string{"wins", "points_scored"}); got != -1 {
		t.Fatalf("points should decide: %d", got)
	}
	if got := rules.TieBreak(a, a, nil); got != 0 {
		t.Fatalf("identical athletes should tie: %d", got)
	}
}
