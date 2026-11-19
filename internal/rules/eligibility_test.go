package rules_test

import (
	"testing"
	"time"

	"github.com/restful-otaku/j3/internal/rules"
)

func testCategories() []rules.Category {
	return []rules.Category{
		{"w76", "Adult -76 kg", "weight", "male", 0, 76, ""},
		{"w82", "Adult -82 kg", "weight", "male", 76, 82, ""},
		{"w100", "Adult -100 kg", "weight", "male", 95, 110, ""},
		{"u18", "Youth U18", "age", "open", 0, 18, ""},
		{"adult", "Adult", "age", "open", 18, 99, ""},
		{"blue", "Blue belt", "belt", "open", 6, 6, "br-blue"},
		{"kata", "Kata Open", "combined", "open", 0, 0, ""},
		{"women", "Women -58 kg", "weight", "female", 0, 58, ""},
	}
}

func TestCompatibleStrict(t *testing.T) {
	stats := rules.Stats{
		DOB: "1998-04-12", WeightKg: 74, HeightCm: 175, BeltOrder: 6, Gender: "male",
	}
	got := rules.Compatible(testCategories(), stats, time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC), false)
	seen := map[string]bool{}
	for _, e := range got {
		seen[e.CategoryID] = true
		if !e.Strict {
			t.Errorf("unexpected move-up without consent: %s", e.CategoryID)
		}
	}
	for _, id := range []string{"w76", "adult", "blue", "kata"} {
		if !seen[id] {
			t.Errorf("missing eligible category %s", id)
		}
	}
	if seen["women"] {
		t.Error("male athlete offered a female category")
	}
}

func TestCompatibleMoveUpWithConsent(t *testing.T) {
	stats := rules.Stats{
		DOB: "1998-04-12", WeightKg: 90, HeightCm: 180, BeltOrder: 6, Gender: "male",
	}
	got := rules.Compatible(testCategories(), stats, time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC), true)
	found := false
	for _, e := range got {
		if e.CategoryID == "w100" && e.MoveUp {
			found = true
		}
	}
	if !found {
		t.Fatal("heavier division not offered with consent")
	}
}
