// Package rules contains pure, deterministic eligibility and pool algorithms.
package rules

import (
	"sort"
	"time"
)

// Stats are the canonical metric values captured at sign-up.
type Stats struct {
	DOB       string
	WeightKg  float64
	HeightCm  float64
	BeltOrder int
	Gender    string
}

// Category is the data needed to decide eligibility.
type Category struct {
	ID         string
	Name       string
	Kind       string
	Gender     string
	Min        float64
	Max        float64
	BeltRankID string
}

// Eligibility describes an offered category and why it was offered.
type Eligibility struct {
	CategoryID string `json:"categoryId"`
	Name       string `json:"name"`
	Strict     bool   `json:"strict"`
	MoveUp     bool   `json:"moveUp"`
	Reason     string `json:"reason"`
}

func ageOn(dob string, on time.Time) (int, bool) {
	birth, err := time.Parse("2006-01-02", dob)
	if err != nil || birth.After(on) {
		return 0, false
	}
	age := on.Year() - birth.Year()
	if on.Month() < birth.Month() || (on.Month() == birth.Month() && on.Day() < birth.Day()) {
		age--
	}
	return age, true
}

func genderMatches(categoryGender, athleteGender string) bool {
	if categoryGender == "" || categoryGender == "open" || categoryGender == "coed" {
		return true
	}
	if athleteGender == "" || athleteGender == "open" || athleteGender == "coed" {
		return true
	}
	return categoryGender == athleteGender
}

// Compatible returns categories the athlete fits. Strict matches come first;
// when consentMoveUp is true, nearby heavier/older divisions are appended.
func Compatible(categories []Category, stats Stats, on time.Time, consentMoveUp bool) []Eligibility {
	age, hasAge := ageOn(stats.DOB, on)
	result := make([]Eligibility, 0, len(categories))
	type candidate struct {
		cat Category
		min float64
		why string
	}
	moveUps := []candidate{}

	for _, cat := range categories {
		if !genderMatches(cat.Gender, stats.Gender) {
			continue
		}
		switch {
		case cat.Kind == "belt" || (cat.BeltRankID != "" && cat.Max == 0):
			if stats.BeltOrder == 0 {
				continue
			}
			if stats.BeltOrder >= int(cat.Min) && (cat.Max == 0 || stats.BeltOrder <= int(cat.Max)) {
				result = append(result, Eligibility{cat.ID, cat.Name, true, false, "belt grade in range"})
			}
		case cat.Kind == "age":
			if !hasAge {
				continue
			}
			if age >= int(cat.Min) && age <= int(cat.Max) {
				result = append(result, Eligibility{cat.ID, cat.Name, true, false, "age in range"})
			} else if consentMoveUp && age < int(cat.Min) {
				moveUps = append(moveUps, candidate{cat, cat.Min, "older division with consent"})
			}
		case cat.Kind == "weight" || (cat.Kind == "combined" && cat.Max > 0):
			if stats.WeightKg <= 0 {
				continue
			}
			if stats.WeightKg >= cat.Min && stats.WeightKg <= cat.Max {
				result = append(result, Eligibility{cat.ID, cat.Name, true, false, "weight in range"})
			} else if consentMoveUp && stats.WeightKg < cat.Min {
				moveUps = append(moveUps, candidate{cat, cat.Min, "heavier division with consent"})
			}
		default:
			// Open divisions (e.g. Kata) accept everyone.
			result = append(result, Eligibility{cat.ID, cat.Name, true, false, "open division"})
		}
	}

	if consentMoveUp && len(moveUps) > 0 {
		sort.Slice(moveUps, func(i, j int) bool { return moveUps[i].min < moveUps[j].min })
		for _, c := range moveUps {
			result = append(result, Eligibility{c.cat.ID, c.cat.Name, false, true, c.why})
		}
	}
	return result
}
