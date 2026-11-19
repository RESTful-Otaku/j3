package rules_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/restful-otaku/j3/internal/rules"
)

func eliCategories() []rules.Category {
	return []rules.Category{
		{"w60", "Under 60", "weight", "open", 0, 60, ""},
		{"w70", "Under 70", "weight", "open", 60, 70, ""},
		{"w80", "Under 80", "weight", "open", 70, 80, ""},
		{"w90", "Under 90", "weight", "male", 80, 90, ""},
		{"w120", "Under 120", "weight", "open", 100, 130, ""},
		{"women", "Women", "weight", "female", 0, 90, ""},
		{"u16", "U16", "age", "open", 0, 16, ""},
		{"adult", "Adult", "age", "open", 17, 99, ""},
		{"blue", "Blue", "belt", "open", 6, 6, "br-blue"},
		{"open", "Open", "combined", "open", 0, 0, ""},
	}
}

func TestEligibilityProperties(t *testing.T) {
	on := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	byID := map[string]rules.Category{}
	for _, c := range eliCategories() {
		byID[c.ID] = c
	}

	for i := 0; i < 250; i++ {
		age := 8 + i%33
		stats := rules.Stats{
			DOB:       fmt.Sprintf("%d-01-01", 2026-age),
			WeightKg:  35 + float64(i%70),
			HeightCm:  120 + float64(i%90),
			BeltOrder: 1 + i%11,
			Gender:    []string{"male", "female", "open"}[i%3],
		}
		for _, consent := range []bool{false, true} {
			result := rules.Compatible(eliCategories(), stats, on, consent)
			offered := map[string]bool{}
			for _, item := range result {
				if offered[item.CategoryID] {
					t.Fatalf("duplicate category %s", item.CategoryID)
				}
				offered[item.CategoryID] = true
				cat := byID[item.CategoryID]
				if cat.Gender != "open" && cat.Gender != "coed" &&
					stats.Gender != "open" && stats.Gender != "coed" && cat.Gender != stats.Gender {
					t.Fatalf("gender mismatch for %s", item.CategoryID)
				}
				if !consent && item.MoveUp {
					t.Fatalf("move-up offered without consent: %s", item.CategoryID)
				}
				if cat.Kind == "weight" {
					if item.Strict && !(stats.WeightKg >= cat.Min && stats.WeightKg <= cat.Max) {
						t.Fatalf("strict weight mismatch: %v in %s", stats.WeightKg, item.CategoryID)
					}
					if item.MoveUp && stats.WeightKg >= cat.Min {
						t.Fatalf("move-up is not heavier: %v vs %s", stats.WeightKg, item.CategoryID)
					}
				}
				if cat.Kind == "age" {
					if item.Strict && !(age >= int(cat.Min) && age <= int(cat.Max)) {
						t.Fatalf("strict age mismatch: %d in %s", age, item.CategoryID)
					}
				}
				if cat.Kind == "belt" && item.Strict && stats.BeltOrder != int(cat.Min) {
					t.Fatalf("strict belt mismatch: %d in %s", stats.BeltOrder, item.CategoryID)
				}
			}
		}
	}
}

func TestEligibilityConsentEnablesHeavierDivision(t *testing.T) {
	stats := rules.Stats{DOB: "1995-01-01", WeightKg: 95, HeightCm: 180, BeltOrder: 6, Gender: "male"}
	without := rules.Compatible(eliCategories(), stats, time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC), false)
	with := rules.Compatible(eliCategories(), stats, time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC), true)
	if len(with) <= len(without) {
		t.Fatalf("consent did not add options: %d vs %d", len(with), len(without))
	}
}
