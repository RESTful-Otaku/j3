package rules_test

import (
	"fmt"
	"testing"

	"github.com/restful-otaku/j3/internal/rules"
)

func TestBuildRotation(t *testing.T) {
	fighters := []string{"a", "b", "c", "d", "e", "f"}
	pairs := rules.BuildRotation(fighters, []string{"Mat 1", "Mat 2"}, 0)
	if len(pairs) != 15 {
		t.Fatalf("pairs = %d, want 15", len(pairs))
	}
	seen := map[string]bool{}
	for _, pair := range pairs {
		if pair.Red == pair.Blue {
			t.Fatal("self pairing")
		}
		key := pair.Red + "|" + pair.Blue
		if seen[key] {
			t.Fatalf("duplicate pairing %s", key)
		}
		seen[key] = true
		if pair.Mat == "" {
			t.Fatal("missing mat")
		}
	}
}

func TestBuildRotationOddField(t *testing.T) {
	fighters := []string{"a", "b", "c", "d", "e"}
	pairs := rules.BuildRotation(fighters, []string{"Mat 1"}, 0)
	if len(pairs) != 10 {
		t.Fatalf("pairs = %d, want 10", len(pairs))
	}
	for _, pair := range pairs {
		if pair.Red == "BYE" || pair.Blue == "BYE" {
			t.Fatal("BYE leaked into a pairing")
		}
	}
}

func TestBuildRotationLimitedRounds(t *testing.T) {
	fighters := []string{}
	for i := 0; i < 8; i++ {
		fighters = append(fighters, fmt.Sprintf("f%d", i))
	}
	pairs := rules.BuildRotation(fighters, []string{"Mat 1"}, 3)
	if len(pairs) == 0 || pairs[len(pairs)-1].Round > 3 {
		t.Fatalf("round cap not applied: %+v", pairs)
	}
}
